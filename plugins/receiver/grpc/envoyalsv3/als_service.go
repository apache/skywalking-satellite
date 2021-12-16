// Licensed to Apache Software Foundation (ASF) under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Apache Software Foundation (ASF) licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package envoyalsv3

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/apache/skywalking-satellite/internal/pkg/log"

	"google.golang.org/protobuf/proto"

	"github.com/apache/skywalking-satellite/internal/satellite/module/buffer"
	"github.com/apache/skywalking-satellite/internal/satellite/telemetry"
	"github.com/apache/skywalking-satellite/plugins/server/grpc"

	v3 "skywalking.apache.org/repo/goapi/proto/envoy/service/accesslog/v3"
	v1 "skywalking.apache.org/repo/goapi/satellite/data/v1"
)

const (
	eventName  = "grpc-envoy-als-v3-event"
	alsVersion = "v3"
)

type AlsService struct {
	receiveChannel chan *v1.SniffData
	limiterConfig  buffer.LimiterConfig
	v3.UnimplementedAccessLogServiceServer

	streamingCount        telemetry.Counter
	streamingFailedCount  telemetry.Counter
	streamingToEventCount telemetry.Counter
	activeStreamingCount  telemetry.DynamicGauge
}

func (m *AlsService) init() {
	m.streamingCount = telemetry.NewCounter("als_streaming_receive_count",
		"Total count of the receive stream in the ALS Receiver.", "version", "peer_host")
	m.streamingFailedCount = telemetry.NewCounter("als_streaming_receive_failed_count",
		"Total count of the failed receive message in the ALS Receiver.", "version", "peer_host")
	m.streamingToEventCount = telemetry.NewCounter("als_streaming_to_event_count",
		"Total count of the packaged the ALS streaming in the ALS Receiver.", "version", "peer_host")
	m.activeStreamingCount = telemetry.NewDynamicGauge("als_streaming_active_count",
		"Total active stream count in ALS Receiver", "version", "peer_host")
}

func (m *AlsService) StreamAccessLogs(stream v3.AccessLogService_StreamAccessLogsServer) error {
	messages := make(chan []byte, m.limiterConfig.LimitCount*2)
	limiter := buffer.NewLimiter(m.limiterConfig, func() int {
		return len(messages)
	})
	peer := grpc.GetPeerHostFromStreamContext(stream.Context())
	m.activeStreamingCount.Inc(alsVersion, peer)
	defer m.activeStreamingCount.Dec(alsVersion, peer)

	var identity *v3.StreamAccessLogsMessage_Identifier

	defer limiter.Stop()
	limiter.Start(context.Background(), func() {
		if identity == nil {
			return
		}
		count := len(messages)
		if count == 0 {
			return
		}
		logsMessages := make([][]byte, 0)
		for i := 0; i < count; i++ {
			logsMessages = append(logsMessages, <-messages)
		}

		// process first message identity
		firstMessage := logsMessages[0]
		firstAls := new(v3.StreamAccessLogsMessage)
		if err := proto.Unmarshal(firstMessage, firstAls); err != nil {
			log.Logger.Warnf("could not unmarshal als message: %v", err)
			return
		}
		firstAls.Identifier = identity
		marshal, _ := proto.Marshal(firstAls)
		logsMessages[0] = marshal

		d := &v1.SniffData{
			Name:      eventName,
			Timestamp: time.Now().UnixNano() / 1e6,
			Meta:      nil,
			Type:      v1.SniffType_EnvoyALSV3Type,
			Remote:    true,
			Data: &v1.SniffData_EnvoyALSV3List{
				EnvoyALSV3List: &v1.EnvoyALSV3List{
					Messages: logsMessages,
				},
			},
		}
		m.streamingToEventCount.Inc(alsVersion, peer)
		m.receiveChannel <- d
	})

	var err1 error
	for {
		data := grpc.NewOriginalData(nil)
		err := stream.RecvMsg(data)
		if err != nil {
			m.streamingFailedCount.Inc(alsVersion, peer)
			err1 = err
			break
		}
		if identity == nil {
			item := new(v3.StreamAccessLogsMessage)
			err = proto.Unmarshal(data.Content, item)
			if err != nil {
				return fmt.Errorf("could not umarshal first message, %v", err)
			}
			if item.Identifier == nil {
				return fmt.Errorf("could not found identity in message")
			}
			identity = item.Identifier
		}

		m.streamingCount.Inc(alsVersion, peer)
		messages <- data.Content
		limiter.Check()
	}

	if err1 != io.EOF {
		return err1
	}
	return stream.SendAndClose(&v3.StreamAccessLogsResponse{})
}

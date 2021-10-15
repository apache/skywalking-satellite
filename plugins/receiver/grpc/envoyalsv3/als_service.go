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
	"io"
	"time"

	"github.com/apache/skywalking-satellite/internal/satellite/module/buffer"

	v3 "skywalking.apache.org/repo/goapi/proto/envoy/service/accesslog/v3"
	v1 "skywalking.apache.org/repo/goapi/satellite/data/v1"
)

const eventName = "grpc-envoy-als-v3-event"

type AlsService struct {
	receiveChannel chan *v1.SniffData
	limiterConfig  buffer.LimiterConfig
	v3.UnimplementedAccessLogServiceServer
}

func (m *AlsService) StreamAccessLogs(stream v3.AccessLogService_StreamAccessLogsServer) error {
	messages := make(chan *v3.StreamAccessLogsMessage, m.limiterConfig.LimitCount*2)
	limiter := buffer.NewLimiter(m.limiterConfig, func() int {
		return len(messages)
	})

	var identity *v3.StreamAccessLogsMessage_Identifier

	defer limiter.Stop()
	limiter.Start(context.Background(), func() {
		count := len(messages)
		if count == 0 {
			return
		}
		logsMessages := make([]*v3.StreamAccessLogsMessage, 0)
		for i := 0; i < count; i++ {
			logsMessages = append(logsMessages, <-messages)
		}
		logsMessages[0].Identifier = identity

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
		m.receiveChannel <- d
	})

	var err1 error
	for {
		item, err := stream.Recv()
		if err != nil {
			err1 = err
			break
		}
		if item.Identifier != nil {
			identity = item.Identifier
		}
		messages <- item
		limiter.Check()
	}

	if err1 != io.EOF {
		return err1
	}
	return stream.SendAndClose(&v3.StreamAccessLogsResponse{})
}

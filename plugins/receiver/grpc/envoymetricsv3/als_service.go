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

package envoymetricsv3

import (
	"context"
	"io"
	"time"

	"github.com/apache/skywalking-satellite/internal/satellite/module/buffer"

	v3 "skywalking.apache.org/repo/goapi/proto/envoy/service/metrics/v3"
	v1 "skywalking.apache.org/repo/goapi/satellite/data/v1"
)

const eventName = "grpc-envoy-metrics-v3-event"

type MetricsService struct {
	receiveChannel chan *v1.SniffData
	limiterConfig  buffer.LimiterConfig
	v3.UnimplementedMetricsServiceServer
}

func (m *MetricsService) StreamMetrics(stream v3.MetricsService_StreamMetricsServer) error {
	messages := make(chan *v3.StreamMetricsMessage, m.limiterConfig.LimitCount*2)
	limiter := buffer.NewLimiter(m.limiterConfig, func() int {
		return len(messages)
	})

	var identity *v3.StreamMetricsMessage_Identifier

	defer limiter.Stop()
	limiter.Start(context.Background(), func() {
		count := len(messages)
		if count == 0 {
			return
		}
		metricsMessages := make([]*v3.StreamMetricsMessage, 0)
		for i := 0; i < count; i++ {
			metricsMessages = append(metricsMessages, <-messages)
		}
		metricsMessages[0].Identifier = identity

		d := &v1.SniffData{
			Name:      eventName,
			Timestamp: time.Now().UnixNano() / 1e6,
			Meta:      nil,
			Type:      v1.SniffType_EnvoyMetricsV3Type,
			Remote:    true,
			Data: &v1.SniffData_EnvoyMetricsV3List{
				EnvoyMetricsV3List: &v1.EnvoyMetricsV3List{
					Messages: metricsMessages,
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
	return stream.SendAndClose(&v3.StreamMetricsResponse{})
}

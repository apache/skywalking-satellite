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

package envoyalsv2

import (
	"io"
	"time"

	v2 "skywalking.apache.org/repo/goapi/proto/envoy/service/accesslog/v2"
	v1 "skywalking.apache.org/repo/goapi/satellite/data/v1"
)

const eventName = "grpc-envoyals-v2-event"

type AlsService struct {
	receiveChannel chan *v1.SniffData
	v2.UnimplementedAccessLogServiceServer
}

func (m *AlsService) StreamAccessLogs(stream v2.AccessLogService_StreamAccessLogsServer) error {
	var identifier *v2.StreamAccessLogsMessage_Identifier
	for {
		item, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&v2.StreamAccessLogsResponse{})
		}
		if err != nil {
			return err
		}
		// only first item has identifier property
		// need correlate information to each item
		if item.Identifier != nil {
			identifier = item.Identifier
		}
		item.Identifier = identifier
		d := &v1.SniffData{
			Name:      eventName,
			Timestamp: time.Now().UnixNano() / 1e6,
			Meta:      nil,
			Type:      v1.SniffType_EnvoyALSV2Type,
			Remote:    true,
			Data: &v1.SniffData_EnvoyALSV2{
				EnvoyALSV2: item,
			},
		}
		m.receiveChannel <- d
	}
}

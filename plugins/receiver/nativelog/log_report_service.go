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

package nativelog

import (
	"io"
	"time"

	common "skywalking/network/common/v3"
	logging "skywalking/network/logging/v3"

	"github.com/apache/skywalking-satellite/protocol/gen-codes/satellite/protocol"
)

const eventName = "grpc-log-event"

type LogReportService struct {
	receiveChannel chan *protocol.Event
	logging.UnimplementedLogReportServiceServer
}

func (s *LogReportService) Collect(stream logging.LogReportService_CollectServer) error {
	for {
		logData, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&common.Commands{})
		}
		if err != nil {
			return err
		}
		e := &protocol.Event{
			Name:      eventName,
			Timestamp: time.Now().UnixNano() / 1e6,
			Meta:      nil,
			Type:      protocol.EventType_Logging,
			Remote:    true,
			Data: &protocol.Event_Log{
				Log: logData,
			},
		}
		s.receiveChannel <- e
	}
}

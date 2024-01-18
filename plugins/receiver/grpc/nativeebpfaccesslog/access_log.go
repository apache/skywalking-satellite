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

package nativeebpfaccesslog

import (
	"io"
	"time"

	"github.com/apache/skywalking-satellite/plugins/server/grpc"

	v3 "skywalking.apache.org/repo/goapi/collect/ebpf/accesslog/v3"
	v1 "skywalking.apache.org/repo/goapi/satellite/data/v1"
)

var eventName = "grpc-accesslog-event"

type AccessLogService struct {
	OutputChannel chan *v1.SniffData

	v3.UnimplementedEBPFAccessLogServiceServer
}

func (a *AccessLogService) Collect(stream v3.EBPFAccessLogService_CollectServer) error {
	result := make([][]byte, 0)
	originalData := grpc.NewOriginalData(nil)
	for {
		if err := stream.RecvMsg(originalData); err == io.EOF {
			a.sendData(result)
			return stream.SendAndClose(&v3.EBPFAccessLogDownstream{})
		} else if err != nil {
			a.sendData(result)
			return err
		}
		result = append(result, originalData.Content)
	}
}

func (a *AccessLogService) sendData(dataList [][]byte) {
	e := &v1.SniffData{
		Name:      eventName,
		Timestamp: time.Now().UnixNano() / 1e6,
		Meta:      nil,
		Type:      v1.SniffType_EBPFAccessLogType,
		Remote:    true,
		Data: &v1.SniffData_EBPFAccessLogList{
			EBPFAccessLogList: &v1.EBPFAccessLogList{
				Messages: dataList,
			},
		},
	}
	a.OutputChannel <- e
}

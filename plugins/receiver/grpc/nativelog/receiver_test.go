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
	"context"
	"strconv"
	"testing"
	"time"

	"google.golang.org/protobuf/proto"

	"google.golang.org/grpc"

	common "skywalking.apache.org/repo/goapi/collect/common/v3"
	logging "skywalking.apache.org/repo/goapi/collect/logging/v3"
	v1 "skywalking.apache.org/repo/goapi/satellite/data/v1"

	_ "github.com/apache/skywalking-satellite/internal/satellite/test"
	receiver_grpc "github.com/apache/skywalking-satellite/plugins/receiver/grpc"
)

func TestReceiver_RegisterHandler(t *testing.T) {
	receiver_grpc.TestReceiver(new(Receiver), func(t *testing.T, sequence int, conn *grpc.ClientConn, ctx context.Context) string {
		client := logging.NewLogReportServiceClient(conn)
		data := initData(sequence)
		collect, err := client.Collect(ctx)
		if err != nil {
			t.Fatalf("cannot open the stream send mode: %v", err)
		}
		if err := collect.Send(data); err != nil {
			t.Fatalf("cannot send the data to the server: %v", err)
		}
		if err := collect.CloseSend(); err != nil {
			t.Fatalf("cannot close the stream mode: %v", err)
		}
		return data.String()
	}, func(data *v1.SniffData) string {
		d := new(logging.LogData)
		_ = proto.Unmarshal(data.GetLogList().Logs[0], d)
		return d.String()
	}, t)
}

func initData(sequence int) *logging.LogData {
	seq := strconv.Itoa(sequence)
	return &logging.LogData{
		Timestamp:       time.Now().Unix(),
		Service:         "demo-service" + seq,
		ServiceInstance: "demo-instance" + seq,
		Endpoint:        "demo-endpoint" + seq,
		TraceContext: &logging.TraceContext{
			TraceSegmentId: "mock-segmentId" + seq,
			TraceId:        "mock-traceId" + seq,
			SpanId:         1,
		},
		Tags: &logging.LogTags{
			Data: []*common.KeyStringValuePair{
				{
					Key:   "mock-key" + seq,
					Value: "mock-value" + seq,
				},
			},
		},
		Body: &logging.LogDataBody{
			Type: "mock-type" + seq,
			Content: &logging.LogDataBody_Text{
				Text: &logging.TextLog{
					Text: "this is a mock text mock log" + seq,
				},
			},
		},
	}
}

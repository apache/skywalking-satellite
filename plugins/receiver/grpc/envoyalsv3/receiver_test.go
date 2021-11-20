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
	"strconv"
	"testing"

	"google.golang.org/protobuf/proto"

	"google.golang.org/grpc"

	v3 "skywalking.apache.org/repo/goapi/proto/envoy/service/accesslog/v3"
	v1 "skywalking.apache.org/repo/goapi/satellite/data/v1"

	_ "github.com/apache/skywalking-satellite/internal/satellite/test"
	receiver_grpc "github.com/apache/skywalking-satellite/plugins/receiver/grpc"
)

func TestReceiver_RegisterHandler(t *testing.T) {
	recConf := make(map[string]string, 2)
	recConf["limit_count"] = "1"
	recConf["flush_time"] = "1000"
	receiver_grpc.TestReceiverWithConfig(new(Receiver), recConf, func(t *testing.T, sequence int, conn *grpc.ClientConn, ctx context.Context) string {
		client := v3.NewAccessLogServiceClient(conn)
		data := initData(sequence)
		collect, err := client.StreamAccessLogs(ctx)
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
		m := new(v3.StreamAccessLogsMessage)
		_ = proto.Unmarshal(data.GetEnvoyALSV3List().Messages[0], m)
		return m.String()
	}, t)
}

func initData(sequence int) *v3.StreamAccessLogsMessage {
	return &v3.StreamAccessLogsMessage{
		Identifier: &v3.StreamAccessLogsMessage_Identifier{
			LogName: "test" + strconv.Itoa(sequence),
		},
	}
}

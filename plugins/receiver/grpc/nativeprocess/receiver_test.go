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

package nativeprocess

import (
	"context"
	"testing"

	"google.golang.org/grpc"

	agent "skywalking.apache.org/repo/goapi/collect/ebpf/profiling/process/v3"
	v1 "skywalking.apache.org/repo/goapi/satellite/data/v1"

	_ "github.com/apache/skywalking-satellite/internal/satellite/test"
	receiver_grpc "github.com/apache/skywalking-satellite/plugins/receiver/grpc"
)

func TestReceiver_RegisterHandler(t *testing.T) {
	receiver_grpc.TestReceiver(new(Receiver), func(t *testing.T, sequence int, conn *grpc.ClientConn, ctx context.Context) string {
		client := agent.NewEBPFProcessServiceClient(conn)
		data := initData()
		_, err := client.KeepAlive(ctx, data)
		if err != nil {
			t.Fatalf("cannot send data: %v", err)
		}
		return data.String()
	}, func(data *v1.SniffData) string {
		return data.GetEBPFProcessPingPkgList().String()
	}, t)
}

func initData() *agent.EBPFProcessPingPkgList {
	return &agent.EBPFProcessPingPkgList{
		Processes: []*agent.EBPFProcessPingPkg{
			{
				EntityMetadata: &agent.EBPFProcessEntityMetadata{
					Layer:        "GENERAL",
					ServiceName:  "test-service",
					InstanceName: "test-instance",
					ProcessName:  "test-process",
					Labels: []string{
						"test-label",
					},
				},
			},
		},
	}
}

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
	"testing"
	"time"

	"google.golang.org/grpc"

	common "skywalking.apache.org/repo/goapi/collect/common/v3"
	agent "skywalking.apache.org/repo/goapi/collect/language/agent/v3"
	v1 "skywalking.apache.org/repo/goapi/satellite/data/v1"

	_ "github.com/apache/skywalking-satellite/internal/satellite/test"
	receiver_grpc "github.com/apache/skywalking-satellite/plugins/receiver/grpc"
)

func TestReceiver_RegisterHandler(t *testing.T) {
	receiver_grpc.TestReceiver(new(Receiver), func(t *testing.T, sequence int, conn *grpc.ClientConn, ctx context.Context) string {
		client := agent.NewJVMMetricReportServiceClient(conn)
		data := initData()
		_, err := client.Collect(ctx, data)
		if err != nil {
			t.Fatalf("cannot send data: %v", err)
		}
		return data.String()
	}, func(data *v1.SniffData) string {
		return data.GetJvm().String()
	}, t)
}

func initData() *agent.JVMMetricCollection {
	return &agent.JVMMetricCollection{
		Service:         "demo-service",
		ServiceInstance: "demo-instance",
		Metrics: []*agent.JVMMetric{
			{
				Time: time.Now().Unix() / 1e6,
				Cpu: &common.CPU{
					UsagePercent: 99.9,
				},
				Memory: []*agent.Memory{
					{
						IsHeap:    true,
						Init:      1,
						Max:       2,
						Used:      3,
						Committed: 4,
					},
				},
				MemoryPool: []*agent.MemoryPool{
					{
						Type:      0,
						Init:      1,
						Max:       2,
						Used:      3,
						Committed: 4,
					},
				},
				Gc: []*agent.GC{
					{
						Count: 3,
						Time:  202106111010,
					},
				},
				Thread: &agent.Thread{
					LiveCount:   1,
					PeakCount:   2,
					DaemonCount: 3,
				},
			},
		},
	}
}

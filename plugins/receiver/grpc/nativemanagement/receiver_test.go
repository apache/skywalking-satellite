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

package nativemanagement

import (
	"context"
	"fmt"
	"testing"

	"google.golang.org/grpc"

	common "skywalking.apache.org/repo/goapi/collect/common/v3"
	management "skywalking.apache.org/repo/goapi/collect/management/v3"
	v1 "skywalking.apache.org/repo/goapi/satellite/data/v1"

	_ "github.com/apache/skywalking-satellite/internal/satellite/test"
	receiver_grpc "github.com/apache/skywalking-satellite/plugins/receiver/grpc"
)

func TestReceiver_RegisterHandler_ReportInstance(t *testing.T) {
	receiver_grpc.TestReceiver(new(Receiver), func(t *testing.T, sequence int, conn *grpc.ClientConn, ctx context.Context) string {
		client := management.NewManagementServiceClient(conn)
		properties := &management.InstanceProperties{
			Service:         fmt.Sprintf("service_%d", sequence),
			ServiceInstance: fmt.Sprintf("instance_%d", sequence),
			Properties:      []*common.KeyStringValuePair{},
		}
		commands, err := client.ReportInstanceProperties(ctx, properties)
		if err != nil {
			t.Fatalf("cannot send the data to the server: %v", err)
		}
		if commands == nil {
			t.Fatalf("report instance result is nil")
		}
		return properties.String()
	}, func(data *v1.SniffData) string {
		return data.GetInstance().String()
	}, t)
}

func TestReceiver_RegisterHandler_InstancePing(t *testing.T) {
	receiver_grpc.TestReceiver(new(Receiver), func(t *testing.T, sequence int, conn *grpc.ClientConn, ctx context.Context) string {
		client := management.NewManagementServiceClient(conn)
		instancePing := &management.InstancePingPkg{
			Service:         fmt.Sprintf("service_%d", sequence),
			ServiceInstance: fmt.Sprintf("instance_%d", sequence),
		}
		commands, err := client.KeepAlive(ctx, instancePing)
		if err != nil {
			t.Fatalf("cannot send the data to the server: %v", err)
		}
		if commands == nil {
			t.Fatalf("instance ping result is nil")
		}
		return instancePing.String()
	}, func(data *v1.SniffData) string {
		return data.GetInstancePing().String()
	}, t)
}

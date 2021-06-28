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

package nativecds

import (
	"context"
	"fmt"
	"testing"

	v3 "skywalking.apache.org/repo/goapi/collect/agent/configuration/v3"

	common "skywalking.apache.org/repo/goapi/collect/common/v3"

	"google.golang.org/grpc"

	v1 "skywalking.apache.org/repo/goapi/satellite/data/v1"

	_ "github.com/apache/skywalking-satellite/internal/satellite/test"
	receiver_grpc "github.com/apache/skywalking-satellite/plugins/receiver/grpc"
)

func TestReceiver_RegisterHandler(t *testing.T) {
	receiver_grpc.TestReceiverWithSync(new(Receiver), func(t *testing.T, sequence int, conn *grpc.ClientConn, sendData *string, ctx context.Context) {
		client := v3.NewConfigurationDiscoveryServiceClient(conn)
		data := &v3.ConfigurationSyncRequest{
			Service: fmt.Sprintf("service-%d", sequence),
			Uuid:    "",
		}
		*sendData = data.String()
		_, err := client.FetchConfigurations(ctx, data)
		if err != nil {
			t.Fatalf("cannot send data: %v", err)
		}
	}, func(data *v1.SniffData) string {
		return data.GetConfigurationSyncRequest().String()
	}, &v1.SniffData{
		Data: &v1.SniffData_Commands{
			Commands: &common.Commands{},
		},
	}, t)
}

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

package nativeevent

import (
	"context"
	"strconv"
	"testing"
	"time"

	"google.golang.org/grpc"

	nativeevent "skywalking.apache.org/repo/goapi/collect/event/v3"
	v1 "skywalking.apache.org/repo/goapi/satellite/data/v1"

	_ "github.com/apache/skywalking-satellite/internal/satellite/test"
	receiver_grpc "github.com/apache/skywalking-satellite/plugins/receiver/grpc"
)

func TestReceiver_RegisterHandler(t *testing.T) {
	receiver_grpc.TestReceiver(new(Receiver), func(t *testing.T, sequence int, conn *grpc.ClientConn, ctx context.Context) string {
		client := nativeevent.NewEventServiceClient(conn)
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
		return data.GetEvent().String()
	}, t)
}

func initData(sequence int) *nativeevent.Event {
	seq := strconv.Itoa(sequence)
	return &nativeevent.Event{
		StartTime: time.Now().Unix() / 1e6,
		EndTime:   time.Now().Unix() / 1e6,
		Uuid:      "12345" + seq,
		Source: &nativeevent.Source{
			Service:         "demo-service" + seq,
			ServiceInstance: "demo-instance" + seq,
			Endpoint:        "demo-endpoint" + seq,
		},
		Name:    "test-name" + seq,
		Type:    nativeevent.Type_Error,
		Message: "test message" + seq,
	}
}

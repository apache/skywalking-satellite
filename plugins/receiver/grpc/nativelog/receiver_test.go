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
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"google.golang.org/grpc"

	common "skywalking/network/common/v3"
	logging "skywalking/network/logging/v3"

	"github.com/apache/skywalking-satellite/internal/pkg/plugin"
	_ "github.com/apache/skywalking-satellite/internal/satellite/test"
	receiver "github.com/apache/skywalking-satellite/plugins/receiver/api"
	server "github.com/apache/skywalking-satellite/plugins/server/api"
	grpcserver "github.com/apache/skywalking-satellite/plugins/server/grpc"
	"github.com/apache/skywalking-satellite/protocol/gen-codes/satellite/protocol"
)

func TestReceiver_RegisterHandler(t *testing.T) {
	Init()
	r := initReceiver(make(plugin.Config), t)
	s := initServer(make(plugin.Config), t)
	r.RegisterHandler(s.GetServer())
	_ = s.Start()
	time.Sleep(time.Second)
	defer func() {
		if err := s.Close(); err != nil {
			t.Fatalf("cannot close the sever: %v", err)
		}
	}()
	client := initClient(t)
	for i := 0; i < 10; i++ {
		data := initData(i)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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
		newData := <-r.Channel()
		if !cmp.Equal(newData.Data.(*protocol.Event_Log).Log.String(), data.String()) {
			t.Fatalf("the sent data is not equal to the received data\n, "+
				"want data %s\n, but got %s\n", data.String(), newData.String())
		}
		cancel()
	}
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
		Tags: []*common.KeyStringValuePair{
			{
				Key:   "mock-key" + seq,
				Value: "mock-value" + seq,
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

func initClient(t *testing.T) logging.LogReportServiceClient {
	conn, err := grpc.Dial("localhost:8000", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		t.Fatalf("cannot init the grpc client: %v", err)
	}
	return logging.NewLogReportServiceClient(conn)
}

func Init() {
	plugin.RegisterPluginCategory(reflect.TypeOf((*server.Server)(nil)).Elem())
	plugin.RegisterPluginCategory(reflect.TypeOf((*receiver.Receiver)(nil)).Elem())
	plugin.RegisterPlugin(new(grpcserver.Server))
	plugin.RegisterPlugin(new(Receiver))
}

func initServer(cfg plugin.Config, t *testing.T) server.Server {
	cfg[plugin.NameField] = grpcserver.Name
	q := server.GetServer(cfg)
	if q == nil {
		t.Fatalf("cannot get a grpc server from the registry")
	}
	if err := q.Prepare(); err != nil {
		t.Fatalf("cannot perpare the grpc server: %v", err)
	}
	return q
}

func initReceiver(cfg plugin.Config, t *testing.T) receiver.Receiver {
	cfg[plugin.NameField] = "grpc-nativelog-receiver"
	q := receiver.GetReceiver(cfg)
	if q == nil {
		t.Fatalf("cannot get grpclog-receiver from the registry")
	}
	return q
}

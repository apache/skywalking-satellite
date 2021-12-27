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

package grpc

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"reflect"
	"testing"
	"time"

	v1 "skywalking.apache.org/repo/goapi/satellite/data/v1"

	"github.com/apache/skywalking-satellite/internal/pkg/plugin"
	receiver "github.com/apache/skywalking-satellite/plugins/receiver/api"
	server "github.com/apache/skywalking-satellite/plugins/server/api"
	grpc_server "github.com/apache/skywalking-satellite/plugins/server/grpc"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// TestReceiver help to testing grpc receiver
func TestReceiver(rec receiver.Receiver,
	dataGenerator func(t *testing.T, sequence int, conn *grpc.ClientConn, ctx context.Context) string,
	snifferConvertor func(data *v1.SniffData) string, t *testing.T) {
	TestReceiverWithConfig(rec, make(map[string]string), dataGenerator, snifferConvertor, t)
}

// TestReceiverWithConfig help to testing grpc receiver with customize config
func TestReceiverWithConfig(rec receiver.Receiver, recConf map[string]string,
	dataGenerator func(t *testing.T, sequence int, conn *grpc.ClientConn, ctx context.Context) string,
	snifferConvertor func(data *v1.SniffData) string, t *testing.T) {
	Init(rec)
	grpcPort := randomGrpcPort()
	receiverConfig := make(plugin.Config)
	for k, v := range recConf {
		receiverConfig[k] = v
	}
	r := initReceiver(receiverConfig, t, rec)
	s := initServer(make(plugin.Config), grpcPort, t)
	r.RegisterHandler(s.GetServer())
	_ = s.Start()
	time.Sleep(time.Second)
	defer func() {
		if err := s.Close(); err != nil {
			t.Fatalf("cannot close the sever: %v", err)
		}
	}()
	conn := initConnection(grpcPort, t)
	for i := 0; i < 10; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		var data string
		errorMsg := ""
		go func() {
			newData := <-r.Channel()
			// await data content
			time.Sleep(time.Millisecond * 100)
			if !cmp.Equal(snifferConvertor(newData), data) {
				errorMsg = fmt.Sprintf("the sent data is not equal to the received data\n, "+
					"want data %s\n, but got %s\n", data, newData.String())
			}
			cancel()
		}()
		data = dataGenerator(t, i, conn, ctx)
		<-ctx.Done()
		if errorMsg != "" {
			t.Fatalf(errorMsg)
		}
	}
}

// TestReceiverWithSync help to testing grpc receiver
func TestReceiverWithSync(rec receiver.Receiver,
	dataGenerator func(t *testing.T, sequence int, conn *grpc.ClientConn, sendData *string, ctx context.Context),
	snifferConvertor func(data *v1.SniffData) string, mockResp *v1.SniffData, t *testing.T) {
	Init(rec)
	grpcPort := randomGrpcPort()
	r := initReceiver(make(plugin.Config), t, rec)
	s := initServer(make(plugin.Config), grpcPort, t)
	r.RegisterHandler(s.GetServer())
	time.Sleep(time.Second)
	defer func() {
		if err := s.Close(); err != nil {
			t.Fatalf("cannot close the sever: %v", err)
		}
	}()
	_ = s.Start()

	var data string
	invoker := syncInvoker{snifferConvertor: snifferConvertor, mockResp: mockResp, data: &data}
	r.RegisterSyncInvoker(&invoker)
	conn := initConnection(grpcPort, t)
	for i := 0; i < 10; i++ {
		dataGenerator(t, i, conn, &data, context.Background())
		if invoker.errorMsg != "" {
			t.Fatalf(invoker.errorMsg)
		}
	}
}

type syncInvoker struct {
	snifferConvertor func(data *v1.SniffData) string
	mockResp         *v1.SniffData
	data             *string
	errorMsg         string
}

func (s *syncInvoker) SyncInvoke(event *v1.SniffData) (*v1.SniffData, error) {
	// await data content
	time.Sleep(time.Millisecond * 100)
	if !cmp.Equal(s.snifferConvertor(event), *s.data) {
		s.errorMsg = fmt.Sprintf("the sent data is not equal to the received data\n, "+
			"want data %s\n, but got %s\n", *s.data, event.String())
		return nil, nil
	}
	return s.mockResp, nil
}

func initConnection(grpcPort int, t *testing.T) *grpc.ClientConn {
	conn, err := grpc.Dial(fmt.Sprintf("localhost:%d", grpcPort), grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		t.Fatalf("cannot init the grpc client: %v", err)
	}
	return conn
}

func Init(receiverPlugin plugin.Plugin) {
	plugin.RegisterPluginCategory(reflect.TypeOf((*server.Server)(nil)).Elem())
	plugin.RegisterPluginCategory(reflect.TypeOf((*receiver.Receiver)(nil)).Elem())
	plugin.RegisterPlugin(new(grpc_server.Server))
	plugin.RegisterPlugin(receiverPlugin)
}

func initServer(cfg plugin.Config, grpcPort int, t *testing.T) server.Server {
	cfg[plugin.NameField] = grpc_server.Name
	cfg["address"] = fmt.Sprintf(":%d", grpcPort)
	q := server.GetServer(cfg)
	if err := q.Prepare(); err != nil {
		t.Fatalf("cannot perpare the grpc server: %v", err)
	}
	return q
}

func randomGrpcPort() int {
	b := new(big.Int).SetInt64(int64(65535 - 1000))
	i, err := rand.Int(rand.Reader, b)
	if err != nil {
		fmt.Printf("Can't generate random value: %v, %v", i, err)
		return -1
	}
	return int(i.Int64() + 1000)
}

func initReceiver(cfg plugin.Config, _ *testing.T, receive receiver.Receiver) receiver.Receiver {
	cfg[plugin.NameField] = receive.Name()
	return receiver.GetReceiver(cfg)
}

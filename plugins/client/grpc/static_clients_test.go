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

	common "skywalking.apache.org/repo/goapi/collect/common/v3"
	agent "skywalking.apache.org/repo/goapi/collect/language/agent/v3"

	"google.golang.org/grpc"

	"github.com/apache/skywalking-satellite/internal/pkg/log"
	"github.com/apache/skywalking-satellite/internal/pkg/plugin"
	_ "github.com/apache/skywalking-satellite/internal/satellite/telemetry/none"
	client "github.com/apache/skywalking-satellite/plugins/client/api"
	receiver "github.com/apache/skywalking-satellite/plugins/receiver/api"
	server "github.com/apache/skywalking-satellite/plugins/server/api"
	grpc_server "github.com/apache/skywalking-satellite/plugins/server/grpc"
)

type JVMReportService struct {
	receiveCount int
	agent.UnimplementedJVMMetricReportServiceServer
}

func (j *JVMReportService) Collect(_ context.Context, jvm *agent.JVMMetricCollection) (*common.Commands, error) {
	j.receiveCount++
	return &common.Commands{}, nil
}

func TestStaticServer(t *testing.T) {
	serverCount := 2
	Init()

	// init all servers
	servers, ports := initServers(serverCount, t)
	receivers := make([]*JVMReportService, serverCount)
	for inx, s := range servers {
		reportService := &JVMReportService{receiveCount: 0}
		receivers[inx] = reportService
		agent.RegisterJVMMetricReportServiceServer(s.GetServer().(*grpc.Server), reportService)

		if err := s.Start(); err != nil {
			t.Errorf("start client error: %v", err)
		}
	}
	defer func() {
		for _, s := range servers {
			s.Close()
		}
	}()

	// init client
	c := initClient(ports, t)

	// wait all channel being connected (connect by async)
	time.Sleep(time.Second * 1)

	// send request
	jvmClient := agent.NewJVMMetricReportServiceClient(c.GetConnectedClient().(*grpc.ClientConn))
	for inx := 0; inx < serverCount; inx++ {
		if _, err := jvmClient.Collect(context.Background(), &agent.JVMMetricCollection{}); err != nil {
			t.Errorf("send request error: %v", err)
		}
	}

	// check all receiver must have received data
	for inx, receiver := range receivers {
		if receiver.receiveCount <= 0 {
			t.Errorf("check result failed, client index: %d", inx)
		}
	}
}

func Init() {
	log.Init(new(log.LoggerConfig))
	plugin.RegisterPluginCategory(reflect.TypeOf((*server.Server)(nil)).Elem())
	plugin.RegisterPluginCategory(reflect.TypeOf((*client.Client)(nil)).Elem())
	plugin.RegisterPluginCategory(reflect.TypeOf((*receiver.Receiver)(nil)).Elem())
	plugin.RegisterPlugin(new(grpc_server.Server))
	plugin.RegisterPlugin(new(Client))
}

func initServers(serverCount int, t *testing.T) (servers []server.Server, ports []int) {
	for inx := 0; inx < serverCount; inx++ {
		cfg := make(plugin.Config)
		cfg[plugin.NameField] = grpc_server.Name
		port := randomGrpcPort()
		cfg["address"] = fmt.Sprintf(":%d", port)
		q := server.GetServer(cfg)
		if err := q.Prepare(); err != nil {
			t.Fatalf("cannot perpare the grpc server: %v", err)
		}

		servers = append(servers, q)
		ports = append(ports, port)
	}

	return servers, ports
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

func initClient(ports []int, t *testing.T) client.Client {
	cfg := make(plugin.Config)
	cfg[plugin.NameField] = Name
	serverList := ""
	for inx := range ports {
		if inx > 0 {
			serverList += ","
		}
		serverList += fmt.Sprintf("%s%d", "0.0.0.0:", ports[inx])
	}
	cfg["server_addr"] = serverList
	q := client.GetClient(cfg)
	if err := q.Prepare(); err != nil {
		t.Errorf("prepare client error: %v", err)
	}
	if err := q.Start(); err != nil {
		t.Errorf("start client error: %v", err)
	}
	return q
}

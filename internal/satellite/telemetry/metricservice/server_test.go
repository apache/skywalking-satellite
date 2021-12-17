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

package metricservice

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/apache/skywalking-satellite/internal/pkg/log"
	"github.com/apache/skywalking-satellite/internal/pkg/plugin"
	"github.com/apache/skywalking-satellite/internal/satellite/config"
	"github.com/apache/skywalking-satellite/internal/satellite/sharing"
	"github.com/apache/skywalking-satellite/internal/satellite/telemetry"
	client "github.com/apache/skywalking-satellite/plugins/client/api"
	grpc_client "github.com/apache/skywalking-satellite/plugins/client/grpc"
	receiver "github.com/apache/skywalking-satellite/plugins/receiver/api"
	server "github.com/apache/skywalking-satellite/plugins/server/api"
	grpc_server "github.com/apache/skywalking-satellite/plugins/server/grpc"

	"google.golang.org/grpc"

	v3 "skywalking.apache.org/repo/goapi/collect/language/agent/v3"
)

type MeterHandler struct {
	sendCount int
	v3.UnimplementedMeterReportServiceServer
}

func (m *MeterHandler) CollectBatch(batch v3.MeterReportService_CollectBatchServer) error {
	for {
		_, err := batch.Recv()
		if err != nil {
			return err
		}
		m.sendCount++
	}
}

func initTelemetryServer(port int, meterHandler *MeterHandler) (*grpc.Server, error) {
	s := grpc.NewServer()
	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	v3.RegisterMeterReportServiceServer(s, meterHandler)

	go func() {
		err = s.Serve(listen)
		if err != nil {
			log.Logger.Errorf("could not start gRPC server: %v", err)
		}
	}()
	return s, err
}

func initGRPCClient(port int) error {
	plugin.RegisterPluginCategory(reflect.TypeOf((*server.Server)(nil)).Elem())
	plugin.RegisterPluginCategory(reflect.TypeOf((*client.Client)(nil)).Elem())
	plugin.RegisterPluginCategory(reflect.TypeOf((*receiver.Receiver)(nil)).Elem())
	plugin.RegisterPlugin(new(grpc_server.Server))
	plugin.RegisterPlugin(new(grpc_client.Client))

	sharing.Load(&config.SharingConfig{
		Clients: []plugin.Config{
			map[string]interface{}{
				"plugin_name": grpc_client.Name,
				"server_addr": fmt.Sprintf("%s%d", "0.0.0.0:", port),
			},
		},
	})
	if err := sharing.Prepare(); err != nil {
		return fmt.Errorf("prepare client error: %v", err)
	}
	if err := sharing.Start(); err != nil {
		return fmt.Errorf("start client error: %v", err)
	}
	return nil
}

func TestMetricsService(t *testing.T) {
	log.Init(new(log.LoggerConfig))
	grpcPort := randomGrpcPort()
	handler := &MeterHandler{}

	// init server
	s, err := initTelemetryServer(grpcPort, handler)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Stop()

	// init telemetry
	name := "metrics_service"
	telemetry.Register(name, &Server{}, false)
	c := new(telemetry.Config)
	c.ExportType = name
	c.MetricsService.Interval = 1
	c.MetricsService.ClientName = grpc_client.Name
	if err := telemetry.Init(c); err != nil {
		t.Fatal(err)
	}

	// init sharding client
	if err := initGRPCClient(grpcPort); err != nil {
		t.Fatal(err)
	}
	if err := telemetry.AfterShardingStart(); err != nil {
		t.Fatal(err)
	}

	// add simple counter
	telemetry.NewCounter("test", "").Inc()

	time.Sleep(time.Second * 5)

	if handler.sendCount == 0 {
		t.Fatal("could not receive the metrics")
	}
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

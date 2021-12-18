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
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/apache/skywalking-satellite/internal/pkg/log"
	"github.com/apache/skywalking-satellite/internal/satellite/sharing"
	"github.com/apache/skywalking-satellite/internal/satellite/telemetry"
	client "github.com/apache/skywalking-satellite/plugins/client/api"
	"github.com/apache/skywalking-satellite/plugins/client/grpc/lb"
	server_grpc "github.com/apache/skywalking-satellite/plugins/server/grpc"

	"google.golang.org/grpc"

	v3 "skywalking.apache.org/repo/goapi/collect/language/agent/v3"
)

func init() {
	telemetry.Register("metrics_service", &Server{}, false)
}

type Server struct {
	telemetry.MetricsServiceConfig

	service  string
	instance string
	metrics  map[string]Metric
	lock     sync.Mutex

	prevServerAddr string
	meterClient    v3.MeterReportServiceClient
	reportStream   v3.MeterReportService_CollectBatchClient
	ctx            context.Context
	cancel         context.CancelFunc
}

func (s *Server) Start(config *telemetry.Config) error {
	s.metrics = make(map[string]Metric)
	s.MetricsServiceConfig = config.MetricsService
	s.service = config.Service
	s.instance = config.Instance
	s.ctx, s.cancel = context.WithCancel(context.Background())
	return nil
}

func (s *Server) AfterSharingStart() error {
	plugin := sharing.Manager[s.MetricsServiceConfig.ClientName]
	if plugin == nil {
		return fmt.Errorf("could not fould client %s", s.MetricsServiceConfig.ClientName)
	}
	grpcClient, ok := plugin.(client.Client)
	if !ok {
		return fmt.Errorf("the client is not grpc client")
	}
	conn := grpcClient.GetConnectedClient().(*grpc.ClientConn)
	s.meterClient = v3.NewMeterReportServiceClient(conn)

	go func() {
		ticker := time.NewTicker(time.Duration(s.Interval) * time.Second)
		for {
			select {
			case <-ticker.C:
				if err := s.sendMetrics(); err != nil {
					log.Logger.Warnf("send satellite metrics failure: %v", err)
				}
			case <-context.Background().Done():
				return
			case <-s.ctx.Done():
				s.cancel()
				return
			}
		}
	}()
	return nil
}

func (s *Server) sendMetrics() error {
	if s.reportStream == nil {
		if err := s.openBatchStream(); err != nil {
			return err
		}
	}
	appender := &MetricsAppender{
		time:    time.Now().UnixNano() / int64(time.Millisecond),
		metrics: make([]*v3.MeterData, 0),
	}
	for _, metric := range s.metrics {
		metric.WriteMetric(appender)
	}

	if len(appender.metrics) == 0 {
		return nil
	}

	appender.metrics[0].Service = s.service
	appender.metrics[0].ServiceInstance = s.instance
	if err := s.reportStream.Send(&v3.MeterDataCollection{MeterData: appender.metrics}); err != nil {
		if openErr := s.openBatchStream(); openErr != nil {
			log.Logger.Warnf("detect send message error and reopen stream failure: %v", openErr)
		}
	}
	return nil
}

func (s *Server) openBatchStream() error {
	if s.reportStream != nil {
		_, err := s.reportStream.CloseAndRecv()
		if err != nil {
			log.Logger.Warnf("close satellite meter protocol error: %v", err)
		}
	}

	if meterStream, err := s.meterClient.CollectBatch(lb.WithLoadBalanceConfig(context.Background(), "metricsService", s.prevServerAddr)); err == nil {
		s.reportStream = meterStream
		s.prevServerAddr = server_grpc.GetPeerAddressFromStreamContext(meterStream.Context())
	} else {
		return fmt.Errorf("could not start metrics service stream: %v", err)
	}
	return nil
}

func (s *Server) Close() error {
	s.cancel()
	if s.reportStream != nil {
		if _, err := s.reportStream.CloseAndRecv(); err != nil {
			log.Logger.Warnf("error close the meter stream in satellite metrics service: %v", err)
		}
	}
	return nil
}

func (s *Server) Register(name string, metric Metric) {
	s.metrics[name] = metric
}

type MetricsAppender struct {
	time    int64
	metrics []*v3.MeterData
}

func (a *MetricsAppender) appendSingleValue(name string, labels []*v3.Label, val float64) {
	a.metrics = append(a.metrics, &v3.MeterData{
		Timestamp: a.time,
		Metric: &v3.MeterData_SingleValue{
			SingleValue: &v3.MeterSingleValue{
				Name:   name,
				Labels: labels,
				Value:  val,
			},
		},
	})
}

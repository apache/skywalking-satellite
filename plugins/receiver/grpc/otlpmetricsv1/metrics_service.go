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

package otlpmetricsv1

import (
	"context"
	"time"

	metrics "skywalking.apache.org/repo/goapi/proto/opentelemetry/proto/collector/metrics/v1"
	sniffer "skywalking.apache.org/repo/goapi/satellite/data/v1"
)

const eventName = "grpc-envoy-metrics-v3-event"

type MetricsService struct {
	receiveChannel chan *sniffer.SniffData
	metrics.MetricsServiceServer
}

func (m *MetricsService) Export(ctx context.Context, req *metrics.ExportMetricsServiceRequest) (*metrics.ExportMetricsServiceResponse, error) {
	e := &sniffer.SniffData{
		Name:      eventName,
		Timestamp: time.Now().UnixNano() / 1e6,
		Meta:      nil,
		Type:      sniffer.SniffType_OpenTelementryMetricsV1Type,
		Remote:    true,
		Data: &sniffer.SniffData_OpenTelementryMetricsV1Request{
			OpenTelementryMetricsV1Request: req,
		},
	}

	m.receiveChannel <- e
	return &metrics.ExportMetricsServiceResponse{}, nil
}

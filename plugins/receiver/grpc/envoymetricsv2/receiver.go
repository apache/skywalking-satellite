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

package envoymetricsv2

import (
	"github.com/apache/skywalking-satellite/internal/pkg/config"
	module "github.com/apache/skywalking-satellite/internal/satellite/module/api"
	"github.com/apache/skywalking-satellite/internal/satellite/module/buffer"
	forwarder "github.com/apache/skywalking-satellite/plugins/forwarder/api"
	"github.com/apache/skywalking-satellite/plugins/forwarder/grpc/envoymetricsv2"
	"github.com/apache/skywalking-satellite/plugins/receiver/grpc"

	v2 "skywalking.apache.org/repo/goapi/proto/envoy/service/metrics/v2"
	v1 "skywalking.apache.org/repo/goapi/satellite/data/v1"
)

const (
	Name     = "grpc-envoy-metrics-v2-receiver"
	ShowName = "GRPC Envoy Metrics v2 Receiver"
)

type Receiver struct {
	config.CommonFields
	grpc.CommonGRPCReceiverFields
	service *MetricsService

	LimitConfig buffer.LimiterConfig `mapstructure:",squash"`
}

func (r *Receiver) Name() string {
	return Name
}

func (r *Receiver) ShowName() string {
	return ShowName
}

func (r *Receiver) Description() string {
	return "This is a receiver for Envoy Metrics format, " +
		"which is defined at https://github.com/envoyproxy/envoy/blob/v1.17.4/api/envoy/service/metrics/v2/metrics_service.proto."
}

func (r *Receiver) DefaultConfig() string {
	return `
# The time interval between two flush operations. And the time unit is millisecond.
flush_time: 1000
# The max cache count when receive the message
limit_count: 500
`
}

func (r *Receiver) RegisterHandler(server interface{}) {
	r.CommonGRPCReceiverFields = *grpc.InitCommonGRPCReceiverFields(server)
	r.service = &MetricsService{receiveChannel: r.OutputChannel, limiterConfig: r.LimitConfig}
	v2.RegisterMetricsServiceServer(r.Server, r.service)
}

func (r *Receiver) RegisterSyncInvoker(_ module.SyncInvoker) {
}

func (r *Receiver) Channel() <-chan *v1.SniffData {
	return r.OutputChannel
}

func (r *Receiver) SupportForwarders() []forwarder.Forwarder {
	return []forwarder.Forwarder{
		new(envoymetricsv2.Forwarder),
	}
}

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

package envoyalsv3

import (
	"github.com/apache/skywalking-satellite/internal/pkg/config"
	module "github.com/apache/skywalking-satellite/internal/satellite/module/api"
	"github.com/apache/skywalking-satellite/internal/satellite/module/buffer"
	forwarder "github.com/apache/skywalking-satellite/plugins/forwarder/api"
	"github.com/apache/skywalking-satellite/plugins/forwarder/grpc/envoyalsv3"
	"github.com/apache/skywalking-satellite/plugins/receiver/grpc"

	v3 "skywalking.apache.org/repo/goapi/proto/envoy/service/accesslog/v3"
	v1 "skywalking.apache.org/repo/goapi/satellite/data/v1"
)

const Name = "grpc-envoy-als-v3-receiver"

type Receiver struct {
	config.CommonFields
	grpc.CommonGRPCReceiverFields
	service *AlsService

	LimitConfig buffer.LimiterConfig `mapstructure:",squash"`
}

func (r *Receiver) Name() string {
	return Name
}

func (r *Receiver) Description() string {
	return "This is a receiver for Envoy ALS format, " +
		"which is defined at https://github.com/envoyproxy/envoy/blob/" +
		"3791753e94edbac8a90c5485c68136886c40e719/api/envoy/config/accesslog/v3/accesslog.proto."
}

func (r *Receiver) DefaultConfig() string {
	return ""
}

func (r *Receiver) RegisterHandler(server interface{}) {
	r.CommonGRPCReceiverFields = *grpc.InitCommonGRPCReceiverFields(server)
	r.service = &AlsService{receiveChannel: r.OutputChannel, limiterConfig: r.LimitConfig}
	v3.RegisterAccessLogServiceServer(r.Server, r.service)
}

func (r *Receiver) RegisterSyncInvoker(_ module.SyncInvoker) {
}

func (r *Receiver) Channel() <-chan *v1.SniffData {
	return r.OutputChannel
}

func (r *Receiver) SupportForwarders() []forwarder.Forwarder {
	return []forwarder.Forwarder{
		new(envoyalsv3.Forwarder),
	}
}

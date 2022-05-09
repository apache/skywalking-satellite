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

package nativeebpfprofiling

import (
	v3 "skywalking.apache.org/repo/goapi/collect/ebpf/profiling/v3"
	v1 "skywalking.apache.org/repo/goapi/satellite/data/v1"

	"github.com/apache/skywalking-satellite/internal/pkg/config"
	module "github.com/apache/skywalking-satellite/internal/satellite/module/api"
	forwarder "github.com/apache/skywalking-satellite/plugins/forwarder/api"
	forwarder_nativeebpfprofiling "github.com/apache/skywalking-satellite/plugins/forwarder/grpc/nativeebpfprofiling"
	grpcreceiver "github.com/apache/skywalking-satellite/plugins/receiver/grpc"
)

const (
	Name     = "grpc-native-ebpf-profiling-receiver"
	ShowName = "GRPC Native EBFP Profiling Receiver"
)

type Receiver struct {
	config.CommonFields
	grpcreceiver.CommonGRPCReceiverFields
	service *ProfilingReportService // The gRPC request handler for profiling data.
}

func (r *Receiver) Name() string {
	return Name
}

func (r *Receiver) ShowName() string {
	return ShowName
}

func (r *Receiver) Description() string {
	return "This is a receiver for SkyWalking native process format, " +
		"which is defined at https://github.com/apache/skywalking-data-collect-protocol/blob/master/ebpf/profiling/Process.proto."
}

func (r *Receiver) DefaultConfig() string {
	return ""
}

func (r *Receiver) RegisterHandler(server interface{}) {
	r.CommonGRPCReceiverFields = *grpcreceiver.InitCommonGRPCReceiverFields(server)
	r.service = &ProfilingReportService{receiveChannel: r.OutputChannel}
	v3.RegisterEBPFProfilingServiceServer(r.Server, r.service)
}

func (r *Receiver) RegisterSyncInvoker(invoker module.SyncInvoker) {
	r.service.SyncInvoker = invoker
}

func (r *Receiver) Channel() <-chan *v1.SniffData {
	return r.OutputChannel
}

func (r *Receiver) SupportForwarders() []forwarder.Forwarder {
	return []forwarder.Forwarder{
		new(forwarder_nativeebpfprofiling.Forwarder),
	}
}

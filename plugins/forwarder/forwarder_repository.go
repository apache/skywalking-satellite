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

package forwarder

import (
	"reflect"

	"github.com/apache/skywalking-satellite/plugins/forwarder/grpc/envoyalsv2"
	"github.com/apache/skywalking-satellite/plugins/forwarder/grpc/envoyalsv3"
	"github.com/apache/skywalking-satellite/plugins/forwarder/grpc/envoymetricsv2"
	"github.com/apache/skywalking-satellite/plugins/forwarder/grpc/envoymetricsv3"
	grpc_nativecds "github.com/apache/skywalking-satellite/plugins/forwarder/grpc/nativecds"
	grpc_nativeclr "github.com/apache/skywalking-satellite/plugins/forwarder/grpc/nativeclr"
	grpc_nativeebpfprofiling "github.com/apache/skywalking-satellite/plugins/forwarder/grpc/nativeebpfprofiling"
	grpc_nativeevent "github.com/apache/skywalking-satellite/plugins/forwarder/grpc/nativeevent"
	grpc_nativejvm "github.com/apache/skywalking-satellite/plugins/forwarder/grpc/nativejvm"
	grpc_nativelog "github.com/apache/skywalking-satellite/plugins/forwarder/grpc/nativelog"
	grpc_nativemanagement "github.com/apache/skywalking-satellite/plugins/forwarder/grpc/nativemanagement"
	grpc_meter "github.com/apache/skywalking-satellite/plugins/forwarder/grpc/nativemeter"
	grpc_nativeprocess "github.com/apache/skywalking-satellite/plugins/forwarder/grpc/nativeprocess"
	grpc_nativeprofile "github.com/apache/skywalking-satellite/plugins/forwarder/grpc/nativeprofile"
	grpc_nativetracing "github.com/apache/skywalking-satellite/plugins/forwarder/grpc/nativetracing"
	"github.com/apache/skywalking-satellite/plugins/forwarder/grpc/otlpmetricsv1"
	kafka_nativelog "github.com/apache/skywalking-satellite/plugins/forwarder/kafka/nativelog"

	"github.com/apache/skywalking-satellite/internal/pkg/plugin"
	"github.com/apache/skywalking-satellite/plugins/forwarder/api"
)

// RegisterForwarderPlugins register the used filter plugins.
func RegisterForwarderPlugins() {
	plugin.RegisterPluginCategory(reflect.TypeOf((*api.Forwarder)(nil)).Elem())
	forwarders := []api.Forwarder{
		// Please register the forwarder plugins at here.
		new(kafka_nativelog.Forwarder),
		new(grpc_nativelog.Forwarder),
		new(grpc_meter.Forwarder),
		new(grpc_nativemanagement.Forwarder),
		new(grpc_nativetracing.Forwarder),
		new(grpc_nativeprofile.Forwarder),
		new(grpc_nativecds.Forwarder),
		new(grpc_nativeevent.Forwarder),
		new(grpc_nativejvm.Forwarder),
		new(grpc_nativeclr.Forwarder),
		new(grpc_nativeprocess.Forwarder),
		new(grpc_nativeebpfprofiling.Forwarder),
		new(envoyalsv2.Forwarder),
		new(envoyalsv3.Forwarder),
		new(envoymetricsv2.Forwarder),
		new(envoymetricsv3.Forwarder),
		new(otlpmetricsv1.Forwarder),
	}
	for _, forwarder := range forwarders {
		plugin.RegisterPlugin(forwarder)
	}
}

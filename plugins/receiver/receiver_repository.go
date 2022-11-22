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

package receiver

import (
	"reflect"

	"github.com/apache/skywalking-satellite/internal/pkg/plugin"
	"github.com/apache/skywalking-satellite/plugins/receiver/api"
	"github.com/apache/skywalking-satellite/plugins/receiver/grpc/envoyalsv2"
	"github.com/apache/skywalking-satellite/plugins/receiver/grpc/envoyalsv3"
	"github.com/apache/skywalking-satellite/plugins/receiver/grpc/envoymetricsv2"
	"github.com/apache/skywalking-satellite/plugins/receiver/grpc/envoymetricsv3"
	grpcnativecds "github.com/apache/skywalking-satellite/plugins/receiver/grpc/nativecds"
	grpcnativeclr "github.com/apache/skywalking-satellite/plugins/receiver/grpc/nativeclr"
	grpcnativeebpfprofiling "github.com/apache/skywalking-satellite/plugins/receiver/grpc/nativeebpfprofiling"
	grpcnativeevent "github.com/apache/skywalking-satellite/plugins/receiver/grpc/nativeevent"
	grpcnativejvm "github.com/apache/skywalking-satellite/plugins/receiver/grpc/nativejvm"
	grpcnavtivelog "github.com/apache/skywalking-satellite/plugins/receiver/grpc/nativelog"
	grpcnativemanagement "github.com/apache/skywalking-satellite/plugins/receiver/grpc/nativemanagement"
	grpcnativemeter "github.com/apache/skywalking-satellite/plugins/receiver/grpc/nativemeter"
	grpcnativeprocess "github.com/apache/skywalking-satellite/plugins/receiver/grpc/nativeprocess"
	grpcnativeprofile "github.com/apache/skywalking-satellite/plugins/receiver/grpc/nativeprofile"
	grpcnativetracing "github.com/apache/skywalking-satellite/plugins/receiver/grpc/nativetracing"
	"github.com/apache/skywalking-satellite/plugins/receiver/grpc/otlpmetricsv1"
	httpnavtivelog "github.com/apache/skywalking-satellite/plugins/receiver/http/nativcelog"
)

// RegisterReceiverPlugins register the used receiver plugins.
func RegisterReceiverPlugins() {
	plugin.RegisterPluginCategory(reflect.TypeOf((*api.Receiver)(nil)).Elem())
	receivers := []api.Receiver{
		// Please register the receiver plugins at here.
		new(grpcnavtivelog.Receiver),
		new(grpcnativemanagement.Receiver),
		new(grpcnativetracing.Receiver),
		new(grpcnativeprofile.Receiver),
		new(grpcnativecds.Receiver),
		new(httpnavtivelog.Receiver),
		new(grpcnativejvm.Receiver),
		new(grpcnativeclr.Receiver),
		new(grpcnativeevent.Receiver),
		new(grpcnativemeter.Receiver),
		new(grpcnativeprocess.Receiver),
		new(grpcnativeebpfprofiling.Receiver),
		new(envoyalsv2.Receiver),
		new(envoyalsv3.Receiver),
		new(envoymetricsv2.Receiver),
		new(envoymetricsv3.Receiver),
		new(otlpmetricsv1.Receiver),
	}
	for _, receiver := range receivers {
		plugin.RegisterPlugin(receiver)
	}
}

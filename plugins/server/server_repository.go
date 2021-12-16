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

package server

import (
	"reflect"

	"github.com/apache/skywalking-satellite/internal/pkg/plugin"
	"github.com/apache/skywalking-satellite/plugins/server/api"
	"github.com/apache/skywalking-satellite/plugins/server/grpc"
	"github.com/apache/skywalking-satellite/plugins/server/http"
)

// RegisterServerPlugins register the used server plugins.
func RegisterServerPlugins() {
	plugin.RegisterPluginCategory(reflect.TypeOf((*api.Server)(nil)).Elem())
	servers := []api.Server{
		// Please register the server plugins at here.
		new(grpc.Server),
		new(http.Server),
	}
	for _, server := range servers {
		plugin.RegisterPlugin(server)
	}
}

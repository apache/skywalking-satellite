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

package client

import (
	"reflect"

	"github.com/apache/skywalking-satellite/plugins/client/grpc"

	"github.com/apache/skywalking-satellite/internal/pkg/plugin"
	"github.com/apache/skywalking-satellite/plugins/client/api"
	"github.com/apache/skywalking-satellite/plugins/client/kafka"
)

// RegisterClientPlugins register the used client plugins.
func RegisterClientPlugins() {
	plugin.RegisterPluginCategory(reflect.TypeOf((*api.Client)(nil)).Elem())
	clients := []api.Client{
		// Please register the client plugins at here.
		new(kafka.Client),
		new(grpc.Client),
	}
	for _, client := range clients {
		plugin.RegisterPlugin(client)
	}
}

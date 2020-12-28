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

package api

import (
	"reflect"

	"github.com/apache/skywalking-satellite/internal/pkg/plugin"
	"github.com/apache/skywalking-satellite/protocol/gen-codes/satellite/protocol"
)

// Receiver is a plugin interface, that defines new collectors.
type Receiver interface {
	plugin.Plugin

	// RegisterHandler register  a handler to the server, such as to handle a gRPC or an HTTP request
	RegisterHandler(server interface{})

	// Channel would be put a data when the receiver receives an APM data.
	Channel() <-chan *protocol.Event
}

// Get an initialized receiver plugin.
func GetReceiver(config plugin.Config) Receiver {
	return plugin.Get(reflect.TypeOf((*Receiver)(nil)).Elem(), config).(Receiver)
}

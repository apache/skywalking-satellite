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

	v1 "skywalking.apache.org/repo/goapi/satellite/data/v1"

	"github.com/apache/skywalking-satellite/internal/pkg/plugin"
	module "github.com/apache/skywalking-satellite/internal/satellite/module/api"
	forwarder "github.com/apache/skywalking-satellite/plugins/forwarder/api"
)

// Receiver is a plugin interface, that defines new collectors.
type Receiver interface {
	plugin.Plugin

	// RegisterHandler register  a handler to the server, such as to handle a gRPC or an HTTP request
	RegisterHandler(server interface{})

	// RegisterSyncInvoker register the sync invoker, receive event and sync invoke to sender
	RegisterSyncInvoker(invoker module.SyncInvoker)

	// Channel would be put a data when the receiver receives an APM data.
	Channel() <-chan *v1.SniffData

	// SupportForwarders should provider all forwarder support current receiver
	SupportForwarders() []forwarder.Forwarder
}

// GetReceiver gets an initialized receiver plugin.
func GetReceiver(config plugin.Config) Receiver {
	return plugin.Get(reflect.TypeOf((*Receiver)(nil)).Elem(), config).(Receiver)
}

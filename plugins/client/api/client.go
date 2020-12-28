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
)

// The client statuses.
const (
	_ ClientStatus = iota
	Connected
	Disconnect
)

// ClientStatus represents the status of the client.
type ClientStatus int8

// Client is a plugin interface, that defines new clients, such as gRPC client and Kafka client.
type Client interface {
	plugin.SharingPlugin

	// GetConnection returns the connected client to publish events.
	GetConnectedClient() interface{}
	// RegisterListener register a listener to listen the client status.
	RegisterListener(chan<- ClientStatus)
}

// Get an initialized client plugin.
func GetClient(config plugin.Config) Client {
	return plugin.Get(reflect.TypeOf((*Client)(nil)).Elem(), config).(Client)
}

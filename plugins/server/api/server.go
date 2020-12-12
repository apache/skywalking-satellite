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

import "github.com/apache/skywalking-satellite/internal/pkg/plugin"

// Server is a plugin interface, that defines new servers, such as gRPC server and http server.
type Server interface {
	plugin.SharingPlugin

	// Initialize would do some preparation works.
	Initialize()

	// Start a server to receive the input APM data.
	Start() error
}

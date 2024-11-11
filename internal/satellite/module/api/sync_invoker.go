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
	"google.golang.org/grpc"
	v1 "skywalking.apache.org/repo/goapi/satellite/data/v1"
)

type SyncInvoker interface {
	// SyncInvoke means synchronized process event.
	// The returned result grpc.ClientStream is the stream initiated by satellite to oap server,
	// which is used to provide bidirectional stream support
	SyncInvoke(d *v1.SniffData) (*v1.SniffData, grpc.ClientStream, error)
}

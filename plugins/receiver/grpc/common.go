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

package grpc

import (
	"fmt"
	"reflect"

	"google.golang.org/grpc"

	v1 "skywalking.apache.org/repo/goapi/satellite/data/v1"
)

type CommonGRPCReceiverFields struct {
	Server        *grpc.Server
	OutputChannel chan *v1.SniffData // The channel is to bridge the LogReportService and the Gatherer to delivery the data.
}

// InitCommonGRPCReceiverFields init the common fields for gRPC receivers.
func InitCommonGRPCReceiverFields(server interface{}) *CommonGRPCReceiverFields {
	s, ok := server.(*grpc.Server)
	if !ok {
		panic(fmt.Errorf("registerHandler does not support %s", reflect.TypeOf(server).String()))
	}
	return &CommonGRPCReceiverFields{
		Server:        s,
		OutputChannel: make(chan *v1.SniffData),
	}
}

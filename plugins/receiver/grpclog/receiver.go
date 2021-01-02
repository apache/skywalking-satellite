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

package grpclog

import (
	"fmt"
	"reflect"

	"google.golang.org/grpc"

	logging "skywalking/network/logging/v3"

	"github.com/apache/skywalking-satellite/internal/pkg/config"
	"github.com/apache/skywalking-satellite/internal/pkg/plugin"
	"github.com/apache/skywalking-satellite/protocol/gen-codes/satellite/protocol"
)

type Receiver struct {
	config.CommonFields
	MaxBufferSpaces int `mapstructure:"max_buffer_spaces"` // The max buffer events.

	server        *grpc.Server
	service       *LogReportService    // The gRPC request handler for logData.
	outputChannel chan *protocol.Event // The channel is to bridge the LogReportService and the Gatherer to delivery the data.

}

func (r *Receiver) Name() string {
	return plugin.GetPluginName(r)
}

func (r *Receiver) Description() string {
	return "This is a receiver for SkyWalking native logging format, " +
		"which is defined at https://github.com/apache/skywalking-data-collect-protocol/blob/master/logging/Logging.proto."
}

func (r *Receiver) DefaultConfig() string {
	return `
# The max buffer events to process flow surge.
max_buffer_spaces: 1
`
}

func (r *Receiver) RegisterHandler(server interface{}) {
	s, ok := server.(*grpc.Server)
	if !ok {
		panic(fmt.Errorf("registerHandler does not support %s", reflect.TypeOf(server).String()))
	}
	r.server = s
	r.outputChannel = make(chan *protocol.Event, r.MaxBufferSpaces)
	r.service = &LogReportService{receiveChannel: r.outputChannel}
	logging.RegisterLogReportServiceServer(r.server, r.service)
}

func (r *Receiver) Channel() <-chan *protocol.Event {
	return r.outputChannel
}

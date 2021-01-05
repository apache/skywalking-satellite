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

package nativelog

import (
	logging "skywalking/network/logging/v3"

	"github.com/apache/skywalking-satellite/internal/pkg/config"
	grpcreceiver "github.com/apache/skywalking-satellite/plugins/receiver/grpc"
	"github.com/apache/skywalking-satellite/protocol/gen-codes/satellite/protocol"
)

const Name = "grpc-nativelog-receiver"

type Receiver struct {
	config.CommonFields
	grpcreceiver.CommonGRPCReceiverFields
	service *LogReportService // The gRPC request handler for logData.
}

func (r *Receiver) Name() string {
	return Name
}

func (r *Receiver) Description() string {
	return "This is a receiver for SkyWalking native logging format, " +
		"which is defined at https://github.com/apache/skywalking-data-collect-protocol/blob/master/logging/Logging.proto."
}

func (r *Receiver) DefaultConfig() string {
	return ""
}

func (r *Receiver) RegisterHandler(server interface{}) {
	r.CommonGRPCReceiverFields = *grpcreceiver.InitCommonGRPCReceiverFields(server)
	r.service = &LogReportService{receiveChannel: r.OutputChannel}
	logging.RegisterLogReportServiceServer(r.Server, r.service)
}

func (r *Receiver) Channel() <-chan *protocol.Event {
	return r.OutputChannel
}

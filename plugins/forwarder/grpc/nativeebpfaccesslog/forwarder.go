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

package nativeebpfaccesslog

import (
	"context"
	"fmt"
	"io"
	"reflect"

	"google.golang.org/grpc"

	"github.com/apache/skywalking-satellite/internal/pkg/config"
	"github.com/apache/skywalking-satellite/internal/pkg/log"
	"github.com/apache/skywalking-satellite/internal/satellite/event"
	server_grpc "github.com/apache/skywalking-satellite/plugins/server/grpc"

	v3 "skywalking.apache.org/repo/goapi/collect/ebpf/accesslog/v3"
	v1 "skywalking.apache.org/repo/goapi/satellite/data/v1"
)

const (
	Name     = "native-ebpf-accesslog-grpc-forwarder"
	ShowName = "GRPC Native EBFP Access Log Forwarder"
)

type Forwarder struct {
	config.CommonFields

	accessClient v3.EBPFAccessLogServiceClient
}

func (f *Forwarder) Name() string {
	return Name
}

func (f *Forwarder) ShowName() string {
	return ShowName
}

func (f *Forwarder) Description() string {
	return "This is a synchronization grpc forwarder with the SkyWalking native eBPF access log protocol."
}

func (f *Forwarder) DefaultConfig() string {
	return ``
}

func (f *Forwarder) Prepare(connection interface{}) error {
	client, ok := connection.(*grpc.ClientConn)
	if !ok {
		return fmt.Errorf("the %s only accepts a grpc client, but received a %s",
			f.Name(), reflect.TypeOf(connection).String())
	}
	f.accessClient = v3.NewEBPFAccessLogServiceClient(client)
	return nil
}

func (f *Forwarder) Forward(batch event.BatchEvents) (err error) {
	var stream v3.EBPFAccessLogService_CollectClient
	for _, e := range batch {
		data, ok := e.GetData().(*v1.SniffData_EBPFAccessLogList)
		if !ok {
			continue
		}

		stream, err = f.accessClient.Collect(context.Background())
		if err != nil {
			log.Logger.Errorf("open grpc stream error: %v", err)
			return err
		}

		streamClosed := false
		for _, message := range data.EBPFAccessLogList.Messages {
			err := stream.SendMsg(server_grpc.NewOriginalData(message))
			if err != nil {
				log.Logger.Errorf("%s send log data error: %v", f.Name(), err)
				f.closeStream(stream)
				streamClosed = true
				break
			}
		}

		if !streamClosed {
			f.closeStream(stream)
		}
	}

	return nil
}

func (f *Forwarder) closeStream(stream v3.EBPFAccessLogService_CollectClient) {
	_, err := stream.CloseAndRecv()
	if err != nil && err != io.EOF {
		log.Logger.Errorf("%s close stream error: %v", f.Name(), err)
	}
}

func (f *Forwarder) ForwardType() v1.SniffType {
	return v1.SniffType_EBPFAccessLogType
}

func (f *Forwarder) SyncForward(e *v1.SniffData) (*v1.SniffData, grpc.ClientStream, error) {
	return nil, nil, fmt.Errorf("unsupport sync forward")
}

func (f *Forwarder) SupportedSyncInvoke() bool {
	return false
}

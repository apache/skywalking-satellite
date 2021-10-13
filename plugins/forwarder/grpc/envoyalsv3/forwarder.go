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

package envoyalsv3

import (
	"context"
	"fmt"
	"io"
	"reflect"

	v3 "skywalking.apache.org/repo/goapi/proto/envoy/service/accesslog/v3"
	v1 "skywalking.apache.org/repo/goapi/satellite/data/v1"

	"google.golang.org/grpc"

	"github.com/apache/skywalking-satellite/internal/pkg/config"
	"github.com/apache/skywalking-satellite/internal/pkg/log"
	"github.com/apache/skywalking-satellite/internal/satellite/event"
)

const Name = "envoy-als-v3-grpc-forwarder"

type Forwarder struct {
	config.CommonFields
	alsClient v3.AccessLogServiceClient
}

func (f *Forwarder) Name() string {
	return Name
}

func (f *Forwarder) Description() string {
	return "This is a synchronization ALS v3 grpc forwarder with the Envoy ALS protocol."
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
	f.alsClient = v3.NewAccessLogServiceClient(client)
	return nil
}

func (f *Forwarder) Forward(batch event.BatchEvents) error {
	streamMap := make(map[*v3.StreamAccessLogsMessage_Identifier]v3.AccessLogService_StreamAccessLogsClient)
	defer func() {
		for _, stream := range streamMap {
			err := closeStream(stream)
			if err != nil {
				log.Logger.Warnf("%s close stream error: %v", f.Name(), err)
			}
		}
	}()
	for _, e := range batch {
		data, ok := e.GetData().(*v1.SniffData_EnvoyALSV3)
		if !ok {
			continue
		}
		stream := streamMap[data.EnvoyALSV3.Identifier]
		if stream == nil {
			curStream, err := f.alsClient.StreamAccessLogs(context.Background())
			if err != nil {
				log.Logger.Errorf("open grpc stream error %v", err)
				return err
			}
			streamMap[data.EnvoyALSV3.Identifier] = curStream
			stream = curStream
		} else {
			// only first message have identifier
			data.EnvoyALSV3.Identifier = nil
		}

		err := stream.Send(data.EnvoyALSV3)
		if err != nil {
			log.Logger.Errorf("%s send meter data error: %v", f.Name(), err)
			return err
		}
	}
	return nil
}

func closeStream(stream v3.AccessLogService_StreamAccessLogsClient) error {
	_, err := stream.CloseAndRecv()
	if err != nil && err != io.EOF {
		return err
	}
	return nil
}

func (f *Forwarder) ForwardType() v1.SniffType {
	return v1.SniffType_EnvoyALSV3Type
}

func (f *Forwarder) SyncForward(_ *v1.SniffData) (*v1.SniffData, error) {
	return nil, fmt.Errorf("unsupport sync forward")
}

func (f *Forwarder) SupportedSyncInvoke() bool {
	return false
}

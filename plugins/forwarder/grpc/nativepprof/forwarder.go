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

package nativepprof

import (
	"context"
	"fmt"
	"reflect"

	client_grpc "github.com/apache/skywalking-satellite/plugins/client/grpc"
	server_grpc "github.com/apache/skywalking-satellite/plugins/server/grpc"

	"google.golang.org/grpc"
	pprof "skywalking.apache.org/repo/goapi/collect/pprof/v10"
	v1 "skywalking.apache.org/repo/goapi/satellite/data/v1"

	"github.com/apache/skywalking-satellite/internal/pkg/config"
	"github.com/apache/skywalking-satellite/internal/pkg/log"
	"github.com/apache/skywalking-satellite/internal/satellite/event"
)

const (
	Name     = "native-pprof-grpc-forwarder"
	ShowName = "Native Pprof GRPC Forwarder"
)

type Forwarder struct {
	config.CommonFields

	pprofClient pprof.PprofTaskClient
}

func (f *Forwarder) Name() string     { return Name }
func (f *Forwarder) ShowName() string { return ShowName }
func (f *Forwarder) Description() string {
	return "This is a grpc forwarder with the SkyWalking native pprof protocol."
}
func (f *Forwarder) DefaultConfig() string { return `` }

func (f *Forwarder) Prepare(connection interface{}) error {
	client, ok := connection.(*grpc.ClientConn)
	if !ok {
		return fmt.Errorf("the %s only accepts a grpc client, but received a %s",
			f.Name(), reflect.TypeOf(connection).String())
	}
	f.pprofClient = pprof.NewPprofTaskClient(client)
	return nil
}

func (f *Forwarder) Forward(_ event.BatchEvents) error {
	return fmt.Errorf("unsupport forward")
}

func (f *Forwarder) SyncForward(e *v1.SniffData) (*v1.SniffData, grpc.ClientStream, error) {
	switch requestData := e.GetData().(type) {
	case *v1.SniffData_PprofTaskCommandQuery:
		cmds, err := f.pprofClient.GetPprofTaskCommands(context.Background(), requestData.PprofTaskCommandQuery)
		if err != nil {
			return nil, nil, err
		}
		return &v1.SniffData{Data: &v1.SniffData_Commands{Commands: cmds}}, nil, nil

	case *v1.SniffData_PprofData:
		ctx := context.WithValue(context.Background(), client_grpc.CtxBidirectionalStreamKey, true)
		stream, err := f.pprofClient.Collect(ctx)
		if err != nil {
			log.Logger.Errorf("%s open collect stream error: %v", f.Name(), err)
			return nil, nil, err
		}
		metaData := server_grpc.NewOriginalData(requestData.PprofData)
		if err = stream.SendMsg(metaData); err != nil {
			log.Logger.Errorf("%s send meta data error: %v", f.Name(), err)
			f.closeStream(stream)
			return nil, nil, err
		}
		resp, err := stream.Recv()
		if err != nil {
			log.Logger.Errorf("%s receive meta data response error: %v", f.Name(), err)
			f.closeStream(stream)
			return nil, nil, err
		}
		return &v1.SniffData{
			Data: &v1.SniffData_PprofCollectionResponse{PprofCollectionResponse: resp},
		}, stream, nil
	}

	return nil, nil, fmt.Errorf("unsupport data")
}

func (f *Forwarder) ForwardType() v1.SniffType { return v1.SniffType_PprofType }
func (f *Forwarder) SupportedSyncInvoke() bool { return true }

func (f *Forwarder) closeStream(stream pprof.PprofTask_CollectClient) {
	if err := stream.CloseSend(); err != nil {
		log.Logger.Errorf("%s close stream error: %v", f.Name(), err)
	}
}

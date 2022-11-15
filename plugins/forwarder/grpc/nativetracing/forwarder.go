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

package nativetracing

import (
	"context"
	"fmt"
	"io"
	"reflect"

	"github.com/hashicorp/go-multierror"

	"google.golang.org/grpc"

	"github.com/apache/skywalking-satellite/internal/pkg/config"
	"github.com/apache/skywalking-satellite/internal/pkg/log"
	"github.com/apache/skywalking-satellite/internal/satellite/event"
	server_grpc "github.com/apache/skywalking-satellite/plugins/server/grpc"

	v3 "skywalking.apache.org/repo/goapi/collect/common/v3"
	agent "skywalking.apache.org/repo/goapi/collect/language/agent/v3"
	v1 "skywalking.apache.org/repo/goapi/satellite/data/v1"
)

const (
	Name     = "native-tracing-grpc-forwarder"
	ShowName = "Native Tracing GRPC Forwarder"
)

type Forwarder struct {
	config.CommonFields

	tracingClient       agent.TraceSegmentReportServiceClient
	attachedEventClient agent.SpanAttachedEventReportServiceClient
}

type streaming interface {
	SendMsg(m interface{}) error
	CloseAndRecv() (*v3.Commands, error)
}

func (f *Forwarder) Name() string {
	return Name
}

func (f *Forwarder) ShowName() string {
	return ShowName
}

func (f *Forwarder) Description() string {
	return "This is a synchronization grpc forwarder with the SkyWalking native tracing protocol."
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
	f.tracingClient = agent.NewTraceSegmentReportServiceClient(client)
	f.attachedEventClient = agent.NewSpanAttachedEventReportServiceClient(client)
	return nil
}

func (f *Forwarder) Forward(batch event.BatchEvents) error {
	var tracingStream agent.TraceSegmentReportService_CollectClient
	var spanStream agent.SpanAttachedEventReportService_CollectClient

	defer func() {
		if err := closeStream(tracingStream, spanStream); err != nil {
			log.Logger.Errorf("%s close stream error: %v", f.Name(), err)
		}
	}()
	var err error
	var stream streaming
	var streamData *server_grpc.OriginalData
	for _, e := range batch {
		switch data := e.GetData().(type) {
		case *v1.SniffData_Segment:
			if tracingStream == nil {
				tracingStream, err = f.tracingClient.Collect(context.Background())
				if err != nil {
					log.Logger.Errorf("open grpc stream error %v", err)
					return err
				}
			}
			stream = tracingStream
			streamData = server_grpc.NewOriginalData(data.Segment)
		case *v1.SniffData_SpanAttachedEvent:
			if spanStream == nil {
				spanStream, err = f.attachedEventClient.Collect(context.Background())
				if err != nil {
					log.Logger.Errorf("open grpc stream error %v", err)
					return err
				}
			}
			stream = spanStream
			streamData = server_grpc.NewOriginalData(data.SpanAttachedEvent)
		default:
			continue
		}

		err = stream.SendMsg(streamData)
		if err != nil {
			log.Logger.Errorf("%s send log data error: %v", f.Name(), err)
			return err
		}
	}
	return closeStream(tracingStream, spanStream)
}

func closeStream(streams ...streaming) error {
	var err error
	for _, s := range streams {
		if s == nil {
			continue
		}
		if _, e := s.CloseAndRecv(); e != nil && e != io.EOF {
			err = multierror.Append(err, e)
		}
	}
	return err
}

func (f *Forwarder) ForwardType() v1.SniffType {
	return v1.SniffType_TracingType
}

func (f *Forwarder) SyncForward(_ *v1.SniffData) (*v1.SniffData, error) {
	return nil, fmt.Errorf("unsupport sync forward")
}

func (f *Forwarder) SupportedSyncInvoke() bool {
	return false
}

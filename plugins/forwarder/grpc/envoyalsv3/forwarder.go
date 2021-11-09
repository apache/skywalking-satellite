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

	"github.com/apache/skywalking-satellite/internal/satellite/telemetry"

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

	eventReadySendCount        *telemetry.Counter
	eventSendFinishedCount     *telemetry.Counter
	messageReadySendCount      *telemetry.Counter
	messageSendFinishedCount   *telemetry.Counter
	streamingReadySendCount    *telemetry.Counter
	streamingSendFinishedCount *telemetry.Counter

	forwardConnectTime *telemetry.Timer
	forwardSendTime    *telemetry.Timer
	forwardCloseTime   *telemetry.Timer
}

func (f *Forwarder) init() {
	f.eventReadySendCount = telemetry.NewCounter("als_event_ready_send", "Total count of the ALS event ready send.")
	f.eventSendFinishedCount = telemetry.NewCounter("als_event_send_finished", "Total count of the ALS event send finished.")
	f.messageReadySendCount = telemetry.NewCounter("als_message_ready_send", "Total count of the ALS message ready send.")
	f.messageSendFinishedCount = telemetry.NewCounter("als_message_send_finished", "Total count of the ALS event sned finished.")
	f.streamingReadySendCount = telemetry.NewCounter("als_streaming_ready_send", "Total count of the ALS streaming ready send.")
	f.streamingSendFinishedCount = telemetry.NewCounter("als_streaming_send_finished", "Total count of the ALS streaming send finished.")

	f.forwardConnectTime = telemetry.NewTimer("als_forward_connect_time", "Total time of the open ALS streaming.")
	f.forwardSendTime = telemetry.NewTimer("als_forward_send_time", "Total time of the ALS send message.")
	f.forwardCloseTime = telemetry.NewTimer("als_forward_close_time", "Total time of the ALS streaming close.")
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
	f.init()
	return nil
}

func (f *Forwarder) Forward(batch event.BatchEvents) error {
	f.eventReadySendCount.Add(float64(len(batch)))
	for _, e := range batch {
		data, _ := e.GetData().(*v1.SniffData_EnvoyALSV3List)
		f.streamingReadySendCount.Add(float64(len(data.EnvoyALSV3List.Messages)))
		for _, message := range data.EnvoyALSV3List.Messages {
			if message.GetHttpLogs() != nil {
				f.messageReadySendCount.Add(float64(len(message.GetHttpLogs().LogEntry)))
			} else if message.GetTcpLogs() != nil {
				f.messageReadySendCount.Add(float64(len(message.GetTcpLogs().LogEntry)))
			}
		}
	}

	for _, e := range batch {
		data, ok := e.GetData().(*v1.SniffData_EnvoyALSV3List)
		if !ok {
			continue
		}
		// open stream
		timeRecord := f.forwardConnectTime.Start()
		stream, err := f.alsClient.StreamAccessLogs(context.Background())
		timeRecord.Stop()
		if err != nil {
			log.Logger.Errorf("open grpc stream error %v", err)
			return err
		}

		// send message
		timeRecord = f.forwardSendTime.Start()
		for _, message := range data.EnvoyALSV3List.Messages {
			err := stream.Send(message)
			if err != nil {
				log.Logger.Errorf("%s send envoy ALS v3 data error: %v", f.Name(), err)
				f.closeStream(stream)
				return err
			}
			if message.GetHttpLogs() != nil {
				f.messageSendFinishedCount.Add(float64(len(message.GetHttpLogs().LogEntry)))
			} else if message.GetTcpLogs() != nil {
				f.messageSendFinishedCount.Add(float64(len(message.GetTcpLogs().LogEntry)))
			}
		}
		timeRecord.Stop()

		// close stream
		timeRecord = f.forwardCloseTime.Start()
		f.closeStream(stream)
		timeRecord.Stop()
		f.eventSendFinishedCount.Inc()
		f.streamingSendFinishedCount.Add(float64(len(data.EnvoyALSV3List.Messages)))
	}
	return nil
}

func (f *Forwarder) closeStream(stream v3.AccessLogService_StreamAccessLogsClient) {
	_, err := stream.CloseAndRecv()
	if err != nil && err != io.EOF {
		log.Logger.Warnf("%s close stream error: %v", f.Name(), err)
	}
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

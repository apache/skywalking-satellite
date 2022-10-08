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

package nativcelog

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"testing"
	"time"

	common "skywalking.apache.org/repo/goapi/collect/common/v3"
	logging "skywalking.apache.org/repo/goapi/collect/logging/v3"
	v1 "skywalking.apache.org/repo/goapi/satellite/data/v1"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/proto"

	"encoding/json"

	"github.com/apache/skywalking-satellite/internal/pkg/plugin"
	_ "github.com/apache/skywalking-satellite/internal/satellite/test"
	receiver "github.com/apache/skywalking-satellite/plugins/receiver/api"
	server "github.com/apache/skywalking-satellite/plugins/server/api"
	httpserver "github.com/apache/skywalking-satellite/plugins/server/http"
)

func TestReceiver_http_RegisterHandler(t *testing.T) {
	Init()
	r := initReceiver(make(plugin.Config), t)
	s := initServer(make(plugin.Config), t)
	r.RegisterHandler(s.GetServer())
	err := s.Start()
	if err != nil {
		t.Fatalf(err.Error())
	}
	time.Sleep(time.Second)
	defer func() {
		if err := s.Close(); err != nil {
			t.Fatalf("cannot close the http sever: %v", err)
		}
	}()
	for i := 0; i < 10; i++ {
		data := initData(i)
		dataBytes, err := proto.Marshal(data)
		client := http.Client{Timeout: 5 * time.Second}
		if err != nil {
			t.Fatalf("cannot marshal the data: %v", err)
		}
		go func() {
			resp, err := client.Post("http://localhost:12800/logging", "application/json", bytes.NewBuffer(dataBytes))
			if err != nil {
				fmt.Printf("cannot request the http-server , error: %v", err)
			}
			defer resp.Body.Close()
			result, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf("cannot get response from request, error: %v ", err.Error())
			}
			var response Response
			_ = json.Unmarshal(result, &response)
			if !cmp.Equal(response.Status, success) {
				panic("the response should be success, but failing")
			}
		}()

		newData := <-r.Channel()
		d := new(logging.LogData)
		_ = proto.Unmarshal(newData.Data.(*v1.SniffData_LogList).LogList.Logs[0], d)
		if !cmp.Equal(d.String(), data.String()) {
			t.Fatalf("the sent data is not equal to the received data\n, "+
				"want data %s\n, but got %s\n", data.String(), newData.String())
		}
	}
}

func TestReceiver_http_RegisterHandler_failed(t *testing.T) {
	Init()
	r := initReceiver(make(plugin.Config), t)
	s := initServer(make(plugin.Config), t)
	r.RegisterHandler(s.GetServer())
	err := s.Start()
	if err != nil {
		t.Fatalf(err.Error())
	}
	time.Sleep(time.Second)
	defer func() {
		if err = s.Close(); err != nil {
			t.Fatalf("cannot close the http sever: %v", err)
		}
	}()
	data := initData(0)
	dataBytes, err := json.Marshal(data)
	client := http.Client{Timeout: 5 * time.Second}
	if err != nil {
		t.Fatalf("cannot marshal the data: %v", err)
	}
	resp, err := client.Post("http://localhost:12800/logging", "application/json", bytes.NewBuffer(dataBytes))
	if err != nil {
		fmt.Printf("cannot request the http-server , error: %v", err)
	}
	defer resp.Body.Close()
	result, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("cannot get response from request, error: %v ", err.Error())
	}
	var response Response
	_ = json.Unmarshal(result, &response)
	if !cmp.Equal(response.Status, failing) {
		panic("the response should be failing, but success")
	}
}

func initData(sequence int) *logging.LogData {
	seq := strconv.Itoa(sequence)
	return &logging.LogData{
		Timestamp:       time.Now().Unix(),
		Service:         "demo-service" + seq,
		ServiceInstance: "demo-instance" + seq,
		Endpoint:        "demo-endpoint" + seq,
		TraceContext: &logging.TraceContext{
			TraceSegmentId: "mock-segmentId" + seq,
			TraceId:        "mock-traceId" + seq,
			SpanId:         1,
		},
		Tags: &logging.LogTags{
			Data: []*common.KeyStringValuePair{
				{
					Key:   "mock-key" + seq,
					Value: "mock-value" + seq,
				},
			},
		},
		Body: &logging.LogDataBody{
			Type: "mock-type" + seq,
			Content: &logging.LogDataBody_Text{
				Text: &logging.TextLog{
					Text: "this is a mock text mock log" + seq,
				},
			},
		},
	}
}

func Init() {
	plugin.RegisterPluginCategory(reflect.TypeOf((*server.Server)(nil)).Elem())
	plugin.RegisterPluginCategory(reflect.TypeOf((*receiver.Receiver)(nil)).Elem())
	plugin.RegisterPlugin(new(httpserver.Server))
	plugin.RegisterPlugin(new(Receiver))
}

func initServer(cfg plugin.Config, t *testing.T) server.Server {
	cfg[plugin.NameField] = httpserver.Name
	q := server.GetServer(cfg)
	if q == nil {
		t.Fatalf("cannot get a http server from the registry")
	}
	if err := q.Prepare(); err != nil {
		t.Fatalf("cannot perpare the http server: %v", err)
	}
	return q
}

func initReceiver(cfg plugin.Config, t *testing.T) receiver.Receiver {
	cfg[plugin.NameField] = Name
	q := receiver.GetReceiver(cfg)
	if q == nil {
		t.Fatalf("cannot get http-log-receiver from the registry")
	}
	return q
}

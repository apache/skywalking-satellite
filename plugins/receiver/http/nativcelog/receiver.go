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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	module "github.com/apache/skywalking-satellite/internal/satellite/module/api"

	"google.golang.org/protobuf/proto"

	"github.com/apache/skywalking-satellite/internal/pkg/config"
	"github.com/apache/skywalking-satellite/internal/pkg/log"
	forwarder "github.com/apache/skywalking-satellite/plugins/forwarder/api"
	frowarder_nativelog "github.com/apache/skywalking-satellite/plugins/forwarder/grpc/nativelog"
	http_server "github.com/apache/skywalking-satellite/plugins/server/http"

	logging "skywalking.apache.org/repo/goapi/collect/logging/v3"
	v1 "skywalking.apache.org/repo/goapi/satellite/data/v1"
)

const (
	Name      = "http-native-log-receiver"
	ShowName  = "HTTP Native Log Receiver"
	eventName = "http-nativelog-event"
	success   = "success"
	failing   = "failing"
)

type Receiver struct {
	config.CommonFields
	// config
	URI     string `mapstructure:"uri"`
	Timeout int    `mapstructure:"timeout"`
	// components
	Server        *http_server.Server
	OutputChannel chan *v1.SniffData
}

type Response struct {
	Status string `json:"status"`
	Msg    string `json:"msg"`
}

func (r *Receiver) Name() string {
	return Name
}

func (r *Receiver) ShowName() string {
	return ShowName
}

func (r *Receiver) Description() string {
	return "This is a receiver for SkyWalking http logging format, " +
		"which is defined at https://github.com/apache/skywalking-data-collect-protocol/blob/master/logging/Logging.proto."
}

func (r *Receiver) DefaultConfig() string {
	return `
# The native log request URI.
uri: "/logging"
# The request timeout seconds.
timeout: 5
`
}

func (r *Receiver) RegisterHandler(server interface{}) {
	r.Server = server.(*http_server.Server)
	r.OutputChannel = make(chan *v1.SniffData)
	r.Server.Server.Handle(r.URI, r.httpHandler())
}

func (r *Receiver) RegisterSyncInvoker(_ module.SyncInvoker) {
}

func ResponseWithJSON(rsp http.ResponseWriter, response *Response, code int) {
	rsp.WriteHeader(code)
	_ = json.NewEncoder(rsp).Encode(response)
}

func (r *Receiver) httpHandler() http.Handler {
	h := http.HandlerFunc(func(rsp http.ResponseWriter, req *http.Request) {
		rsp.Header().Set("Content-Type", "application/json")
		b, err := io.ReadAll(req.Body)
		if err != nil {
			log.Logger.Errorf("get http body error: %v", err)
			response := &Response{Status: failing, Msg: err.Error()}
			ResponseWithJSON(rsp, response, http.StatusBadRequest)
			return
		}
		var data logging.LogData
		err = proto.Unmarshal(b, &data)
		if err != nil {
			response := &Response{Status: failing, Msg: err.Error()}
			ResponseWithJSON(rsp, response, http.StatusInternalServerError)
			return
		}
		e := &v1.SniffData{
			Name:      eventName,
			Timestamp: time.Now().UnixNano() / 1e6,
			Meta:      nil,
			Type:      v1.SniffType_Logging,
			Remote:    true,
			Data: &v1.SniffData_LogList{
				LogList: &v1.BatchLogList{
					Logs: [][]byte{b},
				},
			},
		}
		r.OutputChannel <- e
		response := &Response{Status: success, Msg: success}
		ResponseWithJSON(rsp, response, http.StatusOK)
	})
	return http.TimeoutHandler(h, time.Duration(r.Timeout)*time.Second, fmt.Sprintf("Exceeded configured timeout of %d seconds", r.Timeout))
}

func (r *Receiver) Channel() <-chan *v1.SniffData {
	return r.OutputChannel
}

func (r *Receiver) SupportForwarders() []forwarder.Forwarder {
	return []forwarder.Forwarder{
		new(frowarder_nativelog.Forwarder),
	}
}

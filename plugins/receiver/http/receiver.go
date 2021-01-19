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

package http

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	logging "skywalking/network/logging/v3"

	"google.golang.org/protobuf/proto"

	"encoding/json"

	"github.com/apache/skywalking-satellite/internal/pkg/config"
	"github.com/apache/skywalking-satellite/internal/pkg/log"
	http_server "github.com/apache/skywalking-satellite/plugins/server/http"
	"github.com/apache/skywalking-satellite/protocol/gen-codes/satellite/protocol"
)

const (
	Name      = "http-log-receiver"
	eventName = "http-log-event"
	timeout   = 5 * time.Second
	Success   = "success"
	Failing   = "failing"
)

type Receiver struct {
	config.CommonFields
	Server        *http_server.Server
	OutputChannel chan *protocol.Event
	URI           string `mapstructure:"uri"`
}

type Response struct {
	Status string `json:"status"`
	Msg    string `json:"msg"`
}

func (r *Receiver) Name() string {
	return Name
}

func (r *Receiver) Description() string {
	return "This is a receiver for SkyWalking http logging format, " +
		"which is defined at https://github.com/apache/skywalking-data-collect-protocol/blob/master/logging/Logging.proto."
}

func (r *Receiver) DefaultConfig() string {
	return `
# The http server uri .
uri: "/logging"
`
}

func (r *Receiver) RegisterHandler(server interface{}) {
	r.Server = server.(*http_server.Server)
	r.OutputChannel = make(chan *protocol.Event)
	r.Server.Server.Handle(r.URI, httpHandler(r))
}

func httpHandler(r *Receiver) http.Handler {
	h := http.HandlerFunc(func(rsp http.ResponseWriter, req *http.Request) {
		rsp.Header().Set("Content-Type", "application/json")
		b, err := ioutil.ReadAll(req.Body)
		if err != nil {
			log.Logger.Errorf("get http body error: %v", err)
			response := &Response{Status: Failing, Msg: err.Error()}
			json.NewEncoder(rsp).Encode(response)
			http.Error(rsp, err.Error(), http.StatusBadRequest)
			return
		}
		var data logging.LogData
		err = proto.Unmarshal(b, &data)
		if err != nil {
			response := &Response{Status: Failing, Msg: err.Error()}
			json.NewEncoder(rsp).Encode(response)
			http.Error(rsp, err.Error(), http.StatusInternalServerError)
			return
		}
		e := &protocol.Event{
			Name:      eventName,
			Timestamp: time.Now().UnixNano() / 1e6,
			Meta:      nil,
			Type:      protocol.EventType_Logging,
			Remote:    true,
			Data: &protocol.Event_Log{
				Log: &data,
			},
		}
		r.OutputChannel <- e
		response := &Response{Status: Success, Msg: Success}
		json.NewEncoder(rsp).Encode(response)
		return
	})
	return http.TimeoutHandler(h, timeout, fmt.Sprintf("Exceeded configured timeout of %v \n", timeout))
}

func (r *Receiver) Channel() <-chan *protocol.Event {
	return r.OutputChannel
}

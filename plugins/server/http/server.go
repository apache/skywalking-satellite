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
	"net/http"

	"github.com/apache/skywalking-satellite/internal/pkg/config"
	"github.com/apache/skywalking-satellite/internal/pkg/log"
)

const Name = "http-server"

type Server struct {
	config.CommonFields
	Address string `mapstructure:"address"`
	URI     string `mapstructure:"uri"`

	Server *http.ServeMux // The http server.
}

func (s *Server) Name() string {
	return Name
}

func (s *Server) Description() string {
	return "this is a http server for receive logs."
}

func (s *Server) DefaultConfig() string {
	return `
# The http server address.
address: ":8080"
# The http server .
uri: "/logging"
`
}

func (s *Server) Prepare() error {
	s.Server = http.NewServeMux()
	return nil
}

func (s *Server) Start() error {
	go func() {
		err := http.ListenAndServe(s.Address, s.Server)
		if err != nil {
			log.Logger.Errorf("start http server error: %v", err)
		}
	}()
	return nil
}

func (s *Server) Close() error {
	return nil
}

func (s *Server) GetServer() interface{} {
	return s
}

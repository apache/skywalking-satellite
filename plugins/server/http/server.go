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

const (
	Name     = "http-server"
	ShowName = "HTTP Server"
)

type Server struct {
	config.CommonFields
	Address string         `mapstructure:"address"`
	Server  *http.ServeMux // The http server.
}

func (s *Server) Name() string {
	return Name
}

func (s *Server) ShowName() string {
	return ShowName
}

func (s *Server) Description() string {
	return "This is a sharing plugin, which would start a http server."
}

func (s *Server) DefaultConfig() string {
	return `
# The http server address.
address: ":12800"
`
}

func (s *Server) Prepare() error {
	s.Server = http.NewServeMux()
	return nil
}

func (s *Server) Start() error {
	log.Logger.WithField("address", s.Address).Info("http server is starting...")
	go func() {
		err := http.ListenAndServe(s.Address, s.Server)
		if err != nil {
			log.Logger.WithField("address", s.Address).Infof("http server has failure when starting: %v", err)
		}
	}()
	return nil
}

func (s *Server) Close() error {
	log.Logger.Info("http server is closed")
	return nil
}

func (s *Server) GetServer() interface{} {
	return s
}

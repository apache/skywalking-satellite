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

package grpc

import (
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/apache/skywalking-satellite/internal/pkg/config"
	"github.com/apache/skywalking-satellite/internal/pkg/log"
)

const (
	Name     = "grpc-server"
	ShowName = "GRPC Server"
)

type Server struct {
	config.CommonFields
	Address              string `mapstructure:"address"`                // The address of grpc server.
	Network              string `mapstructure:"network"`                // The network of grpc.
	MaxRecvMsgSize       int    `mapstructure:"max_recv_msg_size"`      // The max size of the received log.
	MaxConcurrentStreams uint32 `mapstructure:"max_concurrent_streams"` // The max concurrent stream channels.
	TLSCertFile          string `mapstructure:"tls_cert_file"`          // The TLS cert file path.
	TLSKeyFile           string `mapstructure:"tls_key_file"`           // The TLS key file path.

	AcceptLimit AcceptConnectionConfig `mapstructure:"accept_limit"` // To Accept Connection Limiter when reach the resource

	// components
	server   *grpc.Server
	listener net.Listener
}

func (s *Server) Name() string {
	return Name
}

func (s *Server) ShowName() string {
	return ShowName
}

func (s *Server) Description() string {
	return "This is a sharing plugin, which would start a gRPC server."
}

func (s *Server) DefaultConfig() string {
	return `
# The address of grpc server. Default value is :11800
address: :11800
# The network of grpc. Default value is :tcp
network: tcp
# The max size of receiving log. Default value is 2M. The unit is Byte.
max_recv_msg_size: 2097152
# The max concurrent stream channels.
max_concurrent_streams: 32
# The TLS cert file path.
tls_cert_file: ""
# The TLS key file path.
tls_key_file: ""
# To Accept Connection Limiter when reach the resource
accept_limit:
  # The max CPU utilization limit
  cpu_utilization: 75
  # The max connection count
  connection_count: 4000
`
}

func (s *Server) Prepare() error {
	var options []grpc.ServerOption
	if s.TLSCertFile != "" && s.TLSKeyFile != "" {
		if c, err := credentials.NewServerTLSFromFile(s.TLSCertFile, s.TLSKeyFile); err == nil {
			options = append(options, grpc.Creds(c))
		} else {
			log.Logger.Errorf("error in checking TLS files: %v", err)
			return err
		}
	}
	options = append(options, grpc.MaxRecvMsgSize(s.MaxRecvMsgSize), grpc.MaxConcurrentStreams(s.MaxConcurrentStreams))
	s.server = grpc.NewServer(options...)
	listener, err := NewConnectionManager(s.Network, s.Address, s.AcceptLimit)
	if err != nil {
		log.Logger.Errorf("grpc server cannot be created: %v", err)
		return err
	}
	s.listener = listener
	return nil
}

func (s *Server) Start() error {
	go func() {
		log.Logger.WithField("address", s.Address).Info("grpc server is starting...")
		if err := s.server.Serve(s.listener); err != nil {
			log.Logger.WithField("address", s.Address).Infof("grpc server has failure when starting: %v", err)
		}
	}()
	return nil
}

func (s *Server) Close() error {
	s.server.Stop()
	log.Logger.Info("grpc server is closed")
	return nil
}

func (s *Server) GetServer() interface{} {
	return s.server
}

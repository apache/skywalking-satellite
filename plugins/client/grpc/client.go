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
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"

	"github.com/apache/skywalking-satellite/internal/pkg/config"
	"github.com/apache/skywalking-satellite/internal/pkg/log"
	"github.com/apache/skywalking-satellite/plugins/client/api"
	"github.com/apache/skywalking-satellite/plugins/client/grpc/resolvers"
)

const (
	Name     = "grpc-client"
	ShowName = "GRPC Client"
)

type Client struct {
	config.CommonFields
	// server finder config
	ServerFinderConfig resolvers.ServerFinderConfig `mapstructure:",squash"`

	EnableTLS          bool   `mapstructure:"enable_TLS"`           // Enable TLS connect to server
	ClientPemPath      string `mapstructure:"client_pem_path"`      // The file path of client.pem. The config only works when opening the TLS switch.
	ClientKeyPath      string `mapstructure:"client_key_path"`      // The file path of client.key. The config only works when opening the TLS switch.
	CaPemPath          string `mapstructure:"ca_pem_path"`          // The file path oca.pem. The config only works when opening the TLS switch.
	InsecureSkipVerify bool   `mapstructure:"insecure_skip_verify"` // Controls whether a client verifies the server's certificate chain and host name.
	Authentication     string `mapstructure:"authentication"`       // The auth value when send request
	CheckPeriod        int    `mapstructure:"check_period"`         // How frequently to check the connection(second)

	// components
	status    api.ClientStatus
	client    *grpc.ClientConn
	listeners []chan<- api.ClientStatus
	ctx       context.Context    // Parent ctx
	cancel    context.CancelFunc // Parent ctx cancel function
}

func (c *Client) Name() string {
	return Name
}

func (c *Client) ShowName() string {
	return ShowName
}

func (c *Client) Description() string {
	return "The gRPC client is a sharing plugin to keep connection with the gRPC server and delivery the data to it."
}

func (c *Client) DefaultConfig() string {
	return `
# The gRPC client finder type
finder_type: "static"

# The gRPC server address (default localhost:11800), multiple addresses are split by ",".
server_addr: localhost:11800

# The gRPC kubernetes server address finder
kubernetes_config:
  # The kind of resource
  kind: pod
  # The resource namespaces
  namespaces:
    - default
  # How to get the address exported port
  extra_port:
    # Resource target port
    port: 11800

# The TLS switch (default false).
enable_TLS: false

# The file path of client.pem. The config only works when opening the TLS switch.
client_pem_path: ""

# The file path of client.key. The config only works when opening the TLS switch.
client_key_path: ""

# The file path oca.pem. The config only works when opening the TLS switch.
ca_pem_path: ""

# InsecureSkipVerify controls whether a client verifies the server's certificate chain and host name.
insecure_skip_verify: true

# The auth value when send request
authentication: ""

# How frequently to check the connection(second)
check_period: 5
`
}

func (c *Client) Prepare() error {
	// config
	cfg, err := c.loadConfig()
	if err != nil {
		return fmt.Errorf("cannot init the grpc client: %v", err)
	}

	// logger
	grpclog.SetLoggerV2(&logrusGrpcLoggerV2{log.Logger.WithFields(logrus.Fields{
		"client_name": Name,
	})})

	// server address resolver
	resolvers.RegisterAllGrpcResolvers()

	// connect to server
	target, err := resolvers.BuildTarget(&c.ServerFinderConfig)
	if err != nil {
		return fmt.Errorf("cannot build grpc target: %v", err)
	}
	client, err := grpc.Dial(target, *cfg...)
	if err != nil {
		return fmt.Errorf("cannot connect to grpc server: %v", err)
	}

	c.client = client
	c.status = api.Connected
	c.ctx, c.cancel = context.WithCancel(context.Background())
	c.listeners = make([]chan<- api.ClientStatus, 0)
	return nil
}

func (c *Client) Close() error {
	c.cancel()
	defer log.Logger.Info("grpc client is closed")
	return c.client.Close()
}

func (c *Client) GetConnectedClient() interface{} {
	return c.client
}

func (c *Client) RegisterListener(listener chan<- api.ClientStatus) {
	c.listeners = append(c.listeners, listener)
}

func (c *Client) Start() error {
	go c.snifferChannelStatus()
	return nil
}

// grpc log adaptor
type logrusGrpcLoggerV2 struct {
	*logrus.Entry
}

func (l *logrusGrpcLoggerV2) V(level int) bool {
	return l.Logger.IsLevelEnabled(logrus.Level(level))
}

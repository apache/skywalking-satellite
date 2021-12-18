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
	"sync"

	"github.com/apache/skywalking-satellite/internal/pkg/log"
)

type ConnectionManager struct {
	net.Listener
	acceptLimiter *AcceptLimiter
}

func NewConnectionManager(network, address string, acceptConnectionConfig AcceptConnectionConfig) (*ConnectionManager, error) {
	listen, err := net.Listen(network, address)
	if err != nil {
		return nil, err
	}

	limiter, err := NewAcceptLimiter(acceptConnectionConfig)
	if err != nil {
		return nil, err
	}
	return &ConnectionManager{listen, limiter}, nil
}

func (c *ConnectionManager) Accept() (net.Conn, error) {
	conn, err := c.Listener.Accept()
	if err != nil {
		return nil, err
	}

	if !c.acceptLimiter.CouldHandleConnection() {
		conn.Close()
		log.Logger.Warnf("out of accept limit, drop the connection: %v->%v, environment: cpuUtilization: %f, connectionCount: %d",
			conn.RemoteAddr(), conn.LocalAddr(), c.acceptLimiter.CurrentCPU, c.acceptLimiter.ActiveConnection)
		return nil, &outOfLimit{}
	}
	return &ConnectionWrapper{Conn: conn, manager: c}, nil
}

func (c *ConnectionManager) notifyCloseConnection() {
	c.acceptLimiter.CloseConnection()
}

type ConnectionWrapper struct {
	net.Conn
	manager   *ConnectionManager
	closeOnce sync.Once
}

func (c *ConnectionWrapper) Close() error {
	defer c.CloseNotify()
	return c.Conn.Close()
}

func (c *ConnectionWrapper) CloseNotify() {
	c.closeOnce.Do(func() {
		c.manager.notifyCloseConnection()
	})
}

type outOfLimit struct {
	error
}

func (o *outOfLimit) Temporary() bool {
	return true
}

func (o *outOfLimit) Error() string {
	return "out of limit"
}

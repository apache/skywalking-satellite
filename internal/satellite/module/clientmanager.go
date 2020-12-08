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

package module

import (
	"context"
	"sync"
	"time"

	"github.com/apache/skywalking-satellite/internal/pkg/constant"
	"github.com/apache/skywalking-satellite/internal/pkg/log"
	"github.com/apache/skywalking-satellite/internal/satellite/config"
	client "github.com/apache/skywalking-satellite/plugins/client/api"
)

// The specific client statuses.
const (
	_ ClientStatus = iota
	Connected
	Disconnect
)

// ClientStatus represents the status of the client.
type ClientStatus int8

// ClientManager is a module plugin to control the connection with the outer service.
type ClientManager struct {
	sync.Mutex

	// config
	config *config.ClientManagerConfig

	// dependency plugins
	runningClient client.Client

	// self components
	listeners []chan ClientStatus // the sender client status listeners

	// Metrics:
	retryCount int64
	status     ClientStatus
}

func (c *ClientManager) Name() string {
	return constant.ClientManagerModule
}

func (c *ClientManager) Description() string {
	return "keep connection with external services, such as Kafka and OAP."
}

func (c *ClientManager) Config() config.ModuleConfig {
	return c.config
}

// Init ClientManager, dependency plugins and self components.
func (c *ClientManager) Init(cfg config.ModuleConfig) {
	log.Logger.Infof("%s module of %s namespace is being initialized", c.Name(), c.config.NameSpace())
	c.config = cfg.(*config.ClientManagerConfig)
	c.runningClient = client.GetClient(c.config.ClientConfig)
	c.listeners = []chan ClientStatus{}
}

// Prepare makes the connection with external services, such as Kafka and OAP.
func (c *ClientManager) Prepare() error {
	log.Logger.Infof("%s module of %s namespace is in preparing stage", c.Name(), c.config.NameSpace())
	return c.initializeClient()
}

// Boot start ClientManager to maintain the connection with external services.
func (c *ClientManager) Boot(ctx context.Context) {
	log.Logger.Infof("%s module of %s namespace is running", c.Name(), c.config.NameSpace())
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		timeTicker := time.NewTicker(time.Duration(c.config.RetryInterval) * time.Millisecond)
		for {
			select {
			case <-timeTicker.C:
				if err := c.connectClient(); err != nil {
					log.Logger.Errorf("cannot make a connection with the %s client", c.runningClient.Name())
				}
			case <-ctx.Done():
				c.Shutdown()
				return
			}
		}
	}()
	wg.Wait()
}

// Shutdown close the connection and listeners.
func (c *ClientManager) Shutdown() {
	log.Logger.Infof("%s module of %s namespace is closing", c.Name(), c.config.NameSpace())
	for _, listener := range c.listeners {
		close(listener)
	}
	if err := c.runningClient.Close(); err != nil {
		log.Logger.Errorf("an error occurring when closing %s client: %v", c.runningClient.Name(), err)
	}
}

// RegisterListener adds the listener to listen to the status of the client.
func (c *ClientManager) RegisterListener(listener chan ClientStatus) {
	c.Lock()
	defer c.Unlock()
	c.listeners = append(c.listeners, listener)
}

// GetClient returns a connected client when . Otherwise, would return a nil client.
func (c *ClientManager) GetConnectedClient() interface{} {
	return c.runningClient.GetConnectedClient()
}

// ReportError reports the client is disconnect.
func (c *ClientManager) ReportError() {
	c.Lock()
	defer c.Unlock()
	if c.status == Connected {
		c.status = Disconnect
		c.notify()
	}
}

// initializeClient initialize the connection with external services and retry one time when initialize failed.
func (c *ClientManager) initializeClient() error {
	c.Lock()
	defer c.Unlock()
	if err := c.connectClient(); err != nil {
		log.Logger.Infof("preparing to reconnect with %s client,retrying in 10s", c.runningClient.Name())
		time.Sleep(10 * time.Second)
		return c.connectClient()
	}
	return nil
}

// connectClient would make a connection with external services, such as Kafka and OAP. When successfully connected,
// ClientManager would notify the connected status to all senders.
func (c *ClientManager) connectClient() error {
	c.Lock()
	defer c.Unlock()
	if c.runningClient.IsConnected() {
		return nil
	}
	log.Logger.Infof("preparing to connect with %s client", c.runningClient.Name())
	if c.runningClient.IsConnected() {
		return nil
	}
	c.retryCount++
	err := c.runningClient.Connect()
	if err == nil {
		c.status = Connected
		c.notify()
		log.Logger.Infof("successfully connected to %s client", c.runningClient.Name())
	}
	return err
}

// notify the current status to all listeners.
func (c *ClientManager) notify() {
	c.Lock()
	defer c.Unlock()
	for _, listener := range c.listeners {
		listener <- c.status
	}
}

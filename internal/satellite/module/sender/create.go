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

package sender

import (
	"github.com/apache/skywalking-satellite/internal/pkg/log"
	"github.com/apache/skywalking-satellite/internal/satellite/event"
	"github.com/apache/skywalking-satellite/internal/satellite/module/buffer"
	gatherer "github.com/apache/skywalking-satellite/internal/satellite/module/gatherer/api"
	"github.com/apache/skywalking-satellite/internal/satellite/module/sender/api"
	"github.com/apache/skywalking-satellite/internal/satellite/sharing"
	client "github.com/apache/skywalking-satellite/plugins/client/api"
	fallbacker "github.com/apache/skywalking-satellite/plugins/fallbacker/api"
	forwarder "github.com/apache/skywalking-satellite/plugins/forwarder/api"
)

// NewSender crate a Sender.
func NewSender(cfg *api.SenderConfig, g gatherer.Gatherer) api.Sender {
	log.Logger.Infof("sender module of %s namespace is being initialized", cfg.PipeName)
	s := &Sender{
		config:            cfg,
		runningForwarders: []forwarder.Forwarder{},
		runningFallbacker: fallbacker.GetFallbacker(cfg.FallbackerConfig),
		runningClient:     sharing.Manager[cfg.ClientName].(client.Client),
		gatherer:          g,
		logicInput:        nil,
		physicalInput:     make(chan *event.OutputEventContext),
		listener:          make(chan client.ClientStatus),
		flushChannel:      make(chan *buffer.BatchBuffer, 1),
		buffer:            buffer.NewBatchBuffer(cfg.MaxBufferSize),
	}
	for _, c := range s.config.ForwardersConfig {
		s.runningForwarders = append(s.runningForwarders, forwarder.GetForwarder(c))
	}
	return s
}

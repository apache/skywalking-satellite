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
	"github.com/apache/skywalking-satellite/internal/pkg/event"
	"github.com/apache/skywalking-satellite/internal/pkg/log"
	"github.com/apache/skywalking-satellite/internal/satellite/config"
	"github.com/apache/skywalking-satellite/internal/satellite/module/buffer"
	fallbacker "github.com/apache/skywalking-satellite/plugins/fallbacker/api"
	forwarder "github.com/apache/skywalking-satellite/plugins/forwarder/api"
)

// Sender is the forward module in Satellite.
type Sender struct {
	// config
	config *config.SenderConfig

	// dependency plugins
	runningForwarders []forwarder.Forwarder
	runningFallbacker fallbacker.Fallbacker

	// dependency modules
	gatherer      *Gatherer
	clientManager *ClientManager

	// self components
	Input        chan *event.OutputEventContext // logic input channel
	input        chan *event.OutputEventContext // physical input channel
	listener     chan ClientStatus              // client status listener
	flushChannel chan *buffer.BatchBuffer       // forwarder flush channel
	buffer       *buffer.BatchBuffer            // cache the downstream input data
}

func (s *Sender) Name() string {
	return constant.SenderModule
}

func (s *Sender) Description() string {
	return "forward the input events to external services, such as Kafka and OAP"
}

func (s *Sender) Config() config.ModuleConfig {
	return s.config
}

// Init Sender, dependency plugins and self components.
func (s *Sender) Init(cfg config.ModuleConfig) {
	log.Logger.Infof("%s module of %s namespace is being initialized", s.Name(), s.config.NameSpace())
	s.config = cfg.(*config.SenderConfig)
	s.runningFallbacker = fallbacker.GetFallbacker(s.config.FallbackerConfig)
	s.runningForwarders = []forwarder.Forwarder{}
	for _, c := range s.config.ForwardersConfig {
		s.runningForwarders = append(s.runningForwarders, forwarder.GetForwarder(c))
	}
	s.input = make(chan *event.OutputEventContext)
	s.Input = s.input
	s.listener = make(chan ClientStatus)
	s.flushChannel = make(chan *buffer.BatchBuffer, 1)
	s.buffer = buffer.NewBatchBuffer(s.config.MaxBufferSize)
}

// Prepare inject the dependency modules and register the client status listener to the clientManager.
func (s *Sender) Prepare() error {
	log.Logger.Infof("%s module of %s namespace is in preparing stage", s.Name(), s.config.NameSpace())
	s.clientManager = GetRunningModule(s.config.NameSpace(), constant.ClientManagerModule).(*ClientManager)
	s.gatherer = GetRunningModule(s.config.NameSpace(), constant.GathererModule).(*Gatherer)
	s.clientManager.RegisterListener(s.listener)
	return nil
}

// Boot fetches the downstream input data and forward to external services, such as Kafka and OAP receiver.
func (s *Sender) Boot(ctx context.Context) {
	log.Logger.Infof("%s module of %s namespace is running", s.Name(), s.config.NameSpace())
	var wg sync.WaitGroup
	wg.Add(2)
	// 1. keep fetching the downstream data when client connected, and put it into BatchBuffer.
	// 2. When reaches the buffer limit or receives a timer flush signal, and put BatchBuffer into flushChannel.
	go func() {
		defer wg.Done()
		timeTicker := time.NewTicker(time.Duration(s.config.FlushTime) * time.Millisecond)
		for {
			select {
			case status := <-s.listener:
				switch status {
				case Connected:
					log.Logger.Infof("%s module of %s namespace is notified the connection is connected", s.Name(), s.config.NameSpace())
					s.Input = s.input
				case Disconnect:
					log.Logger.Infof("%s module of %s namespace is notified the connection is disconnected", s.Name(), s.config.NameSpace())
					s.Input = nil
				}
			case <-timeTicker.C:
				if s.buffer.Len() > s.config.MinFlushEvents {
					s.flushChannel <- s.buffer
					s.buffer = buffer.NewBatchBuffer(s.config.MaxBufferSize)
				}
			case e := <-s.Input:
				s.buffer.Add(e)
				if s.buffer.Len() == s.config.MaxBufferSize {
					s.flushChannel <- s.buffer
					s.buffer = buffer.NewBatchBuffer(s.config.MaxBufferSize)
				}
			case <-ctx.Done():
				s.Input = nil
				return
			}
		}
	}()
	// Keep fetching BatchBuffer to forward.
	go func() {
		defer wg.Done()
		for {
			select {
			case b := <-s.flushChannel:
				s.consume(b)
			case <-ctx.Done():
				s.Shutdown()
				return
			}
		}
	}()
	wg.Wait()
}

// Shutdown closes the channels and tries to force forward the events in the buffer.
func (s *Sender) Shutdown() {
	log.Logger.Infof("%s module of %s namespace is closing", s.Name(), s.config.NameSpace())
	close(s.input)
	for buf := range s.flushChannel {
		s.consume(buf)
	}
	s.consume(s.buffer)
	close(s.flushChannel)
}

// consume would forward the events by type and ack this batch.
func (s *Sender) consume(batch *buffer.BatchBuffer) {
	log.Logger.Infof("%s module of %s namespace is flushing a new batch buffer. the start offset is %d, and the batch size is %d",
		s.Name(), s.config.NameSpace(), batch.First(), batch.BatchSize())
	var events = make(map[event.Type]event.BatchEvents)
	for i := 0; i < batch.Len(); i++ {
		eventContext := batch.Buf()[i]
		for _, e := range eventContext.Context {
			if e.IsRemote() {
				events[e.Type()] = append(events[e.Type()], e)
			}
		}
	}

	for _, f := range s.runningForwarders {
		for t, batchEvents := range events {
			if f.ForwardType() == t {
				s.runningFallbacker.FallBack(batchEvents, s.clientManager.GetConnectedClient(), f.Forward, func() {
					s.clientManager.ReportError()
				})
			}
		}
	}
	s.gatherer.Ack(batch.First(), batch.BatchSize())
}

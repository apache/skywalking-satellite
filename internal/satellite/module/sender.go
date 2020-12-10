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

	"github.com/apache/skywalking-satellite/internal/pkg/event"
	"github.com/apache/skywalking-satellite/internal/pkg/log"
	"github.com/apache/skywalking-satellite/internal/pkg/plugin"
	"github.com/apache/skywalking-satellite/internal/satellite/module/api"
	"github.com/apache/skywalking-satellite/internal/satellite/module/buffer"
	fallbacker "github.com/apache/skywalking-satellite/plugins/fallbacker/api"
	forwarder "github.com/apache/skywalking-satellite/plugins/forwarder/api"
)

type SenderConfig struct {
	api.ModuleCommonConfig
	// plugins config
	ForwardersConfig []plugin.Config `mapstructure:"forwarders"` // forwarder plugins config
	FallbackerConfig plugin.Config   `mapstructure:"fallbacker"` // fallbacker plugins config

	// self config
	MaxBufferSize  int `mapstructure:"max_buffer_size"`  // the max buffer capacity
	MinFlushEvents int `mapstructure:"min_flush_events"` // the min flush events when receives a timer flush signal
	FlushTime      int `mapstructure:"flush_time"`       // the period flush time
}

// Sender is the forward module in Satellite.
type Sender struct {
	// config
	config *SenderConfig

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

// Init Sender, dependency plugins and self components.
func NewSender(cfg *SenderConfig, gatherer *Gatherer, manager *ClientManager) *Sender {
	log.Logger.Infof("sender module of %s namespace is being initialized", cfg.RunningNamespace)
	s := &Sender{
		gatherer:          gatherer,
		clientManager:     manager,
		config:            cfg,
		flushChannel:      make(chan *buffer.BatchBuffer, 1),
		listener:          make(chan ClientStatus),
		input:             make(chan *event.OutputEventContext),
		runningFallbacker: fallbacker.GetFallbacker(cfg.FallbackerConfig),
		runningForwarders: []forwarder.Forwarder{},
		buffer:            buffer.NewBatchBuffer(cfg.MaxBufferSize),
		Input:             nil,
	}
	for _, c := range s.config.ForwardersConfig {
		s.runningForwarders = append(s.runningForwarders, forwarder.GetForwarder(c))
	}
	return s
}

// Prepare register the client status listener to the client manager and open input channel.
func (s *Sender) Prepare() error {
	log.Logger.Infof("sender module of %s namespace is preparing", s.config.RunningNamespace)
	s.clientManager.RegisterListener(s.listener)
	s.Input = s.input
	return nil
}

// Boot fetches the downstream input data and forward to external services, such as Kafka and OAP receiver.
func (s *Sender) Boot(ctx context.Context) {
	log.Logger.Infof("sender module of %s namespace is running", s.config.RunningNamespace)
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
					log.Logger.Infof("sender module of %s namespace is notified the connection is connected", s.config.RunningNamespace)
					s.Input = s.input
				case Disconnect:
					log.Logger.Infof("sender module of %s namespace is notified the connection is disconnected", s.config.RunningNamespace)
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
	log.Logger.Infof("sender module of %s namespace is closing", s.config.RunningNamespace)
	close(s.input)
	for buf := range s.flushChannel {
		s.consume(buf)
	}
	s.consume(s.buffer)
	close(s.flushChannel)
}

// consume would forward the events by type and ack this batch.
func (s *Sender) consume(batch *buffer.BatchBuffer) {
	log.Logger.Infof("sender module of %s namespace is flushing a new batch buffer. the start offset is %d, and the batch size is %d",
		s.config.RunningNamespace, batch.First(), batch.BatchSize())
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
			if f.ForwardType() != t {
				continue
			}
			if err := f.Forward(s.clientManager.GetConnectedClient(), batchEvents); err == nil {
				continue
			}
			if !s.runningFallbacker.FallBack(batchEvents, s.clientManager.GetConnectedClient(), f.Forward) {
				if s.clientManager.runningClient.IsConnected() {
					s.clientManager.ReportError()
				}
			}
		}
	}
	s.gatherer.Ack(batch.First(), batch.BatchSize())
}

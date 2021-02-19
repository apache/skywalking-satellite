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
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/apache/skywalking-satellite/internal/pkg/log"
	"github.com/apache/skywalking-satellite/internal/satellite/event"
	"github.com/apache/skywalking-satellite/internal/satellite/module/buffer"
	gatherer "github.com/apache/skywalking-satellite/internal/satellite/module/gatherer/api"
	"github.com/apache/skywalking-satellite/internal/satellite/module/sender/api"
	"github.com/apache/skywalking-satellite/internal/satellite/telemetry"
	client "github.com/apache/skywalking-satellite/plugins/client/api"
	fallbacker "github.com/apache/skywalking-satellite/plugins/fallbacker/api"
	forwarder "github.com/apache/skywalking-satellite/plugins/forwarder/api"
	"github.com/apache/skywalking-satellite/protocol/gen-codes/satellite/protocol"
)

// Sender is the forward module in Satellite.
type Sender struct {
	// config
	config *api.SenderConfig

	// dependency plugins
	runningForwarders []forwarder.Forwarder
	runningFallbacker fallbacker.Fallbacker
	runningClient     client.Client

	// dependency modules
	gatherer gatherer.Gatherer

	// self components
	logicInput    chan *event.OutputEventContext // logic input channel
	physicalInput chan *event.OutputEventContext // physical input channel
	listener      chan client.ClientStatus       // client status listener
	flushChannel  chan *buffer.BatchBuffer       // forwarder flush channel
	buffer        *buffer.BatchBuffer            // cache the downstream input data

	// metrics
	sendCounter *telemetry.Counter
}

// Prepare register the client status listener to the client manager and open input channel.
func (s *Sender) Prepare() error {
	log.Logger.WithField("pipe", s.config.PipeName).Info("sender module is preparing...")
	s.runningClient.RegisterListener(s.listener)
	s.logicInput = s.physicalInput
	for _, runningForwarder := range s.runningForwarders {
		err := runningForwarder.Prepare(s.runningClient.GetConnectedClient())
		if err != nil {
			return err
		}
	}
	s.sendCounter = telemetry.NewCounter("sender_output_count", "Total number of the output count in the Sender.", "pipe", "status", "type")
	return nil
}

// Boot fetches the downstream input data and forward to external services, such as Kafka and OAP receiver.
func (s *Sender) Boot(ctx context.Context) {
	log.Logger.WithField("pipe", s.config.PipeName).Info("sender module is starting...")
	var wg sync.WaitGroup
	wg.Add(2)
	// 1. keep fetching the downstream data when client connected, and put it into BatchBuffer.
	// 2. When reaches the buffer limit or receives a timer flush signal, and put BatchBuffer into flushChannel.
	go func() {
		defer wg.Done()
		childCtx, cancel := context.WithCancel(ctx)
		timeTicker := time.NewTicker(time.Duration(s.config.FlushTime) * time.Millisecond)
		for {
			select {
			case status := <-s.listener:
				switch status {
				case client.Connected:
					log.Logger.Infof("sender module of %s namespace is notified the connection connected", s.config.PipeName)
					s.logicInput = s.physicalInput
				case client.Disconnect:
					log.Logger.Infof("sender module of %s namespace is notified the connection disconnected", s.config.PipeName)
					s.logicInput = nil
				}
			case <-timeTicker.C:
				if s.buffer.Len() > s.config.MinFlushEvents {
					s.flushChannel <- s.buffer
					s.buffer = buffer.NewBatchBuffer(s.config.MaxBufferSize)
				}
			case e := <-s.logicInput:
				s.buffer.Add(e)
				if s.buffer.Len() == s.config.MaxBufferSize {
					s.flushChannel <- s.buffer
					s.buffer = buffer.NewBatchBuffer(s.config.MaxBufferSize)
				}
			case <-childCtx.Done():
				cancel()
				s.logicInput = nil
				return
			}
		}
	}()
	// Keep fetching BatchBuffer to forward.
	go func() {
		defer wg.Done()
		childCtx, cancel := context.WithCancel(ctx)
		for {
			select {
			case b := <-s.flushChannel:
				s.consume(b)
			case <-childCtx.Done():
				cancel()
				s.Shutdown()
				return
			}
		}
	}()
	wg.Wait()
}

// Shutdown closes the channels and tries to force forward the events in the buffer.
func (s *Sender) Shutdown() {
	log.Logger.WithField("pipe", s.config.PipeName).Info("sender module is closing")
	close(s.logicInput)
	for buf := range s.flushChannel {
		s.consume(buf)
	}
	s.consume(s.buffer)
	close(s.flushChannel)
}

// consume would forward the events by type and ack this batch.
func (s *Sender) consume(batch *buffer.BatchBuffer) {
	log.Logger.WithFields(logrus.Fields{
		"pipe":   s.config.PipeName,
		"offset": batch.Last(),
		"size":   batch.Len(),
	}).Info("sender module is flushing a new batch buffer.")
	var events = make(map[protocol.EventType]event.BatchEvents)
	for i := 0; i < batch.Len(); i++ {
		eventContext := batch.Buf()[i]
		for _, e := range eventContext.Context {
			if e.Remote {
				events[e.Type] = append(events[e.Type], e)
			}
		}
	}
	for _, f := range s.runningForwarders {
		for t, batchEvents := range events {
			if f.ForwardType() != t {
				continue
			}
			if err := f.Forward(batchEvents); err == nil {
				s.sendCounter.Add(float64(len(batchEvents)), s.config.PipeName, "success", f.ForwardType().String())
				continue
			}
			if !s.runningFallbacker.FallBack(batchEvents, f.Forward) {
				s.sendCounter.Add(float64(len(batchEvents)), s.config.PipeName, "failure", f.ForwardType().String())
			}
		}
	}
	s.gatherer.Ack(batch.Last())
}

func (s *Sender) InputDataChannel() chan<- *event.OutputEventContext {
	return s.logicInput
}

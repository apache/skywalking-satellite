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
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"

	v1 "skywalking.apache.org/repo/goapi/satellite/data/v1"

	"github.com/apache/skywalking-satellite/internal/pkg/log"
	"github.com/apache/skywalking-satellite/internal/satellite/event"
	module "github.com/apache/skywalking-satellite/internal/satellite/module/api"
	"github.com/apache/skywalking-satellite/internal/satellite/module/buffer"
	gatherer "github.com/apache/skywalking-satellite/internal/satellite/module/gatherer/api"
	"github.com/apache/skywalking-satellite/internal/satellite/module/sender/api"
	"github.com/apache/skywalking-satellite/internal/satellite/telemetry"
	client "github.com/apache/skywalking-satellite/plugins/client/api"
	fallbacker "github.com/apache/skywalking-satellite/plugins/fallbacker/api"
	forwarder "github.com/apache/skywalking-satellite/plugins/forwarder/api"
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
	input        chan *event.OutputEventContext // physical input channel
	listener     chan client.ClientStatus       // client status listener
	flushChannel chan *buffer.BatchBuffer       // forwarder flush channel
	buffer       *buffer.BatchBuffer            // cache the downstream input data
	blocking     int32                          // the status of input channel

	// metrics
	sendCounter *telemetry.Counter
}

// Prepare register the client status listener to the client manager and open input channel.
func (s *Sender) Prepare() error {
	log.Logger.WithField("pipe", s.config.PipeName).Info("sender module is preparing...")
	s.runningClient.RegisterListener(s.listener)
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
	wg.Add(3)
	go s.store(ctx, &wg)
	go s.listen(ctx, &wg)
	go s.flush(ctx, &wg)
	wg.Wait()
}

// store data.
// 1. keep fetching the downstream data when client connected, and put it into BatchBuffer.
// 2. When reaches the buffer limit or receives a timer flush signal, and put BatchBuffer into flushChannel.
func (s *Sender) store(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	defer log.Logger.WithField("pipe", s.config.PipeName).Infof("store routine closed")
	childCtx, _ := context.WithCancel(ctx) // nolint
	timeTicker := time.NewTicker(time.Duration(s.config.FlushTime) * time.Millisecond)
	for {
		// blocking output when disconnecting.
		if atomic.LoadInt32(&s.blocking) == 1 {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		select {
		case <-childCtx.Done():
			return
		case <-timeTicker.C:
			if s.buffer.Len() >= s.config.MinFlushEvents {
				s.flushChannel <- s.buffer
				s.buffer = buffer.NewBatchBuffer(s.config.MaxBufferSize)
			}
		case e := <-s.input:
			if e == nil {
				continue
			}
			s.buffer.Add(e)
			if s.buffer.Len() == s.config.MaxBufferSize {
				s.flushChannel <- s.buffer
				s.buffer = buffer.NewBatchBuffer(s.config.MaxBufferSize)
			}
		}
	}
}

// Listen the client status.
func (s *Sender) listen(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	defer log.Logger.WithField("pipe", s.config.PipeName).Infof("listen routine closed")
	childCtx, _ := context.WithCancel(ctx) // nolint
	for {
		select {
		case <-childCtx.Done():
			return
		case status := <-s.listener:
			switch status {
			case client.Connected:
				log.Logger.WithField("pipe", s.config.PipeName).Info("the client connection of the sender module connected")
				atomic.StoreInt32(&s.blocking, 0)
			case client.Disconnect:
				log.Logger.WithField("pipe", s.config.PipeName).Info("the client connection of the sender module disconnected")
				atomic.StoreInt32(&s.blocking, 1)
			}
		}
	}
}

// Keep fetching BatchBuffer to forward.
func (s *Sender) flush(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	defer log.Logger.WithField("pipe", s.config.PipeName).Infof("flush routine closed")
	childCtx, _ := context.WithCancel(ctx) // nolint
	for {
		select {
		case <-childCtx.Done():
			s.Shutdown()
			return
		case b := <-s.flushChannel:
			s.consume(b)
		}
	}
}

// Shutdown closes the channels and tries to force forward the events in the buffer.
func (s *Sender) Shutdown() {
	log.Logger.WithField("pipe", s.config.PipeName).Info("sender module is closing")
	close(s.input)
	ticker := time.NewTicker(module.ShutdownHookTime)
	for {
		select {
		case <-ticker.C:
			s.consume(s.buffer)
			return
		case b := <-s.flushChannel:
			s.consume(b)
		}
	}
}

// consume would forward the events by type and ack this batch.
func (s *Sender) consume(batch *buffer.BatchBuffer) {
	if batch.Len() == 0 {
		return
	}
	log.Logger.WithFields(logrus.Fields{
		"pipe":   s.config.PipeName,
		"offset": batch.Last(),
		"size":   batch.Len(),
	}).Info("sender module is flushing a new batch buffer.")
	var events = make(map[v1.SniffType]event.BatchEvents)
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
	return s.input
}

func (s *Sender) SyncInvoke(d *v1.SniffData) (*v1.SniffData, error) {
	supportSyncInvoke := make([]forwarder.Forwarder, 0)
	for inx := range s.runningForwarders {
		if s.runningForwarders[inx].SupportedSyncInvoke() {
			supportSyncInvoke = append(supportSyncInvoke, s.runningForwarders[inx])
		}
	}
	if len(supportSyncInvoke) > 1 {
		return nil, fmt.Errorf("only support single forwarder")
	} else if len(supportSyncInvoke) == 0 {
		return nil, fmt.Errorf("could not found forwarder")
	}
	return supportSyncInvoke[0].SyncForward(d)
}

func (s *Sender) DependencyInjection(modules ...module.Module) {
	s.gatherer = modules[0].(gatherer.Gatherer)
}

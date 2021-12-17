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
	"errors"
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

var defaultSenderFlushTime = 1000

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
	inputs       []chan *event.OutputEventContext // physical input channel of partitions
	listener     chan client.ClientStatus         // client status listener
	flushChannel []chan *buffer.BatchBuffer       // forwarder flush channel
	buffers      []*buffer.BatchBuffer            // cache the downstream petitioned input data
	blocking     int32                            // the status of input channel
	shutdownOnce sync.Once

	// metrics
	sendCounter telemetry.Counter
}

// Prepare register the client status listener to the client manager and open partitioned input channel.
func (s *Sender) Prepare() error {
	log.Logger.WithField("pipe", s.config.PipeName).Info("sender module is preparing...")
	s.runningClient.RegisterListener(s.listener)
	for _, runningForwarder := range s.runningForwarders {
		err := runningForwarder.Prepare(s.runningClient.GetConnectedClient())
		if err != nil {
			return err
		}
	}
	s.inputs = make([]chan *event.OutputEventContext, s.gatherer.PartitionCount())
	s.buffers = make([]*buffer.BatchBuffer, s.gatherer.PartitionCount())
	s.flushChannel = make([]chan *buffer.BatchBuffer, s.gatherer.PartitionCount())
	for partition := 0; partition < s.gatherer.PartitionCount(); partition++ {
		s.inputs[partition] = make(chan *event.OutputEventContext)
		s.buffers[partition] = buffer.NewBatchBuffer(s.config.MaxBufferSize)
		s.flushChannel[partition] = make(chan *buffer.BatchBuffer)
	}
	s.sendCounter = telemetry.NewCounter("sender_output_count", "Total number of the output count in the Sender.", "pipe", "status", "type")
	return nil
}

// Boot fetches the downstream partitioned input data and forward to external services, such as Kafka and OAP receiver.
func (s *Sender) Boot(ctx context.Context) {
	log.Logger.WithField("pipe", s.config.PipeName).Info("sender module is starting...")
	var wg sync.WaitGroup
	wg.Add(2*s.gatherer.PartitionCount() + 1)
	go s.listen(ctx, &wg)
	for partition := 0; partition < s.gatherer.PartitionCount(); partition++ {
		go s.store(ctx, partition, &wg)
		go s.flush(ctx, partition, &wg)
	}
	wg.Wait()
}

// store data.
// 1. keep fetching the downstream data when client connected, and put it into BatchBuffer.
// 2. When reaches the buffer limit or receives a timer flush signal, and put BatchBuffer into flushChannel.
func (s *Sender) store(ctx context.Context, partition int, wg *sync.WaitGroup) {
	defer wg.Done()
	defer log.Logger.WithField("pipe", s.config.PipeName).Infof("store routine closed")
	childCtx, _ := context.WithCancel(ctx) // nolint
	flushTime := s.config.FlushTime
	if flushTime <= 0 {
		flushTime = defaultSenderFlushTime
	}
	timeTicker := time.NewTicker(time.Duration(flushTime) * time.Millisecond)
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
			if s.buffers[partition].Len() >= s.config.MinFlushEvents {
				s.flushChannel[partition] <- s.buffers[partition]
				s.buffers[partition] = buffer.NewBatchBuffer(s.config.MaxBufferSize)
			}
		case e := <-s.inputs[partition]:
			if e == nil {
				continue
			}
			s.buffers[partition].Add(e)
			if s.buffers[partition].Len() == s.config.MaxBufferSize {
				s.flushChannel[partition] <- s.buffers[partition]
				s.buffers[partition] = buffer.NewBatchBuffer(s.config.MaxBufferSize)
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
func (s *Sender) flush(ctx context.Context, partition int, wg *sync.WaitGroup) {
	defer wg.Done()
	defer log.Logger.WithField("pipe", s.config.PipeName).Infof("flush routine closed")
	childCtx, _ := context.WithCancel(ctx) // nolint
	for {
		select {
		case <-childCtx.Done():
			s.Shutdown()
			return
		case b := <-s.flushChannel[partition]:
			s.consume(b)
		}
	}
}

// Shutdown closes the channels and tries to force forward the events in the buffer.
func (s *Sender) Shutdown() {
	s.shutdownOnce.Do(func() {
		s.shutdown0()
	})
}

func (s *Sender) shutdown0() {
	log.Logger.WithField("pipe", s.config.PipeName).Info("sender module is closing")
	for _, in := range s.inputs {
		close(in)
	}
	var wg sync.WaitGroup
	finished := make(chan struct{}, 1)
	wg.Add(len(s.flushChannel))
	for partition := range s.buffers {
		go func(p int) {
			defer wg.Done()
			s.consume(s.buffers[p])
		}(partition)
	}
	go func() {
		wg.Wait()
		close(finished)
	}()

	ticker := time.NewTicker(module.ShutdownHookTime)
	select {
	case <-ticker.C:
		for _, buffer := range s.buffers {
			s.consume(buffer)
		}
		return
	case <-finished:
		return
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

func (s *Sender) InputDataChannel(partition int) chan<- *event.OutputEventContext {
	return s.inputs[partition]
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

func (s *Sender) SetGatherer(m module.Module) error {
	if g, ok := m.(gatherer.Gatherer); ok {
		s.gatherer = g
		return nil
	}

	return errors.New("set gatherer only supports to inject gatherer module")
}

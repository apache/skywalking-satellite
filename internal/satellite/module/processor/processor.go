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

package processor

import (
	"context"
	"errors"
	"sync"

	"github.com/apache/skywalking-satellite/internal/pkg/log"
	"github.com/apache/skywalking-satellite/internal/satellite/event"
	"github.com/apache/skywalking-satellite/internal/satellite/module/api"
	gatherer "github.com/apache/skywalking-satellite/internal/satellite/module/gatherer/api"
	processor "github.com/apache/skywalking-satellite/internal/satellite/module/processor/api"
	sender "github.com/apache/skywalking-satellite/internal/satellite/module/sender/api"
	filter "github.com/apache/skywalking-satellite/plugins/filter/api"

	v1 "skywalking.apache.org/repo/goapi/satellite/data/v1"
)

// Processor is the processing module in Satellite.
type Processor struct {
	// config
	config *processor.ProcessorConfig

	// dependency plugins
	runningFilters []filter.Filter

	// dependency modules
	sender   sender.Sender
	gatherer gatherer.Gatherer
}

func (p *Processor) Prepare() error {
	return nil
}

// Boot fetches the data of Queue, does a series of processing, and then sends to Sender.
func (p *Processor) Boot(ctx context.Context) {
	log.Logger.WithField("pipe", p.config.PipeName).Info("processor module is starting...")
	var wg sync.WaitGroup
	wg.Add(p.gatherer.PartitionCount())
	for partition := 0; partition < p.gatherer.PartitionCount(); partition++ {
		p.processPerPartition(ctx, partition, &wg)
	}
	wg.Wait()
}

func (p *Processor) processPerPartition(ctx context.Context, partition int, wg *sync.WaitGroup) {
	go func() {
		childCtx, cancel := context.WithCancel(ctx)
		defer wg.Done()
		for {
			select {
			// receive the input event from the output channel of the gatherer
			case e := <-p.gatherer.OutputDataChannel(partition):
				c := &event.OutputEventContext{
					Offset:  e.Offset,
					Context: make(map[string]*v1.SniffData),
				}
				c.Put(e.Event)
				// processing the event with filters, that put the necessary events to OutputEventContext.
				for _, f := range p.runningFilters {
					f.Process(c)
				}
				// send the final context that contains many events to the sender.
				p.sender.InputDataChannel(partition) <- c
			case <-childCtx.Done():
				cancel()
				p.Shutdown()
				return
			}
		}
	}()
}

func (p *Processor) Shutdown() {
}

func (p *Processor) SyncInvoke(d *v1.SniffData) (*v1.SniffData, error) {
	// direct send data to sender
	return p.sender.SyncInvoke(d)
}

func (p *Processor) SetGatherer(m api.Module) error {
	if g, ok := m.(gatherer.Gatherer); ok {
		p.gatherer = g
		return nil
	}

	return errors.New("set gatherer only supports to inject gatherer module")
}

func (p *Processor) SetSender(m api.Module) error {
	if s, ok := m.(sender.Sender); ok {
		p.sender = s
		return nil
	}

	return errors.New("set sender only supports to inject sender module")
}

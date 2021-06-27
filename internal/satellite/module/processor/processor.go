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
	"sync"

	"github.com/apache/skywalking-satellite/internal/pkg/log"
	"github.com/apache/skywalking-satellite/internal/satellite/event"
	module "github.com/apache/skywalking-satellite/internal/satellite/module/api"
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
	wg.Add(1)
	go func() {
		childCtx, cancel := context.WithCancel(ctx)
		defer wg.Done()
		for {
			select {
			// receive the input event from the output channel of the gatherer
			case e := <-p.gatherer.OutputDataChannel():
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
				p.sender.InputDataChannel() <- c
			case <-childCtx.Done():
				cancel()
				p.Shutdown()
				return
			}
		}
	}()
	wg.Wait()
}

func (p *Processor) Shutdown() {
}

func (p *Processor) SyncInvoke(d *v1.SniffData) (*v1.SniffData, error) {
	// direct send data to sender
	return p.sender.SyncInvoke(d)
}

func (p *Processor) DependencyInjection(modules ...module.Module) {
	for _, m := range modules {
		switch t := m.(type) {
		case gatherer.Gatherer:
			p.gatherer = t
		case sender.Sender:
			p.sender = t
		}
	}
}

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

	"github.com/apache/skywalking-satellite/internal/pkg/constant"
	"github.com/apache/skywalking-satellite/internal/pkg/event"
	"github.com/apache/skywalking-satellite/internal/pkg/log"
	"github.com/apache/skywalking-satellite/internal/satellite/config"
	filter "github.com/apache/skywalking-satellite/plugins/filter/api"
)

// Processor is the processing module in Satellite.
type Processor struct {
	// config
	config *config.ProcessorConfig

	// dependency plugins
	runningFilters []filter.Filter

	// dependency modules
	sender   *Sender
	gatherer *Gatherer
}

func (p *Processor) Name() string {
	return constant.ProcessorModule
}

func (p *Processor) Description() string {
	return "processor module is the core processing module in Satellite"
}

func (p *Processor) Config() config.ModuleConfig {
	return p.config
}

// Init Processor and dependency plugins
func (p *Processor) Init(cfg config.ModuleConfig) {
	log.Logger.Infof("%s module of %s namespace is being initialized", p.Name(), p.config.NameSpace())
	p.config = cfg.(*config.ProcessorConfig)
	p.runningFilters = []filter.Filter{}
	for _, c := range p.config.FilterConfig {
		p.runningFilters = append(p.runningFilters, filter.GetFilter(c))
	}
}

// Prepare inject the dependency modules.
func (p *Processor) Prepare() error {
	log.Logger.Infof("%s module of %s namespace is in preparing stage", p.Name(), p.config.NameSpace())
	p.sender = GetRunningModule(p.config.NameSpace(), constant.SenderModule).(*Sender)
	p.gatherer = GetRunningModule(p.config.NameSpace(), constant.GathererModule).(*Gatherer)
	return nil
}

// BOOT fetches the data of Queue, does a series of processing, and then sends to Sender.
func (p *Processor) Boot(ctx context.Context) {
	log.Logger.Infof("%s module of %s namespace is running", p.Name(), p.config.NameSpace())
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		for {
			// fetch a new event from Queue of Gatherer
			e, offset, err := p.gatherer.runningQueue.Consumer().Dequeue()
			if err != nil {
				// todo add metrics
				log.Logger.Errorf("cannot get event from queue in %s namespace, error is: %v",
					p.config.NameSpace(), err)
				continue
			}
			c := &event.OutputEventContext{
				Offset:  offset,
				Context: make(map[string]event.Event),
			}
			// processing the event with filters, that put the necessary events to OutputEventContext.
			c.Put(e)
			for _, f := range p.runningFilters {
				f.Process(c)
			}

			select {
			// put result input the Input channel of Sender
			case p.sender.Input <- c:
			case <-ctx.Done():
				p.Shutdown()
				return
			}
		}
	}()
	wg.Wait()
}

func (p *Processor) Shutdown() {
	log.Logger.Infof("%s module of %s namespace is closing", p.Name(), p.config.NameSpace())
}

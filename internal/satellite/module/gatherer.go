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
	"github.com/apache/skywalking-satellite/internal/pkg/log"
	"github.com/apache/skywalking-satellite/internal/satellite/config"
	collector "github.com/apache/skywalking-satellite/plugins/collector/api"
	queue "github.com/apache/skywalking-satellite/plugins/queue/api"
)

// Gatherer is the APM data collection module in Satellite.
type Gatherer struct {
	// config
	config *config.GathererConfig

	// dependency plugins
	runningCollector collector.Collector
	runningQueue     queue.Queue
}

func (g *Gatherer) Name() string {
	return constant.GathererModule
}

func (g *Gatherer) Description() string {
	return "gatherer is the APM data collection module in Satellite, which supports Log, Trace, and Metrics scopes."
}

func (g *Gatherer) Config() config.ModuleConfig {
	return g.config
}

// Init Gatherer, dependency plugins.
func (g *Gatherer) Init(cfg config.ModuleConfig) {
	log.Logger.Infof("%s module of %s namespace is being initialized", g.Name(), g.config.NameSpace())
	g.config = cfg.(*config.GathererConfig)
	g.runningCollector = collector.GetCollector(g.config.CollectorConfig)
	g.runningQueue = queue.GetQueue(g.config.CollectorConfig)
}

// Prepare starts Collector and Queue.
func (g *Gatherer) Prepare() error {
	log.Logger.Infof("%s module of %s namespace is in preparing stage", g.Name(), g.config.NameSpace())
	if err := g.runningCollector.Prepare(); err != nil {
		log.Logger.Infof("%s collector of %s namespace was failed to initialize", g.runningCollector.Name(), g.config.NameSpace())
		return err
	}

	if err := g.runningQueue.Prepare(); err != nil {
		log.Logger.Infof("the %s queue of %s namespace was failed to initialize", g.runningQueue.Name(), g.config.NameSpace())
		return err
	}
	return nil
}

// Boot fetches Collector input data and pushes it into Queue.
func (g *Gatherer) Boot(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for {
			select {
			case e := <-g.runningCollector.EventChannel():
				err := g.runningQueue.Publisher().Enqueue(e)
				if err != nil {
					// todo add abandonedCount metrics
					log.Logger.Errorf("cannot put event into queue in %s namespace, error is: %v", g.config.NameSpace(), err)
				}
			case <-ctx.Done():
				g.Shutdown()
				return
			}
		}
	}()
	wg.Wait()
}

// Shutdown close Collector and Queue.
func (g *Gatherer) Shutdown() {
	log.Logger.Infof("%s module of %s namespace is closing", g.Name(), g.config.NameSpace())
	if err := g.runningCollector.Close(); err != nil {
		log.Logger.Errorf("failure occurs when closing %s collector  in %s namespace, error is: %v", g.runningCollector.Name(), g.config.NameSpace(), err)
	}
	if err := g.runningQueue.Close(); err != nil {
		log.Logger.Errorf("failure occurs when closing %s queue  in %s namespace, error is: %v", g.runningQueue.Name(), g.config.NameSpace(), err)
	}
}

// Ack some events according to the startOffset and the BatchSize
func (g *Gatherer) Ack(startOffset int64, batchSize int) {
	<-g.runningQueue.Ack(startOffset, batchSize)
}

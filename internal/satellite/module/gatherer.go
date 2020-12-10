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

	"github.com/apache/skywalking-satellite/internal/pkg/log"
	"github.com/apache/skywalking-satellite/internal/pkg/plugin"
	"github.com/apache/skywalking-satellite/internal/satellite/module/api"
	collector "github.com/apache/skywalking-satellite/plugins/collector/api"
	queue "github.com/apache/skywalking-satellite/plugins/queue/api"
)

type GathererConfig struct {
	api.ModuleCommonConfig
	// plugins config
	CollectorConfig plugin.Config `mapstructure:"collector"` // collector plugin config
	QueueConfig     plugin.Config `mapstructure:"queue"`     // queue plugin config
}

// Gatherer is the APM data collection module in Satellite.
type Gatherer struct {
	// config
	config *GathererConfig

	// dependency plugins
	runningCollector collector.Collector
	runningQueue     queue.Queue
}

// Init Gatherer, dependency plugins.
func NewGatherer(cfg *GathererConfig) *Gatherer {
	log.Logger.Infof("gatherer module of %s namespace is being initialized", cfg.RunningNamespace)
	return &Gatherer{
		config:           cfg,
		runningQueue:     queue.GetQueue(cfg.CollectorConfig),
		runningCollector: collector.GetCollector(cfg.CollectorConfig),
	}
}

func (g *Gatherer) Prepare() error {
	log.Logger.Infof("gatherer module of %s namespace is preparing", g.config.RunningNamespace)
	if err := g.runningCollector.Prepare(); err != nil {
		log.Logger.Infof("%s collector of %s namespace was failed to initialize", g.runningCollector.Name(), g.config.RunningNamespace)
		return err
	}
	if err := g.runningQueue.Prepare(); err != nil {
		log.Logger.Infof("the %s queue of %s namespace was failed to initialize", g.runningQueue.Name(), g.config.RunningNamespace)
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
					log.Logger.Errorf("cannot put event into queue in %s namespace, error is: %v", g.config.RunningNamespace, err)
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
	log.Logger.Infof("gatherer module of %s namespace is closing", g.config.RunningNamespace)
	if err := g.runningCollector.Close(); err != nil {
		log.Logger.Errorf("failure occurs when closing %s collector  in %s namespace, error is: %v",
			g.runningCollector.Name(), g.config.RunningNamespace, err)
	}
	if err := g.runningQueue.Close(); err != nil {
		log.Logger.Errorf("failure occurs when closing %s queue  in %s namespace, error is: %v", g.runningQueue.Name(), g.config.RunningNamespace, err)
	}
}

// Ack some events according to the startOffset and the BatchSize
func (g *Gatherer) Ack(startOffset int64, batchSize int) {
	<-g.runningQueue.Ack(startOffset, batchSize)
}

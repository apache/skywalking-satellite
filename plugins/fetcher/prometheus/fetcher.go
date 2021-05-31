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

package prometheus

import (
	"context"

	"github.com/apache/skywalking-satellite/internal/pkg/config"
	"github.com/apache/skywalking-satellite/internal/pkg/log"
	"github.com/apache/skywalking-satellite/internal/satellite/event"
	"github.com/apache/skywalking-satellite/protocol/gen-codes/satellite/protocol"

	"github.com/prometheus/prometheus/scrape"

	"go.uber.org/zap"

	promConfig "github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/discovery"
)

const (
	Name      = "prometheus-metrics-fetcher"
	eventName = "prometheus-metrics-event"
)

// Fetcher is the struct for Prometheus fetcher
type Fetcher struct {
	config.CommonFields
	// config is the top level configuration of prometheus
	ScrapeConfigs []*promConfig.ScrapeConfig `mapstructure:"scrape_configs"`
	// events
	OutputEvents event.BatchEvents
	// outputChannel
	OutputChannel chan *protocol.Event
}

func (f Fetcher) Name() string {
	return Name
}

func (f Fetcher) Description() string {
	return "This is a fetcher for Skywalking prometheus metrics format, " +
		"which will translate Prometheus metrics to Skywalking meter system."
}

func (f Fetcher) DefaultConfig() string {
	return `
## some config here
scrape_configs:
 - job_name: 'prometheus'
   static_configs:
   - targets: ["foo:9090", "bar:9090"]
`
}

func (f Fetcher) Prepare() {}

func (f Fetcher) Fetch() event.BatchEvents {
	// config of scraper
	c := make(map[string]discovery.Configs)
	for _, v := range f.ScrapeConfigs {
		c[v.JobName] = v.ServiceDiscoveryConfigs
	}

	// manager
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	manager := discovery.NewManager(ctx, nil)
	if err := manager.ApplyConfig(c); err != nil {
		log.Logger.Error("prometheus discovery config error", zap.Error(err))
	}
	// manager start
	go func() {
		if err := manager.Run(); err != nil {
			log.Logger.Error("Discovery manager run failed", zap.Error(err))
		}
	}()
	// queue store
	qs := NewQueueStore(ctx, true, Name, f.OutputChannel)
	scrapeManager := scrape.NewManager(nil, qs)
	qs.SetScrapeManager(scrapeManager)
	cfg := &promConfig.Config{ScrapeConfigs: f.ScrapeConfigs}
	if err := scrapeManager.ApplyConfig(cfg); err != nil {
		log.Logger.Error("scrape failed", zap.Error(err))
	}
	// stop scrape
	go func() {
		if err := scrapeManager.Run(manager.SyncCh()); err != nil {
			log.Logger.Error("scrape failed", zap.Error(err))
		}
	}()
	// do not need to return events
	return nil
}

func (f Fetcher) Channel() <-chan *protocol.Event {
	return f.OutputChannel
}

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
	"time"

	"go.uber.org/zap"

	"github.com/prometheus/prometheus/discovery"
	"github.com/prometheus/prometheus/scrape"

	yaml "gopkg.in/yaml.v2"

	"github.com/apache/skywalking-satellite/internal/pkg/config"
	"github.com/apache/skywalking-satellite/internal/pkg/log"
	"github.com/apache/skywalking-satellite/internal/satellite/event"
	v1 "skywalking.apache.org/repo/goapi/satellite/data/v1"

	promConfig "github.com/prometheus/prometheus/config"
)

const (
	Name      = "prometheus-metrics-fetcher"
	eventName = "prometheus-metrics-event"
)

type scrapeConfig struct {
	JobName        string                   `yaml:"job_name" mapstructure:"job_name"`
	ScrapeInterval time.Duration            `yaml:"scrape_interval,omitempty" mapstructure:"scrape_interval,omitempty"`
	StaticConfigs  []map[string]interface{} `yaml:"static_configs" mapstructure:"static_configs"`
	MetricsPath    string                   `yaml:"metrics_path,omitempty" mapstructure:"metrics_path,omitempty"`
}

// Fetcher is the struct for Prometheus fetcher
type Fetcher struct {
	config.CommonFields
	// config is the top level configuratScrapeConfigsMapion of prometheus
	ScrapeConfigsMap []*scrapeConfig `mapstructure:"scrape_configs" yaml:"scrape_configs"`

	ScrapeConfigs []*promConfig.ScrapeConfig
	// events
	OutputEvents event.BatchEvents
	// outputChannel
	OutputChannel chan *v1.SniffData

	cancelFunc context.CancelFunc
}

func (f *Fetcher) Name() string {
	return Name
}

func (f *Fetcher) Description() string {
	return "This is a fetcher for Skywalking prometheus metrics format, " +
		"which will translate Prometheus metrics to Skywalking meter system."
}

func (f *Fetcher) DefaultConfig() string {
	return `
## some config here
scrape_configs:
 - job_name: 'prometheus'
   metrics_path: '/metrics'
   scrape_interval: 10s
   static_configs:
   - targets: ['127.0.0.1:2020']
`
}

func (f *Fetcher) Prepare() {}

func (f *Fetcher) Fetch() event.BatchEvents {
	ctx, cancel := context.WithCancel(context.Background())
	f.cancelFunc = cancel
	// yaml
	configDeclare := make(map[string]interface{})
	configDeclare["scrape_configs"] = f.ScrapeConfigsMap
	configBytes, err := yaml.Marshal(configDeclare)
	if err != nil {
		log.Logger.Fatal("prometheus fetcher configure failed", err.Error())
	}
	log.Logger.Debug(string(configBytes))
	configStruct, err := promConfig.Load(string(configBytes))
	if err != nil {
		log.Logger.Fatal("prometheus fetcher configure load failed", err.Error())
	}
	f.ScrapeConfigs = configStruct.ScrapeConfigs
	return fetch(ctx, f.ScrapeConfigs, f.OutputChannel)
}

func (f *Fetcher) Channel() <-chan *v1.SniffData {
	return f.OutputChannel
}

func (f *Fetcher) Shutdown(context.Context) error {
	f.cancelFunc()
	return nil
}

func fetch(ctx context.Context, scrapeConfigs []*promConfig.ScrapeConfig, outputChannel chan *v1.SniffData) event.BatchEvents {
	// config of scraper
	c := make(map[string]discovery.Configs)
	for _, v := range scrapeConfigs {
		c[v.JobName] = v.ServiceDiscoveryConfigs
	}
	// manager
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
	qs := NewQueueStore(ctx, true, Name, outputChannel)
	scrapeManager := scrape.NewManager(nil, qs)
	qs.SetScrapeManager(scrapeManager)
	cfg := &promConfig.Config{ScrapeConfigs: scrapeConfigs}
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

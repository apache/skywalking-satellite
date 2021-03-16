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
	"fmt"

	"github.com/prometheus/prometheus/scrape"

	"github.com/ghodss/yaml"

	"github.com/apache/skywalking-satellite/internal/pkg/config"
	"github.com/apache/skywalking-satellite/internal/satellite/event"
	promConfig "github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/discovery"
	sd_config "github.com/prometheus/prometheus/discovery/config"
)

const (
	Name      = "prometheus-metrics-fetcher"
	eventName = "prometheus-metrics-event"
	success   = "success"
	failing   = "failing"
)

// Fetcher is the struct for Prometheus fetcher
type Fetcher struct {
	config.CommonFields
	// config
	Configs []MetricsConfig
	// components
	OutputEvents event.BatchEvents
}

type OriginPrometheus struct {
	promConfig *promConfig.Config
}

// MetricsConfig is the struct for Prometheus fetcher
type MetricsConfig struct {
	Endpoint string   `json:"endpoint" yaml:"endpoint"`
	TLS      bool     `json:"tls" yaml:"tls"`
	Metrics  []string `json:"metrics" yaml:"metrics"`
}

func (f Fetcher) Name() string {
	return Name
}

func (f Fetcher) Description() string {
	return "This is a fetcher for Skywalking prometheus metrics format, " +
		"which will translate Prometheus metrics to OpenTelemetry struct."
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

func (f Fetcher) Prepare() {
	// todo do we need prepare() ?
}

func (f Fetcher) Fetch() event.BatchEvents {
	// config
	cfg := &promConfig.Config{}
	sc := f.DefaultConfig()
	if err := yaml.Unmarshal([]byte(sc), cfg); err != nil {
		fmt.Errorf(err.Error())
	}
	// manage config
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	manager := discovery.NewManager(ctx, nil)
	fmt.Print(manager)
	c := make(map[string]sd_config.ServiceDiscoveryConfig)
	for _, v := range cfg.ScrapeConfigs {
		c[v.JobName] = v.ServiceDiscoveryConfig
	}

	if err := manager.ApplyConfig(c); err != nil {
		// log err
	}
	go func() {
		if err := manager.Run(); err != nil {
			// logger.Error("Discovery manager failed", zap.Error(err))
			// report error
		}
	}()

	// fetch metrics from prometheus endpoints && translate to OTLP
	//var jobsMap *JobsMap
	qs := &QueueStore{}
	scrapeManager := scrape.NewManager(nil, qs)
	if err := scrapeManager.ApplyConfig(cfg); err != nil {
		// report err and return
	}

	go func() {
		if err := scrapeManager.Run(manager.SyncCh()); err != nil {
			// logger.Error("Scrape manager failed", zap.Error(err))
			//report error
		}
	}()

	// Add to queue

	events := event.BatchEvents{}
	return events
}

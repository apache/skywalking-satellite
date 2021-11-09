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

	"github.com/apache/skywalking-satellite/internal/pkg/config"
	"github.com/apache/skywalking-satellite/internal/pkg/log"
	"github.com/apache/skywalking-satellite/internal/satellite/event"
	forwarder "github.com/apache/skywalking-satellite/plugins/forwarder/api"
	"github.com/apache/skywalking-satellite/plugins/forwarder/grpc/nativemeter"

	promConfig "github.com/prometheus/prometheus/config"
	"github.com/prometheus/prometheus/discovery"
	_ "github.com/prometheus/prometheus/discovery/install" // Need the init() func in this package to register service discovery implement.
	"github.com/prometheus/prometheus/scrape"
	yaml "gopkg.in/yaml.v3"
	v1 "skywalking.apache.org/repo/goapi/satellite/data/v1"
)

const (
	Name      = "prometheus-metrics-fetcher"
	ShowName  = "Prometheus Metrics Fetcher"
	eventName = "prometheus-metrics-event"
)

type scrapeConfig struct {
	JobName             string                   `yaml:"job_name" mapstructure:"job_name"`
	ScrapeInterval      time.Duration            `yaml:"scrape_interval,omitempty" mapstructure:"scrape_interval,omitempty"`
	StaticConfigs       []map[string]interface{} `yaml:"static_configs,omitempty" mapstructure:"static_configs,omitempty"`
	MetricsPath         string                   `yaml:"metrics_path,omitempty" mapstructure:"metrics_path,omitempty"`
	TLSConfig           tlsConfig                `yaml:"tls_config,omitempty" mapstructure:"tls_config,omitempty"`
	BearerTokenFile     string                   `yaml:"bearer_token_file,omitempty" mapstructure:"bearer_token_file,omitempty"`
	KubernetesSdConfigs []kubernetesSdConfig     `yaml:"kubernetes_sd_configs,omitempty" mapstructure:"kubernetes_sd_configs,omitempty"`
	RelabelConfigs      []relabelConfig          `yaml:"relabel_configs,omitempty" mapstructure:"relabel_configs,omitempty"`
}

type tlsConfig struct {
	CaFile string `yaml:"ca_file" mapstructure:"ca_file"`
}

// kubernetes_sd_configs []kubernetesSdConfig
type kubernetesSdConfig struct {
	Role      string     `yaml:"role,omitempty" mapstructure:"role,omitempty"`
	Selectors []selector `yaml:"selectors,omitempty" mapstructure:"selectors,omitempty"`
}

// relabel_configs []relabelConfig
type relabelConfig struct {
	SourceLabels []string `yaml:"source_labels,omitempty" mapstructure:"source_labels,omitempty"`
	Separator    string   `yaml:"separator,omitempty" mapstructure:"separator,omitempty"`
	Regex        string   `yaml:"regex,omitempty" mapstructure:"regex,omitempty"`
	TargetLabel  string   `yaml:"target_label,omitempty" mapstructure:"target_label,omitempty"`
	Replacement  string   `yaml:"replacement,omitempty" mapstructure:"replacement,omitempty"`
	Action       string   `yaml:"action,omitempty" mapstructure:"action,omitempty"`
}

type selector struct {
	Role  string `yaml:"role,omitempty" mapstructure:"role,omitempty"`
	Label string `yaml:"label,omitempty" mapstructure:"label,omitempty"`
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
}

func (f *Fetcher) Name() string {
	return Name
}

func (f *Fetcher) ShowName() string {
	return ShowName
}

func (f *Fetcher) Description() string {
	return "This is a fetcher for Skywalking prometheus metrics format, " +
		"which will translate Prometheus metrics to Skywalking meter system."
}

func (f *Fetcher) DefaultConfig() string {
	return `
# scrape_configs is the scrape configuration of prometheus 
# which is fully compatible with prometheus scrap.
scrape_configs:
# job_name will be used as the label in prometheus, 
# and in skywalking meterdata will be used as the service name.
# Set the scrape interval through scrape_interval
# static_configs is the service list of metrics server
- job_name: 'prometheus'
  metrics_path: '/metrics'
  scrape_interval: 10s
  static_configs:
    - targets:
      - "127.0.0.1:9100"
# In K8S, service discovery needs to be used to obtain the metrics server list.
# Configure and select related pods through kubernetes_sd_configs.selectors
# Because K8S resource permissions are involved, K8S serviceaccount needs to be configured
# tls_config is the certificate assigned by K8S to satellite, and generally does not need to be changed.
- job_name: 'prometheus-k8s'
  metrics_path: '/metrics'
  scrape_interval: 10s
  tls_config:
    ca_file: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
  bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token
  kubernetes_sd_configs:
  - role: pod
    selectors:
    - role: pod
      label: "app=prometheus"
`
}

func (f *Fetcher) Prepare() {}

func (f *Fetcher) Fetch(ctx context.Context) {
	f.OutputChannel = make(chan *v1.SniffData, 100)
	f.ScrapeConfig(ctx)
	fetch(ctx, f.ScrapeConfigs, f.OutputChannel)
}

func (f *Fetcher) ScrapeConfig(ctx context.Context) {
	// yaml
	configDeclare := make(map[string]interface{})
	log.Logger.Info(f.ScrapeConfigs)
	configDeclare["scrape_configs"] = f.ScrapeConfigsMap
	configBytes, err := yaml.Marshal(configDeclare)
	if err != nil {
		log.Logger.Fatal("prometheus fetcher configure failed", err.Error())
	}
	log.Logger.Debug(string(configBytes))
	configStruct, err := promConfig.Load(string(configBytes))
	if err != nil {
		log.Logger.Fatal("prometheus fetcher configure load failed ", err.Error())
	}
	f.ScrapeConfigs = configStruct.ScrapeConfigs
}

func (f *Fetcher) Channel() <-chan *v1.SniffData {
	return f.OutputChannel
}

func (f *Fetcher) Shutdown(ctx context.Context) error {
	ctx.Done()
	return nil
}

func fetch(ctx context.Context, scrapeConfigs []*promConfig.ScrapeConfig, outputChannel chan *v1.SniffData) {
	// config of scraper
	c := make(map[string]discovery.Configs)
	for _, v := range scrapeConfigs {
		c[v.JobName] = v.ServiceDiscoveryConfigs
	}
	// manager
	manager := discovery.NewManager(ctx, nil)
	if err := manager.ApplyConfig(c); err != nil {
		log.Logger.Fatalf("prometheus discovery config error %s", err.Error())
	}
	// manager start
	go func() {
		if err := manager.Run(); err != nil {
			log.Logger.Errorf("Discovery manager run failed, error %s", err.Error())
		}
	}()
	// queue store
	qs := NewQueueStore(ctx, true, Name, outputChannel)
	scrapeManager := scrape.NewManager(nil, qs)
	qs.SetScrapeManager(scrapeManager)
	cfg := &promConfig.Config{ScrapeConfigs: scrapeConfigs}
	if err := scrapeManager.ApplyConfig(cfg); err != nil {
		log.Logger.Fatalf("scrape failed, error: %s", err.Error())
	}
	// stop scrape
	go func() {
		if err := scrapeManager.Run(manager.SyncCh()); err != nil {
			log.Logger.Errorf("scrape failed, error: %s", err.Error())
		}
	}()
}

func (f *Fetcher) SupportForwarders() []forwarder.Forwarder {
	return []forwarder.Forwarder{
		new(nativemeter.Forwarder),
	}
}

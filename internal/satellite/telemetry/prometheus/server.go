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
	"net/http"
	"sync"

	"github.com/apache/skywalking-satellite/internal/satellite/telemetry"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/apache/skywalking-satellite/internal/pkg/log"
)

func init() {
	telemetry.Register("prometheus", &Server{}, false)
}

type Server struct {
	telemetry.PrometheusConfig

	Gatherer           prometheus.Gatherer // The gatherer is for fetching metrics from the registry.
	registry           *prometheus.Registry
	registerer         prometheus.Registerer // The register is for adding metrics to the registry.
	collectorContainer map[string]telemetry.Metric
	lock               sync.Mutex

	server *http.ServeMux // The prometheus server.
}

func (s *Server) Start(config *telemetry.Config) error {
	s.PrometheusConfig = config.Prometheus

	labels := make(map[string]string)
	if config.Cluster != "" {
		labels["cluster"] = config.Cluster
	}
	if config.Service != "" {
		labels["service"] = config.Service
	}
	if config.Instance != "" {
		labels["instance"] = config.Instance
	}

	s.registry = prometheus.NewRegistry()
	s.registerer = prometheus.WrapRegistererWith(labels, s.registry)
	s.Gatherer = s.registry
	s.collectorContainer = make(map[string]telemetry.Metric)

	s.server = http.NewServeMux()
	// add go info metrics.
	s.Register(s.WithMeta("processor_collector", prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{})),
		s.WithMeta("go_collector", prometheus.NewGoCollector()))
	// register prometheus metrics exporter handler.
	s.server.Handle(s.Endpoint, promhttp.HandlerFor(s.Gatherer, promhttp.HandlerOpts{ErrorLog: log.Logger}))
	go func() {
		log.Logger.WithField("address", s.Address).Info("prometheus server is starting...")
		err := http.ListenAndServe(s.Address, s.server)
		if err != nil {
			log.Logger.WithField("address", s.Address).Infof("prometheus server has failure when starting: %v", err)
		}
	}()
	return nil
}

func (s *Server) AfterSharingStart() error {
	return nil
}

func (s *Server) Close() error {
	log.Logger.Info("prometheus server is closed")
	return nil
}

// Register registers the metric meta to the registerer.
func (s *Server) Register(meta ...SelfTelemetryMetaFunc) {
	for _, telemetryMeta := range meta {
		name, collector := telemetryMeta()
		s.registerer.MustRegister(collector)
		log.Logger.WithField("telemetry_name", name).Info("self telemetry register success")
	}
}

// SelfTelemetryMetaFunc returns the metric name and the metric instance.
type SelfTelemetryMetaFunc func() (string, prometheus.Collector)

// WithMeta is used as the param of the Register function.
func (s *Server) WithMeta(name string, collector prometheus.Collector) SelfTelemetryMetaFunc {
	return func() (string, prometheus.Collector) {
		return name, collector
	}
}

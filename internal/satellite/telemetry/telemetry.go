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

package telemetry

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/apache/skywalking-satellite/internal/pkg/log"
)

// registerer is the global metrics center for collecting the telemetry data in core modules or plugins.
var (
	Gatherer           prometheus.Gatherer // The gatherer is for fetching metrics from the registry.
	registry           *prometheus.Registry
	registerer         prometheus.Registerer // The register is for adding metrics to the registry.
	collectorContainer map[string]Collector
	lock               sync.Mutex
)

// register the metric meta to the registerer.
func Register(meta ...SelfTelemetryMetaFunc) {
	for _, telemetryMeta := range meta {
		name, collector := telemetryMeta()
		registerer.MustRegister(collector)
		log.Logger.WithField("telemetry_name", name).Info("self telemetry register success")
	}
}

// SelfTelemetryMetaFunc returns the metric name and the metric instance.
type SelfTelemetryMetaFunc func() (string, prometheus.Collector)

// WithMeta is used as the param of the Register function.
func WithMeta(name string, collector prometheus.Collector) SelfTelemetryMetaFunc {
	return func() (string, prometheus.Collector) {
		return name, collector
	}
}

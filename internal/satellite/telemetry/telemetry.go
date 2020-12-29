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
	"github.com/prometheus/client_golang/prometheus"
)

// Registerer is the global metrics center for collecting the telemetry data in core modules or plugins.
var (
	registry   *prometheus.Registry
	Registerer prometheus.Registerer // The register is for adding metrics to the registry.
	Gatherer   prometheus.Gatherer   // The gatherer is for fetching metrics from the registry.
)

// Config defines the common telemetry labels.
type Config struct {
	Cluster  string `mapstructure:"cluster"`  // The cluster name.
	Service  string `mapstructure:"service"`  // The service name.
	Instance string `mapstructure:"instance"` // The instance name.
}

// Init create the global telemetry center according to the config.
func Init(c *Config) {
	labels := make(map[string]string)
	if c.Service != "" {
		labels["service"] = c.Service
	}
	if c.Cluster != "" {
		labels["cluster"] = c.Cluster
	}
	if c.Instance != "" {
		labels["instance"] = c.Instance
	}
	registry = prometheus.NewRegistry()
	Registerer = prometheus.WrapRegistererWith(labels, registry)
	Gatherer = registry
}

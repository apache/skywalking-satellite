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
	"fmt"
)

var (
	servers       = make(map[string]Server)
	currentServer Server
	defaultServer Server
)

// Config defines the common telemetry labels.
type Config struct {
	Cluster  string `mapstructure:"cluster"`  // The cluster name.
	Service  string `mapstructure:"service"`  // The service name.
	Instance string `mapstructure:"instance"` // The instance name.

	// Telemetry export type, support "prometheus", "metrics_service" or "none"
	ExportType string `mapstructure:"export_type"`
	// Export telemetry data through Prometheus server, only works on "export_type=prometheus".
	Prometheus PrometheusConfig `mapstructure:"prometheus"`
	// Export telemetry data through native meter format to OAP backend, only works on "export_type=metrics_service".
	MetricsService MetricsServiceConfig `mapstructure:"metrics_service"`
}

type PrometheusConfig struct {
	Address  string `mapstructure:"address"`  // The prometheus server address.
	Endpoint string `mapstructure:"endpoint"` // The prometheus server metrics endpoint.
}

type MetricsServiceConfig struct {
	ClientName   string `mapstructure:"client_name"`   // The grpc-client plugin name, using the SkyWalking native batch meter protocol
	Interval     int    `mapstructure:"interval"`      // The interval second for sending metrics
	MetricPrefix string `mapstructure:"metric_prefix"` // The prefix of telemetry metric name
}

type Server interface {
	Start(config *Config) error
	AfterSharingStart() error
	Close() error

	NewCounter(name, help string, labels ...string) Counter
	NewGauge(name, help string, getter func() float64, labels ...string) Gauge
	NewDynamicGauge(name, help string, labels ...string) DynamicGauge
	NewTimer(name, help string, labels ...string) Timer
}

func Register(name string, server Server, isDefault bool) {
	servers[name] = server
	if isDefault {
		defaultServer = server
	}
}

// Init create the global telemetry center according to the config.
func Init(c *Config) error {
	currentServer = servers[c.ExportType]
	if currentServer == nil {
		return fmt.Errorf("could not found telemetry exporter: %s", c.ExportType)
	}

	return currentServer.Start(c)
}

func AfterShardingStart() error {
	return getServer().AfterSharingStart()
}

func Close() error {
	return getServer().Close()
}

func NewCounter(name, help string, labels ...string) Counter {
	return getServer().NewCounter(name, help, labels...)
}

func NewGauge(name, help string, getter func() float64, labels ...string) Gauge {
	return getServer().NewGauge(name, help, getter, labels...)
}

func NewDynamicGauge(name, help string, labels ...string) DynamicGauge {
	return getServer().NewDynamicGauge(name, help, labels...)
}

func NewTimer(name, help string, labels ...string) Timer {
	return getServer().NewTimer(name, help, labels...)
}

func getServer() Server {
	if currentServer == nil {
		currentServer = defaultServer
	}
	return currentServer
}

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
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
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

	// Telemetry export type, support "prometheus", "metrics_service" or "none", multiple split by ","
	ExportType string `mapstructure:"export_type"`
	// Export telemetry data through Prometheus server, only works on "export_type=prometheus".
	Prometheus PrometheusConfig `mapstructure:"prometheus"`
	// Export telemetry data through native meter format to OAP backend, only works on "export_type=metrics_service".
	MetricsService MetricsServiceConfig `mapstructure:"metrics_service"`
	// Export pprof service for detect performance issue
	PProfService PProfServiceConfig `mapstructure:"pprof"`
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

type PProfServiceConfig struct {
	Address string `mapstructure:"address"` // The pprof server address.
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
	types := strings.Split(c.ExportType, ",")
	exportServers := make([]Server, 0)
	for _, t := range types {
		server := servers[t]
		if server == nil {
			return fmt.Errorf("could not found telemetry exporter: %s", t)
		}
		if e := server.Start(c); e != nil {
			return e
		}
		exportServers = append(exportServers, server)
	}

	if len(exportServers) > 1 {
		currentServer = &MultipleServer{Servers: exportServers}
	} else if len(exportServers) == 1 {
		currentServer = exportServers[0]
	}
	return nil
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

type MultipleServer struct {
	Servers []Server
}

func (s *MultipleServer) Start(config *Config) error {
	var err error
	for _, ss := range s.Servers {
		if e := ss.Start(config); e != nil {
			err = multierror.Append(err, e)
		}
	}
	return err
}

func (s *MultipleServer) AfterSharingStart() error {
	var err error
	for _, ss := range s.Servers {
		if e := ss.AfterSharingStart(); e != nil {
			err = multierror.Append(err, e)
		}
	}
	return err
}

func (s *MultipleServer) Close() error {
	var err error
	for _, ss := range s.Servers {
		if e := ss.Close(); e != nil {
			err = multierror.Append(err, e)
		}
	}
	return err
}

func (s *MultipleServer) NewCounter(name, help string, labels ...string) Counter {
	result := make([]Counter, 0)
	for _, ss := range s.Servers {
		result = append(result, ss.NewCounter(name, help, labels...))
	}
	return &MultipleCounter{Counters: result}
}

func (s *MultipleServer) NewGauge(name, help string, getter func() float64, labels ...string) Gauge {
	result := make([]Gauge, 0)
	for _, ss := range s.Servers {
		result = append(result, ss.NewGauge(name, help, getter, labels...))
	}
	return &MultipleGauge{Gauges: result}
}

func (s *MultipleServer) NewDynamicGauge(name, help string, labels ...string) DynamicGauge {
	result := make([]DynamicGauge, 0)
	for _, ss := range s.Servers {
		result = append(result, ss.NewDynamicGauge(name, help, labels...))
	}
	return &MultipleDynamicGauge{Gauges: result}
}

func (s *MultipleServer) NewTimer(name, help string, labels ...string) Timer {
	result := make([]Timer, 0)
	for _, ss := range s.Servers {
		result = append(result, ss.NewTimer(name, help, labels...))
	}
	return &MultipleTimer{Timers: result}
}

type MultipleCounter struct {
	Counters []Counter
}

func (m *MultipleCounter) Inc(labelValues ...string) {
	for _, c := range m.Counters {
		c.Inc(labelValues...)
	}
}

func (m *MultipleCounter) Add(val float64, labelValues ...string) {
	for _, c := range m.Counters {
		c.Add(val, labelValues...)
	}
}

type MultipleGauge struct {
	Gauges []Gauge
}

type MultipleDynamicGauge struct {
	Gauges []DynamicGauge
}

func (g *MultipleDynamicGauge) Inc(labelValues ...string) {
	for _, ga := range g.Gauges {
		ga.Inc(labelValues...)
	}
}

func (g *MultipleDynamicGauge) Dec(labelValues ...string) {
	for _, ga := range g.Gauges {
		ga.Dec(labelValues...)
	}
}

type MultipleTimer struct {
	Timers []Timer
}

func (m *MultipleTimer) Start(labelValues ...string) TimeRecorder {
	recorders := make([]TimeRecorder, 0)
	for _, t := range m.Timers {
		recorders = append(recorders, t.Start(labelValues...))
	}
	return &MultipleTimeRecorder{Recorders: recorders}
}

func (m *MultipleTimer) AddTime(t time.Duration, labelValues ...string) {
	for _, ti := range m.Timers {
		ti.AddTime(t, labelValues...)
	}
}

type MultipleTimeRecorder struct {
	Recorders []TimeRecorder
}

func (m *MultipleTimeRecorder) Stop() {
	for _, r := range m.Recorders {
		r.Stop()
	}
}

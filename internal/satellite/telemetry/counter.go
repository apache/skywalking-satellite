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

import "github.com/prometheus/client_golang/prometheus"

// The counter metric.
type Counter struct {
	Collector
	name    string // The name of counter.
	counter *prometheus.CounterVec
}

// NewCounter create a new counter if no metric with the same name exists.
func NewCounter(name, help string, labels ...string) *Counter {
	lock.Lock()
	defer lock.Unlock()
	collector, ok := collectorContainer[name]
	if !ok {
		counter := &Counter{
			name: name,
			counter: prometheus.NewCounterVec(prometheus.CounterOpts{
				Name: name,
				Help: help,
			}, labels),
		}
		Register(WithMeta(name, counter.counter))
		collectorContainer[name] = counter
		collector = counter
	}
	return collector.(*Counter)
}

// Add one.
func (c *Counter) Inc(labelValues ...string) {
	c.counter.WithLabelValues(labelValues...).Inc()
}

// Add float value.
func (c *Counter) Add(val float64, labelValues ...string) {
	c.counter.WithLabelValues(labelValues...).Add(val)
}

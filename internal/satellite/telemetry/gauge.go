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

type Gauge struct {
	Collector
	name  string
	gauge prometheus.GaugeFunc
}

func NewGauge(name, help string, getter func() float64, labels ...string) *Gauge {
	lock.Lock()
	defer lock.Unlock()
	rebuildName, rebuildLabels := rebuildGaugeNameAndLabels(name, labels...)
	collector, ok := collectorContainer[rebuildName]
	if !ok {
		gauge := &Gauge{
			name: name,
			gauge: prometheus.NewGaugeFunc(prometheus.GaugeOpts{
				Name:        name,
				Help:        help,
				ConstLabels: rebuildLabels,
			}, getter),
		}
		Register(WithMeta(rebuildName, gauge.gauge))
		collectorContainer[rebuildName] = gauge
		collector = gauge
	}
	return collector.(*Gauge)
}

func rebuildGaugeNameAndLabels(name string, labels ...string) (gaugeName string, labelsMap map[string]string) {
	resultName := name
	resultLabels := make(map[string]string)
	for inx := 0; inx < len(labels); inx += 2 {
		resultName += "_" + labels[inx] + "_" + labels[inx+1]
		resultLabels[labels[inx]] = labels[inx+1]
	}

	return resultName, resultLabels
}

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
	"fmt"
	v3 "skywalking/network/language/agent/v3"
	"strings"

	"github.com/prometheus/common/model"

	"github.com/prometheus/prometheus/pkg/labels"
)

const (
	metricsSuffixCount  = "_count"
	metricsSuffixBucket = "_bucket"
	metricsSuffixSum    = "_sum"
	startTimeMetricName = "process_start_time_seconds"
	scrapeUpMetricName  = "up"
)

var (
	trimmableSuffixes = []string{metricsSuffixBucket, metricsSuffixCount, metricsSuffixSum}
)

type metricBuilder struct {
	hasData            bool
	hasInternalMetric  bool
	mc                 MetadataCache
	metrics            []*v3.MeterData
	numTimeseries      int
	useStartTimeMetric bool
	startTime          float64
	currentMf          MetricFamily
}

func newMetricBuilder(mc MetadataCache, useStartTimeMetric bool) *metricBuilder {
	return &metricBuilder{
		mc:                 mc,
		metrics:            make([]*v3.MeterData, 0),
		numTimeseries:      0,
		useStartTimeMetric: useStartTimeMetric,
	}
}

func (b *metricBuilder) AddDataPoint(ls labels.Labels, t int64, v float64) error {
	metricName := ls.Get(model.MetricNameLabel)
	switch {
	case metricName == "":
		b.numTimeseries++
		return fmt.Errorf("errMetricNameNotFound")
	case isInternalMetric(metricName):
		// ignore internal metrics
		return nil
	case b.useStartTimeMetric && b.matchStartTimeMetric(metricName):
		b.startTime = v
	}

	b.hasData = true

	// check if the same metric_family
	if b.currentMf != nil && !b.currentMf.IsSameFamily(metricName) {
		m := b.currentMf.ToMetric()
		if m != nil {
			for _, mv := range m {
				b.metrics = append(b.metrics, mv)
			}
		}
		b.currentMf = newMetricFamily(metricName, b.mc)
	} else if b.currentMf == nil {
		b.currentMf = newMetricFamily(metricName, b.mc)
	}

	return b.currentMf.Add(metricName, ls, t, v)
}

func isInternalMetric(metricName string) bool {
	if metricName == scrapeUpMetricName || strings.HasPrefix(metricName, "scrape_") {
		return true
	}
	return false
}

func (b *metricBuilder) matchStartTimeMetric(metricName string) bool {
	return metricName == startTimeMetricName
}

func (b *metricBuilder) Build() (*v3.MeterDataCollection, int, error) {
	result := &v3.MeterDataCollection{}
	if !b.hasData {
		if b.hasInternalMetric {
			return result, 0, nil
		}
		return nil, 0, fmt.Errorf("errNoDataToBuild")
	}

	if b.currentMf != nil {
		m := b.currentMf.ToMetric()
		if m != nil {
			for _, v := range m {
				b.metrics = append(b.metrics, v)
			}
		}
		b.currentMf = nil
	}
	result.MeterData = b.metrics
	return result, b.numTimeseries, nil
}

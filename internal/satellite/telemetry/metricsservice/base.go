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

package metricsservice

import (
	"errors"
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"unicode/utf8"
	"unsafe"

	"github.com/apache/skywalking-satellite/internal/satellite/telemetry"

	v3 "skywalking.apache.org/repo/goapi/collect/language/agent/v3"
)

var errInconsistentCardinality = errors.New("inconsistent label cardinality")

// Inspired by Prometheus(prometheus/client_golang), adding labels data is more efficient and saves memory
const (
	offset64           = 14695981039346656037
	prime64            = 1099511628211
	SeparatorByte byte = 255
)

type Metric interface {
	telemetry.Metric
	WriteMetric(appender *MetricsAppender)
}

type SubMetric interface {
	WriteMetric(base *BaseMetric, labels []*v3.Label, appender *MetricsAppender)
}

type BaseMetric struct {
	Metric

	Name         string
	LabelKeys    []string
	NewSubMetric func(labelValues ...string) SubMetric

	curry   []curriedLabelValue
	mtx     sync.RWMutex
	metrics map[uint64][]metricWithLabelValues
}

func NewBaseMetric(name string, labels []string, newSubMetric func(labelValues ...string) SubMetric) *BaseMetric {
	return &BaseMetric{
		Name:         name,
		LabelKeys:    labels,
		NewSubMetric: newSubMetric,
		metrics:      map[uint64][]metricWithLabelValues{},
	}
}

func (b *BaseMetric) WriteMetric(appender *MetricsAppender) {
	for _, m := range b.metrics {
		for _, metric := range m {
			labels := make([]*v3.Label, len(b.LabelKeys))
			for i := range b.LabelKeys {
				labels[i] = &v3.Label{
					Name:  b.LabelKeys[i],
					Value: metric.values[i],
				}
			}
			metric.metric.WriteMetric(b, labels, appender)
		}
	}
}

func (b *BaseMetric) GetMetricWithLabelValues(lvs ...string) (SubMetric, error) {
	h, err := b.hashLabelValues(lvs)
	if err != nil {
		return nil, err
	}

	return b.getOrCreateMetricWithLabelValues(h, lvs, b.curry), nil
}

type metricWithLabelValues struct {
	values []string
	metric SubMetric
}

// curriedLabelValue sets the curried value for a label at the given index.
type curriedLabelValue struct {
	index int
	value string
}

func validateLabelValues(vals []string, expectedNumberOfValues int) error {
	if len(vals) != expectedNumberOfValues {
		return fmt.Errorf(
			"%s: expected %d label values but got %d in %#v",
			errInconsistentCardinality, expectedNumberOfValues,
			len(vals), vals,
		)
	}

	for _, val := range vals {
		if !utf8.ValidString(val) {
			return fmt.Errorf("label value %q is not valid UTF-8", val)
		}
	}

	return nil
}

func (b *BaseMetric) hashLabelValues(vals []string) (uint64, error) {
	if err := validateLabelValues(vals, len(b.LabelKeys)-len(b.curry)); err != nil {
		return 0, err
	}

	var (
		h             = hashNew()
		curry         = b.curry
		iVals, iCurry int
	)
	for i := 0; i < len(b.LabelKeys); i++ {
		if iCurry < len(curry) && curry[iCurry].index == i {
			h = hashAdd(h, curry[iCurry].value)
			iCurry++
		} else {
			h = hashAdd(h, vals[iVals])
			iVals++
		}
		h = hashAddByte(h, SeparatorByte)
	}
	return h, nil
}

func (b *BaseMetric) getOrCreateMetricWithLabelValues(
	hash uint64, lvs []string, curry []curriedLabelValue,
) SubMetric {
	b.mtx.RLock()
	metric, ok := b.getMetricWithHashAndLabelValues(hash, lvs, curry)
	b.mtx.RUnlock()
	if ok {
		return metric
	}

	b.mtx.Lock()
	defer b.mtx.Unlock()
	metric, ok = b.getMetricWithHashAndLabelValues(hash, lvs, curry)
	if !ok {
		inlinedLVs := inlineLabelValues(lvs, curry)
		metric = b.NewSubMetric(inlinedLVs...)
		b.metrics[hash] = append(b.metrics[hash], metricWithLabelValues{values: inlinedLVs, metric: metric})
	}
	return metric
}

func (b *BaseMetric) getMetricWithHashAndLabelValues(
	h uint64, lvs []string, curry []curriedLabelValue,
) (SubMetric, bool) {
	metrics, ok := b.metrics[h]
	if ok {
		if i := findMetricWithLabelValues(metrics, lvs, curry); i < len(metrics) {
			return metrics[i].metric, true
		}
	}
	return nil, false
}

func inlineLabelValues(lvs []string, curry []curriedLabelValue) []string {
	labelValues := make([]string, len(lvs)+len(curry))
	var iCurry, iLVs int
	for i := range labelValues {
		if iCurry < len(curry) && curry[iCurry].index == i {
			labelValues[i] = curry[iCurry].value
			iCurry++
			continue
		}
		labelValues[i] = lvs[iLVs]
		iLVs++
	}
	return labelValues
}

func findMetricWithLabelValues(
	metrics []metricWithLabelValues, lvs []string, curry []curriedLabelValue,
) int {
	for i, metric := range metrics {
		if matchLabelValues(metric.values, lvs, curry) {
			return i
		}
	}
	return len(metrics)
}

func matchLabelValues(values, lvs []string, curry []curriedLabelValue) bool {
	if len(values) != len(lvs)+len(curry) {
		return false
	}
	var iLVs, iCurry int
	for i, v := range values {
		if iCurry < len(curry) && curry[iCurry].index == i {
			if v != curry[iCurry].value {
				return false
			}
			iCurry++
			continue
		}
		if v != lvs[iLVs] {
			return false
		}
		iLVs++
	}
	return true
}

// hashNew initializies a new fnv64a hash value.
func hashNew() uint64 {
	return offset64
}

// hashAdd adds a string to a fnv64a hash value, returning the updated hash.
func hashAdd(h uint64, s string) uint64 {
	for i := 0; i < len(s); i++ {
		h ^= uint64(s[i])
		h *= prime64
	}
	return h
}

// hashAddByte adds a byte to a fnv64a hash value, returning the updated hash.
func hashAddByte(h uint64, b byte) uint64 {
	h ^= uint64(b)
	h *= prime64
	return h
}

func addFloat64(addr *float64, delta float64) {
	for {
		old := math.Float64frombits(atomic.LoadUint64((*uint64)(unsafe.Pointer(addr))))
		newVal := old + delta
		if atomic.CompareAndSwapUint64(
			(*uint64)(unsafe.Pointer(addr)),
			math.Float64bits(old),
			math.Float64bits(newVal),
		) {
			break
		}
	}
}

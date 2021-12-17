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

package metricservice

import (
	"math"
	"sync/atomic"
	"unsafe"

	"github.com/apache/skywalking-satellite/internal/satellite/telemetry"

	v3 "skywalking.apache.org/repo/goapi/collect/language/agent/v3"
)

type Counter struct {
	BaseMetric
}

type subCounter struct {
	val float64
}

// NewCounter create a new counter if no metric with the same name exists.
func (s *Server) NewCounter(name, _ string, labels ...string) telemetry.Counter {
	s.lock.Lock()
	defer s.lock.Unlock()
	metric, ok := s.metrics[name]
	if !ok {
		metric = &Counter{
			*NewBaseMetric(name, labels, func(labelValues ...string) SubMetric {
				return &subCounter{0}
			}),
		}
		s.Register(name, metric)
	}
	return metric.(telemetry.Counter)
}

func (c *Counter) Inc(labelValues ...string) {
	if counter, err := c.GetMetricWithLabelValues(labelValues...); err != nil {
		panic(err)
	} else {
		addFloat64(&(counter.(*subCounter).val), 1)
	}
}

func (c *Counter) Add(val float64, labelValues ...string) {
	if counter, err := c.GetMetricWithLabelValues(labelValues...); err != nil {
		panic(err)
	} else {
		addFloat64(&(counter.(*subCounter).val), val)
	}
}

func (c *subCounter) WriteMetric(base *BaseMetric, labels []*v3.Label, appender *MetricsAppender) {
	appender.appendSingleValue(base.Name, labels, math.Float64frombits(atomic.LoadUint64((*uint64)(unsafe.Pointer(&c.val)))))
}

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
	"sync"
	"time"

	"github.com/apache/skywalking-satellite/internal/satellite/telemetry"
)

var timerLocker sync.Mutex

type Timer struct {
	BaseMetric

	SumCounter   *Counter
	CountCounter *Counter
}

type TimeRecorder struct {
	timer       *Timer
	startTime   time.Time
	labelValues []string
}

func (s *Server) NewTimer(name, help string, labels ...string) telemetry.Timer {
	timerLocker.Lock()
	defer timerLocker.Unlock()

	metric, ok := s.metrics[name]
	if !ok {
		metric = &Timer{
			*NewBaseMetric(name, nil, func(labelValues ...string) SubMetric {
				return nil
			}),
			s.NewCounter(name+"_sum", help, labels...).(*Counter),
			s.NewCounter(name+"_count", help, labels...).(*Counter),
		}
		s.Register(name, metric)
	}
	return metric.(telemetry.Timer)
}

func (t *Timer) WriteMetric(appender *MetricsAppender) {
	t.SumCounter.WriteMetric(appender)
	t.CountCounter.WriteMetric(appender)
}

func (t *Timer) Start(labelValues ...string) telemetry.TimeRecorder {
	return &TimeRecorder{
		timer:       t,
		startTime:   time.Now(),
		labelValues: labelValues,
	}
}

func (t *Timer) AddTime(d time.Duration, labelValues ...string) {
	t.SumCounter.Add(float64(d.Milliseconds()), labelValues...)
	t.CountCounter.Inc(labelValues...)
}

// Stop the time and record the time
func (t *TimeRecorder) Stop() {
	t.timer.AddTime(time.Since(t.startTime), t.labelValues...)
}

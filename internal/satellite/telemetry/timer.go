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
	"time"
)

var timerLocker sync.Mutex

type Timer struct {
	Collector
	name         string
	sumCounter   *Counter
	countCounter *Counter
}

type TimeRecorder struct {
	timer       *Timer
	startTime   time.Time
	labelValues []string
}

// NewCounter create a new counter if no metric with the same name exists.
func NewTimer(name, help string, labels ...string) *Timer {
	timerLocker.Lock()
	defer timerLocker.Unlock()

	collector, ok := collectorContainer[name]
	if !ok {
		timer := &Timer{
			name:         name,
			sumCounter:   NewCounter(name+"_sum", help, labels...),
			countCounter: NewCounter(name+"_count", help, labels...),
		}
		collectorContainer[name] = timer
		collector = timer
	}
	return collector.(*Timer)
}

// Start a new time recorder
func (c *Timer) Start(labelValues ...string) *TimeRecorder {
	return &TimeRecorder{
		timer:       c,
		startTime:   time.Now(),
		labelValues: labelValues,
	}
}

// AddTime add a new duration and count
func (c *Timer) AddTime(t time.Duration, labelValues ...string) {
	c.sumCounter.Add(float64(t.Milliseconds()), labelValues...)
	c.countCounter.Inc(labelValues...)
}

// Stop the time and record the time
func (c *TimeRecorder) Stop() {
	c.timer.AddTime(time.Since(c.startTime), c.labelValues...)
}

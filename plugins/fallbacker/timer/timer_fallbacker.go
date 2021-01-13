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

package timer

import (
	"time"

	"github.com/apache/skywalking-satellite/internal/pkg/config"
	"github.com/apache/skywalking-satellite/internal/satellite/event"
	"github.com/apache/skywalking-satellite/plugins/forwarder/api"
)

// Fallbacker is a timer fallbacker when forward fails. `latencyFactor` is the standard retry duration,
// and the time for each retry is expanded by 2 times until the number of retries reaches the maximum.
type Fallbacker struct {
	config.CommonFields
	maxTimes      int `mapstructure:"max_times"`
	latencyFactor int `mapstructure:"latency_factor"`
}

func (t *Fallbacker) Name() string {
	return "timer-fallbacker"
}

func (t *Fallbacker) Description() string {
	return "this is a timer fallback trigger when forward fails."
}

func (t *Fallbacker) DefaultConfig() string {
	return `
max_times: 3
latency_factor: 2000
`
}

func (t *Fallbacker) FallBack(batch event.BatchEvents, forward api.ForwardFunc) bool {
	currentLatency := t.latencyFactor
	for i := 1; i < t.maxTimes; i++ {
		time.Sleep(time.Duration(currentLatency) * time.Millisecond)
		if err := forward(batch); err != nil {
			currentLatency *= 2
		} else {
			return true
		}
	}
	return false
}

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

	"github.com/apache/skywalking-satellite/internal/pkg/event"
	"github.com/apache/skywalking-satellite/plugins/forwarder/api"
)

// Fallbacker is a timer fallbacker when forward fails. `latencyFactor` is the standard retry duration,
// and the time for each retry is expanded by 2 times until the number of retries reaches the maximum.
type Fallbacker struct {
	maxTimes      int `mapstructure:"max_times"`
	latencyFactor int `mapstructure:"latency_factor"`
}

func (t *Fallbacker) Name() string {
	return "timer-fallbacker"
}

func (t *Fallbacker) Description() string {
	return "this is a timer trigger when forward fails."
}

func (t *Fallbacker) DefaultConfig() string {
	return `
max_times: 3
latency_factor: 2000
`
}

func (t *Fallbacker) FallBack(batch event.BatchEvents, connection interface{}, forward api.ForwardFunc) {
	if err := forward(connection, batch); err != nil {
		count := 1
		currentLatency := count * t.latencyFactor
		for count < t.maxTimes {
			time.Sleep(time.Duration(currentLatency) * time.Millisecond)
			if err := forward(connection, batch); err != nil {
				currentLatency *= 2
			} else {
				break
			}
		}
	}
}

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

package buffer

import (
	"context"
	"time"
)

type LimiterConfig struct {
	// The time interval between two flush operations. And the time unit is millisecond.
	FlushTime int `mapstructure:"flush_time" yaml:"flush_time"`
	// The max cache count when receive the message
	LimitCount int `mapstructure:"limit_count" yaml:"limit_count"`
}

type Flusher func()
type Checker func() int

type Limiter struct {
	Config       LimiterConfig
	checker      Checker
	stopChannel  chan struct{}
	flushChannel chan struct{}
}

func NewLimiter(config LimiterConfig, checker Checker) *Limiter {
	return &Limiter{
		Config:       config,
		checker:      checker,
		stopChannel:  make(chan struct{}),
		flushChannel: make(chan struct{}),
	}
}

func (l *Limiter) Start(ctx context.Context, flush Flusher) {
	go func() {
		childCtx, cancel := context.WithCancel(ctx)
		timer := time.NewTimer(time.Duration(l.Config.FlushTime) * time.Millisecond)

		defer cancel()
		for {
			timer.Reset(time.Duration(l.Config.FlushTime) * time.Millisecond)
			select {
			case <-timer.C:
				flush()
			case <-l.flushChannel:
				flush()

			case <-l.stopChannel:
				flush()
				return
			case <-childCtx.Done():
				flush()
				return
			}
		}
	}()
}

func (l *Limiter) Check() {
	if l.checker() >= l.Config.LimitCount {
		l.flushChannel <- struct{}{}
	}
}

func (l *Limiter) Stop() {
	l.stopChannel <- struct{}{}
}

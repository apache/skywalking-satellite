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
	"testing"
	"time"
)

func TestLimitor(t *testing.T) {
	tests := []struct {
		name       string
		limitCount int
		flushTime  int
		sendCount  int
		isFlush    bool
	}{
		{
			name:       "reach limit to flush",
			limitCount: 5,
			flushTime:  5000,
			sendCount:  5,
			isFlush:    true,
		},
		{
			name:       "not reach limit",
			limitCount: 5,
			flushTime:  5000,
			sendCount:  2,
			isFlush:    false,
		},
		{
			name:       "using flush time to flush",
			limitCount: 5,
			flushTime:  100,
			sendCount:  1,
			isFlush:    true,
		},
	}
	for _, ts := range tests {
		t.Run(ts.name, func(t *testing.T) {
			conf := LimiterConfig{LimitCount: ts.limitCount, FlushTime: ts.flushTime}
			limiter := NewLimiter(conf, func() int {
				return ts.sendCount
			})

			flushChannel := make(chan struct{}, 1)
			defer limiter.Stop()
			limiter.Start(context.Background(), func() {
				flushChannel <- struct{}{}
			})
			limiter.Check()

			beenFlushed := false
			select {
			case <-flushChannel:
				beenFlushed = true
			case <-time.After(time.Second * 1):
			}

			if ts.isFlush != beenFlushed {
				t.Fail()
			}
		})
	}
}

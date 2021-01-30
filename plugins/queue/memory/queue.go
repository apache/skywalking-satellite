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

package memory

import (
	"fmt"
	"sync/atomic"

	"github.com/apache/skywalking-satellite/internal/pkg/config"
	"github.com/apache/skywalking-satellite/internal/satellite/event"
	"github.com/apache/skywalking-satellite/plugins/queue/api"
	"github.com/apache/skywalking-satellite/protocol/gen-codes/satellite/protocol"
)

const (
	Name = "memory-queue"
	// discard strategy
	discardOldest = "DISCARD_OLDEST"
	discardLatest = "DISCARD_LATEST"
)

type Queue struct {
	config.CommonFields
	// config
	EventBufferSize int64  `mapstructure:"event_buffer_size"` // The maximum buffer event size.
	DiscardStrategy string `mapstructure:"discard_strategy"`  // The discard strategy.

	// components
	queue []*protocol.Event
	// The position continuously increasing, but don't worry, it can run for another 1067519911 days at 10W OPS.
	readPos  int64
	writePos int64
	count    int64
}

func (q *Queue) Name() string {
	return Name
}

func (q *Queue) Description() string {
	return "this is a memory queue to buffer the input event."
}

func (q *Queue) DefaultConfig() string {
	return `
# The maximum buffer event size.
event_buffer_size: 5000
# The discard strategy when facing the full condition.
# There are 2 strategies, which are DISCARD_OLDEST and DISCARD_LATEST. 
discard_strategy: DISCARD_OLDEST
`
}

func (q *Queue) Initialize() error {
	if q.EventBufferSize <= 0 {
		return fmt.Errorf("the size of the memory queue must be positive")
	}
	if q.DiscardStrategy != discardLatest && q.DiscardStrategy != discardOldest {
		return fmt.Errorf("%s discard strategy is not supported in the memory queue", q.DiscardStrategy)
	}
	q.queue = make([]*protocol.Event, q.EventBufferSize)
	return nil
}

func (q *Queue) Push(e *protocol.Event) error {
	if q.isFull() {
		switch q.DiscardStrategy {
		case discardLatest:
			return api.ErrFull
		case discardOldest:
			atomic.AddInt64(&q.readPos, 1)
		}
	} else {
		atomic.AddInt64(&q.count, 1)
	}
	q.queue[q.writePos%q.count] = e
	q.writePos++
	return nil
}

func (q *Queue) Pop() (*api.SequenceEvent, error) {
	if q.isEmpty() {
		return nil, api.ErrEmpty
	}
	e := &api.SequenceEvent{
		Event: q.queue[q.readPos],
	}
	atomic.AddInt64(&q.readPos, 1)
	atomic.AddInt64(&q.count, -1)
	return e, nil
}

func (q *Queue) Close() error {
	return nil
}

func (q *Queue) Ack(_ event.Offset) {
}

func (q *Queue) isEmpty() bool {
	return q.count == 0
}

func (q *Queue) isFull() bool {
	return q.count == q.EventBufferSize
}

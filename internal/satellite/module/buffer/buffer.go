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
	"sync"

	"github.com/apache/skywalking-satellite/internal/pkg/event"
	"github.com/apache/skywalking-satellite/internal/pkg/log"
)

// BatchBuffer is a buffer to cache the input data in Sender.
type BatchBuffer struct {
	sync.Mutex                             // local
	buf        []*event.OutputEventContext // cache
	first      int64                       // the first OutputEventContext offset
	last       int64                       // the last OutputEventContext offset
	size       int                         // usage size
	cap        int                         // the max capacity
}

// NewBuffer crate a new BatchBuffer according to the capacity param.
func NewBatchBuffer(capacity int) *BatchBuffer {
	return &BatchBuffer{
		buf:   make([]*event.OutputEventContext, capacity),
		first: 0,
		last:  0,
		size:  0,
		cap:   capacity,
	}
}

// Buf returns the cached data in BatchBuffer.
func (b *BatchBuffer) Buf() []*event.OutputEventContext {
	b.Lock()
	defer b.Unlock()
	return b.buf
}

// First returns the first OutputEventContext offset.
func (b *BatchBuffer) First() int64 {
	b.Lock()
	defer b.Unlock()
	return b.first
}

// Len returns the usage size.
func (b *BatchBuffer) Len() int {
	b.Lock()
	defer b.Unlock()
	return b.size
}

// BatchSize returns the offset increment of the cached data.
func (b *BatchBuffer) BatchSize() int {
	b.Lock()
	defer b.Unlock()
	return int(b.last - b.first + 1)
}

// Add adds a new data input buffer.
func (b *BatchBuffer) Add(data *event.OutputEventContext) {
	b.Lock()
	defer b.Unlock()
	if b.size == b.cap {
		log.Logger.Errorf("cannot add one item to the fulling BatchBuffer, the capacity is %d", b.cap)
		return
	} else if data.Offset <= 0 {
		log.Logger.Errorf("cannot add one item to BatchBuffer because the input data is illegal, the offset is %d", data.Offset)
		return
	}
	if b.size == 0 {
		b.first = data.Offset
	} else {
		b.last = data.Offset
	}
	b.buf[b.size] = data
	b.size++
}

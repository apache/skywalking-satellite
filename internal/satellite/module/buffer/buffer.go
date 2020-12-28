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
	"github.com/apache/skywalking-satellite/internal/pkg/log"
	"github.com/apache/skywalking-satellite/internal/satellite/event"
)

// BatchBuffer is a buffer to cache the input data in Sender.
type BatchBuffer struct {
	buf   []*event.OutputEventContext // cache
	first event.Offset                // the first OutputEventContext offset
	last  event.Offset                // the last OutputEventContext offset
	size  int                         // usage size
	cap   int                         // the max capacity
}

// NewBuffer crate a new BatchBuffer according to the capacity param.
func NewBatchBuffer(capacity int) *BatchBuffer {
	return &BatchBuffer{
		buf:   make([]*event.OutputEventContext, capacity),
		first: "",
		last:  "",
		size:  0,
		cap:   capacity,
	}
}

// Buf returns the cached data in BatchBuffer.
func (b *BatchBuffer) Buf() []*event.OutputEventContext {
	return b.buf
}

// First returns the first OutputEventContext offset.
func (b *BatchBuffer) First() event.Offset {
	return b.first
}

// Last returns the last OutputEventContext offset.
func (b *BatchBuffer) Last() event.Offset {
	return b.last
}

// Len returns the usage size.
func (b *BatchBuffer) Len() int {
	return b.size
}

// Add adds a new data input buffer.
func (b *BatchBuffer) Add(data *event.OutputEventContext) {
	if b.size == b.cap {
		log.Logger.Errorf("cannot add one item to the fulling BatchBuffer, the capacity is %d", b.cap)
		return
	} else if data.Offset == "" {
		log.Logger.Errorf("cannot add one item to BatchBuffer because the input data is illegal, the offset is empty")
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

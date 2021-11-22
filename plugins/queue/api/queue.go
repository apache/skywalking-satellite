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

package api

import (
	"reflect"

	"github.com/apache/skywalking-satellite/internal/pkg/plugin"
	"github.com/apache/skywalking-satellite/internal/satellite/event"

	v1 "skywalking.apache.org/repo/goapi/satellite/data/v1"
)

// Queue is a plugin interface, that defines new queues.
type Queue interface {
	plugin.Plugin

	// Initialize creates the queue.
	Initialize() error

	// Enqueue a inputEvent into the queue.
	Enqueue(event *v1.SniffData) error

	// Dequeue returns a SequenceEvent when Queue is not empty,
	Dequeue() (*SequenceEvent, error)

	// Close would close the queue.
	Close() error

	// Ack the lastOffset
	Ack(lastOffset *event.Offset)

	// TotalSize total capacity of queue
	TotalSize() int64

	// UsedCount used count of queue
	UsedCount() int64

	// IsFull the queue is full
	IsFull() bool
}

// SequenceEvent is a wrapper to pass the event and the offset.
type SequenceEvent struct {
	Event  *v1.SniffData
	Offset event.Offset
}

// GetQueue an initialized filter plugin.
func GetQueue(config plugin.Config) Queue {
	return plugin.Get(reflect.TypeOf((*Queue)(nil)).Elem(), config).(Queue)
}

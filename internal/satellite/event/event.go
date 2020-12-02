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

package event

import (
	"time"

	"github.com/apache/skywalking-satellite/internal/pkg/event"
)

// Event defines the common fields.
type Event struct {
	name      string
	timestamp time.Time
	meta      map[string]string
	eventType event.Type
	remote    bool
}

// ETBFunc serialize event to bytes.
type ETBFunc func(event event.SerializableEvent) []byte

// BToFunc deserialize bytes to event.
type BToFunc func(bytes []byte) event.SerializableEvent

// StructuredEvent works when the data is a struct type.
type StructuredEvent struct {
	Event
	data interface{}
}

// StructuredEvent works when the data is not a struct type.
type UnstructuredEvent struct {
	Event
	data map[string]interface{}
}

// StructuredEvent works when the data is a struct type in the collector.
type StructuredSerializableEvent struct {
	StructuredEvent
	etb ETBFunc
	bte BToFunc
}

// StructuredEvent works when the data is not a struct type in the collector.
type UnstructuredSerializableEvent struct {
	UnstructuredEvent
	etb ETBFunc
	bte BToFunc
}

func (s *Event) Name() string {
	return s.name
}

func (s *Event) Meta() map[string]string {
	return s.meta
}

func (s *Event) Time() time.Time {
	return s.timestamp
}

func (s *Event) Type() event.Type {
	return s.eventType
}

func (s *Event) IsRemote() bool {
	return s.remote
}

func (u *StructuredEvent) Data() interface{} {
	return u.data
}

func (u *UnstructuredEvent) Data() interface{} {
	return u.data
}

func (s *StructuredSerializableEvent) ToBytes() []byte {
	return s.etb(s)
}

func (s *StructuredSerializableEvent) FromBytes(bytes []byte) event.SerializableEvent {
	return s.bte(bytes)
}

func (u *UnstructuredSerializableEvent) ToBytes() []byte {
	return u.etb(u)
}

func (u *UnstructuredSerializableEvent) FromBytes(bytes []byte) event.SerializableEvent {
	return u.bte(bytes)
}

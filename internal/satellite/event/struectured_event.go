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

	"github.com/apache/skywalking-satellite/internal/pkg/api"
)

// StructuredInputEventToBytesFunc serialize event to bytes.
type StructuredInputEventToBytesFunc func(event *StructuredInputEvent) []byte

// BytesToStructuredInputEventFunc deserialize bytes to event.
type BytesToStructuredInputEventFunc func(bytes []byte) *StructuredInputEvent

// StructuredEvent works when the data is a struct type.
type StructuredEvent struct {
	name      string
	timestamp time.Time
	meta      map[string]string
	data      interface{}
	eventType api.EventType
	remote    bool
}

// StructuredEvent works when the data is a struct type in the collector.
type StructuredInputEvent struct {
	StructuredEvent
	etb StructuredInputEventToBytesFunc
	bte BytesToStructuredInputEventFunc
}

func (s *StructuredEvent) Name() string {
	return s.name
}

func (s *StructuredEvent) Meta() map[string]string {
	return s.meta
}

func (s *StructuredEvent) Data() interface{} {
	return &s.data
}

func (s *StructuredEvent) Time() time.Time {
	return s.timestamp
}

func (s *StructuredEvent) Type() api.EventType {
	return s.eventType
}

func (s *StructuredEvent) IsRemote() bool {
	return s.remote
}

func (s *StructuredInputEvent) ToBytes() []byte {
	return s.etb(s)
}

func (s *StructuredInputEvent) FromBytes(bytes []byte) api.SerializableEvent {
	return s.bte(bytes)
}

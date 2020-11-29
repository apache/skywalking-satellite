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

// UnstructuredInputEventToBytesFunc serialize event to bytes.
type UnstructuredInputEventToBytesFunc func(event *UnstructuredInputEvent) []byte

// BytesToStructuredInputEventFunc deserialize bytes to event.
type BytesToUnstructuredInputEventFunc func(bytes []byte) *UnstructuredInputEvent

// UnstructuredEvent works when the data is a map type.
type UnstructuredEvent struct {
	name      string
	timestamp time.Time
	meta      map[string]string
	data      map[string]interface{}
	eventType api.EventType
	remote    bool
}

// UnstructuredInputEvent works when the data is a map type in the collector.
type UnstructuredInputEvent struct {
	UnstructuredEvent
	etb UnstructuredInputEventToBytesFunc
	bte BytesToUnstructuredInputEventFunc
}

func (s *UnstructuredEvent) Name() string {
	return s.name
}

func (s *UnstructuredEvent) Meta() map[string]string {
	return s.meta
}

func (s *UnstructuredEvent) Data() interface{} {
	return &s.data
}

func (s *UnstructuredEvent) Time() time.Time {
	return s.timestamp
}

func (s *UnstructuredEvent) Type() api.EventType {
	return s.eventType
}

func (s *UnstructuredEvent) Remote() bool {
	return s.remote
}

func (s *UnstructuredInputEvent) ToBytes() []byte {
	return s.etb(s)
}

func (s *UnstructuredInputEvent) FromBytes(bytes []byte) api.SerializationEvent {
	return s.bte(bytes)
}

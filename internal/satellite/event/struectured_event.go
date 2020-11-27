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
	"github.com/apache/skywalking-satellite/internal/pkg/api"
	"time"
)

type StructuredInputEventToBytesFunc func(event *StructuredInputEvent) []byte
type BytesToStructuredInputEventFunc func(bytes []byte) *StructuredInputEvent

type StructuredEvent struct {
	name      string
	timestamp time.Time
	meta      map[string]string
	data      interface{}
	output    bool
}

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

func (s *StructuredEvent) IsOutput() bool {
	return s.output
}

func (s *StructuredInputEvent) ToBytes() []byte {
	return s.etb(s)
}

func (s *StructuredInputEvent) FromBytes(bytes []byte) api.InputEvent {
	return s.bte(bytes)
}
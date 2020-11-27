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

type UnstructuredInputEventToBytesFunc func(event *UnstructuredInputEvent) []byte
type BytesToUnstructuredInputEventFunc func(bytes []byte) *UnstructuredInputEvent

type UnstructuredEvent struct {
	name      string
	timestamp time.Time
	meta      map[string]string
	data      map[string]interface{}
	output    bool
}

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

func (s *UnstructuredEvent) IsOutput() bool {
	return s.output
}

func (s *UnstructuredInputEvent) ToBytes() []byte {
	return s.etb(s)
}

func (s *UnstructuredInputEvent) FromBytes(bytes []byte) api.InputEvent {
	return s.bte(bytes)
}

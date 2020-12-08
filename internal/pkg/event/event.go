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
	"fmt"
	"time"
)

// The event type.
const (
	// Mapping to the type supported by SkyWalking OAP.
	_ Type = iota
	MetricsEvent
	ProfilingEvent
	SegmentEvent
	ManagementEvent
	MeterEvent
	LogEvent
)

type Type int32

// Event that implement this interface would be allowed to transmit in the Satellite.
type Event interface {
	// Name returns the event name.
	Name() string

	// Meta is a pair of key and value to record meta data, such as labels.
	Meta() map[string]string

	// Data returns the wrapped data.
	Data() interface{}

	// Time returns the event time.
	Time() time.Time

	// Type is to distinguish different events.
	Type() Type

	// IsRemote means is a output event when returns true.
	IsRemote() bool
}

// SerializableEvent is used in Collector to bridge Queue.
type SerializableEvent interface {
	Event

	// ToBytes serialize the event to a byte array.
	ToBytes() []byte

	// FromBytes deserialize the byte array to an event.
	FromBytes(bytes []byte) SerializableEvent
}

// BatchEvents is used by Forwarder to forward.
type BatchEvents []Event

// OutputEventContext is a container to store the output context.
type OutputEventContext struct {
	Context map[string]Event
	Offset  int64
}

// Put puts the incoming event into the context when the event is a remote event.
func (c *OutputEventContext) Put(event Event) {
	c.Context[event.Name()] = event
}

// Get returns a event in the context. When the eventName does not exist, a error would be returned.
func (c *OutputEventContext) Get(eventName string) (Event, error) {
	e, ok := c.Context[eventName]
	if !ok {
		err := fmt.Errorf("cannot find the event name in OutputEventContext : %s", eventName)
		return nil, err
	}
	return e, nil
}

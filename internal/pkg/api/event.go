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
	"fmt"
	"time"
)

// Event that implement this interface would be allowed to transmit in the Satellite.
type Event interface {
	// Name is a identify to distinguish different events.
	Name() string

	// Meta is a pair of key and value to record meta data, such as labels.
	Meta() map[string]string

	// Data returns the wrappered data.
	Data() interface{}

	// Time returns the event time.
	Time() time.Time

	// IsOutput indicates that the event can exist in the output context when the return value is true.
	IsOutput() bool
}

// InputEvent is used in Gatherer to bridge Queue.
type InputEvent interface {
	Event

	// ToBytes serialize the event to a byte array.
	ToBytes() []byte

	// FromBytes deserialize the byte array to an event.
	FromBytes(bytes []byte) InputEvent
}

// BatchOutputEvents is batch events to output.
type BatchOutputEvents struct {
	// BatchEvents grouped by event name.
	BatchEvents map[string][]Event

	// The start offset of the batch.
	StartOffset int64

	// The end offset of the batch.
	EndOffset int64
}

// OutputEventContext is a container to store the output context.
type OutputEventContext struct {
	context map[string]Event
}

// Put puts the incoming event into the context when the event allows to output.
func (c *OutputEventContext) Put(event Event) {
	if event.IsOutput() {
		c.context[event.Name()] = event
	}
}

// Get returns a event in the context. When the eventName does not exist, a error would be returned.
func (c *OutputEventContext) Get(eventName string) (Event, error) {
	e, ok := c.context[eventName]
	if !ok {
		err := fmt.Errorf("cannot find the event name in OutputEventContext : %s", eventName)
		return nil, err
	}
	return e, nil
}
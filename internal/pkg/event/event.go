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

	"github.com/apache/skywalking-satellite/protocol/gen-codes/satellite/protocol"
)

type Type int32

// Offset is a generic form, which allows having different definitions in different Queues.
type Offset string

// BatchEvents is used by Forwarder to forward.
type BatchEvents []*protocol.Event

// OutputEventContext is a container to store the output context.
type OutputEventContext struct {
	Context map[string]*protocol.Event
	Offset  Offset
}

// Put puts the incoming event into the context.
func (c *OutputEventContext) Put(event *protocol.Event) {
	c.Context[event.GetName()] = event
}

// Get returns an event in the context. When the eventName does not exist, an error would be returned.
func (c *OutputEventContext) Get(eventName string) (*protocol.Event, error) {
	e, ok := c.Context[eventName]
	if !ok {
		err := fmt.Errorf("cannot find the event name in OutputEventContext : %s", eventName)
		return nil, err
	}
	return e, nil
}

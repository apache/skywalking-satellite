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

	"github.com/apache/skywalking-satellite/internal/pkg/event"
	"github.com/apache/skywalking-satellite/internal/pkg/plugin"
)

// Queue is a plugin interface, that defines new queues.
type Queue interface {
	plugin.Plugin

	// Publisher get the only publisher for the current queue.
	Publisher() QueuePublisher

	// Consumer get the only consumer for the current queue.
	Consumer() QueueConsumer

	// Close would close the queue.
	Close()
}

// QueuePublisher is a plugin interface, that defines new queue publishers.
type QueuePublisher interface {
	// Enqueue push a inputEvent into the queue.
	Enqueue(event *event.SerializableEvent) error
}

// QueueConsumer is a plugin interface, that defines new queue consumers.
type QueueConsumer interface {
	// Dequeue pop an event form the Queue. When the queue is empty, the method would be blocked.
	Dequeue() (event *event.SerializableEvent, offset int64, err error)
}

var QueueCategory = reflect.TypeOf((*Queue)(nil)).Elem()

func GetQueue(pluginName string, config map[string]interface{}) Queue {
	return plugin.Get(QueueCategory, pluginName, config).(Queue)
}

func init() {
	plugin.AddPluginCategory(QueueCategory)
}

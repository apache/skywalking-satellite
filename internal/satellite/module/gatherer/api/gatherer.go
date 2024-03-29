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
	"github.com/apache/skywalking-satellite/internal/satellite/event"
	"github.com/apache/skywalking-satellite/internal/satellite/module/api"
	queue "github.com/apache/skywalking-satellite/plugins/queue/api"
)

// Gatherer is the APM data collection module in Satellite.
type Gatherer interface {
	api.Module
	// PartitionCount is the all partition counter of gatherer. All event is partitioned.
	PartitionCount() int
	// OutputDataChannel is a blocking channel to transfer the apm data to the upstream processor module.
	OutputDataChannel(partition int) <-chan *queue.SequenceEvent
	// Ack the sent offset.
	Ack(lastOffset *event.Offset)
	// Inject the Processor module.
	SetProcessor(processor api.Module) error
}

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

package none

import (
	"github.com/apache/skywalking-satellite/internal/pkg/config"
	"github.com/apache/skywalking-satellite/internal/satellite/event"
	"github.com/apache/skywalking-satellite/plugins/queue/api"

	v1 "skywalking.apache.org/repo/goapi/satellite/data/v1"
)

const (
	Name     = "none-queue"
	ShowName = "None Queue"
)

type Queue struct {
	config.CommonFields
}

func (q *Queue) Name() string {
	return Name
}

func (q *Queue) ShowName() string {
	return ShowName
}

func (q *Queue) Description() string {
	return "This is an empty queue for direct connection protocols, such as SkyWalking native configuration discovery service protocol."
}

func (q *Queue) DefaultConfig() string {
	return ``
}

func (q *Queue) Initialize() error {
	return nil
}

func (q *Queue) Enqueue(e *v1.SniffData) error {
	return api.ErrFull
}

func (q *Queue) Dequeue() (*api.SequenceEvent, error) {
	return nil, api.ErrEmpty
}

func (q *Queue) Close() error {
	return nil
}

func (q *Queue) Ack(_ *event.Offset) {
}

func (q *Queue) TotalSize() int64 {
	return 0
}

func (q *Queue) UsedCount() int64 {
	return 0
}

func (q *Queue) IsFull() bool {
	return false
}

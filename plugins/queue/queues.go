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

//go:build !windows

package queue

import (
	"github.com/apache/skywalking-satellite/plugins/queue/api"
	"github.com/apache/skywalking-satellite/plugins/queue/memory"
	"github.com/apache/skywalking-satellite/plugins/queue/mmap"
	"github.com/apache/skywalking-satellite/plugins/queue/none"
)

var queues = []api.Queue{
	// Please register the queue plugins available on Linux, MacOS here.
	new(memory.Queue),
	new(mmap.Queue),
	new(none.Queue),
}

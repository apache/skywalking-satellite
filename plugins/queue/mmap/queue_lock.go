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

package mmap

import "sync/atomic"

func (q *Queue) lock(segmentID int64) {
	index := q.GetIndex(segmentID)
	q.lockByIndex(index)
}

func (q *Queue) unlock(segmentID int64) {
	index := q.GetIndex(segmentID)
	q.unlockByIndex(index)
}

func (q *Queue) lockByIndex(index int) {
	for !atomic.CompareAndSwapInt32(&q.locker[index], 0, 1) {
	}
}

func (q *Queue) unlockByIndex(index int) {
	for !atomic.CompareAndSwapInt32(&q.locker[index], 1, 0) {
	}
}

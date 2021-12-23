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

import (
	"fmt"
	"os"
	"testing"

	"github.com/apache/skywalking-satellite/internal/pkg/plugin"
	"github.com/apache/skywalking-satellite/plugins/queue/api"
)

type benchmarkParam struct {
	segmentSize      int
	message          int // unit KB
	maxInMemSegments int
	queueCapacity    int
}

var params = []benchmarkParam{
	{segmentSize: 1024 * 128, message: 8, maxInMemSegments: 18, queueCapacity: 10000},
	// compare the influence of the segmentSize.
	{segmentSize: 1024 * 256, message: 8, maxInMemSegments: 10, queueCapacity: 10000},
	{segmentSize: 1024 * 512, message: 8, maxInMemSegments: 6, queueCapacity: 10000},
	// compare the influence of the maxInMemSegments.
	{segmentSize: 1024 * 256, message: 8, maxInMemSegments: 20, queueCapacity: 10000},
	// compare the influence of the message size.
	{segmentSize: 1024 * 128, message: 16, maxInMemSegments: 10, queueCapacity: 10000},
	{segmentSize: 1024 * 128, message: 8, maxInMemSegments: 10, queueCapacity: 100000},
}

func cleanBenchmarkQueue(b *testing.B, q api.Queue) {
	if err := os.RemoveAll(q.(*Queue).queueName); err != nil {
		b.Errorf("cannot remove test queue dir, %v", err)
	}
}

func BenchmarkEnqueue(b *testing.B) {
	for _, param := range params {
		name := fmt.Sprintf("segmentSize: %dKB maxInMemSegments:%d message:%dKB queueCapacity:%d",
			param.segmentSize/1024, param.maxInMemSegments, param.message, param.queueCapacity)
		b.Run(name, func(b *testing.B) {
			q, err := initMmapQueue(plugin.Config{
				"segment_size":            param.segmentSize,
				"max_in_mem_segments":     param.maxInMemSegments,
				"queue_capacity_segments": param.queueCapacity,
			})
			if err != nil {
				b.Fatalf("cannot get a mmap queue: %v", err)
			}
			event := getLargeEvent(param.message)
			b.ReportAllocs()
			b.ResetTimer()
			println()
			for i := 0; i < b.N; i++ {
				if err := q.Enqueue(event); err != nil {
					b.Fatalf("error in pushing: %v", err)
				}
			}
			b.StopTimer()
			_ = q.Close()
			cleanBenchmarkQueue(b, q)
		})
	}
}

func BenchmarkEnqueueAndDequeue(b *testing.B) {
	for _, param := range params {
		name := fmt.Sprintf("segmentSize: %dKB maxInMemSegments:%d message:%dKB queueCapacity:%d",
			param.segmentSize/1024, param.maxInMemSegments, param.message, param.queueCapacity)
		b.Run(name, func(b *testing.B) {
			q, err := initMmapQueue(plugin.Config{
				"segment_size":            param.segmentSize,
				"max_in_mem_segments":     param.maxInMemSegments,
				"queue_capacity_segments": param.queueCapacity,
			})
			if err != nil {
				b.Fatalf("cannot get a mmap queue: %v", err)
			}
			event := getLargeEvent(param.message)
			b.ReportAllocs()
			b.ResetTimer()
			println()
			for i := 0; i < b.N; i++ {
				if err := q.Enqueue(event); err != nil {
					b.Fatalf("error in enqueue: %v", err)
				}
				if _, err := q.Dequeue(); err != nil {
					b.Fatalf("error in enqueue: %v", err)
				}
			}
			b.StopTimer()
			_ = q.Close()
			cleanBenchmarkQueue(b, q)
		})
	}
}

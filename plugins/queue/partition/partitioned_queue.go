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

package partition

import (
	"fmt"
	"reflect"
	"strconv"
	"sync/atomic"

	"github.com/apache/skywalking-satellite/internal/pkg/config"
	"github.com/apache/skywalking-satellite/internal/pkg/plugin"
	"github.com/apache/skywalking-satellite/internal/satellite/event"
	"github.com/apache/skywalking-satellite/internal/satellite/telemetry"
	"github.com/apache/skywalking-satellite/plugins/queue/api"

	v1 "skywalking.apache.org/repo/goapi/satellite/data/v1"
)

type PartitionedQueue struct {
	config.CommonFields
	Partition int `mapstructure:"partition"` // The total partition count.
	config    plugin.Config

	subQueues         []api.Queue
	loadBalancerIndex int32
	totalQueueSize    int64
}

func NewPartitionQueue(c plugin.Config) *PartitionedQueue {
	queue := &PartitionedQueue{config: c}
	plugin.Initializing(queue, c)
	return queue
}

func (p *PartitionedQueue) Initialize() error {
	queues := make([]api.Queue, p.Partition)
	pipeName := p.PipeName
	for partition := 0; partition < p.Partition; partition++ {
		p.config["pipe_name"] = fmt.Sprintf("%s-%d", p.config["pipe_name"], partition)
		queue := plugin.Get(reflect.TypeOf((*api.Queue)(nil)).Elem(), p.config).(api.Queue)
		if err := queue.Initialize(); err != nil {
			return err
		}
		queues[partition] = queue
		p.totalQueueSize += queue.TotalSize()
	}
	p.subQueues = queues
	p.registerQueueTelemetry(pipeName)
	return nil
}

func (p *PartitionedQueue) DefaultConfig() string {
	return `
# The partition count of queue.
partition: 1
`
}

func (p *PartitionedQueue) Enqueue(e *v1.SniffData) error {
	partition, err := p.findPartition(e)
	if err != nil {
		return err
	}
	return p.subQueues[partition].Enqueue(e)
}

func (p *PartitionedQueue) Dequeue(partition int) (*api.SequenceEvent, error) {
	dequeue, err := p.subQueues[partition].Dequeue()
	if err == nil {
		dequeue.Offset.Partition = partition
	}
	return dequeue, err
}

func (p *PartitionedQueue) TotalPartitionCount() int {
	return len(p.subQueues)
}

func (p *PartitionedQueue) Name() string {
	return "partition-queue"
}

func (p *PartitionedQueue) ShowName() string {
	return "Partition Queue"
}

func (p *PartitionedQueue) Description() string {
	return "The partition queue to management all the sub queues."
}

func (p *PartitionedQueue) Close() error {
	for _, q := range p.subQueues {
		if err := q.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (p *PartitionedQueue) Ack(lastOffset *event.Offset) {
	p.subQueues[lastOffset.Partition].Ack(lastOffset)
}

func (p *PartitionedQueue) findPartition(_ *v1.SniffData) (int, error) {
	if p.Partition == 1 {
		return 0, nil
	}

	// increment
	var partition int
	for {
		result := atomic.AddInt32(&p.loadBalancerIndex, 1)
		partition = int(result)

		if partition < p.Partition {
			break
		} else if atomic.CompareAndSwapInt32(&p.loadBalancerIndex, result, 0) {
			partition = 0
			break
		}
	}

	// check partition is full
	if !p.subQueues[partition].IsFull() {
		return partition, nil
	}
	for addition := 1; addition < p.Partition; addition++ {
		checkPartition := partition + addition
		if checkPartition >= p.Partition {
			checkPartition -= p.Partition
		}
		if !p.subQueues[checkPartition].IsFull() {
			return checkPartition, nil
		}
	}
	return 0, fmt.Errorf("the queue is full")
}

func (p *PartitionedQueue) registerQueueTelemetry(pipeline string) {
	telemetry.NewGauge("pipeline_queue_total_capacity", "The total capacity of pipeline", func() float64 {
		return float64(p.totalQueueSize)
	}, "pipeline", pipeline)

	for partition := 0; partition < len(p.subQueues); partition++ {
		pn := partition
		telemetry.NewGauge("pipeline_queue_partition_size",
			"The current count of elements in the queue partition",
			func() float64 {
				return float64(p.subQueues[pn].UsedCount())
			},
			"pipeline", pipeline, "partition", strconv.Itoa(pn))
	}
}

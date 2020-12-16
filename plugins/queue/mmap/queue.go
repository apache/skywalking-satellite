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

package mmap

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/grandecola/mmap"

	"github.com/apache/skywalking-satellite/internal/pkg/event"
	"github.com/apache/skywalking-satellite/internal/pkg/log"
	"github.com/apache/skywalking-satellite/plugins/queue/api"
	"github.com/apache/skywalking-satellite/plugins/queue/mmap/meta"
)

// Queue is a memory mapped queue to store the input data.
type Queue struct {
	sync.Mutex
	// config
	SegmentSize           int    `mapstructure:"segment_size"`            // The size of each segment. The unit is byte.
	MaxInMemSegments      int    `mapstructure:"max_in_mem_segments"`     // The max num of segments in memory.
	QueueCapacitySegments int    `mapstructure:"queue_capacity_segments"` // The capacity of Queue = segment_size * queue_capacity_segments.
	FlushPeriod           int    `mapstructure:"flush_period"`            // The period flush time. The unit is ms.
	FlushCeilingNum       int    `mapstructure:"flush_ceiling_num"`       // The max number in one flush time.
	MaxEventSize          int    `mapstructure:"max_event_size"`          // The max size of the input event.
	QueueDir              string `mapstructure:"queue_dir"`               // Contains all files in the queue.

	// running components
	meta                   *meta.Metadata // The metadata file.
	segments               []*mmap.File   // The data files.
	mmapCount              int            // The number of the memory mapped files.
	unflushedNum           int            // The unflushed number.
	flushChannel           chan struct{}  // The flushChannel channel would receive a signal when the unflushedNum reach the flush_ceiling_num.
	insufficientMemChannel chan struct{}  // Notify when memory is insufficient
	sufficientMemChannel   chan struct{}  // Notify when memory is sufficient

	// control components
	ctx        context.Context    // Parent ctx
	cancel     context.CancelFunc // Parent ctx cancel function
	showDownWg sync.WaitGroup     // The shutdown wait group.

	bufPool *sync.Pool

	encoder *Encoder
	decoder *Decoder
}

func (q *Queue) Name() string {
	return "mmap-queue"
}

func (q *Queue) Description() string {
	return "this is a mmap queue"
}

func (q *Queue) DefaultConfig() string {
	return `
# The size of each segment. Default value is 128K. The unit is Byte.
segment_size: 131072
# The max num of segments in memory. Default value is 10.
max_in_mem_segments: 10
# The capacity of Queue = segment_size * queue_capacity_segments.
queue_capacity_segments: 4000
# The period flush time. The unit is ms. Default value is 1 second.
flush_period: 1000
# The max number in one flush time.  Default value is 10000.
flush_ceiling_num: 10000
# Contains all files in the queue.
queue_dir: satellite-mmap-queue
# The max size of the input event. Default value is 20k.
max_event_size: 20480
`
}

func (q *Queue) Initialize() error {
	q.encoder = NewEncoder()
	q.decoder = NewDecoder()

	q.bufPool = &sync.Pool{New: func() interface{} {
		return new(bytes.Buffer)
	}}
	// the size of each segment file should be a multiple of the page size.
	pageSize := os.Getpagesize()
	if q.SegmentSize%pageSize != 0 {
		q.SegmentSize -= q.SegmentSize % pageSize
	}
	if q.SegmentSize/pageSize == 0 {
		q.SegmentSize = 131072
	}
	// the minimum MaxInMemSegments value should be 4.
	if q.MaxInMemSegments < 4 {
		q.MaxInMemSegments = 4
	}
	// load metadata and override the reading or writing offset by the committed or watermark offset.
	md, err := meta.NewMetaData(q.QueueDir, q.QueueCapacitySegments)
	if err != nil {
		return fmt.Errorf("error in creating the metadata: %v", err)
	}
	q.meta = md
	cmID, cmOffset := md.GetCommittedOffset()
	wmID, wmOffset := md.GetWatermarkOffset()
	md.PutWritingOffset(wmID, wmOffset)
	md.PutReadingOffset(cmID, cmOffset)
	// keep the reading or writing segments in the memory.
	q.segments = make([]*mmap.File, q.QueueCapacitySegments)
	if _, err := q.GetSegment(cmID); err != nil {
		return err
	}
	if _, err := q.GetSegment(wmID); err != nil {
		return err
	}
	// init components
	q.insufficientMemChannel = make(chan struct{})
	q.sufficientMemChannel = make(chan struct{})
	q.flushChannel = make(chan struct{})
	q.ctx, q.cancel = context.WithCancel(context.Background())
	// async supported processes.
	q.showDownWg.Add(2)
	go q.segmentSwapper()
	go q.flush()
	return nil
}

func (q *Queue) Push(e *event.Event) error {
	data, err := q.encoder.serialize(e)
	if err != nil {
		return err
	}
	if len(data) > q.MaxEventSize {
		return fmt.Errorf("cannot push the event to the queue because the size %dB is over ceiling", len(data))
	}
	return q.push(data)
}

func (q *Queue) Pop() (*api.SequenceEvent, error) {
	data, id, offset, err := q.pop()
	if err != nil {
		return nil, err
	}
	e, err := q.decoder.deserialize(data)
	if err != nil {
		return nil, err
	}
	return &api.SequenceEvent{
		Event:  e,
		Offset: q.encodeOffset(id, offset),
	}, nil
}

func (q *Queue) Close() error {
	q.cancel()
	q.showDownWg.Wait()
	for i, segment := range q.segments {
		if segment != nil {
			err := segment.Unmap()
			if err != nil {
				log.Logger.Errorf("cannot unmap the segments: %d, %v", i, err)
			}
		}
	}
	if err := q.meta.Close(); err != nil {
		log.Logger.Errorf("cannot unmap the metadata: %v", err)
	}
	return nil
}

func (q *Queue) Ack(lastOffset event.Offset) {
	id, offset, err := q.decodeOffset(lastOffset)
	if err != nil {
		log.Logger.Errorf("cannot ack queue with the offset:%s", lastOffset)
	}
	q.meta.PutCommittedOffset(id, offset)
}

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
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/grandecola/mmap"

	"google.golang.org/protobuf/proto"

	"github.com/apache/skywalking-satellite/internal/pkg/config"
	"github.com/apache/skywalking-satellite/internal/pkg/log"
	"github.com/apache/skywalking-satellite/internal/satellite/event"
	"github.com/apache/skywalking-satellite/plugins/queue/api"
	"github.com/apache/skywalking-satellite/plugins/queue/mmap/meta"

	v1 "skywalking.apache.org/repo/goapi/satellite/data/v1"
)

const (
	data4KB         = 131072
	minimumSegments = 4
	Name            = "mmap-queue"
	ShowName        = "Mmap Queue"
)

// Queue is a memory mapped queue to store the input data.
type Queue struct {
	config.CommonFields
	SegmentSize           int   `mapstructure:"segment_size"`            // The size of each segment. The unit is byte.
	MaxInMemSegments      int32 `mapstructure:"max_in_mem_segments"`     // The max num of segments in memory.
	QueueCapacitySegments int   `mapstructure:"queue_capacity_segments"` // The capacity of Queue = segment_size * queue_capacity_segments.
	FlushPeriod           int   `mapstructure:"flush_period"`            // The period flush time. The unit is ms.
	FlushCeilingNum       int   `mapstructure:"flush_ceiling_num"`       // The max number in one flush time.
	MaxEventSize          int   `mapstructure:"max_event_size"`          // The max size of the input event.

	// running components
	queueName              string         // The queue name.
	meta                   *meta.Metadata // The metadata file.
	segments               []*mmap.File   // The data files.
	mmapCount              int32          // The number of the memory mapped files.
	unflushedNum           int            // The unflushed number.
	flushChannel           chan struct{}  // The flushChannel channel would receive a signal when the unflushedNum reach the flush_ceiling_num.
	insufficientMemChannel chan struct{}  // Notify when memory is insufficient
	sufficientMemChannel   chan struct{}  // Notify when memory is sufficient
	markReadChannel        chan int64     // Transfer the read segmentID to do ummap operation.
	ready                  bool           // The status of the queue.
	locker                 []int32        // locker
	usedSize               int64          // The used size of queue

	// control components
	ctx        context.Context    // Parent ctx
	cancel     context.CancelFunc // Parent ctx cancel function
	showDownWg sync.WaitGroup     // The shutdown wait group.
}

func (q *Queue) Name() string {
	return Name
}

func (q *Queue) ShowName() string {
	return ShowName
}

func (q *Queue) Description() string {
	return "This is a memory mapped queue to provide the persistent storage for the input event." +
		" Please note that this plugin does not support Windows platform."
}

func (q *Queue) DefaultConfig() string {
	return `
# The size of each segment. Default value is 256K. The unit is Byte.
segment_size: 262114
# The max num of segments in memory. Default value is 10.
max_in_mem_segments: 10
# The capacity of Queue = segment_size * queue_capacity_segments.
queue_capacity_segments: 2000
# The period flush time. The unit is ms. Default value is 1 second.
flush_period: 1000
# The max number in one flush time.  Default value is 10000.
flush_ceiling_num: 10000
# The max size of the input event. Default value is 20k.
max_event_size: 20480
`
}

func (q *Queue) Initialize() error {
	// the size of each segment file should be a multiple of the page size.
	pageSize := os.Getpagesize()
	if q.SegmentSize%pageSize != 0 {
		q.SegmentSize -= q.SegmentSize % pageSize
	}
	if q.SegmentSize/pageSize == 0 {
		q.SegmentSize = data4KB
	}
	// the minimum MaxInMemSegments value should be 4.
	if q.MaxInMemSegments < minimumSegments {
		q.MaxInMemSegments = minimumSegments
	}
	q.queueName = q.Name() + "_" + q.PipeName
	// load metadata and override the reading or writing offset by the committed or watermark offset.
	md, err := meta.NewMetaData(q.queueName, q.QueueCapacitySegments)
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
	q.markReadChannel = make(chan int64, 1)
	q.flushChannel = make(chan struct{})
	q.ctx, q.cancel = context.WithCancel(context.Background())
	// async supported processes.
	q.showDownWg.Add(2)
	q.locker = make([]int32, q.QueueCapacitySegments)
	go q.segmentSwapper()
	go q.flush()
	q.ready = true
	return nil
}

func (q *Queue) Enqueue(e *v1.SniffData) error {
	if !q.ready {
		log.Logger.WithField("pipe", q.CommonFields.PipeName).Warnf("the enqueue operation would be ignored because the queue was closed.")
		return api.ErrClosed
	}
	data, err := proto.Marshal(e)
	if err != nil {
		return err
	}
	if len(data) > q.MaxEventSize {
		return fmt.Errorf("cannot enqueue the event to the queue because the size %dB is over ceiling", len(data))
	}
	return q.enqueue(data)
}

func (q *Queue) Dequeue() (*api.SequenceEvent, error) {
	if !q.ready {
		log.Logger.WithField("pipe", q.CommonFields.PipeName).Warnf("the dequeue operation would be ignored because the queue was closed.")
		return nil, api.ErrClosed
	}
	data, id, offset, err := q.dequeue()
	if err != nil {
		return nil, err
	}
	e := &v1.SniffData{}
	err = proto.Unmarshal(data, e)
	if err != nil {
		return nil, err
	}
	return &api.SequenceEvent{
		Event:  e,
		Offset: q.encodeOffset(id, offset),
	}, nil
}

func (q *Queue) Close() error {
	q.ready = false
	q.cancel()
	q.showDownWg.Wait()
	for i, segment := range q.segments {
		if segment != nil {
			if err := segment.Flush(syscall.MS_SYNC); err != nil {
				log.Logger.Errorf("cannot unmap the segments: %d, %v", i, err)
			}
			if err := segment.Unmap(); err != nil {
				log.Logger.Errorf("cannot unmap the segments: %d, %v", i, err)
			}
		}
	}
	if err := q.meta.Close(); err != nil {
		log.Logger.Errorf("cannot unmap the metadata: %v", err)
	}
	return nil
}

func (q *Queue) Ack(lastOffset *event.Offset) {
	if !q.ready {
		log.Logger.WithFields(logrus.Fields{
			"pipe":   q.CommonFields.PipeName,
			"offset": lastOffset,
		}).Warnf("the ack operation would be ignored because the queue was closed.")
		return
	}
	id, offset, err := q.decodeOffset(lastOffset)
	if err != nil {
		log.Logger.Errorf("cannot ack queue with the offset:%s", lastOffset)
	}
	q.meta.PutCommittedOffset(id, offset)
}

// flush control the flush operation by timer or counter.
func (q *Queue) flush() {
	defer q.showDownWg.Done()
	ctx, _ := context.WithCancel(q.ctx) // nolint
	for {
		timer := time.NewTimer(time.Duration(q.FlushPeriod) * time.Millisecond)
		select {
		case <-q.flushChannel:
			q.doFlush()
			timer.Reset(time.Duration(q.FlushPeriod) * time.Millisecond)
		case <-timer.C:
			q.doFlush()
		case <-ctx.Done():
			q.doFlush()
			return
		}
	}
}

// doFlush flush the segment and meta files to the disk.
func (q *Queue) doFlush() {
	for i := range q.segments {
		q.lockByIndex(i)
		if q.segments[i] == nil {
			q.unlockByIndex(i)
			continue
		}
		if err := q.segments[i].Flush(syscall.MS_SYNC); err != nil {
			log.Logger.Errorf("cannot flush segment file: %v", err)
		}
		q.unlockByIndex(i)
	}
	wid, woffset := q.meta.GetWritingOffset()
	q.meta.PutWatermarkOffset(wid, woffset)
	if err := q.meta.Flush(); err != nil {
		log.Logger.Errorf("cannot flush meta file: %v", err)
	}
}

// isEmpty returns the capacity status
func (q *Queue) isEmpty() bool {
	rid, roffset := q.meta.GetReadingOffset()
	wid, woffset := q.meta.GetWritingOffset()
	return rid == wid && roffset == woffset
}

func (q *Queue) IsFull() bool {
	rid, _ := q.meta.GetReadingOffset()
	wid, _ := q.meta.GetWritingOffset()
	// ensure enough spaces to promise data stability.
	maxWid := rid + int64(q.QueueCapacitySegments) - 1 - int64(q.MaxEventSize/q.SegmentSize)
	return wid >= maxWid
}

// encode the meta to the offset
func (q *Queue) encodeOffset(id, offset int64) event.Offset {
	return event.Offset{Position: strconv.FormatInt(id, 10) + "-" + strconv.FormatInt(offset, 10)}
}

// decode the offset to the meta of the mmap queue.
func (q *Queue) decodeOffset(val *event.Offset) (id, offset int64, err error) {
	arr := strings.Split(val.Position, "-")
	if len(arr) == 2 {
		id, err := strconv.ParseInt(arr[0], 10, 64)
		if err != nil {
			return 0, 0, err
		}
		offset, err := strconv.ParseInt(arr[1], 10, 64)
		if err != nil {
			return 0, 0, err
		}
		return id, offset, nil
	}
	return 0, 0, fmt.Errorf("the input offset string is illegal: %s", val)
}

func (q *Queue) TotalSize() int64 {
	return int64(q.QueueCapacitySegments * q.SegmentSize)
}

func (q *Queue) UsedCount() int64 {
	return q.usedSize
}

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
	"path"
	"strconv"
	"sync/atomic"
	"syscall"

	"github.com/grandecola/mmap"

	"github.com/apache/skywalking-satellite/internal/pkg/log"
	"github.com/apache/skywalking-satellite/plugins/queue/mmap/segment"
)

// GetSegment returns a memory mapped file at the segmentID position.
func (q *Queue) GetSegment(segmentID int64) (*mmap.File, error) {
	if atomic.LoadInt32(&q.mmapCount) >= q.MaxInMemSegments {
		q.insufficientMemChannel <- struct{}{}
		<-q.sufficientMemChannel
	}
	if err := q.mapSegment(segmentID); err != nil {
		return nil, err
	}
	index := q.GetIndex(segmentID)
	if q.segments[index] != nil {
		return q.segments[index], nil
	}
	return nil, fmt.Errorf("cannot get a memory mapped file at %d segment", segmentID)
}

// mapSegment load the segment file reference to the segments.
func (q *Queue) mapSegment(segmentID int64) error {
	index := q.GetIndex(segmentID)
	if q.segments[index] != nil {
		return nil
	}
	filePath := path.Join(q.queueName, strconv.Itoa(index)+segment.FileSuffix)
	file, err := segment.NewSegment(filePath, q.SegmentSize)
	if err != nil {
		return err
	}
	atomic.AddInt32(&q.mmapCount, 1)
	q.segments[index] = file
	return nil
}

// unmapSegment cancel the memory mapped status.
func (q *Queue) unmapSegment(segmentID int64) error {
	index := q.GetIndex(segmentID)
	if q.segments[index] == nil {
		return nil
	}
	if err := q.segments[index].Flush(syscall.MS_SYNC); err != nil {
		return fmt.Errorf("error in flush segemnt when unmapping: %v", err)
	}
	if err := q.segments[index].Unmap(); err != nil {
		return fmt.Errorf("error in unmap segemnt: %v", err)
	}
	atomic.AddInt32(&q.mmapCount, -1)
	q.segments[index] = nil
	return nil
}

// segmentSwapper run with a go routine to ensure the memory cost.
func (q *Queue) segmentSwapper() {
	defer q.showDownWg.Done()
	ctx, _ := context.WithCancel(q.ctx) // nolint
	for {
		select {
		case id := <-q.markReadChannel:
			q.lock(id)
			if q.unmapSegment(id) != nil {
				log.Logger.Errorf("cannot unmap the markread segment: %d", id)
			}
			q.unlock(id)
		case <-q.insufficientMemChannel:
			if q.mmapCount >= q.MaxInMemSegments {
				if q.doSwap() != nil {
					log.Logger.Errorf("cannot get enough memory to receive new data")
				}
			}
			q.sufficientMemChannel <- struct{}{}
		case <-ctx.Done():
			return
		}
	}
}

// doSwap swap the memory mapped files to normal files to promise the memory resources cost.
func (q *Queue) doSwap() error {
	rID, _ := q.meta.GetReadingOffset()
	wID, _ := q.meta.GetWritingOffset()
	logicWID := wID + int64(q.QueueCapacitySegments)
	wIndex := q.GetIndex(wID)
	rIndex := q.GetIndex(rID)
	//  only clear all memory-mapped file when more than 1.5 times MaxInMemSegments.
	clearAll := (wID - rID + 1) > int64(q.MaxInMemSegments)*3/2
	for q.mmapCount >= q.MaxInMemSegments {
		for i := logicWID - 1; i >= 0 && i >= logicWID-int64(q.MaxInMemSegments); i-- {
			if q.GetIndex(i) == wIndex || q.GetIndex(i) == rIndex {
				continue
			}
			if err := q.unmapSegment(i); err != nil {
				return err
			}
			// the writing segment and the reading segment should still in memory.
			// q.MaxInMemSegments/2-1 means keeping half available spaces to receive new data.
			if !clearAll && q.MaxInMemSegments-q.mmapCount >= q.MaxInMemSegments/2-1 {
				return nil
			}
		}
	}
	return nil
}

// GetIndex returns the index of the segments.
func (q *Queue) GetIndex(segmentID int64) int {
	return int(segmentID) % q.QueueCapacitySegments
}

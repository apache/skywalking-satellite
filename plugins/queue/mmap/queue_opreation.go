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
	"context"
	"fmt"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/apache/skywalking-satellite/internal/pkg/event"
	"github.com/apache/skywalking-satellite/internal/pkg/log"
)

const uInt64Size = 8

// flush control the flush operation by timer or counter.
func (q *Queue) flush() {
	defer q.showDownWg.Done()
	ctx, cancel := context.WithCancel(q.ctx)
	defer cancel()
	for {
		timeTicker := time.NewTicker(time.Duration(q.FlushPeriod) * time.Millisecond)
		select {
		case <-q.flushChannel:
			q.doFlush()
		case <-timeTicker.C:
			q.doFlush()
		case <-ctx.Done():
			q.doFlush()
			return
		}
	}
}

// doFlush flush the segment and meta files to the disk.
func (q *Queue) doFlush() {
	q.Lock()
	defer q.Unlock()
	for _, segment := range q.segments {
		if segment == nil {
			continue
		}
		if err := segment.Flush(syscall.MS_SYNC); err != nil {
			log.Logger.Errorf("cannot flush segment file: %v", err)
		}
	}
	wid, woffset := q.meta.GetWritingOffset()
	q.meta.PutWatermarkOffset(wid, woffset)
	if err := q.meta.Flush(); err != nil {
		log.Logger.Errorf("cannot flush meta file: %v", err)
	}
}

// push writes the data into the file system. It first writes the length of the data,
// then the data itself. It means the whole data may not exist in the one segments.
func (q *Queue) push(bytes []byte) error {
	if q.isFull() {
		return fmt.Errorf("cannot push data when the queue is full")
	}
	id, offset := q.meta.GetWritingOffset()
	id, offset, err := q.writeLength(len(bytes), id, offset)
	if err != nil {
		return err
	}
	id, offset, err = q.writeBytes(bytes, id, offset)
	if err != nil {
		return err
	}
	q.meta.PutWritingOffset(id, offset)
	q.unflushedNum++
	if q.unflushedNum == q.FlushCeilingNum {
		q.flushChannel <- struct{}{}
		q.unflushedNum = 0
	}
	return nil
}

// pop reads the data from the file system. It first reads the length of the data,
// then the data itself. It means the whole data may not exist in the one segments.
func (q *Queue) pop() (data []byte, rid, roffset int64, err error) {
	if q.isEmpty() {
		return nil, 0, 0, fmt.Errorf("cannot read data when the queue is empty")
	}
	id, offset := q.meta.GetReadingOffset()
	id, offset, length, err := q.readLength(id, offset)
	if err != nil {
		return nil, 0, 0, err
	}
	bytes, id, offset, err := q.readBytes(id, offset, length)
	if err != nil {
		return nil, 0, 0, err
	}
	q.meta.PutReadingOffset(id, offset)
	return bytes, id, offset, nil
}

// readBytes reads bytes into the memory mapped file.
func (q *Queue) readBytes(id, offset int64, length int) (data []byte, newID, newOffset int64, err error) {
	counter := 0
	res := make([]byte, length)
	for {
		segment, err := q.GetSegment(id)
		if err != nil {
			return nil, 0, 0, err
		}
		readBytes, err := segment.ReadAt(res[counter:], offset)
		if err != nil {
			return nil, 0, 0, err
		}
		counter += readBytes
		offset += int64(readBytes)
		if offset == int64(q.SegmentSize) {
			id, offset = id+1, 0
		}
		if counter == length {
			break
		}
	}
	return res, id, offset, nil
}

// readLength reads the data length with 8 Bits spaces.
func (q *Queue) readLength(id, offset int64) (newID, newOffset int64, length int, err error) {
	if offset+uInt64Size > int64(q.SegmentSize) {
		id, offset = id+1, 0
	}
	segment, err := q.GetSegment(id)
	if err != nil {
		return 0, 0, 0, err
	}
	num := segment.ReadUint64At(offset)
	offset += uInt64Size
	if offset == int64(q.SegmentSize) {
		id, offset = id+1, 0
	}
	return id, offset, int(num), nil
}

// writeLength write the data length with 8 Bits spaces.
func (q *Queue) writeLength(length int, id, offset int64) (newID, newOffset int64, err error) {
	if offset+uInt64Size > int64(q.SegmentSize) {
		id, offset = id+1, 0
	}
	segment, err := q.GetSegment(id)
	if err != nil {
		return 0, 0, err
	}
	segment.WriteUint64At(uint64(length), offset)
	offset += uInt64Size
	if offset == int64(q.SegmentSize) {
		id, offset = id+1, 0
	}
	return id, offset, nil
}

// writeBytes writes bytes into the memory mapped file.
func (q *Queue) writeBytes(bytes []byte, id, offset int64) (newID, newOffset int64, err error) {
	counter := 0
	length := len(bytes)

	for {
		segment, err := q.GetSegment(id)
		if err != nil {
			return 0, 0, err
		}
		writtenBytes, err := segment.WriteAt(bytes[counter:], offset)
		if err != nil {
			return 0, 0, err
		}
		counter += writtenBytes
		offset += int64(writtenBytes)
		if offset == int64(q.SegmentSize) {
			id, offset = id+1, 0
		}
		if counter == length {
			break
		}
	}
	return id, offset, nil
}

// isEmpty returns the capacity status
func (q *Queue) isEmpty() bool {
	rid, roffset := q.meta.GetReadingOffset()
	wid, woffset := q.meta.GetWritingOffset()
	return rid == wid && roffset == woffset
}

// isEmpty returns the capacity status
func (q *Queue) isFull() bool {
	rid, _ := q.meta.GetReadingOffset()
	wid, _ := q.meta.GetWritingOffset()
	// ensure enough spaces to promise data stability.
	maxWid := rid + int64(q.QueueCapacitySegments) - 1 - int64(q.MaxEventSize/q.SegmentSize)
	return wid >= maxWid
}

// encode the meta to the offset
func (q *Queue) encodeOffset(id, offset int64) event.Offset {
	return event.Offset(strconv.FormatInt(id, 10) + "-" + strconv.FormatInt(offset, 10))
}

// decode the offset to the meta of the mmap queue.
func (q *Queue) decodeOffset(val event.Offset) (id, offset int64, err error) {
	arr := strings.Split(string(val), "-")
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

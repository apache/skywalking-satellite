// MIT License
//
// Copyright (c) 2018 Aman Mangal
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

//go:build !windows

package mmap

import (
	"sync/atomic"

	"github.com/apache/skywalking-satellite/plugins/queue/api"
)

// Because the design of the mmap-queue in Satellite references the design of the
// bigqueue(https://github.com/grandecola/bigqueue), the queue operation file retains
// the original author license.
//
// The reason why we reference the source codes of bigqueue rather than using the lib
// is the file queue in Satellite is like following.
// 1. Only one consumer and publisher in the Satellite queue.
// 2. Reusing files strategy is required to reduce the creation times in the Satellite queue.
// 3. More complex OFFSET design is needed to ensure the final stability of data.

const uInt64Size = 8

// enqueue writes the data into the file system. It first writes the length of the data,
// then the data itself. It means the whole data may not exist in the one segments.
func (q *Queue) enqueue(bytes []byte) error {
	if q.IsFull() {
		return api.ErrFull
	}
	id, offset := q.meta.GetWritingOffset()
	byteSize := len(bytes)
	id, offset, err := q.writeLength(byteSize, id, offset)
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
	atomic.AddInt64(&q.usedSize, int64(byteSize))
	return nil
}

// dequeue reads the data from the file system. It first reads the length of the data,
// then the data itself. It means the whole data may not exist in the one segments.
func (q *Queue) dequeue() (data []byte, rid, roffset int64, err error) {
	if q.isEmpty() {
		return nil, 0, 0, api.ErrEmpty
	}
	preID, preOffset := q.meta.GetReadingOffset()
	id, offset, length, err := q.readLength(preID, preOffset)
	if err != nil {
		return nil, 0, 0, err
	}
	bytes, id, offset, err := q.readBytes(id, offset, length)
	if err != nil {
		return nil, 0, 0, err
	}
	q.meta.PutReadingOffset(id, offset)
	if id != preID {
		q.markReadChannel <- preID
	}
	atomic.AddInt64(&q.usedSize, int64(-len(bytes)))
	return bytes, id, offset, nil
}

// readBytes reads bytes into the memory mapped file.
func (q *Queue) readBytes(id, offset int64, length int) (data []byte, newID, newOffset int64, err error) {
	counter := 0
	res := make([]byte, length)
	for {
		q.lock(id)
		segment, err := q.GetSegment(id)
		if err != nil {
			return nil, 0, 0, err
		}
		readBytes, err := segment.ReadAt(res[counter:], offset)
		q.unlock(id)
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
	q.lock(id)
	segment, err := q.GetSegment(id)
	if err != nil {
		return 0, 0, 0, err
	}
	num := segment.ReadUint64At(offset)
	q.unlock(id)
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
	q.lock(id)
	segment, err := q.GetSegment(id)
	if err != nil {
		return 0, 0, err
	}
	segment.WriteUint64At(uint64(length), offset)
	q.unlock(id)
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
		q.lock(id)
		segment, err := q.GetSegment(id)
		if err != nil {
			return 0, 0, err
		}
		writtenBytes, err := segment.WriteAt(bytes[counter:], offset)
		q.unlock(id)
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

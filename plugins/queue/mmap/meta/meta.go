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

package meta

import (
	"fmt"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/grandecola/mmap"

	"github.com/apache/skywalking-satellite/plugins/queue/mmap/segment"
)

const (
	metaSize    = 80
	metaName    = "meta.dat"
	metaVersion = 1
)

// Offset position
const (
	versionPos = iota * 8
	widPos
	woffsetPos
	wmidPos
	wmoffsetPos
	cidPos
	coffsetPos
	ridPos
	roffsetPos
	capacityPos
)

// Metadata only needs 80B to store the Metadata for the pipe. But for memory alignment,
// it takes at least one memory page size, which is generally 4K.
//
// [    8Bit   ][  8Bit ][  8Bit ][  8Bit ][  8Bit ][  8Bit ][  8Bit ][  8Bit ][  8Bit ][  8Bit  ]
// [metaVersion][  ID   ][ offset][  ID   ][ offset][  ID   ][ offset][  ID   ][ offset][capacity]
// [metaVersion][writing   offset][watermark offset][committed offset][reading   offset][capacity]
type Metadata struct {
	metaFile *mmap.File
	name     string
	size     int
	capacity int
	lock     sync.RWMutex
}

// NewMetaData read or create a Metadata with supported metaVersion
func NewMetaData(metaDir string, capacity int) (*Metadata, error) {
	path := filepath.Join(metaDir, metaName)
	metaFile, err := segment.NewSegment(path, metaSize)
	if err != nil {
		return nil, fmt.Errorf("error in crating the Metadata memory mapped file: %v", err)
	}

	m := &Metadata{
		metaFile: metaFile,
		name:     metaName,
		size:     metaSize,
		capacity: capacity,
	}

	v := m.GetVersion()
	if v != 0 && v != metaVersion {
		return nil, fmt.Errorf("metadata metaVersion is not matching, the Metadata metaVersion is %d", v)
	}
	c := m.GetCapacity()
	if c != 0 && c != capacity {
		return nil, fmt.Errorf("metadata catapacity is not equal to the old capacity, the old capacity is %d", c)
	}
	m.PutVersion(metaVersion)
	m.PutCapacity(int64(capacity))
	return m, nil
}

// GetVersion returns the meta version.
func (m *Metadata) GetVersion() int {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return int(m.metaFile.ReadUint64At(versionPos))
}

// PutVersion put the version into the memory mapped file.
func (m *Metadata) PutVersion(version int64) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.metaFile.WriteUint64At(uint64(version), versionPos)
}

// GetWritingOffset returns the writing offset, which contains the segment ID and the offset of the segment.
func (m *Metadata) GetWritingOffset() (segmentID, offset int64) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return int64(m.metaFile.ReadUint64At(widPos)), int64(m.metaFile.ReadUint64At(woffsetPos))
}

// PutWritingOffset put the segment ID and the offset of the segment into the writing offset.
func (m *Metadata) PutWritingOffset(segmentID, offset int64) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.metaFile.WriteUint64At(uint64(segmentID), widPos)
	m.metaFile.WriteUint64At(uint64(offset), woffsetPos)
}

// GetWatermarkOffset returns the watermark offset, which contains the segment ID and the offset of the segment.
func (m *Metadata) GetWatermarkOffset() (segmentID, offset int64) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return int64(m.metaFile.ReadUint64At(wmidPos)), int64(m.metaFile.ReadUint64At(wmoffsetPos))
}

// PutWatermarkOffset put the segment ID and the offset of the segment into the watermark offset.
func (m *Metadata) PutWatermarkOffset(segmentID, offset int64) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.metaFile.WriteUint64At(uint64(segmentID), wmidPos)
	m.metaFile.WriteUint64At(uint64(offset), wmoffsetPos)
}

// GetCommittedOffset returns the committed offset, which contains the segment ID and the offset of the segment.
func (m *Metadata) GetCommittedOffset() (segmentID, offset int64) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return int64(m.metaFile.ReadUint64At(cidPos)), int64(m.metaFile.ReadUint64At(coffsetPos))
}

// PutCommittedOffset put the segment ID and the offset of the segment into the committed offset.
func (m *Metadata) PutCommittedOffset(segmentID, offset int64) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.metaFile.WriteUint64At(uint64(segmentID), cidPos)
	m.metaFile.WriteUint64At(uint64(offset), coffsetPos)
}

// GetReadingOffset returns the reading offset, which contains the segment ID and the offset of the segment.
func (m *Metadata) GetReadingOffset() (segmentID, offset int64) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return int64(m.metaFile.ReadUint64At(ridPos)), int64(m.metaFile.ReadUint64At(roffsetPos))
}

// PutReadingOffset put the segment ID and the offset of the segment into the reading offset.
func (m *Metadata) PutReadingOffset(segmentID, offset int64) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.metaFile.WriteUint64At(uint64(segmentID), ridPos)
	m.metaFile.WriteUint64At(uint64(offset), roffsetPos)
}

// GetCapacity returns the capacity of the queue.
func (m *Metadata) GetCapacity() int {
	m.lock.RLock()
	defer m.lock.RUnlock()
	return int(m.metaFile.ReadUint64At(capacityPos))
}

// PutCapacity put the capacity into the memory mapped file.
func (m *Metadata) PutCapacity(version int64) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.metaFile.WriteUint64At(uint64(version), capacityPos)
}

// Flush the memory mapped file to the disk.
func (m *Metadata) Flush() error {
	m.lock.Lock()
	defer m.lock.Unlock()
	return m.metaFile.Flush(syscall.MS_SYNC)
}

// Close do Flush operation and unmap the memory mapped file.
func (m *Metadata) Close() error {
	m.lock.Lock()
	defer m.lock.Unlock()
	if err := m.metaFile.Flush(syscall.MS_SYNC); err != nil {
		return err
	}
	return m.metaFile.Unmap()
}

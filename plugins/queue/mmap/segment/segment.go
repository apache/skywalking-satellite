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

package segment

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/grandecola/mmap"
)

const FileSuffix = "_segment.dat"

// NewSegment returns a pointer to a memory mapped file according to the given file name and file size.
// The size of each segment file should be a multiple of the page size.
func NewSegment(name string, size int) (*mmap.File, error) {
	name, err := filepath.Abs(name)
	if err != nil {
		return nil, fmt.Errorf("error in getting the absolute path of the segment file : %v", err)
	}
	paths, _ := filepath.Split(name)
	_, err = os.Stat(paths)
	if err != nil && os.IsNotExist(err) && os.MkdirAll(paths, 0o744) != nil {
		return nil, fmt.Errorf("error in creating the parent dirs of the segment file : %v", err)
	}

	file, err := os.OpenFile(name, os.O_CREATE|os.O_RDWR, 0o744)
	if err != nil {
		return nil, fmt.Errorf("error in opening segment file : %v", err)
	}

	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("error in reading info of the segment file : %v", err)
	}

	if stat.Size() != int64(size) {
		if file.Truncate(int64(size)) != nil {
			return nil, fmt.Errorf("error in truncating file: %v", err)
		}
	}

	segment, err := mmap.NewSharedFileMmap(file, 0, size, syscall.PROT_READ|syscall.PROT_WRITE)
	if err != nil {
		return nil, fmt.Errorf("error in creating the mmap segment file : %v", err)
	}
	if err := file.Close(); err != nil {
		return nil, fmt.Errorf("error in closing the segment file : %v", err)
	}
	return segment, nil
}

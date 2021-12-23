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
	"os"
	"testing"

	"github.com/grandecola/mmap"
)

func Test_newSegmentWithOldFile(t *testing.T) {
	type args struct {
		fileName     string
		originalSize int
		neededSize   int
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "less than the needed size",
			args: args{
				fileName:     "temp.segment",
				originalSize: os.Getpagesize(),
				neededSize:   os.Getpagesize() * 2,
			},
			want:    int64(os.Getpagesize() * 2),
			wantErr: false,
		},
		{
			name: "equal to the needed size",
			args: args{
				fileName:     "temp2.segment",
				originalSize: os.Getpagesize() * 2,
				neededSize:   os.Getpagesize() * 2,
			},
			want:    int64(os.Getpagesize() * 2),
			wantErr: false,
		},
		{
			name: "larger than the needed size",
			args: args{
				fileName:     "temp3.segment",
				originalSize: os.Getpagesize() * 3,
				neededSize:   os.Getpagesize() * 2,
			},
			want:    int64(os.Getpagesize() * 2),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if file, err := os.Create(tt.args.fileName); err != nil {
				t.Errorf("cannot create the original file: %v", err)
				return
			} else if err := file.Truncate(int64(tt.args.originalSize)); err != nil {
				t.Errorf("cannot set the original file size: %v", err)
			}

			got, err := NewSegment(tt.args.fileName, tt.args.neededSize)
			defer clean(got, tt.args.fileName, t)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewSegment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if file, err := os.Open(tt.args.fileName); err != nil {
				t.Errorf("cannot open the mmap file: %v", err)
			} else if stat, err := file.Stat(); err != nil {
				t.Errorf("cannot read mmap file info: %v", err)
			} else if stat.Size() != tt.want {
				t.Errorf("want file size is %d ,but got %d", tt.want, stat.Size())
			}
		})
	}
}

func Test_newSegmentSize(t *testing.T) {
	type args struct {
		fileName string
		size     int
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{
			name: "equal to page size",
			args: args{
				fileName: "temp2.segment",
				size:     os.Getpagesize() * 2,
			},
			want:    int64(os.Getpagesize() * 2),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewSegment(tt.args.fileName, tt.args.size)
			defer clean(got, tt.args.fileName, t)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewSegment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if file, err := os.Open(tt.args.fileName); err != nil {
				t.Errorf("cannot open the mmap file: %v", err)
			} else if stat, err := file.Stat(); err != nil {
				t.Errorf("cannot read mmap file info: %v", err)
			} else if stat.Size() != tt.want {
				t.Errorf("want file size is %d ,but got %d", tt.want, stat.Size())
			}
		})
	}
}

func Test_newSegmentMultiDir(t *testing.T) {
	const testDir = "testQueue"
	defer func() {
		if err := os.RemoveAll(testDir); err != nil {
			t.Errorf("cannot clean the testqueue dir: %v", err)
			return
		}
	}()

	type args struct {
		fileName string
		size     int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test multi dir",
			args: args{
				fileName: testDir + "/temp.segment",
				size:     10,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewSegment(tt.args.fileName, tt.args.size)
			defer clean(got, tt.args.fileName, t)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewSegment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func clean(file *mmap.File, fileName string, t *testing.T) {
	if file == nil {
		return
	}
	if err := file.Unmap(); err != nil {
		t.Errorf("unmap segment file error: %v", err)
	}
	if err := os.Remove(fileName); err != nil {
		t.Errorf("delete segment file error: %v", err)
	}
}

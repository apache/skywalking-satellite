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
	"os"
	"reflect"
	"testing"
)

type args struct {
	metaVersion int
	capacity    int
	// ID
	writingSegID   int64
	readingSegID   int64
	committedSegID int64
	watermarkSegID int64
	// offset
	writingSegOffset   int64
	readingSegOffset   int64
	committedSegOffset int64
	watermarkSegOffset int64
}

type tests struct {
	name    string
	args    args
	want    args
	wantErr bool
}

func Test_newMetaData(t *testing.T) {
	const testDir = "testMeta"
	const preCapacity = 500
	params := []tests{
		buildMetaverionErrorTest(),
		buildNormalTest(),
		buildCapacitynErrorTest(),
	}

	for _, tt := range params {
		t.Run(tt.name, func(t *testing.T) {
			// clean
			defer func() {
				if err := os.RemoveAll(testDir); err != nil {
					t.Errorf("remove Metadata dir error: %v", err)
				}
			}()

			got, err := NewMetaData(testDir, preCapacity)
			if err != nil {
				t.Errorf("cannot create Metadata file: %v", err)
				return
			}

			// write args
			got.PutVersion(int64(tt.args.metaVersion))
			got.PutWritingOffset(tt.args.writingSegID, tt.args.writingSegOffset)
			got.PutReadingOffset(tt.args.readingSegID, tt.args.readingSegOffset)
			got.PutCommittedOffset(tt.args.committedSegID, tt.args.committedSegOffset)
			got.PutWatermarkOffset(tt.args.watermarkSegID, tt.args.watermarkSegOffset)
			if got.Close() != nil {
				t.Errorf("cannot Close the Metadata file: %v", err)
				return
			}

			oldMeta, err := NewMetaData(testDir, tt.args.capacity)
			if err != nil {
				if tt.wantErr {
					return
				}
				t.Errorf("cannot read old Metadata file: %v", err)
				return
			}

			// read args
			wmID, wmOffset := oldMeta.GetWatermarkOffset()
			cID, cOffset := oldMeta.GetCommittedOffset()
			rID, rOffset := oldMeta.GetReadingOffset()
			wID, wOffset := oldMeta.GetWritingOffset()

			readArgs := args{
				metaVersion:        oldMeta.GetVersion(),
				capacity:           oldMeta.GetCapacity(),
				watermarkSegID:     wmID,
				writingSegID:       wID,
				readingSegID:       rID,
				committedSegID:     cID,
				watermarkSegOffset: wmOffset,
				readingSegOffset:   rOffset,
				committedSegOffset: cOffset,
				writingSegOffset:   wOffset,
			}
			if !reflect.DeepEqual(readArgs, tt.want) {
				t.Errorf("want meta info is [%+v]\n ,but got [%+v]", tt.want, readArgs)
			}
		})
	}
}

func buildMetaverionErrorTest() tests {
	return tests{
		name: "wrong version",
		args: args{
			metaVersion:        2,
			capacity:           500,
			watermarkSegID:     1,
			writingSegID:       2,
			readingSegID:       3,
			committedSegID:     4,
			watermarkSegOffset: 10,
			readingSegOffset:   20,
			committedSegOffset: 30,
			writingSegOffset:   40,
		},
		want: args{
			metaVersion:        2,
			capacity:           500,
			watermarkSegID:     1,
			writingSegID:       2,
			readingSegID:       3,
			committedSegID:     4,
			watermarkSegOffset: 10,
			readingSegOffset:   20,
			committedSegOffset: 30,
			writingSegOffset:   40,
		},
		wantErr: true,
	}
}

func buildCapacitynErrorTest() tests {
	return tests{
		name: "wrong version",
		args: args{
			metaVersion:        1,
			capacity:           600,
			watermarkSegID:     1,
			writingSegID:       2,
			readingSegID:       3,
			committedSegID:     4,
			watermarkSegOffset: 10,
			readingSegOffset:   20,
			committedSegOffset: 30,
			writingSegOffset:   40,
		},
		want: args{
			metaVersion:        1,
			capacity:           600,
			watermarkSegID:     1,
			writingSegID:       2,
			readingSegID:       3,
			committedSegID:     4,
			watermarkSegOffset: 10,
			readingSegOffset:   20,
			committedSegOffset: 30,
			writingSegOffset:   40,
		},
		wantErr: true,
	}
}

func buildNormalTest() tests {
	return tests{
		name: "correct version",
		args: args{
			metaVersion:        1,
			capacity:           500,
			watermarkSegID:     2,
			writingSegID:       3,
			readingSegID:       4,
			committedSegID:     5,
			watermarkSegOffset: 6,
			readingSegOffset:   7,
			committedSegOffset: 8,
			writingSegOffset:   9,
		},
		want: args{
			metaVersion:        1,
			capacity:           500,
			watermarkSegID:     2,
			writingSegID:       3,
			readingSegID:       4,
			committedSegID:     5,
			watermarkSegOffset: 6,
			readingSegOffset:   7,
			committedSegOffset: 8,
			writingSegOffset:   9,
		},
		wantErr: true,
	}
}

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

package logger

import (
	"reflect"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func TestFormatter_Format(t *testing.T) {
	Init(SetLogPattern("[%time][%level][%field] - %msg"), SetTimePattern("2006-01-02 15:04:05,001"))
	type args struct {
		entry *logrus.Entry
	}

	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "logWithEmptyFields",
			want: []byte("[2020-12-12 12:12:12,012][trace][] - entry1"),
			args: args{
				entry: func() *logrus.Entry {
					entry := Log.WithTime(time.Date(2020, 12, 12, 12, 12, 12, 12, time.Local).Local())
					entry.Level = logrus.TraceLevel
					entry.Message = "entry1"
					return entry
				}(),
			},
		},
		{
			name: "logWithFields",
			want: []byte("[2020-12-12 12:12:12,012][warning][a=b] - entry2"),
			args: args{
				entry: func() *logrus.Entry {
					entry := Log.WithField("a", "b").WithTime(time.Date(2020, 12, 12, 12, 12, 12, 12, time.Local).Local())
					entry.Level = logrus.WarnLevel
					entry.Message = "entry2"
					return entry
				}(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := Log.Formatter
			got, _ := f.Format(tt.args.entry)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Format() got = %s, want %s", got, tt.want)
			}
		})
	}
}

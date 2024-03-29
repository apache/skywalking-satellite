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

package buffer

import (
	"testing"

	"github.com/apache/skywalking-satellite/internal/pkg/log"
	"github.com/apache/skywalking-satellite/internal/satellite/event"
)

func TestNewBuffer(t *testing.T) {
	buffer := NewBatchBuffer(3)
	tests := []struct {
		name string
		args *event.OutputEventContext
		want int
	}{
		{
			name: "add-1",
			args: &event.OutputEventContext{Offset: &event.Offset{Position: "1"}},
			want: 1,
		},
		{
			name: "add-2",
			args: &event.OutputEventContext{Offset: &event.Offset{Position: "1"}},
			want: 2,
		},
		{
			name: "add-3",
			args: &event.OutputEventContext{Offset: &event.Offset{Position: "1"}},
			want: 3,
		},
		{
			name: "add-4",
			args: &event.OutputEventContext{Offset: &event.Offset{Position: "1"}},
			want: 3,
		},
		{
			name: "add-5",
			args: &event.OutputEventContext{Offset: &event.Offset{Position: "1"}},
			want: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer.Add(tt.args)
			if got := buffer.Len(); got != tt.want {
				t.Errorf("Buffer Len() = %v, want %v", got, tt.want)
			}
		})
	}
}

func init() {
	log.Init(&log.LoggerConfig{})
}

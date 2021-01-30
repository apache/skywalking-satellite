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

package memory

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/apache/skywalking-satellite/internal/pkg/plugin"
	_ "github.com/apache/skywalking-satellite/internal/satellite/test"
	"github.com/apache/skywalking-satellite/plugins/queue/api"
	"github.com/apache/skywalking-satellite/protocol/gen-codes/satellite/protocol"
)

func initMemoryQueue(cfg plugin.Config) (*Queue, error) {
	plugin.RegisterPluginCategory(reflect.TypeOf((*api.Queue)(nil)).Elem())
	plugin.RegisterPlugin(&Queue{})
	var config plugin.Config = map[string]interface{}{
		plugin.NameField: Name,
	}
	for k, v := range cfg {
		config[k] = v
	}
	q := api.GetQueue(config)
	if q == nil {
		return nil, fmt.Errorf("cannot get a memoory queue from the registry")
	}
	if err := q.Initialize(); err != nil {
		return nil, fmt.Errorf("queue cannot initialize: %v", err)
	}
	return q.(*Queue), nil
}

func TestQueue_Push_Strategy(t *testing.T) {
	const num = 5
	tests := []struct {
		name    string
		args    plugin.Config
		wantErr bool
	}{
		{
			name: "test_lost_the_oldest_one_discard_strategy",
			args: map[string]interface{}{
				"event_buffer_size": num,
				"discard_strategy":  discardOldest,
			},
			wantErr: false,
		},
		{
			name: "test_lost_the_new_one_discard_strategy",
			args: map[string]interface{}{
				"event_buffer_size": num,
				"discard_strategy":  discardLatest,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q, err := initMemoryQueue(tt.args)
			if err != nil {
				t.Fatalf("cannot init the memory queue: %v", err)
			}
			for i := 0; i < num; i++ {
				if err := q.Push(new(protocol.Event)); err != nil {
					t.Fatalf("cannot push event to the queue: %v", err)
				}
			}
			if err := q.Push(new(protocol.Event)); (err != nil) != tt.wantErr {
				t.Errorf("Push() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestQueue_Push(t *testing.T) {
	const num = 10
	q, err := initMemoryQueue(map[string]interface{}{
		"event_buffer_size": num,
		"discard_strategy":  discardLatest,
	})
	if err != nil {
		t.Fatalf("cannot init the memory queue: %v", err)
	}

	for i := 0; i < num; i++ {
		if err := q.Push(new(protocol.Event)); err != nil {
			t.Fatalf("the push want seuccess but failure: %v", err)
		}
	}
	if err := q.Push(new(protocol.Event)); err == nil {
		t.Fatalf("the push want failure but success")
	}
	for i := 0; i < num; i++ {
		if e, err := q.Pop(); err != nil {
			t.Fatalf("the pop want seuccess but failure: %v", err)
		} else if e == nil {
			t.Fatalf("the pop want a event but got nil")
		}
	}
	if _, err := q.Pop(); err == nil {
		t.Fatalf("the pop want error but success: %v", err)
	}
}

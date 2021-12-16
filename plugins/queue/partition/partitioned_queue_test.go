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

package partition

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/apache/skywalking-satellite/internal/pkg/log"
	"github.com/apache/skywalking-satellite/internal/pkg/plugin"
	"github.com/apache/skywalking-satellite/internal/satellite/telemetry"
	_ "github.com/apache/skywalking-satellite/internal/satellite/telemetry/none"
	"github.com/apache/skywalking-satellite/plugins/queue/api"
	"github.com/apache/skywalking-satellite/plugins/queue/memory"

	v1 "skywalking.apache.org/repo/goapi/satellite/data/v1"
)

func init() {
	log.Init(&log.LoggerConfig{})
	c := &telemetry.Config{}
	c.ExportType = "none"
	if err := telemetry.Init(c); err != nil {
		panic(err)
	}
}

func initPartitionQueue(queueName string, cfg plugin.Config) (*PartitionedQueue, error) {
	plugin.RegisterPluginCategory(reflect.TypeOf((*api.Queue)(nil)).Elem())
	plugin.RegisterPlugin(&memory.Queue{})
	var config plugin.Config = map[string]interface{}{
		plugin.NameField: queueName,
	}
	for k, v := range cfg {
		config[k] = v
	}
	q := NewPartitionQueue(config)
	if q == nil {
		return nil, fmt.Errorf("cannot get a partition queue from the registry")
	}
	if err := q.Initialize(); err != nil {
		return nil, fmt.Errorf("queue cannot initialize: %v", err)
	}
	return q, nil
}

func initData(count int) []*v1.SniffData {
	result := make([]*v1.SniffData, count)
	for inx := range result {
		result[inx] = &v1.SniffData{Name: strconv.Itoa(inx)}
	}
	return result
}

func TestPartitionQueue(t *testing.T) {
	tests := []struct {
		name              string
		bufferSize        int
		partitionCount    int
		initDataCount     int
		loadBalancerIndex int
		testCount         int
		enqueueError      bool
	}{
		{
			name:              "normal simple queue",
			bufferSize:        10,
			partitionCount:    1,
			initDataCount:     2,
			loadBalancerIndex: 2,
			testCount:         6,
			enqueueError:      false,
		},
		{
			name:              "normal multiple queue",
			bufferSize:        10,
			partitionCount:    5,
			initDataCount:     10,
			loadBalancerIndex: 0,
			testCount:         10,
			enqueueError:      false,
		},
		{
			name:              "single partition full",
			bufferSize:        2,
			partitionCount:    5,
			initDataCount:     8,
			loadBalancerIndex: 2,
			testCount:         2,
			enqueueError:      false,
		},
		{
			name:              "all partition full",
			bufferSize:        2,
			partitionCount:    5,
			initDataCount:     10,
			loadBalancerIndex: 1,
			testCount:         1,
			enqueueError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queue, err := initPartitionQueue("memory-queue", map[string]interface{}{
				"event_buffer_size": tt.bufferSize,
				"partition":         tt.partitionCount,
			})
			if err != nil {
				t.Fatal(err)
			}

			initDates := initData(tt.initDataCount)
			for _, d := range initDates {
				if err2 := queue.Enqueue(d); err2 != nil {
					t.Fatal(err2)
				}
			}

			queue.loadBalancerIndex = int32(tt.loadBalancerIndex)

			testDates := initData(tt.testCount)
			for _, d := range testDates {
				err = queue.Enqueue(d)
			}
			if tt.enqueueError && err == nil {
				t.Fatalf("should contains enqueue error")
			} else if !tt.enqueueError && err != nil {
				t.Fatalf("should not contain enqueue error, error: %v", err)
			}
		})
	}
}

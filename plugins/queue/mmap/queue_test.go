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

package mmap

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"google.golang.org/protobuf/proto"

	common "skywalking.apache.org/repo/goapi/collect/common/v3"
	logging "skywalking.apache.org/repo/goapi/collect/logging/v3"
	v1 "skywalking.apache.org/repo/goapi/satellite/data/v1"

	"github.com/google/go-cmp/cmp"

	"github.com/apache/skywalking-satellite/internal/pkg/plugin"
	_ "github.com/apache/skywalking-satellite/internal/satellite/test"
	"github.com/apache/skywalking-satellite/plugins/queue/api"
)

func initMmapQueue(cfg plugin.Config) (*Queue, error) {
	plugin.RegisterPluginCategory(reflect.TypeOf((*api.Queue)(nil)).Elem())
	plugin.RegisterPlugin(&Queue{})
	var config plugin.Config = map[string]interface{}{
		plugin.NameField: Name,
	}
	for k, v := range cfg {
		config[k] = v
	}
	q := api.GetQueue(config)
	q.(*Queue).PipeName = "test-pipe"
	if q == nil {
		return nil, fmt.Errorf("cannot get a default config mmap queue from the registry")
	}
	if err := q.Initialize(); err != nil {
		return nil, fmt.Errorf("queue cannot initialize: %v", err)
	}
	return q.(*Queue), nil
}

func cleanTestQueue(t *testing.T, q api.Queue) {
	if err := os.RemoveAll(q.(*Queue).queueName); err != nil {
		t.Errorf("cannot remove test queue dir, %v", err)
	}
}

func getBatchEvents(count int) []*v1.SniffData {
	var slice []*v1.SniffData
	for i := 0; i < count; i++ {
		log := &logging.LogData{
			Service:         "mock-service",
			ServiceInstance: "mock-serviceInstance",
			Timestamp:       time.Date(2020, 12, 20, 12, 12, 12, 0, time.UTC).Unix(),
			Endpoint:        "mock-endpoint",
			Tags:            &logging.LogTags{},
			TraceContext: &logging.TraceContext{
				TraceId:        "traceId",
				TraceSegmentId: "trace-segmentId",
				SpanId:         12,
			},
			Body: &logging.LogDataBody{
				Type: "body-type",
				Content: &logging.LogDataBody_Text{
					Text: &logging.TextLog{
						Text: getNKData(2) + strconv.Itoa(i),
					},
				},
			},
		}
		logBytes, _ := proto.Marshal(log)
		slice = append(slice, &v1.SniffData{
			Name:      "event" + strconv.Itoa(i),
			Timestamp: time.Now().Unix(),
			Meta: map[string]string{
				"meta": "mval" + strconv.Itoa(i),
			},
			Type:   v1.SniffType_Logging,
			Remote: true,
			Data: &v1.SniffData_LogList{
				LogList: &v1.BatchLogList{Logs: [][]byte{logBytes}},
			},
		},
		)
	}
	return slice
}

func getNKData(n int) string {
	return strings.Repeat("a", n*1024)
}

func getLargeEvent(n int) *v1.SniffData {
	log := &logging.LogData{
		Service:         "mock-service",
		ServiceInstance: "mock-serviceInstance",
		Timestamp:       time.Date(2020, 12, 20, 12, 12, 12, 0, time.UTC).Unix(),
		Endpoint:        "mock-endpoint",
		Tags: &logging.LogTags{
			Data: []*common.KeyStringValuePair{
				{
					Key:   "tags-key",
					Value: "tags-val",
				},
			},
		},
		TraceContext: &logging.TraceContext{
			TraceId:        "traceId",
			TraceSegmentId: "trace-segmentId",
			SpanId:         12,
		},
		Body: &logging.LogDataBody{
			Type: "body-type",
			Content: &logging.LogDataBody_Text{
				Text: &logging.TextLog{
					Text: getNKData(n),
				},
			},
		},
	}
	logBytes, _ := proto.Marshal(log)
	return &v1.SniffData{
		Name:      "largeEvent",
		Timestamp: time.Now().Unix(),
		Meta: map[string]string{
			"meta": "largeEvent",
		},
		Type:   v1.SniffType_Logging,
		Remote: true,
		Data: &v1.SniffData_LogList{
			LogList: &v1.BatchLogList{Logs: [][]byte{logBytes}},
		},
	}
}

func TestQueue_Normal(t *testing.T) {
	q, err := initMmapQueue(plugin.Config{})
	defer cleanTestQueue(t, q)
	if err != nil {
		t.Fatalf("error in initializing the mmap queue: %v", err)
	}
	events := getBatchEvents(10)
	for _, e := range events {
		if err = q.Enqueue(e); err != nil {
			t.Errorf("queue cannot enqueue one event: %+v", err)
		}
	}
	for i := 0; i < 10; i++ {
		sequenceEvent, err := q.Dequeue()
		if err != nil {
			t.Errorf("error in fetching data from queue: %v", err)
		} else if !cmp.Equal(events[i].String(), sequenceEvent.Event.String()) {
			t.Errorf("history data and fetching data is not equal\n,history:%+v\n. dequeue data:%+v\n", events[i], sequenceEvent.Event)
		}
	}
}

func TestQueue_ReadHistory(t *testing.T) {
	cfg := plugin.Config{
		"segment_size": 10240,
	}

	q, err := initMmapQueue(cfg)
	defer cleanTestQueue(t, q)
	if err != nil {
		t.Fatalf("error in initializing the mmap queue: %v", err)
	}
	// close the queue to create a history empty queue.
	if err := q.Close(); err != nil {
		t.Fatalf("error in closing queue, %v", err)
	}

	// test cases.
	batchSize := 10
	batchNum := 100
	events := getBatchEvents(batchSize * batchNum)

	// Insert batchNum pieces of data in batchNum times
	for i := 0; i < batchSize; i++ {
		// recreate the queue
		q, err := initMmapQueue(cfg)
		if err != nil {
			t.Fatalf("error in initializing the mmap queue: %v", err)
		}
		for j := 0; j < batchNum; j++ {
			index := i*batchSize + j
			if err = q.Enqueue(events[index]); err != nil {
				t.Errorf("queue cannot enqueue one event: %+v", err)
			}
		}
		if err := q.Close(); err != nil {
			t.Fatalf("error in closing queue, %v", err)
		}
	}

	// Read batchNum pieces of data in batchNum times
	for i := 0; i < batchSize; i++ {
		// recreate the queue
		q, err := initMmapQueue(cfg)
		if err != nil {
			t.Fatalf("error in initializing the mmap queue: %v", err)
		}
		for j := 0; j < batchNum; j++ {
			index := i*batchSize + j
			sequenceEvent, err := q.Dequeue()
			if err != nil {
				t.Errorf("error in fetching data from queue: %v", err)
			} else if cmp.Equal(events[index].String(), sequenceEvent.Event.String()) {
				q.Ack(&sequenceEvent.Offset)
			} else {
				t.Errorf("history data and fetching data is not equal\n,history:%+v\n. dequeue data:%+v\n", events[index], sequenceEvent.Event)
			}
		}
		if err := q.Close(); err != nil {
			t.Fatalf("error in closing queue, %v", err)
		}
	}
}

func TestQueue_PushOverCeilingMsg(t *testing.T) {
	cfg := plugin.Config{
		"segment_size":   10240,
		"max_event_size": 1024 * 8,
	}
	largeEvent := getLargeEvent(20)
	q, err := initMmapQueue(cfg)
	if err != nil {
		t.Fatalf("cannot get a mmap queue: %v", err)
	}
	defer cleanTestQueue(t, q)
	err = q.Enqueue(largeEvent)
	if err == nil {
		t.Fatalf("The insertion of the over ceiling event is not as expected")
	} else {
		fmt.Printf("want err: %v", err)
	}
}

func TestQueue_FlushWhenReachNum(t *testing.T) {
	cfg := plugin.Config{
		"segment_size":      10240,
		"flush_ceiling_num": 5,
		"flush_period":      1000 * 60,
	}
	q, err := initMmapQueue(cfg)
	if err != nil {
		t.Fatalf("cannot get a mmap queue: %v", err)
	}
	defer cleanTestQueue(t, q)
	events := getBatchEvents(5)

	for _, e := range events {
		err = q.Enqueue(e)
		if err != nil {
			t.Errorf("queue cannot enqueue one event: %+v", err)
		}
	}
	time.Sleep(time.Second)
	wID, wOffset := q.meta.GetWritingOffset()
	wmID, wmOffset := q.meta.GetWatermarkOffset()
	if wID != wmID || wOffset != wmOffset {
		t.Fatalf("the flush operation was not invoking when reach the flush_ceiling_num.")
	}
}

func TestQueue_FlushPeriod(t *testing.T) {
	cfg := plugin.Config{
		"segment_size":      10240,
		"flush_ceiling_num": 50,
		"flush_period":      1000 * 1,
	}
	q, err := initMmapQueue(cfg)
	if err != nil {
		t.Fatalf("cannot get a mmap queue: %v", err)
	}
	defer cleanTestQueue(t, q)
	events := getBatchEvents(5)

	for _, e := range events {
		err = q.Enqueue(e)
		if err != nil {
			t.Errorf("queue cannot enqueue one event: %+v", err)
		}
	}
	time.Sleep(time.Second * 2)
	wID, wOffset := q.meta.GetWritingOffset()
	wmID, wmOffset := q.meta.GetWatermarkOffset()
	if wID != wmID || wOffset != wmOffset {
		t.Fatalf("the flush operation was not invoking when reach the flush_ceiling_num.")
	}
}

func TestQueue_MemCost(t *testing.T) {
	cfg := plugin.Config{
		"segment_size":        1024 * 4,
		"max_in_mem_segments": 8,
	}
	q, err := initMmapQueue(cfg)
	if err != nil {
		t.Fatalf("cannot get a mmap queue: %v", err)
	}
	defer cleanTestQueue(t, q)
	events := getBatchEvents(20)
	var memcost []int32
	for _, e := range events {
		err = q.Enqueue(e)
		memcost = append(memcost, q.mmapCount)
		if err != nil {
			t.Errorf("queue cannot enqueue one event: %+v", err)
		}
	}
	want := []int32{
		1, 2, 2, 3, 3, 4, 4, 5, 5, 6, 6, 7, 8, 5, 6, 6, 7, 7, 8, 5,
	}
	if !cmp.Equal(want, memcost) {
		t.Fatalf("the memory cost trends are not in line with expectations,\n want: %v,\n but got: %v", want, memcost)
	}
}

func TestQueue_OverSegmentEvent(t *testing.T) {
	cfg := plugin.Config{
		"segment_size": 1024 * 4,
	}
	q, err := initMmapQueue(cfg)
	if err != nil {
		t.Fatalf("cannot get a mmap queue: %v", err)
	}
	defer cleanTestQueue(t, q)
	size := 10
	wantPos := size * 1024 / q.SegmentSize
	largeEvent := getLargeEvent(size)
	err = q.Enqueue(largeEvent)
	if err != nil {
		t.Errorf("queue cannot enqueue one event: %+v", err)
	}
	id, _ := q.meta.GetWritingOffset()
	if int(id) != wantPos {
		t.Fatalf("the writing offset id should at %d segment, but at %d", id, wantPos)
	}
}

func TestQueue_ReusingFiles(t *testing.T) {
	cfg := plugin.Config{
		"segment_size":            1024 * 4,
		"queue_capacity_segments": 5,
		"max_event_size":          1024 * 3,
	}
	q, err := initMmapQueue(cfg)
	if err != nil {
		t.Fatalf("cannot get a mmap queue: %v", err)
	}
	defer cleanTestQueue(t, q)

	for i := 0; i < 100; i++ {
		err = q.Enqueue(getLargeEvent(2))
		if err != nil {
			t.Errorf("queue cannot enqueue one event: %+v", err)
		}
		_, err := q.Dequeue()
		if err != nil {
			t.Errorf("error in fetching data from queue: %v", err)
		}
	}
	rid, roffset := q.meta.GetReadingOffset()
	wid, woffset := q.meta.GetWritingOffset()
	fmt.Printf("rid:%d, roffset:%d, wid:%d, woffset:%d\n", rid, roffset, wid, woffset)
	if int(wid) <= q.QueueCapacitySegments || int(rid) <= q.QueueCapacitySegments {
		t.Errorf("cannot valid reusing files")
	}
}

func TestQueue_Empty(t *testing.T) {
	cfg := plugin.Config{
		"segment_size":            1024 * 4,
		"queue_capacity_segments": 10,
	}
	q, err := initMmapQueue(cfg)
	if err != nil {
		t.Fatalf("cannot get a mmap queue: %v", err)
	}
	defer cleanTestQueue(t, q)
	for _, e := range getBatchEvents(3) {
		err = q.Enqueue(e)
		if err != nil {
			t.Errorf("queue cannot enqueue one event: %+v", err)
		}
	}
	for i := 0; i < 3; i++ {
		if _, err = q.Dequeue(); err != nil {
			t.Errorf("error in fetching data from queue: %v", err)
		}
	}
	_, err = q.Dequeue()
	if err != nil && err.Error() != "cannot read data when the queue is empty" {
		t.Fatalf("not except err: %v", err)
	}
}

func TestQueue_Full(t *testing.T) {
	cfg := plugin.Config{
		"segment_size":            1024 * 4,
		"queue_capacity_segments": 10,
	}
	q, err := initMmapQueue(cfg)
	if err != nil {
		t.Fatalf("cannot get a mmap queue: %v", err)
	}
	defer cleanTestQueue(t, q)
	for _, e := range getBatchEvents(8) {
		err = q.Enqueue(e)
		if err != nil {
			t.Errorf("queue cannot enqueue one event: %+v", err)
		}
	}
	err = q.Enqueue(getLargeEvent(2))
	if err != nil && err.Error() != "cannot write data when the queue is full" {
		t.Fatalf("not except err: %v", err)
	}
}

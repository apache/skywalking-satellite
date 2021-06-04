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
package prometheus

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/storage"
	v1 "skywalking.apache.org/repo/goapi/satellite/data/v1"
)

// QueueAppender appender with queue
type QueueAppender struct {
	Ctx                context.Context
	Ms                 *metadataService
	isNew              bool
	job                string
	instance           string
	metricBuilder      *metricBuilder
	useStartTimeMetric bool
	OutputChannel      chan *v1.SniffData
}

// NewQueueAppender construct QueueAppender
func NewQueueAppender(ctx context.Context, ms *metadataService, oc chan *v1.SniffData, useStartTimeMetric bool) *QueueAppender {
	return &QueueAppender{Ctx: ctx, Ms: ms, OutputChannel: oc, isNew: true, useStartTimeMetric: useStartTimeMetric}
}

func (qa *QueueAppender) initAppender(ls labels.Labels) error {
	job, instance := ls.Get(model.JobLabel), ls.Get(model.InstanceLabel)
	if job == "" || instance == "" {
		// errNoJobInstance
		return fmt.Errorf("errNoJobInstance")
	}
	// discover the binding target when this method is called for the first time during a transaction
	mc, err := qa.Ms.Get(job, instance)
	if err != nil {
		return err
	}
	qa.job = job
	qa.instance = instance
	qa.metricBuilder = newMetricBuilder(mc, qa.useStartTimeMetric)
	qa.isNew = false
	return nil
}

var _ storage.Appender = (*QueueAppender)(nil)

// always returns 0 to disable label caching
func (qa *QueueAppender) Add(ls labels.Labels, t int64, v float64) (uint64, error) {
	if math.IsNaN(v) {
		return 0, nil
	}
	select {
	case <-qa.Ctx.Done():
		return 0, fmt.Errorf("errTransactionAborted")
	default:
	}
	if qa.isNew {
		if err := qa.initAppender(ls); err != nil {
			return 0, err
		}
	}

	return 0, qa.metricBuilder.AddDataPoint(ls, t, v)
}

// always returns error since we do not cache
func (qa *QueueAppender) AddFast(_ uint64, _ int64, _ float64) error {
	return storage.ErrNotFound
}

// submit metrics data to consumers
func (qa *QueueAppender) Commit() error {
	// 1. convert to meter
	meterCollection, _, _ := qa.metricBuilder.Build()
	// 2. send metrics to queue
	for _, meterData := range meterCollection.GetMeterData() {
		meterData.Service = qa.job
		meterData.ServiceInstance = qa.instance
		e := &v1.SniffData{
			Name:      eventName,
			Timestamp: time.Now().UnixNano() / 1e6,
			Meta:      nil,
			Type:      v1.SniffType_MeterType,
			Remote:    true,
			Data: &v1.SniffData_Meter{
				Meter: meterData,
			},
		}
		qa.OutputChannel <- e
	}
	return nil
}

func (qa *QueueAppender) Rollback() error {
	return nil
}

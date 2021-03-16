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
//
// Refers to https://github.com/open-telemetry/opentelemetry-collector [Apache-2.0 License]

package prometheus

import (
	"context"
	"errors"
	"sync/atomic"

	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/storage"

	"github.com/prometheus/prometheus/scrape"
)

const (
	runningStateInit = iota
	runningStateReady
	runningStateStop
)

var noop = &noopAppender{}

type QueueStore struct {
	ctx                  context.Context
	running              int32
	mc                   *metadataService
	useStartTimeMetric   bool
	startTimeMetricRegex string
	receiverName         string
}

// NewQueueStore construct QueueStore
func NewQueueStore(ctx context.Context, useStartTimeMetric bool, startTimeMetricRegex string, receiverName string) *QueueStore {
	return &QueueStore{
		ctx:                  ctx,
		running:              runningStateInit,
		useStartTimeMetric:   useStartTimeMetric,
		startTimeMetricRegex: startTimeMetricRegex,
		receiverName:         receiverName,
	}
}

func (qs *QueueStore) SetScrapeManager(scrapeManager *scrape.Manager) {
	if scrapeManager != nil && atomic.CompareAndSwapInt32(&qs.running, runningStateInit, runningStateReady) {
		qs.mc = &metadataService{sm: scrapeManager}
	}
}

func (qs *QueueStore) Appender() (storage.Appender, error) {
	state := atomic.LoadInt32(&qs.running)
	if state == runningStateReady {
		return NewQueueAppender(qs.ctx, qs.useStartTimeMetric, qs.startTimeMetricRegex, qs.receiverName, qs.mc), nil
	} else if state == runningStateInit {
		panic("ScrapeManager is not set")
	}
	// instead of returning an error, return a dummy appender instead, otherwise it can trigger panic
	return noop, errors.New("noop appender")
}

func (qs *QueueStore) Close() error {
	atomic.CompareAndSwapInt32(&qs.running, runningStateReady, runningStateStop)
	return nil
}

// noopAppender, always return error on any operations
type noopAppender struct{}

func (*noopAppender) Add(labels.Labels, int64, float64) (uint64, error) {
	return 0, errors.New("already stopped")
}

func (*noopAppender) AddFast(labels.Labels, uint64, int64, float64) error {
	return errors.New("already stopped")
}

func (*noopAppender) Commit() error {
	return errors.New("already stopped")
}

func (*noopAppender) Rollback() error {
	return nil
}

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
	"errors"

	"github.com/apache/skywalking-satellite/protocol/gen-codes/satellite/protocol"

	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/scrape"
	"github.com/prometheus/prometheus/storage"
)

var noop = &noopAppender{}

type QueueStore struct {
	ctx                context.Context
	mc                 *metadataService
	useStartTimeMetric bool
	receiverName       string
	OutputChannel      chan *protocol.Event
}

// NewQueueStore construct QueueStore
func NewQueueStore(ctx context.Context, useStartTimeMetric bool, receiverName string, oc chan *protocol.Event) *QueueStore {
	return &QueueStore{
		ctx:                ctx,
		useStartTimeMetric: useStartTimeMetric,
		receiverName:       receiverName,
		OutputChannel:      oc,
	}
}

func (qs *QueueStore) SetScrapeManager(scrapeManager *scrape.Manager) {
	if scrapeManager != nil {
		qs.mc = &metadataService{sm: scrapeManager}
	}
}

func (qs *QueueStore) Appender() (storage.Appender, error) {
	return NewQueueAppender(qs.ctx, qs.mc, qs.OutputChannel), nil
}

func (qs *QueueStore) Close() error {
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

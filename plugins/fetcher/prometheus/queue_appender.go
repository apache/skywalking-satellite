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

	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/storage"
)

// QueueAppender todo appender with queue
type QueueAppender struct {
	Ctx context.Context
	Ms  *metadataService
}

// NewQueueAppender construct QueueAppender
func NewQueueAppender(ctx context.Context, ms *metadataService) *QueueAppender {
	return &QueueAppender{Ctx: ctx, Ms: ms}
}

var _ storage.Appender = (*QueueAppender)(nil)

// always returns 0 to disable label caching
func (qa *QueueAppender) Add(ls labels.Labels, t int64, v float64) (uint64, error) {
	// todo add metrics
	return 0, nil
}

// always returns error since we do not cache
func (qa *QueueAppender) AddFast(_ labels.Labels, _ uint64, _ int64, _ float64) error {
	return storage.ErrNotFound
}

// submit metrics data to consumers
func (qa *QueueAppender) Commit() error {
	// todo send metrics to queue
	return nil
}

func (qa *QueueAppender) Rollback() error {
	return nil
}

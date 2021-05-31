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

	"github.com/apache/skywalking-satellite/protocol/gen-codes/satellite/protocol"

	"github.com/prometheus/prometheus/scrape"
	"github.com/prometheus/prometheus/storage"
)

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

func (qs *QueueStore) Appender(ctx context.Context) storage.Appender {
	return NewQueueAppender(ctx, qs.mc, qs.OutputChannel)
}

func (qs *QueueStore) Close() error {
	return nil
}

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

package gatherer

import (
	"github.com/apache/skywalking-satellite/internal/pkg/log"
	"github.com/apache/skywalking-satellite/internal/satellite/module/gatherer/api"
	"github.com/apache/skywalking-satellite/internal/satellite/sharing"
	fetcher "github.com/apache/skywalking-satellite/plugins/fetcher/api"
	queue "github.com/apache/skywalking-satellite/plugins/queue/api"
	receiver "github.com/apache/skywalking-satellite/plugins/receiver/api"
	server "github.com/apache/skywalking-satellite/plugins/server/api"
)

// NewGatherer returns a gatherer module
func NewGatherer(cfg *api.GathererConfig) api.Gatherer {
	if cfg.ReceiverConfig != nil {
		return newReceiverGatherer(cfg)
	} else if cfg.FetcherConfig != nil {
		return newFetcherGatherer(cfg)
	}
	return nil
}

// newFetcherGatherer crates a gatherer with the fetcher role.
func newFetcherGatherer(cfg *api.GathererConfig) *FetcherGatherer {
	log.Logger.Infof("fetcher gatherer module of %s namespace is being initialized", cfg.PipeName)
	return &FetcherGatherer{
		config:         cfg,
		runningQueue:   queue.GetQueue(cfg.QueueConfig),
		runningFetcher: fetcher.GetFetcher(cfg.FetcherConfig),
		outputChannel:  make(chan *queue.SequenceEvent),
	}
}

// newReceiverGatherer crates a gatherer with the receiver role.
func newReceiverGatherer(cfg *api.GathererConfig) *ReceiverGatherer {
	log.Logger.Infof("receiver gatherer module of %s namespace is being initialized", cfg.PipeName)
	return &ReceiverGatherer{
		config:          cfg,
		runningQueue:    queue.GetQueue(cfg.QueueConfig),
		runningReceiver: receiver.GetReceiver(cfg.ReceiverConfig),
		runningServer:   sharing.Manager[cfg.ServerName].(server.Server),
		outputChannel:   make(chan *queue.SequenceEvent),
	}
}

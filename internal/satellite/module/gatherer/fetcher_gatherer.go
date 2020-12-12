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
	"context"
	"sync"
	"time"

	"github.com/apache/skywalking-satellite/internal/pkg/event"
	"github.com/apache/skywalking-satellite/internal/pkg/log"
	"github.com/apache/skywalking-satellite/internal/satellite/module/gatherer/api"
	fetcher "github.com/apache/skywalking-satellite/plugins/fetcher/api"
	queue "github.com/apache/skywalking-satellite/plugins/queue/api"
)

type FetcherGatherer struct {
	// config
	config *api.GathererConfig

	// dependency plugins
	runningFetcher fetcher.Fetcher
	runningQueue   queue.Queue

	// self components
	outputChannel chan *queue.SequenceEvent
}

func (f *FetcherGatherer) Prepare() error {
	return nil
}

func (f *FetcherGatherer) Boot(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		timeTicker := time.NewTicker(time.Duration(f.config.FetchInterval) * time.Millisecond)
		for {
			select {
			case <-timeTicker.C:
				events := f.runningFetcher.Fetch()
				for _, event := range events {
					err := f.runningQueue.Push(event)
					if err != nil {
						// todo add abandonedCount metrics
						log.Logger.Errorf("cannot put event into queue in %s namespace, %v", f.config.RunningNamespace, err)
					}
				}
			case e := <-f.runningQueue.Pop():
				f.outputChannel <- e
			case <-ctx.Done():
				f.Shutdown()
				return
			}
		}
	}()
	wg.Wait()
}

func (f *FetcherGatherer) Shutdown() {
	log.Logger.Infof("fetcher gatherer module of %s namespace is closing", f.config.RunningNamespace)
	if err := f.runningQueue.Close(); err != nil {
		log.Logger.Errorf("failure occurs when closing %s queue  in %s namespace :%v", f.runningQueue.Name(), f.config.RunningNamespace, err)
	}
}

func (f *FetcherGatherer) OutputDataChannel() <-chan *queue.SequenceEvent {
	return f.outputChannel
}

func (f *FetcherGatherer) Ack(lastOffset event.Offset) {
	f.runningQueue.Ack(lastOffset)
}

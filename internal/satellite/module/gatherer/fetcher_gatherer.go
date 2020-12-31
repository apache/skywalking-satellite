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

	"github.com/prometheus/client_golang/prometheus"

	"github.com/apache/skywalking-satellite/internal/pkg/log"
	"github.com/apache/skywalking-satellite/internal/satellite/event"
	"github.com/apache/skywalking-satellite/internal/satellite/module/gatherer/api"
	"github.com/apache/skywalking-satellite/internal/satellite/telemetry"
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

	// metrics
	fetchCounter       *prometheus.CounterVec
	queueOutputCounter *prometheus.CounterVec
}

func (f *FetcherGatherer) Prepare() error {
	f.fetchCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "gatherer_fetch_count",
		Help: "Total number of the receiving count in the Gatherer.",
	}, []string{"pipe", "status"})
	f.queueOutputCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Subsystem: "fetcher",
		Name:      "queue_output_count",
		Help:      "Total number of the output count in the Queue of Gatherer.",
	}, []string{"pipe", "status"})
	telemetry.Registerer.MustRegister(f.fetchCounter)
	telemetry.Registerer.MustRegister(f.queueOutputCounter)
	return nil
}

func (f *FetcherGatherer) Boot(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		childCtx, cancel := context.WithCancel(ctx)
		timeTicker := time.NewTicker(time.Duration(f.config.FetchInterval) * time.Millisecond)
		for {
			select {
			case <-timeTicker.C:
				events := f.runningFetcher.Fetch()
				for _, e := range events {
					err := f.runningQueue.Push(e)
					f.fetchCounter.WithLabelValues(f.config.PipeName, "all").Inc()
					if err != nil {
						f.fetchCounter.WithLabelValues(f.config.PipeName, "abandoned").Inc()
						log.Logger.Errorf("cannot put event into queue in %s namespace, %v", f.config.PipeName, err)
					}
				}
			case <-childCtx.Done():
				cancel()
				f.Shutdown()
				return
			}
		}
	}()

	go func() {
		defer wg.Done()
		childCtx, cancel := context.WithCancel(ctx)
		for {
			select {
			case <-childCtx.Done():
				cancel()
				f.Shutdown()
				return
			default:
				if e, err := f.runningQueue.Pop(); err == nil {
					f.outputChannel <- e
					f.queueOutputCounter.WithLabelValues(f.config.PipeName, "success").Inc()
				} else if err == queue.ErrEmpty {
					time.Sleep(time.Second)
				} else {
					f.queueOutputCounter.WithLabelValues(f.config.PipeName, "error").Inc()
					log.Logger.Errorf("error in popping from the queue: %v", err)
				}
			}
		}
	}()
	wg.Wait()
}

func (f *FetcherGatherer) Shutdown() {
	log.Logger.Infof("fetcher gatherer module of %s namespace is closing", f.config.PipeName)
	if err := f.runningQueue.Close(); err != nil {
		log.Logger.Errorf("failure occurs when closing %s queue  in %s namespace :%v", f.runningQueue.Name(), f.config.PipeName, err)
	}
}

func (f *FetcherGatherer) OutputDataChannel() <-chan *queue.SequenceEvent {
	return f.outputChannel
}

func (f *FetcherGatherer) Ack(lastOffset event.Offset) {
	f.runningQueue.Ack(lastOffset)
}

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
	queue "github.com/apache/skywalking-satellite/plugins/queue/api"
	receiver "github.com/apache/skywalking-satellite/plugins/receiver/api"
	server "github.com/apache/skywalking-satellite/plugins/server/api"
)

type ReceiverGatherer struct {
	// config
	config *api.GathererConfig

	// dependency plugins
	runningReceiver receiver.Receiver
	runningQueue    queue.Queue
	runningServer   server.Server

	// self components
	outputChannel chan *queue.SequenceEvent
	// metrics
	receiveCounter     *prometheus.CounterVec
	queueOutputCounter *prometheus.CounterVec
}

func (r *ReceiverGatherer) Prepare() error {
	log.Logger.Infof("receiver gatherer module of %s namespace is preparing", r.config.PipeName)
	r.runningReceiver.RegisterHandler(r.runningServer.GetServer())
	if err := r.runningQueue.Initialize(); err != nil {
		log.Logger.Infof("the %s queue of %s namespace was failed to initialize", r.runningQueue.Name(), r.config.PipeName)
		return err
	}
	r.receiveCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "gatherer_receive_count",
		Help: "Total number of the receiving count in the Gatherer.",
	}, []string{"pipe", "status"})
	r.queueOutputCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Subsystem: "receiver",
		Name:      "queue_output_count",
		Help:      "Total number of the output count in the Queue of Gatherer.",
	}, []string{"pipe", "status"})
	telemetry.Registerer.MustRegister(r.receiveCounter)
	log.Logger.Infof("register receiveCounter")
	telemetry.Registerer.MustRegister(r.queueOutputCounter)
	log.Logger.Infof("register queueOutputCounter")
	return nil
}

func (r *ReceiverGatherer) Boot(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		childCtx, cancel := context.WithCancel(ctx)
		defer wg.Done()
		for {
			select {
			case e := <-r.runningReceiver.Channel():
				r.receiveCounter.WithLabelValues(r.config.PipeName, "all").Inc()
				err := r.runningQueue.Push(e)
				if err != nil {
					r.receiveCounter.WithLabelValues(r.config.PipeName, "abandoned").Inc()
					log.Logger.Errorf("cannot put event into queue in %s namespace, error is: %v", r.config.PipeName, err)
				}
			case <-childCtx.Done():
				cancel()
				r.Shutdown()
				return
			}
		}
	}()

	go func() {
		childCtx, cancel := context.WithCancel(ctx)
		defer wg.Done()
		for {
			select {
			case <-childCtx.Done():
				cancel()
				r.Shutdown()
				return
			default:
				if e, err := r.runningQueue.Pop(); err == nil {
					r.outputChannel <- e
					r.queueOutputCounter.WithLabelValues(r.config.PipeName, "success").Inc()
				} else if err == queue.ErrEmpty {
					time.Sleep(time.Second)
				} else {
					r.queueOutputCounter.WithLabelValues(r.config.PipeName, "error").Inc()
					log.Logger.Errorf("error in popping from the queue: %v", err)
				}
			}
		}
	}()
	wg.Wait()
}

func (r *ReceiverGatherer) Shutdown() {
	log.Logger.Infof("receiver gatherer module of %s namespace is closing", r.config.PipeName)
	if err := r.runningQueue.Close(); err != nil {
		log.Logger.Errorf("failure occurs when closing %s queue  in %s namespace, error is: %v", r.runningQueue.Name(), r.config.PipeName, err)
	}
}

func (r *ReceiverGatherer) OutputDataChannel() <-chan *queue.SequenceEvent {
	return r.outputChannel
}

func (r *ReceiverGatherer) Ack(lastOffset event.Offset) {
	r.runningQueue.Ack(lastOffset)
}

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
	"errors"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/apache/skywalking-satellite/internal/pkg/log"
	"github.com/apache/skywalking-satellite/internal/satellite/event"
	module "github.com/apache/skywalking-satellite/internal/satellite/module/api"
	"github.com/apache/skywalking-satellite/internal/satellite/module/gatherer/api"
	processor "github.com/apache/skywalking-satellite/internal/satellite/module/processor/api"
	"github.com/apache/skywalking-satellite/internal/satellite/telemetry"
	queue "github.com/apache/skywalking-satellite/plugins/queue/api"
	"github.com/apache/skywalking-satellite/plugins/queue/partition"
	receiver "github.com/apache/skywalking-satellite/plugins/receiver/api"
	server "github.com/apache/skywalking-satellite/plugins/server/api"

	v1 "skywalking.apache.org/repo/goapi/satellite/data/v1"
)

type ReceiverGatherer struct {
	// config
	config *api.GathererConfig

	// dependency plugins
	runningReceiver receiver.Receiver
	runningQueue    *partition.PartitionedQueue
	runningServer   server.Server

	// self components
	outputChannel []chan *queue.SequenceEvent
	// metrics
	receiveCounter     *telemetry.Counter
	queueOutputCounter *telemetry.Counter

	// sync invoker
	processor processor.Processor
}

func (r *ReceiverGatherer) Prepare() error {
	log.Logger.WithField("pipe", r.config.PipeName).Info("receiver gatherer module is preparing...")
	r.runningReceiver.RegisterHandler(r.runningServer.GetServer())
	if err := r.runningQueue.Initialize(); err != nil {
		log.Logger.WithField("pipe", r.config.PipeName).Infof("the %s queue failed when initializing", r.runningQueue.Name())
		return err
	}
	r.outputChannel = make([]chan *queue.SequenceEvent, r.runningQueue.TotalPartitionCount())
	for p := 0; p < r.runningQueue.TotalPartitionCount(); p++ {
		r.outputChannel[p] = make(chan *queue.SequenceEvent)
	}
	r.receiveCounter = telemetry.NewCounter("gatherer_receive_count", "Total number of the receiving count in the Gatherer.", "pipe", "status")
	r.queueOutputCounter = telemetry.NewCounter("queue_output_count", "Total number of the output count in the Queue of Gatherer.", "pipe", "status")
	return nil
}

func (r *ReceiverGatherer) Boot(ctx context.Context) {
	r.runningReceiver.RegisterSyncInvoker(r)
	var wg sync.WaitGroup
	wg.Add(r.PartitionCount() + 1)
	log.Logger.WithField("pipe", r.config.PipeName).Info("receive_gatherer module is starting...")
	go func() {
		childCtx, cancel := context.WithCancel(ctx)
		defer wg.Done()
		for {
			select {
			case e := <-r.runningReceiver.Channel():
				r.receiveCounter.Inc(r.config.PipeName, "all")
				err := r.runningQueue.Enqueue(e)
				if err != nil {
					r.receiveCounter.Inc(r.config.PipeName, "abandoned")
					log.Logger.WithFields(logrus.Fields{
						"pipe":  r.config.PipeName,
						"queue": r.runningQueue.Name(),
					}).Errorf("error in enqueue: %v", err)
				}
			case <-childCtx.Done():
				cancel()
				return
			}
		}
	}()

	for p := 0; p < r.PartitionCount(); p++ {
		r.consumeQueue(ctx, p, &wg)
	}
	wg.Wait()
}

func (r *ReceiverGatherer) consumeQueue(ctx context.Context, p int, wg *sync.WaitGroup) {
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
				if e, err := r.runningQueue.Dequeue(p); err == nil {
					r.outputChannel[p] <- e
					r.queueOutputCounter.Inc(r.config.PipeName, "success")
				} else if err == queue.ErrEmpty {
					time.Sleep(time.Second)
				} else {
					r.queueOutputCounter.Inc(r.config.PipeName, "error")
					log.Logger.WithFields(logrus.Fields{
						"pipe":  r.config.PipeName,
						"queue": r.runningQueue.Name(),
					}).Errorf("error in dequeue: %v", err)
				}
			}
		}
	}()
}

func (r *ReceiverGatherer) Shutdown() {
	log.Logger.WithField("pipe", r.config.PipeName).Infof("receiver gatherer module is closing...")
	time.Sleep(module.ShutdownHookTime)
	if err := r.runningQueue.Close(); err != nil {
		log.Logger.WithFields(logrus.Fields{
			"pipe":  r.config.PipeName,
			"queue": r.runningQueue.Name(),
		}).Errorf("error in closing: %v", err)
	}
}

func (r *ReceiverGatherer) PartitionCount() int {
	return len(r.outputChannel)
}

func (r *ReceiverGatherer) OutputDataChannel(index int) <-chan *queue.SequenceEvent {
	return r.outputChannel[index]
}

func (r *ReceiverGatherer) Ack(lastOffset *event.Offset) {
	r.runningQueue.Ack(lastOffset)
}

func (r *ReceiverGatherer) SyncInvoke(d *v1.SniffData) (*v1.SniffData, error) {
	return r.processor.SyncInvoke(d)
}

func (r *ReceiverGatherer) SetProcessor(m module.Module) error {
	if p, ok := m.(processor.Processor); ok {
		r.processor = p
		return nil
	}

	return errors.New("set processor only supports to inject processor module")
}

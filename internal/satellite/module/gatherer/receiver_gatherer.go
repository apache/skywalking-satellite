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

	"github.com/apache/skywalking-satellite/internal/pkg/event"
	"github.com/apache/skywalking-satellite/internal/pkg/log"
	"github.com/apache/skywalking-satellite/internal/satellite/module/gatherer/api"
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
}

func (r *ReceiverGatherer) Prepare() error {
	log.Logger.Infof("receiver gatherer module of %s namespace is preparing", r.config.RunningNamespace)
	r.runningReceiver.RegisterHandler(r.runningServer)
	if err := r.runningQueue.Prepare(); err != nil {
		log.Logger.Infof("the %s queue of %s namespace was failed to initialize", r.runningQueue.Name(), r.config.RunningNamespace)
		return err
	}
	return nil
}

func (r *ReceiverGatherer) Boot(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case e := <-r.runningReceiver.Channel():
				err := r.runningQueue.Push(e)
				if err != nil {
					// todo add abandonedCount metrics
					log.Logger.Errorf("cannot put event into queue in %s namespace, error is: %v", r.config.RunningNamespace, err)
				}
			case e := <-r.runningQueue.Pop():
				r.outputChannel <- e
			case <-ctx.Done():
				r.Shutdown()
				return
			}
		}
	}()
	wg.Wait()
}

func (r *ReceiverGatherer) Shutdown() {
	log.Logger.Infof("receiver gatherer module of %s namespace is closing", r.config.RunningNamespace)
	if err := r.runningQueue.Close(); err != nil {
		log.Logger.Errorf("failure occurs when closing %s queue  in %s namespace, error is: %v", r.runningQueue.Name(), r.config.RunningNamespace, err)
	}
}

func (r *ReceiverGatherer) OutputDataChannel() <-chan *queue.SequenceEvent {
	return r.outputChannel
}

func (r *ReceiverGatherer) Ack(lastOffset event.Offset) {
	r.runningQueue.Ack(lastOffset)
}

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

package log

import (
	"fmt"
	"reflect"

	"google.golang.org/protobuf/proto"

	"github.com/Shopify/sarama"

	"github.com/apache/skywalking-satellite/internal/pkg/config"
	"github.com/apache/skywalking-satellite/internal/pkg/log"
	"github.com/apache/skywalking-satellite/internal/satellite/event"
	"github.com/apache/skywalking-satellite/protocol/gen-codes/satellite/protocol"
)

const Name = "log-kafka-forwarder"

type Forwarder struct {
	config.CommonFields
	Topic    string `mapstructure:"topic"` // The forwarder topic.
	producer sarama.SyncProducer
}

func (f *Forwarder) Name() string {
	return Name
}

func (f *Forwarder) Description() string {
	return "this is a synchronization Kafka log forwarder."
}

func (f *Forwarder) DefaultConfig() string {
	return `
# The remote topic. 
topic: "log-topic"
`
}

func (f *Forwarder) Prepare(connection interface{}) error {
	client, ok := connection.(sarama.Client)
	if !ok {
		return fmt.Errorf("the %s is only accepet the kafka client, but receive a %s",
			f.Name(), reflect.TypeOf(connection).String())
	}
	producer, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		return err
	}
	f.producer = producer
	return nil
}

func (f *Forwarder) Forward(batch event.BatchEvents) error {
	var message []*sarama.ProducerMessage
	for _, e := range batch {
		data, ok := e.GetData().(*protocol.Event_Log)
		if !ok {
			continue
		}
		bytes, err := proto.Marshal(data.Log)
		if err != nil {
			log.Logger.Errorf("%s serialize the logData fail: %v", f.Name(), err)
			continue
		}
		message = append(message, &sarama.ProducerMessage{
			Topic: f.Topic,
			Value: sarama.ByteEncoder(bytes),
		})
	}
	return f.producer.SendMessages(message)
}

func (f *Forwarder) ForwardType() protocol.EventType {
	return protocol.EventType_Logging
}

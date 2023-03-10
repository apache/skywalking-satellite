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

package nativemanagement

import (
	"fmt"
	"reflect"

	"github.com/Shopify/sarama"

	"github.com/apache/skywalking-satellite/internal/pkg/config"
	"github.com/apache/skywalking-satellite/internal/satellite/event"
	"google.golang.org/protobuf/proto"

	v1 "skywalking.apache.org/repo/goapi/satellite/data/v1"
)

const (
	Name     = "native-management-kafka-forwarder"
	ShowName = "Native Management Kafka Forwarder"
)

type void struct{}

var empty void

// empty struct for set

type Forwarder struct {
	config.CommonFields
	Topic    string `mapstructure:"topic"` // The forwarder topic.
	producer sarama.SyncProducer
	// managementClient management.ManagementServiceClient
}

func (f *Forwarder) Name() string {
	return Name
}

func (f *Forwarder) ShowName() string {
	return ShowName
}

func (f *Forwarder) Description() string {
	return "This is a synchronization Kafka forwarder with the SkyWalking native management protocol."
}

func (f *Forwarder) DefaultConfig() string {
	return `
# The remote topic. 
topic: "skywalking-managements"
`
}

func (f *Forwarder) Prepare(connection interface{}) error {
	client, ok := connection.(sarama.Client)
	if !ok {
		return fmt.Errorf("the %s is only accept the kafka client, but receive a %s",
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
	pingOnce := make(map[string]void)
	var message []*sarama.ProducerMessage
	for _, e := range batch {

		instanceProperties := e.GetInstance()
		if instanceProperties != nil {
			rawdata, err := proto.Marshal(instanceProperties)
			if err != nil {
				return err
			}
			message = append(message, &sarama.ProducerMessage{
				Topic: f.Topic,
				Key:   sarama.StringEncoder("register-" + instanceProperties.ServiceInstance),
				Value: sarama.ByteEncoder(rawdata),
			})
			continue
		}

		// report instance ping
		instancePing := e.GetInstancePing()
		if instancePing != nil {
			// report once
			instancePingStr := fmt.Sprintf("%s_%s", instancePing.Service, instancePing.ServiceInstance)
			_, exists := pingOnce[instancePingStr]
			if !exists {
				rawdata, err := proto.Marshal(instancePing)
				if err != nil {
					return err
				}
				pingOnce[instancePingStr] = empty

				message = append(message, &sarama.ProducerMessage{
					Topic: f.Topic,
					Key:   sarama.StringEncoder("register-" + instancePing.ServiceInstance),
					Value: sarama.ByteEncoder(rawdata),
				})
			}
			continue
		}

	}
	return f.producer.SendMessages(message)
}

func (f *Forwarder) ForwardType() v1.SniffType {
	return v1.SniffType_ManagementType
}

func (f *Forwarder) SyncForward(_ *v1.SniffData) (*v1.SniffData, error) {
	return nil, fmt.Errorf("unsupport sync forward")
}

func (f *Forwarder) SupportedSyncInvoke() bool {
	return false
}

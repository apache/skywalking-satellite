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

package nativeprofile

//
//import (
//	"fmt"
//	"github.com/Shopify/sarama"
//	"google.golang.org/protobuf/proto"
//	"reflect"
//
//	profile "skywalking.apache.org/repo/goapi/collect/language/profile/v3"
//
//	"github.com/apache/skywalking-satellite/internal/pkg/config"
//	"github.com/apache/skywalking-satellite/internal/satellite/event"
//	v1 "skywalking.apache.org/repo/goapi/satellite/data/v1"
//)
//
//const (
//	Name     = "native-profile-kafka-forwarder"
//	ShowName = "Native Profile Kafka Forwarder"
//)
//
//type Forwarder struct {
//	config.CommonFields
//	Topic         string `mapstructure:"topic"` // The forwarder topic.
//	producer      sarama.SyncProducer
//	profileClient profile.ProfileTaskClient
//}
//
//func (f *Forwarder) Name() string {
//	return Name
//}
//
//func (f *Forwarder) ShowName() string {
//	return ShowName
//}
//
//func (f *Forwarder) Description() string {
//	return "This is a synchronization Kafka forwarder with the SkyWalking native kafka protocol."
//}
//
//func (f *Forwarder) DefaultConfig() string {
//	return `
//# The remote topic.
//topic: "skywalking-profilings"
//`
//}
//
//func (f *Forwarder) Prepare(connection interface{}) error {
//	client, ok := connection.(sarama.Client)
//	if !ok {
//		return fmt.Errorf("the %s is only accepts the kafka client, but receive a %s",
//			f.Name(), reflect.TypeOf(connection).String())
//	}
//	producer, err := sarama.NewSyncProducerFromClient(client)
//	if err != nil {
//		return err
//	}
//	f.producer = producer
//	return nil
//}
//
//func (f *Forwarder) Forward(batch event.BatchEvents) error {
//	var message []*sarama.ProducerMessage
//	for _, e := range batch {
//		data := e.GetData().(*v1.SniffData_Profile)
//		rawdata, ok := proto.Marshal(data.Profile)
//
//		if ok != nil {
//			return ok
//		}
//		//producer.send(new ProducerRecord<>(
//		//                topic,
//		//                object.getTaskId() + object.getSequence(),
//		//                Bytes.wrap(object.toByteArray())
//		//            ));
//		message = append(message, &sarama.ProducerMessage{
//			Topic: f.Topic,
//			Key:   sarama.StringEncoder(data.Profile.GetTaskId() + fmt.Sprint(data.Profile.GetSequence())),
//			Value: sarama.ByteEncoder(rawdata),
//		})
//	}
//	return f.producer.SendMessages(message)
//}
//
//func (f *Forwarder) ForwardType() v1.SniffType {
//	return v1.SniffType_ProfileType
//}
//
//func (f *Forwarder) SyncForward(e *v1.SniffData) (*v1.SniffData, error) {
//	return nil, fmt.Errorf("unsupport sync forward")
//}
//
//func (f *Forwarder) SupportedSyncInvoke() bool {
//	return false
//}

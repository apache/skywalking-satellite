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
	"context"
	"fmt"
	"reflect"

	"google.golang.org/grpc"

	"github.com/apache/skywalking-satellite/internal/pkg/config"
	"github.com/apache/skywalking-satellite/internal/satellite/event"

	management "skywalking.apache.org/repo/goapi/collect/management/v3"
	v1 "skywalking.apache.org/repo/goapi/satellite/data/v1"
)

const (
	Name     = "native-management-grpc-forwarder"
	ShowName = "Native Management GRPC Forwarder"
)

// empty struct for set
type void struct{}

var empty void

type Forwarder struct {
	config.CommonFields

	managementClient management.ManagementServiceClient
}

func (f *Forwarder) Name() string {
	return Name
}

func (f *Forwarder) ShowName() string {
	return ShowName
}

func (f *Forwarder) Description() string {
	return "This is a synchronization grpc forwarder with the SkyWalking native management protocol."
}

func (f *Forwarder) DefaultConfig() string {
	return ``
}

func (f *Forwarder) Prepare(connection interface{}) error {
	client, ok := connection.(*grpc.ClientConn)
	if !ok {
		return fmt.Errorf("the %s only accepts a grpc client, but received a %s",
			f.Name(), reflect.TypeOf(connection).String())
	}
	f.managementClient = management.NewManagementServiceClient(client)
	return nil
}

func (f *Forwarder) Forward(batch event.BatchEvents) error {
	pingOnce := make(map[string]void)
	for _, e := range batch {
		// report instance
		instanceProperties := e.GetInstance()
		if instanceProperties != nil {
			_, err := f.managementClient.ReportInstanceProperties(context.Background(), instanceProperties)
			if err != nil {
				return err
			}
			continue
		}

		// report instance ping
		instancePing := e.GetInstancePing()
		if instancePing != nil {
			// report once
			instancePingStr := fmt.Sprintf("%s_%s", instancePing.Service, instancePing.ServiceInstance)
			_, exists := pingOnce[instancePingStr]
			if !exists {
				_, err := f.managementClient.KeepAlive(context.Background(), instancePing)
				if err != nil {
					return err
				}
				pingOnce[instancePingStr] = empty
			}
			continue
		}
	}
	return nil
}

func (f *Forwarder) ForwardType() v1.SniffType {
	return v1.SniffType_ManagementType
}

func (f *Forwarder) SyncForward(*v1.SniffData) (*v1.SniffData, error) {
	return nil, fmt.Errorf("unsupport sync forward")
}

func (f *Forwarder) SupportedSyncInvoke() bool {
	return false
}

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

package nativecds

import (
	"context"

	v3 "skywalking.apache.org/repo/goapi/collect/agent/configuration/v3"

	module "github.com/apache/skywalking-satellite/internal/satellite/module/api"

	sniffer "skywalking.apache.org/repo/goapi/satellite/data/v1"

	common "skywalking.apache.org/repo/goapi/collect/common/v3"
)

type CDSService struct {
	receiveChannel chan *sniffer.SniffData

	module.SyncInvoker
	v3.UnimplementedConfigurationDiscoveryServiceServer
}

func (p *CDSService) FetchConfigurations(_ context.Context, req *v3.ConfigurationSyncRequest) (*common.Commands, error) {
	event := &sniffer.SniffData{
		Data: &sniffer.SniffData_ConfigurationSyncRequest{
			ConfigurationSyncRequest: req,
		},
	}
	data, err := p.SyncInvoker.SyncInvoke(event)
	if err != nil {
		return nil, err
	}
	return data.GetCommands(), nil
}

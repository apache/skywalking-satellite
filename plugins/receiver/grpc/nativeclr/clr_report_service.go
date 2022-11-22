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

package nativeclr

import (
	"context"
	"time"

	common "skywalking.apache.org/repo/goapi/collect/common/v3"
	agent "skywalking.apache.org/repo/goapi/collect/language/agent/v3"
	v1 "skywalking.apache.org/repo/goapi/satellite/data/v1"
)

const eventName = "grpc-clr-event"

type CLRReportService struct {
	receiveChannel chan *v1.SniffData
	agent.UnimplementedCLRMetricReportServiceServer
}

func (j *CLRReportService) Collect(_ context.Context, clr *agent.CLRMetricCollection) (*common.Commands, error) {
	e := &v1.SniffData{
		Name:      eventName,
		Timestamp: time.Now().UnixNano() / 1e6,
		Meta:      nil,
		Type:      v1.SniffType_CLRMetricType,
		Remote:    true,
		Data: &v1.SniffData_Clr{
			Clr: clr,
		},
	}
	j.receiveChannel <- e
	return &common.Commands{}, nil
}

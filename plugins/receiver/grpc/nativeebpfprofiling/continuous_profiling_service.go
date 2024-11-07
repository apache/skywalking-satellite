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

package nativeebpfprofiling

import (
	"context"

	module "github.com/apache/skywalking-satellite/internal/satellite/module/api"

	common_v3 "skywalking.apache.org/repo/goapi/collect/common/v3"
	v3 "skywalking.apache.org/repo/goapi/collect/ebpf/profiling/v3"
	sniffer "skywalking.apache.org/repo/goapi/satellite/data/v1"
)

type ContinuousProfilingReportService struct {
	module.SyncInvoker

	v3.UnimplementedContinuousProfilingServiceServer
}

func (c *ContinuousProfilingReportService) QueryPolicies(ctx context.Context, query *v3.ContinuousProfilingPolicyQuery) (*common_v3.Commands, error) {
	event := &sniffer.SniffData{
		Data: &sniffer.SniffData_ContinuousProfilingPolicyQuery{
			ContinuousProfilingPolicyQuery: query,
		},
	}
	data, _, err := c.SyncInvoker.SyncInvoke(event)
	if err != nil {
		return nil, err
	}
	return data.GetCommands(), nil
}

func (c *ContinuousProfilingReportService) ReportProfilingTask(ctx context.Context,
	report *v3.ContinuousProfilingReport) (*common_v3.Commands, error) {
	event := &sniffer.SniffData{
		Data: &sniffer.SniffData_ContinuousProfilingReport{
			ContinuousProfilingReport: report,
		},
	}
	data, _, err := c.SyncInvoker.SyncInvoke(event)
	if err != nil {
		return nil, err
	}
	return data.GetCommands(), nil
}

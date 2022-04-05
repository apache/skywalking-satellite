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

	common "skywalking.apache.org/repo/goapi/collect/common/v3"
	management "skywalking.apache.org/repo/goapi/collect/management/v3"
	management_compat "skywalking.apache.org/repo/goapi/collect/management/v3/compat"
)

type ManagementReportServiceCompat struct {
	reportService *ManagementReportService
	management_compat.UnimplementedManagementServiceServer
}

func (m *ManagementReportServiceCompat) ReportInstanceProperties(ctx context.Context, in *management.InstanceProperties) (*common.Commands, error) {
	return m.reportService.ReportInstanceProperties(ctx, in)
}

func (m *ManagementReportServiceCompat) KeepAlive(ctx context.Context, in *management.InstancePingPkg) (*common.Commands, error) {
	return m.reportService.KeepAlive(ctx, in)
}

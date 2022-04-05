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

import (
	"context"

	common "skywalking.apache.org/repo/goapi/collect/common/v3"
	profile "skywalking.apache.org/repo/goapi/collect/language/profile/v3"
	profile_compat "skywalking.apache.org/repo/goapi/collect/language/profile/v3/compat"
)

type ProfileServiceCompat struct {
	reportService *ProfileService
	profile_compat.UnimplementedProfileTaskServer
}

func (p *ProfileServiceCompat) GetProfileTaskCommands(ctx context.Context, q *profile.ProfileTaskCommandQuery) (*common.Commands, error) {
	return p.reportService.GetProfileTaskCommands(ctx, q)
}

func (p *ProfileServiceCompat) CollectSnapshot(stream profile_compat.ProfileTask_CollectSnapshotServer) error {
	return p.reportService.CollectSnapshot(stream)
}

func (p *ProfileServiceCompat) ReportTaskFinish(ctx context.Context, report *profile.ProfileTaskFinishReport) (*common.Commands, error) {
	return p.reportService.ReportTaskFinish(ctx, report)
}

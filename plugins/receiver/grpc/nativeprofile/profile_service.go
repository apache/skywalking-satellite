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
	"io"
	"time"

	module "github.com/apache/skywalking-satellite/internal/satellite/module/api"

	sniffer "skywalking.apache.org/repo/goapi/satellite/data/v1"

	common "skywalking.apache.org/repo/goapi/collect/common/v3"
	profile "skywalking.apache.org/repo/goapi/collect/language/profile/v3"
)

const eventName = "grpc-profile-event"

type ProfileService struct {
	receiveChannel chan *sniffer.SniffData

	module.SyncInvoker
	profile.UnimplementedProfileTaskServer
}

func (p *ProfileService) GetProfileTaskCommands(_ context.Context, q *profile.ProfileTaskCommandQuery) (*common.Commands, error) {
	event := &sniffer.SniffData{
		Data: &sniffer.SniffData_ProfileTaskQuery{
			ProfileTaskQuery: q,
		},
	}
	data, err := p.SyncInvoker.SyncInvoke(event)
	if err != nil {
		return nil, err
	}
	return data.GetCommands(), nil
}

func (p *ProfileService) CollectSnapshot(stream profile.ProfileTask_CollectSnapshotServer) error {
	for {
		snapshot, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&common.Commands{})
		}
		if err != nil {
			return err
		}
		e := &sniffer.SniffData{
			Name:      eventName,
			Timestamp: time.Now().UnixNano() / 1e6,
			Meta:      nil,
			Type:      sniffer.SniffType_ProfileType,
			Remote:    true,
			Data: &sniffer.SniffData_Profile{
				Profile: snapshot,
			},
		}
		p.receiveChannel <- e
	}
}
func (p *ProfileService) ReportTaskFinish(_ context.Context, report *profile.ProfileTaskFinishReport) (*common.Commands, error) {
	event := &sniffer.SniffData{
		Data: &sniffer.SniffData_ProfileTaskFinish{
			ProfileTaskFinish: report,
		},
	}
	data, err := p.SyncInvoker.SyncInvoke(event)
	if err != nil {
		return nil, err
	}
	return data.GetCommands(), nil
}

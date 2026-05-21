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

package nativepprof

import (
	"context"
	"fmt"
	"io"

	"github.com/apache/skywalking-satellite/internal/pkg/log"

	common "skywalking.apache.org/repo/goapi/collect/common/v3"
	pprof "skywalking.apache.org/repo/goapi/collect/pprof/v10"
	v1 "skywalking.apache.org/repo/goapi/satellite/data/v1"

	module "github.com/apache/skywalking-satellite/internal/satellite/module/api"
	"github.com/apache/skywalking-satellite/plugins/server/grpc"
)

type PprofService struct {
	receiveChannel chan *v1.SniffData

	module.SyncInvoker
	pprof.UnimplementedPprofTaskServer
}

func (p *PprofService) GetPprofTaskCommands(_ context.Context,
	query *pprof.PprofTaskCommandQuery,
) (*common.Commands, error) {
	event := &v1.SniffData{
		Data: &v1.SniffData_PprofTaskCommandQuery{
			PprofTaskCommandQuery: query,
		},
	}
	data, _, err := p.SyncInvoke(event)
	if err != nil {
		return nil, err
	}
	return data.GetCommands(), nil
}

func (p *PprofService) Collect(clientStream pprof.PprofTask_CollectServer) error {
	metaData := grpc.NewOriginalData(nil)
	err := clientStream.RecvMsg(metaData)
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return err
	}
	event := &v1.SniffData{
		Data: &v1.SniffData_PprofData{
			PprofData: metaData.Content,
		},
	}
	serverStreamAndResp, serverStream, err := p.SyncInvoke(event)
	if err != nil {
		return fmt.Errorf("satellite send pprof metadata to server but get err: %s", err)
	}
	resp := serverStreamAndResp.GetData().(*v1.SniffData_PprofCollectionResponse)
	err = clientStream.Send(resp.PprofCollectionResponse)
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return err
	}

	for {
		chunk := grpc.NewOriginalData(nil)
		err := clientStream.RecvMsg(chunk)
		if err == io.EOF {
			if cerr := serverStream.CloseSend(); cerr != nil {
				log.Logger.Errorf("pprof service close server stream error: %s", cerr)
			}
			return nil
		}
		if err != nil {
			return err
		}
		if err = serverStream.SendMsg(chunk); err != nil {
			return fmt.Errorf("satellite send pprof chunk to server but get err: %s", err)
		}
	}
}

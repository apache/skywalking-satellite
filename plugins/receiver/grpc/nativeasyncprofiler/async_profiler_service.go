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

package nativeasyncprofiler

import (
	"context"
	"fmt"
	"io"

	"github.com/apache/skywalking-satellite/internal/pkg/log"

	common "skywalking.apache.org/repo/goapi/collect/common/v3"
	asyncprofiler "skywalking.apache.org/repo/goapi/collect/language/asyncprofiler/v10"
	v1 "skywalking.apache.org/repo/goapi/satellite/data/v1"

	module "github.com/apache/skywalking-satellite/internal/satellite/module/api"
	"github.com/apache/skywalking-satellite/plugins/server/grpc"
)

type AsyncProfilerService struct {
	receiveChannel chan *v1.SniffData

	module.SyncInvoker
	asyncprofiler.UnimplementedAsyncProfilerTaskServer
}

func (p *AsyncProfilerService) GetAsyncProfilerTaskCommands(_ context.Context,
	query *asyncprofiler.AsyncProfilerTaskCommandQuery,
) (*common.Commands, error) {
	event := &v1.SniffData{
		Data: &v1.SniffData_AsyncProfilerTaskCommandQuery{
			AsyncProfilerTaskCommandQuery: query,
		},
	}
	data, _, err := p.SyncInvoke(event)
	if err != nil {
		return nil, err
	}
	return data.GetCommands(), nil
}

func (p *AsyncProfilerService) Collect(clientStream asyncprofiler.AsyncProfilerTask_CollectServer) error {
	metaData := grpc.NewOriginalData(nil)
	err := clientStream.RecvMsg(metaData)
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return err
	}
	event := &v1.SniffData{
		Data: &v1.SniffData_AsyncProfilerData{
			AsyncProfilerData: metaData.Content,
		},
	}
	// send metadata to server
	serverStreamAndResp, serverStream, err := p.SyncInvoke(event)
	if err != nil {
		return fmt.Errorf("satellite send metadata to server but get err: %s", err)
	}
	data := serverStreamAndResp.GetData().(*v1.SniffData_AsyncProfilerCollectionResponse)
	// send response to client
	err = clientStream.Send(data.AsyncProfilerCollectionResponse)
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return err
	}

	// receive jfr content and send
	for {
		jfrContent := grpc.NewOriginalData(nil)
		err := clientStream.RecvMsg(jfrContent)
		if err == io.EOF {
			if err = serverStream.CloseSend(); err != nil {
				log.Logger.Errorf("async profiler service close server stream error: %s", err)
			}
			return nil
		}
		if err != nil {
			return err
		}
		if err = serverStream.SendMsg(jfrContent); err != nil {
			return fmt.Errorf("satellite send jfr content to server but get err: %s ", err)
		}
	}
}

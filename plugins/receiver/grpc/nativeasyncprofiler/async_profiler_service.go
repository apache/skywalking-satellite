package nativeasyncprofiler

import (
	"context"
	"io"

	module "github.com/apache/skywalking-satellite/internal/satellite/module/api"
	"github.com/apache/skywalking-satellite/plugins/server/grpc"
	common "skywalking.apache.org/repo/goapi/collect/common/v3"
	asyncprofiler "skywalking.apache.org/repo/goapi/collect/language/asyncprofiler/v10"
	v1 "skywalking.apache.org/repo/goapi/satellite/data/v1"
)

const eventName = "grpc-async-profiler-event"

type AsyncProfilerService struct {
	receiveChannel chan *v1.SniffData

	module.SyncInvoker
	asyncprofiler.UnimplementedAsyncProfilerTaskServer
}

func (p *AsyncProfilerService) GetAsyncProfilerTaskCommands(_ context.Context, query *asyncprofiler.AsyncProfilerTaskCommandQuery) (*common.Commands, error) {
	event := &v1.SniffData{
		Data: &v1.SniffData_AsyncProfilerTaskCommandQuery{
			AsyncProfilerTaskCommandQuery: query,
		},
	}
	data, _, err := p.SyncInvoker.SyncInvoke(event)
	if err != nil {
		return nil, err
	}
	return data.GetCommands(), nil
}

func (p *AsyncProfilerService) Collect(clientStream asyncprofiler.AsyncProfilerTask_CollectServer) error {
	metaData := grpc.NewOriginalData(nil)
	err := clientStream.RecvMsg(metaData)
	if err != nil {
		return err
	}
	event := &v1.SniffData{
		Data: &v1.SniffData_AsyncProfilerData{
			AsyncProfilerData: metaData.Content,
		},
	}

	serverStreamAndResp, serverStream, err := p.SyncInvoker.SyncInvoke(event)
	if err != nil {
		return err
	}
	data := serverStreamAndResp.GetData().(*v1.SniffData_AsyncProfilerCollectionResponse)

	err = clientStream.Send(data.AsyncProfilerCollectionResponse)
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return err
	}

	for {
		jfrContent := grpc.NewOriginalData(nil)
		err := clientStream.RecvMsg(jfrContent)
		if err != nil {
			serverStream.CloseSend()
			if err == io.EOF {
				return nil
			} else {
				return err
			}
		}

		if err = serverStream.SendMsg(jfrContent); err != nil {
			return err
		}
	}
}

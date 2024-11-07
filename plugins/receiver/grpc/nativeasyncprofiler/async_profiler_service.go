package nativeasyncprofiler

import (
	"context"
	"io"
	"time"

	data2 "github.com/apache/skywalking-satellite/internal/data"
	module "github.com/apache/skywalking-satellite/internal/satellite/module/api"
	"github.com/apache/skywalking-satellite/plugins/server/grpc"
	common "skywalking.apache.org/repo/goapi/collect/common/v3"
	asyncprofiler "skywalking.apache.org/repo/goapi/collect/language/asyncprofiler/v10"
	v1 "skywalking.apache.org/repo/goapi/satellite/data/v1"
)

const eventName = "grpc-async-profiler-event"

type AsyncProfilerService struct {
	receiveChannel chan *data2.SniffData

	module.SyncInvoker
	asyncprofiler.UnimplementedAsyncProfilerTaskServer
}

func (p *AsyncProfilerService) GetAsyncProfilerTaskCommands(_ context.Context, query *asyncprofiler.AsyncProfilerTaskCommandQuery) (*common.Commands, error) {
	event := &data2.SniffData{
		SniffData: &v1.SniffData{
			Data: &v1.SniffData_AsyncProfilerTaskCommandQuery{
				AsyncProfilerTaskCommandQuery: query,
			},
		},
	}
	data, err := p.SyncInvoker.SyncInvoke(event)
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
	event := &data2.SniffData{
		SniffData: &v1.SniffData{
			Data: &v1.SniffData_AsyncProfilerData{
				AsyncProfilerData: metaData.Content,
			},
		},
		ClientStream: nil,
	}

	serverStreamAndResp, err := p.SyncInvoker.SyncInvoke(event)
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

	serverStream := serverStreamAndResp.ClientStream

	for {
		jfrContent := grpc.NewOriginalData(nil)
		err := clientStream.RecvMsg(jfrContent)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		e := &data2.SniffData{
			SniffData: &v1.SniffData{
				Name:      eventName,
				Timestamp: time.Now().UnixNano() / 1e6,
				Meta:      nil,
				Type:      v1.SniffType_AsyncProfilerType,
				Remote:    true,
				Data: &v1.SniffData_AsyncProfilerData{
					AsyncProfilerData: jfrContent.Content,
				},
			},
			ClientStream: serverStream,
		}
		p.receiveChannel <- e
	}
}

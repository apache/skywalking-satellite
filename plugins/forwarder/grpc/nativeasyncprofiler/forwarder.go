package nativeasyncprofiler

import (
	"context"
	"fmt"
	server_grpc "github.com/apache/skywalking-satellite/plugins/server/grpc"
	"reflect"

	"github.com/apache/skywalking-satellite/internal/pkg/config"
	"github.com/apache/skywalking-satellite/internal/pkg/log"
	"github.com/apache/skywalking-satellite/internal/satellite/event"
	"google.golang.org/grpc"
	asyncprofiler "skywalking.apache.org/repo/goapi/collect/language/asyncprofiler/v10"
	v1 "skywalking.apache.org/repo/goapi/satellite/data/v1"
)

const (
	Name     = "native-async-profiler-grpc-forwarder"
	ShowName = "Native Async Profiler GRPC Forwarder"
)

type Forwarder struct {
	config.CommonFields

	profilingClient asyncprofiler.AsyncProfilerTaskClient
}

func (f *Forwarder) Name() string {
	return Name
}

func (f *Forwarder) ShowName() string {
	return ShowName
}

func (f *Forwarder) Description() string {
	return "This is a grpc forwarder with the SkyWalking native async profiler protocol."
}

func (f *Forwarder) DefaultConfig() string {
	return ``
}

func (f *Forwarder) Prepare(connection interface{}) error {
	client, ok := connection.(*grpc.ClientConn)
	if !ok {
		return fmt.Errorf("the %s only accepts a grpc client, but received a %s",
			f.Name(), reflect.TypeOf(connection).String())
	}
	f.profilingClient = asyncprofiler.NewAsyncProfilerTaskClient(client)
	return nil
}

func (f *Forwarder) Forward(batch event.BatchEvents) error {
	return fmt.Errorf("unsupport forward")
}

func (f *Forwarder) SyncForward(e *v1.SniffData) (*v1.SniffData, grpc.ClientStream, error) {
	switch requestData := e.GetData().(type) {
	case *v1.SniffData_AsyncProfilerTaskCommandQuery:
		query := requestData.AsyncProfilerTaskCommandQuery
		commands, err := f.profilingClient.GetAsyncProfilerTaskCommands(context.Background(), query)
		if err != nil {
			return nil, nil, err
		}

		return &v1.SniffData{Data: &v1.SniffData_Commands{Commands: commands}}, nil, nil
	case *v1.SniffData_AsyncProfilerData:
		// metadata
		ctx := context.WithValue(context.Background(), "BidirectionalStream", true)
		stream, err := f.profilingClient.Collect(ctx)
		if err != nil {
			log.Logger.Errorf("%s open collect stream error: %v", f.Name(), err)
			return nil, nil, err
		}
		metaData := server_grpc.NewOriginalData(requestData.AsyncProfilerData)
		err = stream.SendMsg(metaData)
		canUpload, err := stream.Recv()
		if err != nil {
			log.Logger.Errorf("%s send meta data error: %v", f.Name(), err)
			f.closeStream(stream)
			return nil, nil, err
		}

		return &v1.SniffData{
			Data: &v1.SniffData_AsyncProfilerCollectionResponse{
				AsyncProfilerCollectionResponse: canUpload,
			},
		}, stream, nil
	}

	return nil, nil, fmt.Errorf("unsupport data")
}

func (f *Forwarder) ForwardType() v1.SniffType {
	return v1.SniffType_AsyncProfilerType
}

func (f *Forwarder) SupportedSyncInvoke() bool {
	return true
}

func (f *Forwarder) closeStream(stream asyncprofiler.AsyncProfilerTask_CollectClient) {
	err := stream.CloseSend()
	if err != nil {
		log.Logger.Errorf("%s close stream error: %v", f.Name(), err)
	}
}

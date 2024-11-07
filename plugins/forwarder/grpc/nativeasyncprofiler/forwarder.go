package nativeasyncprofiler

import (
	"context"
	"fmt"
	"reflect"

	"github.com/apache/skywalking-satellite/internal/data"
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
	for _, e := range batch {
		jfrContent, ok := e.GetData().(*v1.SniffData_AsyncProfilerData)
		stream := e.ClientStream
		if !ok || stream == nil {
			continue
		}

		err := stream.SendMsg(jfrContent)
		if err != nil {
			log.Logger.Errorf("%s send log data error: %v", f.Name(), err)
			err = closeStream(stream)
			if err != nil {
				log.Logger.Errorf("%s close stream error: %v", f.Name(), err)
			}
			return err
		}
		stream.CloseSend()
	}
}

func (f *Forwarder) SyncForward(e *data.SniffData) (*data.SniffData, error) {
	switch requestData := e.GetData().(type) {
	case *v1.SniffData_AsyncProfilerTaskCommandQuery:
		query := requestData.AsyncProfilerTaskCommandQuery
		commands, err := f.profilingClient.GetAsyncProfilerTaskCommands(context.Background(), query)
		if err != nil {
			return nil, err
		}

		command := &data.SniffData{
			SniffData: &v1.SniffData{Data: &v1.SniffData_Commands{Commands: commands}},
		}
		return command, nil
	case *v1.SniffData_AsyncProfilerData:
		// metadata
		stream, err := f.profilingClient.Collect(context.Background())
		if err != nil {
			return nil, err
		}

		err = stream.SendMsg(requestData.AsyncProfilerData)
		if err != nil {
			log.Logger.Errorf("%s send log data error: %v", f.Name(), err)
			err = closeStream(stream)
			if err != nil {
				log.Logger.Errorf("%s close stream error: %v", f.Name(), err)
			}
			return nil, err
		}

		canUpload, err := stream.Recv()
		if err != nil {
			log.Logger.Errorf("%s receive log data error: %v", f.Name(), err)
			err = closeStream(stream)
			if err != nil {
				log.Logger.Errorf("%s close stream error: %v", f.Name(), err)
			}
			return nil, err
		}

		resp := &data.SniffData{
			SniffData: v1.SniffData{
				Data: &v1.SniffData_AsyncProfilerCollectionResponse{
					AsyncProfilerCollectionResponse: canUpload,
				},
			},
			ClientStream: stream,
		}

		return resp, nil
	}

	return nil, fmt.Errorf("unsupport data")
}

func (f *Forwarder) ForwardType() v1.SniffType {
	return v1.SniffType_AsyncProfilerType
}

func (f *Forwarder) SupportedSyncInvoke() bool {
	return true
}

func closeStream(stream asyncprofiler.AsyncProfilerTask_CollectClient) error {
	err := stream.CloseSend()
	if err != nil {
		return err
	}
	return nil
}

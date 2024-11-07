package nativeasyncprofiler

import (
	"github.com/apache/skywalking-satellite/internal/pkg/config"
	module "github.com/apache/skywalking-satellite/internal/satellite/module/api"
	forwarder "github.com/apache/skywalking-satellite/plugins/forwarder/api"
	forwarder_nativeasyncprofiler "github.com/apache/skywalking-satellite/plugins/forwarder/grpc/nativeasyncprofiler"
	"github.com/apache/skywalking-satellite/plugins/receiver/grpc"
	v10 "skywalking.apache.org/repo/goapi/collect/language/asyncprofiler/v10"
	v1 "skywalking.apache.org/repo/goapi/satellite/data/v1"
)

const (
	Name     = "grpc-native-async-profiler-receiver"
	ShowName = "GRPC Native Async Profiler Receiver"
)

type Receiver struct {
	config.CommonFields
	grpc.CommonGRPCReceiverFields
	service *AsyncProfilerService
}

func (r *Receiver) Name() string {
	return Name
}

func (r *Receiver) ShowName() string {
	return ShowName
}

func (r *Receiver) Description() string {
	return "This is a receiver for SkyWalking native async-profiler format, " +
		"which is defined at https://github.com/apache/skywalking-data-collect-protocol/blob/master/asyncprofiler/AsyncProfiler.proto."
}

func (r *Receiver) DefaultConfig() string {
	return ""
}

func (r *Receiver) RegisterHandler(server interface{}) {
	r.CommonGRPCReceiverFields = *grpc.InitCommonGRPCReceiverFields(server)
	r.service = &AsyncProfilerService{receiveChannel: r.OutputChannel}
	v10.RegisterAsyncProfilerTaskServer(r.Server, r.service)
}

func (r *Receiver) RegisterSyncInvoker(invoker module.SyncInvoker) {
	r.service.SyncInvoker = invoker
}

func (r *Receiver) Channel() <-chan *v1.SniffData {
	return r.OutputChannel
}

func (r *Receiver) SupportForwarders() []forwarder.Forwarder {
	return []forwarder.Forwarder{
		new(forwarder_nativeasyncprofiler.Forwarder),
	}
}

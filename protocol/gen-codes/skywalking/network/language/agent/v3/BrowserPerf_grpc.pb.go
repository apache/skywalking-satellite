package v3

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	v3 "skywalking/network/common/v3"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion7

// BrowserPerfServiceClient is the client API for BrowserPerfService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type BrowserPerfServiceClient interface {
	// report once per page
	CollectPerfData(ctx context.Context, in *BrowserPerfData, opts ...grpc.CallOption) (*v3.Commands, error)
	// report one or more error logs for pages, could report multiple times.
	CollectErrorLogs(ctx context.Context, opts ...grpc.CallOption) (BrowserPerfService_CollectErrorLogsClient, error)
}

type browserPerfServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewBrowserPerfServiceClient(cc grpc.ClientConnInterface) BrowserPerfServiceClient {
	return &browserPerfServiceClient{cc}
}

func (c *browserPerfServiceClient) CollectPerfData(ctx context.Context, in *BrowserPerfData, opts ...grpc.CallOption) (*v3.Commands, error) {
	out := new(v3.Commands)
	err := c.cc.Invoke(ctx, "/skywalking.v3.BrowserPerfService/collectPerfData", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *browserPerfServiceClient) CollectErrorLogs(ctx context.Context, opts ...grpc.CallOption) (BrowserPerfService_CollectErrorLogsClient, error) {
	stream, err := c.cc.NewStream(ctx, &_BrowserPerfService_serviceDesc.Streams[0], "/skywalking.v3.BrowserPerfService/collectErrorLogs", opts...)
	if err != nil {
		return nil, err
	}
	x := &browserPerfServiceCollectErrorLogsClient{stream}
	return x, nil
}

type BrowserPerfService_CollectErrorLogsClient interface {
	Send(*BrowserErrorLog) error
	CloseAndRecv() (*v3.Commands, error)
	grpc.ClientStream
}

type browserPerfServiceCollectErrorLogsClient struct {
	grpc.ClientStream
}

func (x *browserPerfServiceCollectErrorLogsClient) Send(m *BrowserErrorLog) error {
	return x.ClientStream.SendMsg(m)
}

func (x *browserPerfServiceCollectErrorLogsClient) CloseAndRecv() (*v3.Commands, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(v3.Commands)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// BrowserPerfServiceServer is the server API for BrowserPerfService service.
// All implementations must embed UnimplementedBrowserPerfServiceServer
// for forward compatibility
type BrowserPerfServiceServer interface {
	// report once per page
	CollectPerfData(context.Context, *BrowserPerfData) (*v3.Commands, error)
	// report one or more error logs for pages, could report multiple times.
	CollectErrorLogs(BrowserPerfService_CollectErrorLogsServer) error
	mustEmbedUnimplementedBrowserPerfServiceServer()
}

// UnimplementedBrowserPerfServiceServer must be embedded to have forward compatible implementations.
type UnimplementedBrowserPerfServiceServer struct {
}

func (UnimplementedBrowserPerfServiceServer) CollectPerfData(context.Context, *BrowserPerfData) (*v3.Commands, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CollectPerfData not implemented")
}
func (UnimplementedBrowserPerfServiceServer) CollectErrorLogs(BrowserPerfService_CollectErrorLogsServer) error {
	return status.Errorf(codes.Unimplemented, "method CollectErrorLogs not implemented")
}
func (UnimplementedBrowserPerfServiceServer) mustEmbedUnimplementedBrowserPerfServiceServer() {}

// UnsafeBrowserPerfServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to BrowserPerfServiceServer will
// result in compilation errors.
type UnsafeBrowserPerfServiceServer interface {
	mustEmbedUnimplementedBrowserPerfServiceServer()
}

func RegisterBrowserPerfServiceServer(s grpc.ServiceRegistrar, srv BrowserPerfServiceServer) {
	s.RegisterService(&_BrowserPerfService_serviceDesc, srv)
}

func _BrowserPerfService_CollectPerfData_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BrowserPerfData)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BrowserPerfServiceServer).CollectPerfData(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/skywalking.v3.BrowserPerfService/collectPerfData",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BrowserPerfServiceServer).CollectPerfData(ctx, req.(*BrowserPerfData))
	}
	return interceptor(ctx, in, info, handler)
}

func _BrowserPerfService_CollectErrorLogs_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(BrowserPerfServiceServer).CollectErrorLogs(&browserPerfServiceCollectErrorLogsServer{stream})
}

type BrowserPerfService_CollectErrorLogsServer interface {
	SendAndClose(*v3.Commands) error
	Recv() (*BrowserErrorLog, error)
	grpc.ServerStream
}

type browserPerfServiceCollectErrorLogsServer struct {
	grpc.ServerStream
}

func (x *browserPerfServiceCollectErrorLogsServer) SendAndClose(m *v3.Commands) error {
	return x.ServerStream.SendMsg(m)
}

func (x *browserPerfServiceCollectErrorLogsServer) Recv() (*BrowserErrorLog, error) {
	m := new(BrowserErrorLog)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

var _BrowserPerfService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "skywalking.v3.BrowserPerfService",
	HandlerType: (*BrowserPerfServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "collectPerfData",
			Handler:    _BrowserPerfService_CollectPerfData_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "collectErrorLogs",
			Handler:       _BrowserPerfService_CollectErrorLogs_Handler,
			ClientStreams: true,
		},
	},
	Metadata: "browser/BrowserPerf.proto",
}

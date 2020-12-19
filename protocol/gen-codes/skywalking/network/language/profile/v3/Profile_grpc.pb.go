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

// ProfileTaskClient is the client API for ProfileTask service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ProfileTaskClient interface {
	// query all sniffer need to execute profile task commands
	GetProfileTaskCommands(ctx context.Context, in *ProfileTaskCommandQuery, opts ...grpc.CallOption) (*v3.Commands, error)
	// collect dumped thread snapshot
	CollectSnapshot(ctx context.Context, opts ...grpc.CallOption) (ProfileTask_CollectSnapshotClient, error)
	// report profiling task finished
	ReportTaskFinish(ctx context.Context, in *ProfileTaskFinishReport, opts ...grpc.CallOption) (*v3.Commands, error)
}

type profileTaskClient struct {
	cc grpc.ClientConnInterface
}

func NewProfileTaskClient(cc grpc.ClientConnInterface) ProfileTaskClient {
	return &profileTaskClient{cc}
}

func (c *profileTaskClient) GetProfileTaskCommands(ctx context.Context, in *ProfileTaskCommandQuery, opts ...grpc.CallOption) (*v3.Commands, error) {
	out := new(v3.Commands)
	err := c.cc.Invoke(ctx, "/skywalking.v3.ProfileTask/getProfileTaskCommands", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *profileTaskClient) CollectSnapshot(ctx context.Context, opts ...grpc.CallOption) (ProfileTask_CollectSnapshotClient, error) {
	stream, err := c.cc.NewStream(ctx, &_ProfileTask_serviceDesc.Streams[0], "/skywalking.v3.ProfileTask/collectSnapshot", opts...)
	if err != nil {
		return nil, err
	}
	x := &profileTaskCollectSnapshotClient{stream}
	return x, nil
}

type ProfileTask_CollectSnapshotClient interface {
	Send(*ThreadSnapshot) error
	CloseAndRecv() (*v3.Commands, error)
	grpc.ClientStream
}

type profileTaskCollectSnapshotClient struct {
	grpc.ClientStream
}

func (x *profileTaskCollectSnapshotClient) Send(m *ThreadSnapshot) error {
	return x.ClientStream.SendMsg(m)
}

func (x *profileTaskCollectSnapshotClient) CloseAndRecv() (*v3.Commands, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(v3.Commands)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *profileTaskClient) ReportTaskFinish(ctx context.Context, in *ProfileTaskFinishReport, opts ...grpc.CallOption) (*v3.Commands, error) {
	out := new(v3.Commands)
	err := c.cc.Invoke(ctx, "/skywalking.v3.ProfileTask/reportTaskFinish", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ProfileTaskServer is the server API for ProfileTask service.
// All implementations must embed UnimplementedProfileTaskServer
// for forward compatibility
type ProfileTaskServer interface {
	// query all sniffer need to execute profile task commands
	GetProfileTaskCommands(context.Context, *ProfileTaskCommandQuery) (*v3.Commands, error)
	// collect dumped thread snapshot
	CollectSnapshot(ProfileTask_CollectSnapshotServer) error
	// report profiling task finished
	ReportTaskFinish(context.Context, *ProfileTaskFinishReport) (*v3.Commands, error)
	mustEmbedUnimplementedProfileTaskServer()
}

// UnimplementedProfileTaskServer must be embedded to have forward compatible implementations.
type UnimplementedProfileTaskServer struct {
}

func (UnimplementedProfileTaskServer) GetProfileTaskCommands(context.Context, *ProfileTaskCommandQuery) (*v3.Commands, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetProfileTaskCommands not implemented")
}
func (UnimplementedProfileTaskServer) CollectSnapshot(ProfileTask_CollectSnapshotServer) error {
	return status.Errorf(codes.Unimplemented, "method CollectSnapshot not implemented")
}
func (UnimplementedProfileTaskServer) ReportTaskFinish(context.Context, *ProfileTaskFinishReport) (*v3.Commands, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ReportTaskFinish not implemented")
}
func (UnimplementedProfileTaskServer) mustEmbedUnimplementedProfileTaskServer() {}

// UnsafeProfileTaskServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ProfileTaskServer will
// result in compilation errors.
type UnsafeProfileTaskServer interface {
	mustEmbedUnimplementedProfileTaskServer()
}

func RegisterProfileTaskServer(s grpc.ServiceRegistrar, srv ProfileTaskServer) {
	s.RegisterService(&_ProfileTask_serviceDesc, srv)
}

func _ProfileTask_GetProfileTaskCommands_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ProfileTaskCommandQuery)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProfileTaskServer).GetProfileTaskCommands(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/skywalking.v3.ProfileTask/getProfileTaskCommands",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ProfileTaskServer).GetProfileTaskCommands(ctx, req.(*ProfileTaskCommandQuery))
	}
	return interceptor(ctx, in, info, handler)
}

func _ProfileTask_CollectSnapshot_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(ProfileTaskServer).CollectSnapshot(&profileTaskCollectSnapshotServer{stream})
}

type ProfileTask_CollectSnapshotServer interface {
	SendAndClose(*v3.Commands) error
	Recv() (*ThreadSnapshot, error)
	grpc.ServerStream
}

type profileTaskCollectSnapshotServer struct {
	grpc.ServerStream
}

func (x *profileTaskCollectSnapshotServer) SendAndClose(m *v3.Commands) error {
	return x.ServerStream.SendMsg(m)
}

func (x *profileTaskCollectSnapshotServer) Recv() (*ThreadSnapshot, error) {
	m := new(ThreadSnapshot)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _ProfileTask_ReportTaskFinish_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ProfileTaskFinishReport)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ProfileTaskServer).ReportTaskFinish(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/skywalking.v3.ProfileTask/reportTaskFinish",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ProfileTaskServer).ReportTaskFinish(ctx, req.(*ProfileTaskFinishReport))
	}
	return interceptor(ctx, in, info, handler)
}

var _ProfileTask_serviceDesc = grpc.ServiceDesc{
	ServiceName: "skywalking.v3.ProfileTask",
	HandlerType: (*ProfileTaskServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "getProfileTaskCommands",
			Handler:    _ProfileTask_GetProfileTaskCommands_Handler,
		},
		{
			MethodName: "reportTaskFinish",
			Handler:    _ProfileTask_ReportTaskFinish_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "collectSnapshot",
			Handler:       _ProfileTask_CollectSnapshot_Handler,
			ClientStreams: true,
		},
	},
	Metadata: "profile/Profile.proto",
}

// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

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
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// TraceSegmentReportServiceClient is the client API for TraceSegmentReportService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type TraceSegmentReportServiceClient interface {
	// Recommended trace segment report channel.
	// gRPC streaming provides better performance.
	// All language agents should choose this.
	Collect(ctx context.Context, opts ...grpc.CallOption) (TraceSegmentReportService_CollectClient, error)
	// An alternative for trace report by using gRPC unary
	// This is provided for some 3rd-party integration, if and only if they prefer the unary mode somehow.
	// The performance of SkyWalking OAP server would be very similar with streaming report,
	// the performance of the network and client side are affected
	CollectInSync(ctx context.Context, in *SegmentCollection, opts ...grpc.CallOption) (*v3.Commands, error)
}

type traceSegmentReportServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewTraceSegmentReportServiceClient(cc grpc.ClientConnInterface) TraceSegmentReportServiceClient {
	return &traceSegmentReportServiceClient{cc}
}

func (c *traceSegmentReportServiceClient) Collect(ctx context.Context, opts ...grpc.CallOption) (TraceSegmentReportService_CollectClient, error) {
	stream, err := c.cc.NewStream(ctx, &TraceSegmentReportService_ServiceDesc.Streams[0], "/skywalking.v3.TraceSegmentReportService/collect", opts...)
	if err != nil {
		return nil, err
	}
	x := &traceSegmentReportServiceCollectClient{stream}
	return x, nil
}

type TraceSegmentReportService_CollectClient interface {
	Send(*SegmentObject) error
	CloseAndRecv() (*v3.Commands, error)
	grpc.ClientStream
}

type traceSegmentReportServiceCollectClient struct {
	grpc.ClientStream
}

func (x *traceSegmentReportServiceCollectClient) Send(m *SegmentObject) error {
	return x.ClientStream.SendMsg(m)
}

func (x *traceSegmentReportServiceCollectClient) CloseAndRecv() (*v3.Commands, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(v3.Commands)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *traceSegmentReportServiceClient) CollectInSync(ctx context.Context, in *SegmentCollection, opts ...grpc.CallOption) (*v3.Commands, error) {
	out := new(v3.Commands)
	err := c.cc.Invoke(ctx, "/skywalking.v3.TraceSegmentReportService/collectInSync", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// TraceSegmentReportServiceServer is the server API for TraceSegmentReportService service.
// All implementations must embed UnimplementedTraceSegmentReportServiceServer
// for forward compatibility
type TraceSegmentReportServiceServer interface {
	// Recommended trace segment report channel.
	// gRPC streaming provides better performance.
	// All language agents should choose this.
	Collect(TraceSegmentReportService_CollectServer) error
	// An alternative for trace report by using gRPC unary
	// This is provided for some 3rd-party integration, if and only if they prefer the unary mode somehow.
	// The performance of SkyWalking OAP server would be very similar with streaming report,
	// the performance of the network and client side are affected
	CollectInSync(context.Context, *SegmentCollection) (*v3.Commands, error)
	mustEmbedUnimplementedTraceSegmentReportServiceServer()
}

// UnimplementedTraceSegmentReportServiceServer must be embedded to have forward compatible implementations.
type UnimplementedTraceSegmentReportServiceServer struct {
}

func (UnimplementedTraceSegmentReportServiceServer) Collect(TraceSegmentReportService_CollectServer) error {
	return status.Errorf(codes.Unimplemented, "method Collect not implemented")
}
func (UnimplementedTraceSegmentReportServiceServer) CollectInSync(context.Context, *SegmentCollection) (*v3.Commands, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CollectInSync not implemented")
}
func (UnimplementedTraceSegmentReportServiceServer) mustEmbedUnimplementedTraceSegmentReportServiceServer() {
}

// UnsafeTraceSegmentReportServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to TraceSegmentReportServiceServer will
// result in compilation errors.
type UnsafeTraceSegmentReportServiceServer interface {
	mustEmbedUnimplementedTraceSegmentReportServiceServer()
}

func RegisterTraceSegmentReportServiceServer(s grpc.ServiceRegistrar, srv TraceSegmentReportServiceServer) {
	s.RegisterService(&TraceSegmentReportService_ServiceDesc, srv)
}

func _TraceSegmentReportService_Collect_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(TraceSegmentReportServiceServer).Collect(&traceSegmentReportServiceCollectServer{stream})
}

type TraceSegmentReportService_CollectServer interface {
	SendAndClose(*v3.Commands) error
	Recv() (*SegmentObject, error)
	grpc.ServerStream
}

type traceSegmentReportServiceCollectServer struct {
	grpc.ServerStream
}

func (x *traceSegmentReportServiceCollectServer) SendAndClose(m *v3.Commands) error {
	return x.ServerStream.SendMsg(m)
}

func (x *traceSegmentReportServiceCollectServer) Recv() (*SegmentObject, error) {
	m := new(SegmentObject)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _TraceSegmentReportService_CollectInSync_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SegmentCollection)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TraceSegmentReportServiceServer).CollectInSync(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/skywalking.v3.TraceSegmentReportService/collectInSync",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TraceSegmentReportServiceServer).CollectInSync(ctx, req.(*SegmentCollection))
	}
	return interceptor(ctx, in, info, handler)
}

// TraceSegmentReportService_ServiceDesc is the grpc.ServiceDesc for TraceSegmentReportService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var TraceSegmentReportService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "skywalking.v3.TraceSegmentReportService",
	HandlerType: (*TraceSegmentReportServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "collectInSync",
			Handler:    _TraceSegmentReportService_CollectInSync_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "collect",
			Handler:       _TraceSegmentReportService_Collect_Handler,
			ClientStreams: true,
		},
	},
	Metadata: "language-agent/Tracing.proto",
}

// Code generated by protoc-gen-go. DO NOT EDIT.
// source: notification/service.proto

package notificationpb

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import notification "github.com/centrifuge/centrifuge-protobufs/gen/go/notification"
import empty "github.com/golang/protobuf/ptypes/empty"
import _ "github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger/options"
import _ "google.golang.org/genproto/googleapis/api/annotations"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for NotificationDummyService service

type NotificationDummyServiceClient interface {
	Notify(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*notification.NotificationMessage, error)
}

type notificationDummyServiceClient struct {
	cc *grpc.ClientConn
}

func NewNotificationDummyServiceClient(cc *grpc.ClientConn) NotificationDummyServiceClient {
	return &notificationDummyServiceClient{cc}
}

func (c *notificationDummyServiceClient) Notify(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*notification.NotificationMessage, error) {
	out := new(notification.NotificationMessage)
	err := grpc.Invoke(ctx, "/notification.NotificationDummyService/Notify", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for NotificationDummyService service

type NotificationDummyServiceServer interface {
	Notify(context.Context, *empty.Empty) (*notification.NotificationMessage, error)
}

func RegisterNotificationDummyServiceServer(s *grpc.Server, srv NotificationDummyServiceServer) {
	s.RegisterService(&_NotificationDummyService_serviceDesc, srv)
}

func _NotificationDummyService_Notify_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(empty.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NotificationDummyServiceServer).Notify(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/notification.NotificationDummyService/Notify",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NotificationDummyServiceServer).Notify(ctx, req.(*empty.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

var _NotificationDummyService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "notification.NotificationDummyService",
	HandlerType: (*NotificationDummyServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Notify",
			Handler:    _NotificationDummyService_Notify_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "notification/service.proto",
}

func init() { proto.RegisterFile("notification/service.proto", fileDescriptor_service_16073442eb60792c) }

var fileDescriptor_service_16073442eb60792c = []byte{
	// 240 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x92, 0xca, 0xcb, 0x2f, 0xc9,
	0x4c, 0xcb, 0x4c, 0x4e, 0x2c, 0xc9, 0xcc, 0xcf, 0xd3, 0x2f, 0x4e, 0x2d, 0x2a, 0xcb, 0x4c, 0x4e,
	0xd5, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0xe2, 0x41, 0x96, 0x93, 0x92, 0x49, 0xcf, 0xcf, 0x4f,
	0xcf, 0x49, 0xd5, 0x4f, 0x2c, 0xc8, 0xd4, 0x4f, 0xcc, 0xcb, 0xcb, 0x2f, 0x01, 0x0b, 0x17, 0x43,
	0xd4, 0x4a, 0x49, 0x43, 0x65, 0xc1, 0xbc, 0xa4, 0xd2, 0x34, 0xfd, 0xd4, 0xdc, 0x82, 0x92, 0x4a,
	0xa8, 0xa4, 0x3c, 0x8a, 0x25, 0xc8, 0x1c, 0xa8, 0x02, 0x1d, 0x30, 0x95, 0xac, 0x9b, 0x9e, 0x9a,
	0xa7, 0x5b, 0x5c, 0x9e, 0x98, 0x9e, 0x9e, 0x5a, 0xa4, 0x9f, 0x5f, 0x00, 0x36, 0x1f, 0xd3, 0x2e,
	0xa3, 0x7e, 0x46, 0x2e, 0x09, 0x3f, 0x24, 0x43, 0x5c, 0x4a, 0x73, 0x73, 0x2b, 0x83, 0x21, 0x4e,
	0x17, 0x2a, 0xe6, 0x62, 0x03, 0xcb, 0x55, 0x0a, 0x89, 0xe9, 0x41, 0xdc, 0xa4, 0x07, 0x73, 0x93,
	0x9e, 0x2b, 0xc8, 0x4d, 0x52, 0x8a, 0x7a, 0x28, 0x2e, 0x40, 0x36, 0xc9, 0x37, 0xb5, 0xb8, 0x38,
	0x31, 0x3d, 0x55, 0x49, 0x6f, 0x92, 0xa3, 0xac, 0x94, 0x34, 0xd8, 0x5c, 0x05, 0x64, 0xc5, 0x0a,
	0xa9, 0x79, 0x29, 0x05, 0xf9, 0x99, 0x79, 0x25, 0x4d, 0x97, 0x9f, 0x4c, 0x66, 0xe2, 0x10, 0x62,
	0xd3, 0x4f, 0x01, 0xa9, 0x71, 0x32, 0xe2, 0x12, 0x48, 0xce, 0xcf, 0x45, 0x31, 0xd7, 0x89, 0x07,
	0xea, 0xa2, 0x00, 0x90, 0xed, 0x01, 0x8c, 0x51, 0x7c, 0xc8, 0xb2, 0x05, 0x49, 0x49, 0x6c, 0x60,
	0x67, 0x19, 0x03, 0x02, 0x00, 0x00, 0xff, 0xff, 0x03, 0xa0, 0xe2, 0xe4, 0x82, 0x01, 0x00, 0x00,
}

// Code generated by protoc-gen-go. DO NOT EDIT.
// source: p2p.proto

/*
Package p2p is a generated protocol buffer package.

It is generated from these files:
	p2p.proto

It has these top-level messages:
	TransmitInvoiceDocument
	TransmitReply
*/
package p2p

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import invoice "github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice"

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

type TransmitInvoiceDocument struct {
	Invoice *invoice.InvoiceDocument `protobuf:"bytes,1,opt,name=invoice" json:"invoice,omitempty"`
}

func (m *TransmitInvoiceDocument) Reset()                    { *m = TransmitInvoiceDocument{} }
func (m *TransmitInvoiceDocument) String() string            { return proto.CompactTextString(m) }
func (*TransmitInvoiceDocument) ProtoMessage()               {}
func (*TransmitInvoiceDocument) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *TransmitInvoiceDocument) GetInvoice() *invoice.InvoiceDocument {
	if m != nil {
		return m.Invoice
	}
	return nil
}

type TransmitReply struct {
	Invoice *invoice.InvoiceDocument `protobuf:"bytes,1,opt,name=invoice" json:"invoice,omitempty"`
}

func (m *TransmitReply) Reset()                    { *m = TransmitReply{} }
func (m *TransmitReply) String() string            { return proto.CompactTextString(m) }
func (*TransmitReply) ProtoMessage()               {}
func (*TransmitReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *TransmitReply) GetInvoice() *invoice.InvoiceDocument {
	if m != nil {
		return m.Invoice
	}
	return nil
}

func init() {
	proto.RegisterType((*TransmitInvoiceDocument)(nil), "p2p.TransmitInvoiceDocument")
	proto.RegisterType((*TransmitReply)(nil), "p2p.TransmitReply")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for P2PService service

type P2PServiceClient interface {
	TransmitInvoice(ctx context.Context, in *TransmitInvoiceDocument, opts ...grpc.CallOption) (*TransmitReply, error)
}

type p2PServiceClient struct {
	cc *grpc.ClientConn
}

func NewP2PServiceClient(cc *grpc.ClientConn) P2PServiceClient {
	return &p2PServiceClient{cc}
}

func (c *p2PServiceClient) TransmitInvoice(ctx context.Context, in *TransmitInvoiceDocument, opts ...grpc.CallOption) (*TransmitReply, error) {
	out := new(TransmitReply)
	err := grpc.Invoke(ctx, "/p2p.P2PService/TransmitInvoice", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for P2PService service

type P2PServiceServer interface {
	TransmitInvoice(context.Context, *TransmitInvoiceDocument) (*TransmitReply, error)
}

func RegisterP2PServiceServer(s *grpc.Server, srv P2PServiceServer) {
	s.RegisterService(&_P2PService_serviceDesc, srv)
}

func _P2PService_TransmitInvoice_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TransmitInvoiceDocument)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(P2PServiceServer).TransmitInvoice(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/p2p.P2PService/TransmitInvoice",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(P2PServiceServer).TransmitInvoice(ctx, req.(*TransmitInvoiceDocument))
	}
	return interceptor(ctx, in, info, handler)
}

var _P2PService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "p2p.P2PService",
	HandlerType: (*P2PServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "TransmitInvoice",
			Handler:    _P2PService_TransmitInvoice_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "p2p.proto",
}

func init() { proto.RegisterFile("p2p.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 195 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x2c, 0x30, 0x2a, 0xd0,
	0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0x2e, 0x30, 0x2a, 0x90, 0x12, 0xcd, 0xcc, 0x2b, 0xcb,
	0xcf, 0x4c, 0x4e, 0xd5, 0x87, 0xd2, 0x10, 0x39, 0x25, 0x5f, 0x2e, 0xf1, 0x90, 0xa2, 0xc4, 0xbc,
	0xe2, 0xdc, 0xcc, 0x12, 0x4f, 0x88, 0x84, 0x4b, 0x7e, 0x72, 0x69, 0x6e, 0x6a, 0x5e, 0x89, 0x90,
	0x11, 0x17, 0x3b, 0x54, 0xad, 0x04, 0xa3, 0x02, 0xa3, 0x06, 0xb7, 0x91, 0x84, 0x1e, 0x4c, 0x2f,
	0x9a, 0xd2, 0x20, 0x98, 0x42, 0x25, 0x67, 0x2e, 0x5e, 0x98, 0x71, 0x41, 0xa9, 0x05, 0x39, 0x95,
	0xe4, 0x18, 0x62, 0x14, 0xcc, 0xc5, 0x15, 0x60, 0x14, 0x10, 0x9c, 0x5a, 0x54, 0x96, 0x99, 0x9c,
	0x2a, 0xe4, 0xca, 0xc5, 0x8f, 0xe6, 0x42, 0x21, 0x19, 0x3d, 0x90, 0xe7, 0x70, 0xb8, 0x5b, 0x4a,
	0x08, 0x45, 0x16, 0xec, 0x0c, 0x25, 0x06, 0x27, 0xf3, 0x28, 0xd3, 0xf4, 0xcc, 0x92, 0x8c, 0xd2,
	0x24, 0xbd, 0xe4, 0xfc, 0x5c, 0x7d, 0xe7, 0xd4, 0xbc, 0x92, 0xa2, 0xcc, 0xb4, 0xd2, 0xf4, 0x54,
	0xcf, 0xbc, 0x64, 0xfd, 0xf4, 0x7c, 0xdd, 0x64, 0xb8, 0x80, 0x3e, 0x12, 0xb3, 0xc0, 0xa8, 0x20,
	0x89, 0x0d, 0x1c, 0x50, 0xc6, 0x80, 0x00, 0x00, 0x00, 0xff, 0xff, 0xf7, 0x75, 0x86, 0x91, 0x51,
	0x01, 0x00, 0x00,
}

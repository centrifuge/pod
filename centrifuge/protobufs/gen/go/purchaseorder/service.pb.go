// Code generated by protoc-gen-go. DO NOT EDIT.
// source: purchaseorder/service.proto

package purchaseorderpb

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import purchaseorder "github.com/centrifuge/centrifuge-protobufs/gen/go/purchaseorder"
import proto1 "github.com/centrifuge/precise-proofs/proofs/proto"
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

type CreatePurchaseOrderProofEnvelope struct {
	DocumentIdentifier   []byte   `protobuf:"bytes,1,opt,name=document_identifier,json=documentIdentifier,proto3" json:"document_identifier,omitempty"`
	Fields               []string `protobuf:"bytes,2,rep,name=fields,proto3" json:"fields,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *CreatePurchaseOrderProofEnvelope) Reset()         { *m = CreatePurchaseOrderProofEnvelope{} }
func (m *CreatePurchaseOrderProofEnvelope) String() string { return proto.CompactTextString(m) }
func (*CreatePurchaseOrderProofEnvelope) ProtoMessage()    {}
func (*CreatePurchaseOrderProofEnvelope) Descriptor() ([]byte, []int) {
	return fileDescriptor_service_758aabd46595c4c6, []int{0}
}
func (m *CreatePurchaseOrderProofEnvelope) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CreatePurchaseOrderProofEnvelope.Unmarshal(m, b)
}
func (m *CreatePurchaseOrderProofEnvelope) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CreatePurchaseOrderProofEnvelope.Marshal(b, m, deterministic)
}
func (dst *CreatePurchaseOrderProofEnvelope) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CreatePurchaseOrderProofEnvelope.Merge(dst, src)
}
func (m *CreatePurchaseOrderProofEnvelope) XXX_Size() int {
	return xxx_messageInfo_CreatePurchaseOrderProofEnvelope.Size(m)
}
func (m *CreatePurchaseOrderProofEnvelope) XXX_DiscardUnknown() {
	xxx_messageInfo_CreatePurchaseOrderProofEnvelope.DiscardUnknown(m)
}

var xxx_messageInfo_CreatePurchaseOrderProofEnvelope proto.InternalMessageInfo

func (m *CreatePurchaseOrderProofEnvelope) GetDocumentIdentifier() []byte {
	if m != nil {
		return m.DocumentIdentifier
	}
	return nil
}

func (m *CreatePurchaseOrderProofEnvelope) GetFields() []string {
	if m != nil {
		return m.Fields
	}
	return nil
}

type PurchaseOrderProof struct {
	DocumentIdentifier   []byte          `protobuf:"bytes,1,opt,name=document_identifier,json=documentIdentifier,proto3" json:"document_identifier,omitempty"`
	FieldProofs          []*proto1.Proof `protobuf:"bytes,2,rep,name=field_proofs,json=fieldProofs,proto3" json:"field_proofs,omitempty"`
	XXX_NoUnkeyedLiteral struct{}        `json:"-"`
	XXX_unrecognized     []byte          `json:"-"`
	XXX_sizecache        int32           `json:"-"`
}

func (m *PurchaseOrderProof) Reset()         { *m = PurchaseOrderProof{} }
func (m *PurchaseOrderProof) String() string { return proto.CompactTextString(m) }
func (*PurchaseOrderProof) ProtoMessage()    {}
func (*PurchaseOrderProof) Descriptor() ([]byte, []int) {
	return fileDescriptor_service_758aabd46595c4c6, []int{1}
}
func (m *PurchaseOrderProof) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PurchaseOrderProof.Unmarshal(m, b)
}
func (m *PurchaseOrderProof) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PurchaseOrderProof.Marshal(b, m, deterministic)
}
func (dst *PurchaseOrderProof) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PurchaseOrderProof.Merge(dst, src)
}
func (m *PurchaseOrderProof) XXX_Size() int {
	return xxx_messageInfo_PurchaseOrderProof.Size(m)
}
func (m *PurchaseOrderProof) XXX_DiscardUnknown() {
	xxx_messageInfo_PurchaseOrderProof.DiscardUnknown(m)
}

var xxx_messageInfo_PurchaseOrderProof proto.InternalMessageInfo

func (m *PurchaseOrderProof) GetDocumentIdentifier() []byte {
	if m != nil {
		return m.DocumentIdentifier
	}
	return nil
}

func (m *PurchaseOrderProof) GetFieldProofs() []*proto1.Proof {
	if m != nil {
		return m.FieldProofs
	}
	return nil
}

type AnchorPurchaseOrderEnvelope struct {
	Document             *purchaseorder.PurchaseOrderDocument `protobuf:"bytes,1,opt,name=document,proto3" json:"document,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                             `json:"-"`
	XXX_unrecognized     []byte                               `json:"-"`
	XXX_sizecache        int32                                `json:"-"`
}

func (m *AnchorPurchaseOrderEnvelope) Reset()         { *m = AnchorPurchaseOrderEnvelope{} }
func (m *AnchorPurchaseOrderEnvelope) String() string { return proto.CompactTextString(m) }
func (*AnchorPurchaseOrderEnvelope) ProtoMessage()    {}
func (*AnchorPurchaseOrderEnvelope) Descriptor() ([]byte, []int) {
	return fileDescriptor_service_758aabd46595c4c6, []int{2}
}
func (m *AnchorPurchaseOrderEnvelope) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_AnchorPurchaseOrderEnvelope.Unmarshal(m, b)
}
func (m *AnchorPurchaseOrderEnvelope) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_AnchorPurchaseOrderEnvelope.Marshal(b, m, deterministic)
}
func (dst *AnchorPurchaseOrderEnvelope) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AnchorPurchaseOrderEnvelope.Merge(dst, src)
}
func (m *AnchorPurchaseOrderEnvelope) XXX_Size() int {
	return xxx_messageInfo_AnchorPurchaseOrderEnvelope.Size(m)
}
func (m *AnchorPurchaseOrderEnvelope) XXX_DiscardUnknown() {
	xxx_messageInfo_AnchorPurchaseOrderEnvelope.DiscardUnknown(m)
}

var xxx_messageInfo_AnchorPurchaseOrderEnvelope proto.InternalMessageInfo

func (m *AnchorPurchaseOrderEnvelope) GetDocument() *purchaseorder.PurchaseOrderDocument {
	if m != nil {
		return m.Document
	}
	return nil
}

type SendPurchaseOrderEnvelope struct {
	// Centrifuge OS Entity of the recipient
	Recipients           [][]byte                             `protobuf:"bytes,1,rep,name=recipients,proto3" json:"recipients,omitempty"`
	Document             *purchaseorder.PurchaseOrderDocument `protobuf:"bytes,10,opt,name=document,proto3" json:"document,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                             `json:"-"`
	XXX_unrecognized     []byte                               `json:"-"`
	XXX_sizecache        int32                                `json:"-"`
}

func (m *SendPurchaseOrderEnvelope) Reset()         { *m = SendPurchaseOrderEnvelope{} }
func (m *SendPurchaseOrderEnvelope) String() string { return proto.CompactTextString(m) }
func (*SendPurchaseOrderEnvelope) ProtoMessage()    {}
func (*SendPurchaseOrderEnvelope) Descriptor() ([]byte, []int) {
	return fileDescriptor_service_758aabd46595c4c6, []int{3}
}
func (m *SendPurchaseOrderEnvelope) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SendPurchaseOrderEnvelope.Unmarshal(m, b)
}
func (m *SendPurchaseOrderEnvelope) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SendPurchaseOrderEnvelope.Marshal(b, m, deterministic)
}
func (dst *SendPurchaseOrderEnvelope) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SendPurchaseOrderEnvelope.Merge(dst, src)
}
func (m *SendPurchaseOrderEnvelope) XXX_Size() int {
	return xxx_messageInfo_SendPurchaseOrderEnvelope.Size(m)
}
func (m *SendPurchaseOrderEnvelope) XXX_DiscardUnknown() {
	xxx_messageInfo_SendPurchaseOrderEnvelope.DiscardUnknown(m)
}

var xxx_messageInfo_SendPurchaseOrderEnvelope proto.InternalMessageInfo

func (m *SendPurchaseOrderEnvelope) GetRecipients() [][]byte {
	if m != nil {
		return m.Recipients
	}
	return nil
}

func (m *SendPurchaseOrderEnvelope) GetDocument() *purchaseorder.PurchaseOrderDocument {
	if m != nil {
		return m.Document
	}
	return nil
}

type GetPurchaseOrderDocumentEnvelope struct {
	DocumentIdentifier   []byte   `protobuf:"bytes,1,opt,name=document_identifier,json=documentIdentifier,proto3" json:"document_identifier,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetPurchaseOrderDocumentEnvelope) Reset()         { *m = GetPurchaseOrderDocumentEnvelope{} }
func (m *GetPurchaseOrderDocumentEnvelope) String() string { return proto.CompactTextString(m) }
func (*GetPurchaseOrderDocumentEnvelope) ProtoMessage()    {}
func (*GetPurchaseOrderDocumentEnvelope) Descriptor() ([]byte, []int) {
	return fileDescriptor_service_758aabd46595c4c6, []int{4}
}
func (m *GetPurchaseOrderDocumentEnvelope) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetPurchaseOrderDocumentEnvelope.Unmarshal(m, b)
}
func (m *GetPurchaseOrderDocumentEnvelope) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetPurchaseOrderDocumentEnvelope.Marshal(b, m, deterministic)
}
func (dst *GetPurchaseOrderDocumentEnvelope) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetPurchaseOrderDocumentEnvelope.Merge(dst, src)
}
func (m *GetPurchaseOrderDocumentEnvelope) XXX_Size() int {
	return xxx_messageInfo_GetPurchaseOrderDocumentEnvelope.Size(m)
}
func (m *GetPurchaseOrderDocumentEnvelope) XXX_DiscardUnknown() {
	xxx_messageInfo_GetPurchaseOrderDocumentEnvelope.DiscardUnknown(m)
}

var xxx_messageInfo_GetPurchaseOrderDocumentEnvelope proto.InternalMessageInfo

func (m *GetPurchaseOrderDocumentEnvelope) GetDocumentIdentifier() []byte {
	if m != nil {
		return m.DocumentIdentifier
	}
	return nil
}

type ReceivedPurchaseOrders struct {
	Purchaseorders       []*purchaseorder.PurchaseOrderDocument `protobuf:"bytes,1,rep,name=purchaseorders,proto3" json:"purchaseorders,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                               `json:"-"`
	XXX_unrecognized     []byte                                 `json:"-"`
	XXX_sizecache        int32                                  `json:"-"`
}

func (m *ReceivedPurchaseOrders) Reset()         { *m = ReceivedPurchaseOrders{} }
func (m *ReceivedPurchaseOrders) String() string { return proto.CompactTextString(m) }
func (*ReceivedPurchaseOrders) ProtoMessage()    {}
func (*ReceivedPurchaseOrders) Descriptor() ([]byte, []int) {
	return fileDescriptor_service_758aabd46595c4c6, []int{5}
}
func (m *ReceivedPurchaseOrders) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ReceivedPurchaseOrders.Unmarshal(m, b)
}
func (m *ReceivedPurchaseOrders) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ReceivedPurchaseOrders.Marshal(b, m, deterministic)
}
func (dst *ReceivedPurchaseOrders) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ReceivedPurchaseOrders.Merge(dst, src)
}
func (m *ReceivedPurchaseOrders) XXX_Size() int {
	return xxx_messageInfo_ReceivedPurchaseOrders.Size(m)
}
func (m *ReceivedPurchaseOrders) XXX_DiscardUnknown() {
	xxx_messageInfo_ReceivedPurchaseOrders.DiscardUnknown(m)
}

var xxx_messageInfo_ReceivedPurchaseOrders proto.InternalMessageInfo

func (m *ReceivedPurchaseOrders) GetPurchaseorders() []*purchaseorder.PurchaseOrderDocument {
	if m != nil {
		return m.Purchaseorders
	}
	return nil
}

func init() {
	proto.RegisterType((*CreatePurchaseOrderProofEnvelope)(nil), "purchaseorder.CreatePurchaseOrderProofEnvelope")
	proto.RegisterType((*PurchaseOrderProof)(nil), "purchaseorder.PurchaseOrderProof")
	proto.RegisterType((*AnchorPurchaseOrderEnvelope)(nil), "purchaseorder.AnchorPurchaseOrderEnvelope")
	proto.RegisterType((*SendPurchaseOrderEnvelope)(nil), "purchaseorder.SendPurchaseOrderEnvelope")
	proto.RegisterType((*GetPurchaseOrderDocumentEnvelope)(nil), "purchaseorder.GetPurchaseOrderDocumentEnvelope")
	proto.RegisterType((*ReceivedPurchaseOrders)(nil), "purchaseorder.ReceivedPurchaseOrders")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// PurchaseOrderDocumentServiceClient is the client API for PurchaseOrderDocumentService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type PurchaseOrderDocumentServiceClient interface {
	CreatePurchaseOrderProof(ctx context.Context, in *CreatePurchaseOrderProofEnvelope, opts ...grpc.CallOption) (*PurchaseOrderProof, error)
	AnchorPurchaseOrderDocument(ctx context.Context, in *AnchorPurchaseOrderEnvelope, opts ...grpc.CallOption) (*purchaseorder.PurchaseOrderDocument, error)
	SendPurchaseOrderDocument(ctx context.Context, in *SendPurchaseOrderEnvelope, opts ...grpc.CallOption) (*purchaseorder.PurchaseOrderDocument, error)
	GetPurchaseOrderDocument(ctx context.Context, in *GetPurchaseOrderDocumentEnvelope, opts ...grpc.CallOption) (*purchaseorder.PurchaseOrderDocument, error)
	GetReceivedPurchaseOrderDocuments(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*ReceivedPurchaseOrders, error)
}

type purchaseOrderDocumentServiceClient struct {
	cc *grpc.ClientConn
}

func NewPurchaseOrderDocumentServiceClient(cc *grpc.ClientConn) PurchaseOrderDocumentServiceClient {
	return &purchaseOrderDocumentServiceClient{cc}
}

func (c *purchaseOrderDocumentServiceClient) CreatePurchaseOrderProof(ctx context.Context, in *CreatePurchaseOrderProofEnvelope, opts ...grpc.CallOption) (*PurchaseOrderProof, error) {
	out := new(PurchaseOrderProof)
	err := c.cc.Invoke(ctx, "/purchaseorder.PurchaseOrderDocumentService/CreatePurchaseOrderProof", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *purchaseOrderDocumentServiceClient) AnchorPurchaseOrderDocument(ctx context.Context, in *AnchorPurchaseOrderEnvelope, opts ...grpc.CallOption) (*purchaseorder.PurchaseOrderDocument, error) {
	out := new(purchaseorder.PurchaseOrderDocument)
	err := c.cc.Invoke(ctx, "/purchaseorder.PurchaseOrderDocumentService/AnchorPurchaseOrderDocument", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *purchaseOrderDocumentServiceClient) SendPurchaseOrderDocument(ctx context.Context, in *SendPurchaseOrderEnvelope, opts ...grpc.CallOption) (*purchaseorder.PurchaseOrderDocument, error) {
	out := new(purchaseorder.PurchaseOrderDocument)
	err := c.cc.Invoke(ctx, "/purchaseorder.PurchaseOrderDocumentService/SendPurchaseOrderDocument", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *purchaseOrderDocumentServiceClient) GetPurchaseOrderDocument(ctx context.Context, in *GetPurchaseOrderDocumentEnvelope, opts ...grpc.CallOption) (*purchaseorder.PurchaseOrderDocument, error) {
	out := new(purchaseorder.PurchaseOrderDocument)
	err := c.cc.Invoke(ctx, "/purchaseorder.PurchaseOrderDocumentService/GetPurchaseOrderDocument", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *purchaseOrderDocumentServiceClient) GetReceivedPurchaseOrderDocuments(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*ReceivedPurchaseOrders, error) {
	out := new(ReceivedPurchaseOrders)
	err := c.cc.Invoke(ctx, "/purchaseorder.PurchaseOrderDocumentService/GetReceivedPurchaseOrderDocuments", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// PurchaseOrderDocumentServiceServer is the server API for PurchaseOrderDocumentService service.
type PurchaseOrderDocumentServiceServer interface {
	CreatePurchaseOrderProof(context.Context, *CreatePurchaseOrderProofEnvelope) (*PurchaseOrderProof, error)
	AnchorPurchaseOrderDocument(context.Context, *AnchorPurchaseOrderEnvelope) (*purchaseorder.PurchaseOrderDocument, error)
	SendPurchaseOrderDocument(context.Context, *SendPurchaseOrderEnvelope) (*purchaseorder.PurchaseOrderDocument, error)
	GetPurchaseOrderDocument(context.Context, *GetPurchaseOrderDocumentEnvelope) (*purchaseorder.PurchaseOrderDocument, error)
	GetReceivedPurchaseOrderDocuments(context.Context, *empty.Empty) (*ReceivedPurchaseOrders, error)
}

func RegisterPurchaseOrderDocumentServiceServer(s *grpc.Server, srv PurchaseOrderDocumentServiceServer) {
	s.RegisterService(&_PurchaseOrderDocumentService_serviceDesc, srv)
}

func _PurchaseOrderDocumentService_CreatePurchaseOrderProof_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreatePurchaseOrderProofEnvelope)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PurchaseOrderDocumentServiceServer).CreatePurchaseOrderProof(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/purchaseorder.PurchaseOrderDocumentService/CreatePurchaseOrderProof",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PurchaseOrderDocumentServiceServer).CreatePurchaseOrderProof(ctx, req.(*CreatePurchaseOrderProofEnvelope))
	}
	return interceptor(ctx, in, info, handler)
}

func _PurchaseOrderDocumentService_AnchorPurchaseOrderDocument_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AnchorPurchaseOrderEnvelope)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PurchaseOrderDocumentServiceServer).AnchorPurchaseOrderDocument(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/purchaseorder.PurchaseOrderDocumentService/AnchorPurchaseOrderDocument",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PurchaseOrderDocumentServiceServer).AnchorPurchaseOrderDocument(ctx, req.(*AnchorPurchaseOrderEnvelope))
	}
	return interceptor(ctx, in, info, handler)
}

func _PurchaseOrderDocumentService_SendPurchaseOrderDocument_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SendPurchaseOrderEnvelope)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PurchaseOrderDocumentServiceServer).SendPurchaseOrderDocument(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/purchaseorder.PurchaseOrderDocumentService/SendPurchaseOrderDocument",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PurchaseOrderDocumentServiceServer).SendPurchaseOrderDocument(ctx, req.(*SendPurchaseOrderEnvelope))
	}
	return interceptor(ctx, in, info, handler)
}

func _PurchaseOrderDocumentService_GetPurchaseOrderDocument_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetPurchaseOrderDocumentEnvelope)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PurchaseOrderDocumentServiceServer).GetPurchaseOrderDocument(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/purchaseorder.PurchaseOrderDocumentService/GetPurchaseOrderDocument",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PurchaseOrderDocumentServiceServer).GetPurchaseOrderDocument(ctx, req.(*GetPurchaseOrderDocumentEnvelope))
	}
	return interceptor(ctx, in, info, handler)
}

func _PurchaseOrderDocumentService_GetReceivedPurchaseOrderDocuments_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(empty.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PurchaseOrderDocumentServiceServer).GetReceivedPurchaseOrderDocuments(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/purchaseorder.PurchaseOrderDocumentService/GetReceivedPurchaseOrderDocuments",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PurchaseOrderDocumentServiceServer).GetReceivedPurchaseOrderDocuments(ctx, req.(*empty.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

var _PurchaseOrderDocumentService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "purchaseorder.PurchaseOrderDocumentService",
	HandlerType: (*PurchaseOrderDocumentServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreatePurchaseOrderProof",
			Handler:    _PurchaseOrderDocumentService_CreatePurchaseOrderProof_Handler,
		},
		{
			MethodName: "AnchorPurchaseOrderDocument",
			Handler:    _PurchaseOrderDocumentService_AnchorPurchaseOrderDocument_Handler,
		},
		{
			MethodName: "SendPurchaseOrderDocument",
			Handler:    _PurchaseOrderDocumentService_SendPurchaseOrderDocument_Handler,
		},
		{
			MethodName: "GetPurchaseOrderDocument",
			Handler:    _PurchaseOrderDocumentService_GetPurchaseOrderDocument_Handler,
		},
		{
			MethodName: "GetReceivedPurchaseOrderDocuments",
			Handler:    _PurchaseOrderDocumentService_GetReceivedPurchaseOrderDocuments_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "purchaseorder/service.proto",
}

func init() {
	proto.RegisterFile("purchaseorder/service.proto", fileDescriptor_service_758aabd46595c4c6)
}

var fileDescriptor_service_758aabd46595c4c6 = []byte{
	// 642 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x95, 0xcf, 0x4f, 0xd4, 0x40,
	0x14, 0xc7, 0x53, 0x48, 0x50, 0x87, 0x45, 0xe3, 0xa0, 0xb8, 0x14, 0xd4, 0xd9, 0x89, 0xc6, 0x0d,
	0x91, 0xd6, 0x20, 0x31, 0x51, 0x2f, 0x2e, 0x42, 0x88, 0x89, 0x89, 0x9b, 0x72, 0x30, 0x31, 0x31,
	0xa4, 0xb4, 0xaf, 0xa5, 0x71, 0xe9, 0x4c, 0x66, 0x06, 0x08, 0x07, 0x2f, 0x1e, 0x4c, 0xbc, 0x19,
	0xfc, 0x5b, 0xfc, 0x0f, 0xbc, 0x7a, 0xf2, 0x5f, 0xf0, 0x0f, 0x31, 0x9d, 0x69, 0x57, 0xa7, 0xdb,
	0x85, 0x95, 0x53, 0x79, 0xbc, 0x1f, 0xdf, 0xcf, 0x9b, 0x79, 0x6f, 0x07, 0x2d, 0xf1, 0x43, 0x11,
	0xed, 0x87, 0x12, 0x98, 0x88, 0x41, 0xf8, 0x12, 0xc4, 0x51, 0x16, 0x81, 0xc7, 0x05, 0x53, 0x0c,
	0xcf, 0x59, 0x4e, 0x77, 0x39, 0x65, 0x2c, 0x1d, 0x80, 0x1f, 0xf2, 0xcc, 0x0f, 0xf3, 0x9c, 0xa9,
	0x50, 0x65, 0x2c, 0x97, 0x26, 0xd8, 0x5d, 0x2a, 0xbd, 0xda, 0xda, 0x3b, 0x4c, 0x7c, 0x38, 0xe0,
	0xea, 0xa4, 0x74, 0x3e, 0xe0, 0x02, 0xa2, 0x4c, 0xc2, 0x2a, 0x17, 0x8c, 0x25, 0xd2, 0xff, 0xfb,
	0x51, 0xcc, 0x18, 0x65, 0xe0, 0x43, 0xfd, 0x89, 0x56, 0x53, 0xc8, 0x57, 0xe5, 0x71, 0x98, 0xa6,
	0x20, 0x7c, 0xc6, 0xb5, 0x4e, 0x83, 0x66, 0xc7, 0xa6, 0xb7, 0x2c, 0x13, 0x42, 0x3f, 0x20, 0xf2,
	0x52, 0x40, 0xa8, 0xa0, 0x5f, 0x3a, 0xdf, 0x14, 0xce, 0x7e, 0x21, 0xb9, 0x95, 0x1f, 0xc1, 0x80,
	0x71, 0xc0, 0x3e, 0x9a, 0x8f, 0x59, 0x74, 0x78, 0x00, 0xb9, 0xda, 0xcd, 0x62, 0xc8, 0x55, 0x96,
	0x64, 0x20, 0xda, 0x0e, 0x71, 0xba, 0xad, 0x00, 0x57, 0xae, 0x57, 0x43, 0x0f, 0x5e, 0x40, 0x33,
	0x49, 0x06, 0x83, 0x58, 0xb6, 0xa7, 0xc8, 0x74, 0xf7, 0x4a, 0x50, 0x5a, 0xf4, 0x18, 0xe1, 0x51,
	0x99, 0xff, 0x2f, 0xff, 0x08, 0xb5, 0x74, 0xc1, 0x5d, 0x73, 0x4c, 0x5a, 0x64, 0x76, 0x6d, 0xce,
	0x33, 0xa6, 0xa7, 0xab, 0x06, 0xb3, 0x3a, 0x44, 0xff, 0x2d, 0xe9, 0x2e, 0x5a, 0xea, 0xe5, 0xd1,
	0x3e, 0x13, 0x96, 0xfc, 0xb0, 0xc1, 0x17, 0xe8, 0x72, 0x25, 0xa3, 0x65, 0x67, 0xd7, 0xee, 0x79,
	0xf6, 0x61, 0x59, 0x79, 0x9b, 0x65, 0x6c, 0x30, 0xcc, 0xa2, 0x1f, 0xd1, 0xe2, 0x0e, 0xe4, 0x71,
	0x73, 0xf9, 0x3b, 0x08, 0x15, 0xd7, 0xcb, 0x33, 0xc8, 0x95, 0x6c, 0x3b, 0x64, 0xba, 0xdb, 0x0a,
	0xfe, 0xf9, 0x8f, 0x25, 0x8f, 0x2e, 0x24, 0xbf, 0x83, 0xc8, 0x36, 0xa8, 0xc6, 0xa8, 0x0b, 0xdf,
	0x22, 0x4d, 0xd0, 0x42, 0x00, 0x11, 0x64, 0x47, 0x60, 0xf7, 0x25, 0xf1, 0x6b, 0x74, 0xd5, 0xe2,
	0x33, 0x4d, 0x4d, 0x8a, 0x5d, 0xcb, 0x5d, 0xfb, 0x72, 0x09, 0x2d, 0x37, 0x46, 0xee, 0x98, 0x6d,
	0xc3, 0x3f, 0x1d, 0xd4, 0x1e, 0x37, 0xa4, 0xd8, 0xaf, 0x69, 0x9e, 0x37, 0xcd, 0x6e, 0xe7, 0x2c,
	0x48, 0x1d, 0x4a, 0xdf, 0x9f, 0xf6, 0x9e, 0xbb, 0x4f, 0x4d, 0x25, 0x49, 0x42, 0x32, 0xc8, 0xa4,
	0x22, 0x2c, 0x21, 0xe5, 0x96, 0x12, 0x33, 0x68, 0x24, 0x61, 0x82, 0xa8, 0x7d, 0x20, 0x92, 0x43,
	0x54, 0x1c, 0x59, 0x4c, 0xcc, 0x9c, 0x7f, 0xfa, 0xf5, 0xfb, 0xdb, 0xd4, 0x22, 0xbd, 0xe1, 0xd7,
	0xd6, 0xaf, 0xc8, 0x7a, 0xe6, 0xac, 0xe0, 0xef, 0x4e, 0xe3, 0x38, 0x56, 0x6d, 0xe3, 0x95, 0x1a,
	0xe1, 0x19, 0xa3, 0xeb, 0x4e, 0x74, 0xe4, 0x74, 0xf3, 0xb4, 0xd7, 0x71, 0xef, 0x9a, 0x3a, 0x24,
	0x24, 0x56, 0x0a, 0xa9, 0xae, 0x5e, 0x63, 0xbb, 0xf4, 0x66, 0x0d, 0x3b, 0xd4, 0x59, 0x05, 0xf7,
	0x0f, 0xa7, 0x61, 0xca, 0x87, 0xd4, 0xdd, 0x1a, 0xc9, 0xd8, 0x7d, 0x98, 0x90, 0xf9, 0xed, 0x69,
	0xef, 0x89, 0xbb, 0x5e, 0x54, 0x19, 0x4b, 0x4c, 0x14, 0xab, 0x5d, 0x00, 0x0f, 0x85, 0x3a, 0xd1,
	0x8d, 0xb4, 0xe9, 0xbc, 0x5f, 0xff, 0xf1, 0xce, 0xe3, 0xa2, 0x8d, 0xaf, 0x0e, 0x6a, 0x8f, 0xdb,
	0x96, 0x91, 0x71, 0x3a, 0x6f, 0xad, 0x26, 0x6c, 0xe6, 0xb6, 0x66, 0xba, 0x45, 0x71, 0x8d, 0x29,
	0x05, 0x55, 0x20, 0x7d, 0x76, 0x50, 0x67, 0x1b, 0x54, 0xe3, 0xba, 0x55, 0x35, 0x24, 0x5e, 0xf0,
	0xcc, 0x1b, 0xe2, 0x55, 0x6f, 0x88, 0xb7, 0x55, 0xbc, 0x21, 0xee, 0xfd, 0x1a, 0x42, 0xf3, 0xd6,
	0x52, 0xaa, 0x19, 0x96, 0xb1, 0x3b, 0xca, 0x50, 0x65, 0x6c, 0xac, 0xa3, 0xeb, 0x11, 0x3b, 0xb0,
	0xeb, 0x6d, 0xb4, 0xca, 0x45, 0xec, 0x17, 0xb2, 0x7d, 0xe7, 0xdd, 0x35, 0xcb, 0xcd, 0xf7, 0xf6,
	0x66, 0x34, 0xd0, 0xe3, 0x3f, 0x01, 0x00, 0x00, 0xff, 0xff, 0xa2, 0x90, 0x04, 0x72, 0x2e, 0x07,
	0x00, 0x00,
}

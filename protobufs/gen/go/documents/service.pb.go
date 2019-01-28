// Code generated by protoc-gen-go. DO NOT EDIT.
// source: document/service.proto

package documentpb

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import _ "github.com/centrifuge/precise-proofs/proofs/proto"
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

type UpdateAccessTokenPayload struct {
	// The document which should contain the access token referenced below
	DelegatingDocumentIdentifier []byte `protobuf:"bytes,1,opt,name=delegating_document_identifier,json=delegatingDocumentIdentifier,proto3" json:"delegating_document_identifier,omitempty"`
	// The access token to be appended to the indicated document above
	AccessTokenParams    *AccessTokenParams `protobuf:"bytes,2,opt,name=access_token_params,json=accessTokenParams,proto3" json:"access_token_params,omitempty"`
	XXX_NoUnkeyedLiteral struct{}           `json:"-"`
	XXX_unrecognized     []byte             `json:"-"`
	XXX_sizecache        int32              `json:"-"`
}

func (m *UpdateAccessTokenPayload) Reset()         { *m = UpdateAccessTokenPayload{} }
func (m *UpdateAccessTokenPayload) String() string { return proto.CompactTextString(m) }
func (*UpdateAccessTokenPayload) ProtoMessage()    {}
func (*UpdateAccessTokenPayload) Descriptor() ([]byte, []int) {
	return fileDescriptor_service_da2bec74e6e4d4b2, []int{0}
}
func (m *UpdateAccessTokenPayload) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_UpdateAccessTokenPayload.Unmarshal(m, b)
}
func (m *UpdateAccessTokenPayload) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_UpdateAccessTokenPayload.Marshal(b, m, deterministic)
}
func (dst *UpdateAccessTokenPayload) XXX_Merge(src proto.Message) {
	xxx_messageInfo_UpdateAccessTokenPayload.Merge(dst, src)
}
func (m *UpdateAccessTokenPayload) XXX_Size() int {
	return xxx_messageInfo_UpdateAccessTokenPayload.Size(m)
}
func (m *UpdateAccessTokenPayload) XXX_DiscardUnknown() {
	xxx_messageInfo_UpdateAccessTokenPayload.DiscardUnknown(m)
}

var xxx_messageInfo_UpdateAccessTokenPayload proto.InternalMessageInfo

func (m *UpdateAccessTokenPayload) GetDelegatingDocumentIdentifier() []byte {
	if m != nil {
		return m.DelegatingDocumentIdentifier
	}
	return nil
}

func (m *UpdateAccessTokenPayload) GetAccessTokenParams() *AccessTokenParams {
	if m != nil {
		return m.AccessTokenParams
	}
	return nil
}

type AccessTokenParams struct {
	// The identity being granted access to the document
	Grantee string `protobuf:"bytes,4,opt,name=grantee,proto3" json:"grantee,omitempty"`
	// Original identifier of the document
	DocumentIdentifier   string   `protobuf:"bytes,2,opt,name=document_identifier,json=documentIdentifier,proto3" json:"document_identifier,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *AccessTokenParams) Reset()         { *m = AccessTokenParams{} }
func (m *AccessTokenParams) String() string { return proto.CompactTextString(m) }
func (*AccessTokenParams) ProtoMessage()    {}
func (*AccessTokenParams) Descriptor() ([]byte, []int) {
	return fileDescriptor_service_da2bec74e6e4d4b2, []int{1}
}
func (m *AccessTokenParams) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_AccessTokenParams.Unmarshal(m, b)
}
func (m *AccessTokenParams) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_AccessTokenParams.Marshal(b, m, deterministic)
}
func (dst *AccessTokenParams) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AccessTokenParams.Merge(dst, src)
}
func (m *AccessTokenParams) XXX_Size() int {
	return xxx_messageInfo_AccessTokenParams.Size(m)
}
func (m *AccessTokenParams) XXX_DiscardUnknown() {
	xxx_messageInfo_AccessTokenParams.DiscardUnknown(m)
}

var xxx_messageInfo_AccessTokenParams proto.InternalMessageInfo

func (m *AccessTokenParams) GetGrantee() string {
	if m != nil {
		return m.Grantee
	}
	return ""
}

func (m *AccessTokenParams) GetDocumentIdentifier() string {
	if m != nil {
		return m.DocumentIdentifier
	}
	return ""
}

type CreateDocumentProofRequest struct {
	Identifier           string   `protobuf:"bytes,1,opt,name=identifier,proto3" json:"identifier,omitempty"`
	Type                 string   `protobuf:"bytes,2,opt,name=type,proto3" json:"type,omitempty"`
	Fields               []string `protobuf:"bytes,3,rep,name=fields,proto3" json:"fields,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *CreateDocumentProofRequest) Reset()         { *m = CreateDocumentProofRequest{} }
func (m *CreateDocumentProofRequest) String() string { return proto.CompactTextString(m) }
func (*CreateDocumentProofRequest) ProtoMessage()    {}
func (*CreateDocumentProofRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_service_da2bec74e6e4d4b2, []int{2}
}
func (m *CreateDocumentProofRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CreateDocumentProofRequest.Unmarshal(m, b)
}
func (m *CreateDocumentProofRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CreateDocumentProofRequest.Marshal(b, m, deterministic)
}
func (dst *CreateDocumentProofRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CreateDocumentProofRequest.Merge(dst, src)
}
func (m *CreateDocumentProofRequest) XXX_Size() int {
	return xxx_messageInfo_CreateDocumentProofRequest.Size(m)
}
func (m *CreateDocumentProofRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_CreateDocumentProofRequest.DiscardUnknown(m)
}

var xxx_messageInfo_CreateDocumentProofRequest proto.InternalMessageInfo

func (m *CreateDocumentProofRequest) GetIdentifier() string {
	if m != nil {
		return m.Identifier
	}
	return ""
}

func (m *CreateDocumentProofRequest) GetType() string {
	if m != nil {
		return m.Type
	}
	return ""
}

func (m *CreateDocumentProofRequest) GetFields() []string {
	if m != nil {
		return m.Fields
	}
	return nil
}

// ResponseHeader contains a set of common fields for most document
type ResponseHeader struct {
	DocumentId           string   `protobuf:"bytes,1,opt,name=document_id,json=documentId,proto3" json:"document_id,omitempty"`
	VersionId            string   `protobuf:"bytes,2,opt,name=version_id,json=versionId,proto3" json:"version_id,omitempty"`
	State                string   `protobuf:"bytes,3,opt,name=state,proto3" json:"state,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ResponseHeader) Reset()         { *m = ResponseHeader{} }
func (m *ResponseHeader) String() string { return proto.CompactTextString(m) }
func (*ResponseHeader) ProtoMessage()    {}
func (*ResponseHeader) Descriptor() ([]byte, []int) {
	return fileDescriptor_service_da2bec74e6e4d4b2, []int{3}
}
func (m *ResponseHeader) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ResponseHeader.Unmarshal(m, b)
}
func (m *ResponseHeader) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ResponseHeader.Marshal(b, m, deterministic)
}
func (dst *ResponseHeader) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ResponseHeader.Merge(dst, src)
}
func (m *ResponseHeader) XXX_Size() int {
	return xxx_messageInfo_ResponseHeader.Size(m)
}
func (m *ResponseHeader) XXX_DiscardUnknown() {
	xxx_messageInfo_ResponseHeader.DiscardUnknown(m)
}

var xxx_messageInfo_ResponseHeader proto.InternalMessageInfo

func (m *ResponseHeader) GetDocumentId() string {
	if m != nil {
		return m.DocumentId
	}
	return ""
}

func (m *ResponseHeader) GetVersionId() string {
	if m != nil {
		return m.VersionId
	}
	return ""
}

func (m *ResponseHeader) GetState() string {
	if m != nil {
		return m.State
	}
	return ""
}

type DocumentProof struct {
	Header               *ResponseHeader `protobuf:"bytes,1,opt,name=header,proto3" json:"header,omitempty"`
	FieldProofs          []*Proof        `protobuf:"bytes,2,rep,name=field_proofs,json=fieldProofs,proto3" json:"field_proofs,omitempty"`
	XXX_NoUnkeyedLiteral struct{}        `json:"-"`
	XXX_unrecognized     []byte          `json:"-"`
	XXX_sizecache        int32           `json:"-"`
}

func (m *DocumentProof) Reset()         { *m = DocumentProof{} }
func (m *DocumentProof) String() string { return proto.CompactTextString(m) }
func (*DocumentProof) ProtoMessage()    {}
func (*DocumentProof) Descriptor() ([]byte, []int) {
	return fileDescriptor_service_da2bec74e6e4d4b2, []int{4}
}
func (m *DocumentProof) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DocumentProof.Unmarshal(m, b)
}
func (m *DocumentProof) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DocumentProof.Marshal(b, m, deterministic)
}
func (dst *DocumentProof) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DocumentProof.Merge(dst, src)
}
func (m *DocumentProof) XXX_Size() int {
	return xxx_messageInfo_DocumentProof.Size(m)
}
func (m *DocumentProof) XXX_DiscardUnknown() {
	xxx_messageInfo_DocumentProof.DiscardUnknown(m)
}

var xxx_messageInfo_DocumentProof proto.InternalMessageInfo

func (m *DocumentProof) GetHeader() *ResponseHeader {
	if m != nil {
		return m.Header
	}
	return nil
}

func (m *DocumentProof) GetFieldProofs() []*Proof {
	if m != nil {
		return m.FieldProofs
	}
	return nil
}

type Proof struct {
	Property string `protobuf:"bytes,1,opt,name=property,proto3" json:"property,omitempty"`
	Value    string `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
	Salt     string `protobuf:"bytes,3,opt,name=salt,proto3" json:"salt,omitempty"`
	// hash is filled if value & salt are not available
	Hash                 string   `protobuf:"bytes,4,opt,name=hash,proto3" json:"hash,omitempty"`
	SortedHashes         []string `protobuf:"bytes,5,rep,name=sorted_hashes,json=sortedHashes,proto3" json:"sorted_hashes,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Proof) Reset()         { *m = Proof{} }
func (m *Proof) String() string { return proto.CompactTextString(m) }
func (*Proof) ProtoMessage()    {}
func (*Proof) Descriptor() ([]byte, []int) {
	return fileDescriptor_service_da2bec74e6e4d4b2, []int{5}
}
func (m *Proof) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Proof.Unmarshal(m, b)
}
func (m *Proof) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Proof.Marshal(b, m, deterministic)
}
func (dst *Proof) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Proof.Merge(dst, src)
}
func (m *Proof) XXX_Size() int {
	return xxx_messageInfo_Proof.Size(m)
}
func (m *Proof) XXX_DiscardUnknown() {
	xxx_messageInfo_Proof.DiscardUnknown(m)
}

var xxx_messageInfo_Proof proto.InternalMessageInfo

func (m *Proof) GetProperty() string {
	if m != nil {
		return m.Property
	}
	return ""
}

func (m *Proof) GetValue() string {
	if m != nil {
		return m.Value
	}
	return ""
}

func (m *Proof) GetSalt() string {
	if m != nil {
		return m.Salt
	}
	return ""
}

func (m *Proof) GetHash() string {
	if m != nil {
		return m.Hash
	}
	return ""
}

func (m *Proof) GetSortedHashes() []string {
	if m != nil {
		return m.SortedHashes
	}
	return nil
}

type CreateDocumentProofForVersionRequest struct {
	Identifier           string   `protobuf:"bytes,1,opt,name=identifier,proto3" json:"identifier,omitempty"`
	Type                 string   `protobuf:"bytes,2,opt,name=type,proto3" json:"type,omitempty"`
	Version              string   `protobuf:"bytes,3,opt,name=version,proto3" json:"version,omitempty"`
	Fields               []string `protobuf:"bytes,4,rep,name=fields,proto3" json:"fields,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *CreateDocumentProofForVersionRequest) Reset()         { *m = CreateDocumentProofForVersionRequest{} }
func (m *CreateDocumentProofForVersionRequest) String() string { return proto.CompactTextString(m) }
func (*CreateDocumentProofForVersionRequest) ProtoMessage()    {}
func (*CreateDocumentProofForVersionRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_service_da2bec74e6e4d4b2, []int{6}
}
func (m *CreateDocumentProofForVersionRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CreateDocumentProofForVersionRequest.Unmarshal(m, b)
}
func (m *CreateDocumentProofForVersionRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CreateDocumentProofForVersionRequest.Marshal(b, m, deterministic)
}
func (dst *CreateDocumentProofForVersionRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CreateDocumentProofForVersionRequest.Merge(dst, src)
}
func (m *CreateDocumentProofForVersionRequest) XXX_Size() int {
	return xxx_messageInfo_CreateDocumentProofForVersionRequest.Size(m)
}
func (m *CreateDocumentProofForVersionRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_CreateDocumentProofForVersionRequest.DiscardUnknown(m)
}

var xxx_messageInfo_CreateDocumentProofForVersionRequest proto.InternalMessageInfo

func (m *CreateDocumentProofForVersionRequest) GetIdentifier() string {
	if m != nil {
		return m.Identifier
	}
	return ""
}

func (m *CreateDocumentProofForVersionRequest) GetType() string {
	if m != nil {
		return m.Type
	}
	return ""
}

func (m *CreateDocumentProofForVersionRequest) GetVersion() string {
	if m != nil {
		return m.Version
	}
	return ""
}

func (m *CreateDocumentProofForVersionRequest) GetFields() []string {
	if m != nil {
		return m.Fields
	}
	return nil
}

func init() {
	proto.RegisterType((*UpdateAccessTokenPayload)(nil), "document.UpdateAccessTokenPayload")
	proto.RegisterType((*AccessTokenParams)(nil), "document.AccessTokenParams")
	proto.RegisterType((*CreateDocumentProofRequest)(nil), "document.CreateDocumentProofRequest")
	proto.RegisterType((*ResponseHeader)(nil), "document.ResponseHeader")
	proto.RegisterType((*DocumentProof)(nil), "document.DocumentProof")
	proto.RegisterType((*Proof)(nil), "document.Proof")
	proto.RegisterType((*CreateDocumentProofForVersionRequest)(nil), "document.CreateDocumentProofForVersionRequest")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// DocumentServiceClient is the client API for DocumentService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type DocumentServiceClient interface {
	CreateDocumentProof(ctx context.Context, in *CreateDocumentProofRequest, opts ...grpc.CallOption) (*DocumentProof, error)
	CreateDocumentProofForVersion(ctx context.Context, in *CreateDocumentProofForVersionRequest, opts ...grpc.CallOption) (*DocumentProof, error)
}

type documentServiceClient struct {
	cc *grpc.ClientConn
}

func NewDocumentServiceClient(cc *grpc.ClientConn) DocumentServiceClient {
	return &documentServiceClient{cc}
}

func (c *documentServiceClient) CreateDocumentProof(ctx context.Context, in *CreateDocumentProofRequest, opts ...grpc.CallOption) (*DocumentProof, error) {
	out := new(DocumentProof)
	err := c.cc.Invoke(ctx, "/document.DocumentService/CreateDocumentProof", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *documentServiceClient) CreateDocumentProofForVersion(ctx context.Context, in *CreateDocumentProofForVersionRequest, opts ...grpc.CallOption) (*DocumentProof, error) {
	out := new(DocumentProof)
	err := c.cc.Invoke(ctx, "/document.DocumentService/CreateDocumentProofForVersion", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// DocumentServiceServer is the server API for DocumentService service.
type DocumentServiceServer interface {
	CreateDocumentProof(context.Context, *CreateDocumentProofRequest) (*DocumentProof, error)
	CreateDocumentProofForVersion(context.Context, *CreateDocumentProofForVersionRequest) (*DocumentProof, error)
}

func RegisterDocumentServiceServer(s *grpc.Server, srv DocumentServiceServer) {
	s.RegisterService(&_DocumentService_serviceDesc, srv)
}

func _DocumentService_CreateDocumentProof_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateDocumentProofRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DocumentServiceServer).CreateDocumentProof(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/document.DocumentService/CreateDocumentProof",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DocumentServiceServer).CreateDocumentProof(ctx, req.(*CreateDocumentProofRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _DocumentService_CreateDocumentProofForVersion_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateDocumentProofForVersionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DocumentServiceServer).CreateDocumentProofForVersion(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/document.DocumentService/CreateDocumentProofForVersion",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DocumentServiceServer).CreateDocumentProofForVersion(ctx, req.(*CreateDocumentProofForVersionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _DocumentService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "document.DocumentService",
	HandlerType: (*DocumentServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateDocumentProof",
			Handler:    _DocumentService_CreateDocumentProof_Handler,
		},
		{
			MethodName: "CreateDocumentProofForVersion",
			Handler:    _DocumentService_CreateDocumentProofForVersion_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "document/service.proto",
}

func init() { proto.RegisterFile("document/service.proto", fileDescriptor_service_da2bec74e6e4d4b2) }

var fileDescriptor_service_da2bec74e6e4d4b2 = []byte{
	// 700 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xa4, 0x54, 0xcd, 0x6e, 0xd3, 0x40,
	0x10, 0x96, 0x93, 0xfe, 0x65, 0x93, 0x52, 0x75, 0x8b, 0xa8, 0x15, 0xda, 0xb2, 0x98, 0x0a, 0xa2,
	0x42, 0x63, 0x14, 0x6e, 0xdc, 0x5a, 0x2a, 0xd4, 0x8a, 0x4b, 0x64, 0x28, 0x48, 0x1c, 0x88, 0xb6,
	0xf6, 0xc4, 0x31, 0xb8, 0x5e, 0xb3, 0xbb, 0x09, 0x8a, 0x2a, 0x0e, 0x70, 0xe0, 0x02, 0xa7, 0xf2,
	0x02, 0xbc, 0x00, 0x4f, 0xc3, 0x85, 0x07, 0xe0, 0x29, 0x38, 0x21, 0xef, 0xae, 0xe3, 0xa4, 0x69,
	0xcb, 0xa1, 0xa7, 0xec, 0xcc, 0x7c, 0x9e, 0xfd, 0xbe, 0x6f, 0x36, 0x83, 0x56, 0x03, 0xe6, 0xf7,
	0x8f, 0x21, 0x91, 0xc2, 0x15, 0xc0, 0x07, 0x91, 0x0f, 0xcd, 0x94, 0x33, 0xc9, 0xf0, 0x42, 0x5e,
	0xa8, 0xaf, 0x85, 0x8c, 0x85, 0x31, 0xb8, 0x34, 0x8d, 0x5c, 0x9a, 0x24, 0x4c, 0x52, 0x19, 0xb1,
	0x44, 0x68, 0x5c, 0xfd, 0x5e, 0xca, 0xc1, 0x8f, 0x04, 0x6c, 0xa7, 0x9c, 0xb1, 0xae, 0x70, 0x8b,
	0x1f, 0xc9, 0x74, 0x60, 0x80, 0x0f, 0xd4, 0x8f, 0xbf, 0x1d, 0x42, 0xb2, 0x2d, 0x3e, 0xd0, 0x30,
	0x04, 0xee, 0xb2, 0x54, 0xb5, 0x9a, 0x6e, 0xeb, 0xfc, 0xb4, 0x90, 0x7d, 0x98, 0x06, 0x54, 0xc2,
	0x8e, 0xef, 0x83, 0x10, 0x2f, 0xd8, 0x3b, 0x48, 0xda, 0x74, 0x18, 0x33, 0x1a, 0xe0, 0x3d, 0xb4,
	0x11, 0x40, 0x0c, 0x21, 0x95, 0x51, 0x12, 0x76, 0x72, 0xa2, 0x9d, 0x28, 0x80, 0x44, 0x46, 0xdd,
	0x08, 0xb8, 0x6d, 0x11, 0xab, 0x51, 0xf3, 0xd6, 0x0a, 0xd4, 0x9e, 0x01, 0x1d, 0x8c, 0x30, 0xf8,
	0x19, 0x5a, 0xa1, 0xaa, 0x77, 0x47, 0x66, 0xcd, 0x3b, 0x29, 0xe5, 0xf4, 0x58, 0xd8, 0x25, 0x62,
	0x35, 0xaa, 0xad, 0x9b, 0xcd, 0xbc, 0x6d, 0x73, 0x82, 0x40, 0x06, 0xf1, 0x96, 0xe9, 0xd9, 0x94,
	0xf3, 0x06, 0x2d, 0x4f, 0xe1, 0xb0, 0x8d, 0xe6, 0x43, 0x4e, 0x13, 0x09, 0x60, 0xcf, 0x10, 0xab,
	0x51, 0xf1, 0xf2, 0x10, 0xbb, 0x68, 0xe5, 0x3c, 0xda, 0x25, 0x85, 0xc2, 0xc1, 0x14, 0x59, 0xa7,
	0x87, 0xea, 0x4f, 0x38, 0x50, 0x09, 0xb9, 0x90, 0x76, 0x66, 0xad, 0x07, 0xef, 0xfb, 0x20, 0x24,
	0xde, 0x40, 0xe8, 0x8c, 0xf8, 0x8a, 0x37, 0x96, 0xc1, 0x18, 0xcd, 0xc8, 0x61, 0x0a, 0xa6, 0xbf,
	0x3a, 0xe3, 0x1b, 0x68, 0xae, 0x1b, 0x41, 0x1c, 0x08, 0xbb, 0x4c, 0xca, 0x8d, 0x8a, 0x67, 0x22,
	0xa7, 0x8b, 0xae, 0x79, 0x20, 0x52, 0x96, 0x08, 0xd8, 0x07, 0x1a, 0x00, 0xc7, 0xb7, 0x50, 0x75,
	0x8c, 0x6c, 0xde, 0xbe, 0x20, 0x89, 0xd7, 0x11, 0x1a, 0x00, 0x17, 0x11, 0x4b, 0xb2, 0xba, 0xbe,
	0xa4, 0x62, 0x32, 0x07, 0x01, 0xbe, 0x8e, 0x66, 0x85, 0xa4, 0x12, 0xec, 0xb2, 0xaa, 0xe8, 0xc0,
	0xe9, 0xa3, 0xc5, 0x09, 0x2d, 0xf8, 0x21, 0x9a, 0xeb, 0xa9, 0x0b, 0xd5, 0x0d, 0xd5, 0x96, 0x5d,
	0x8c, 0x60, 0x92, 0x90, 0x67, 0x70, 0xb8, 0x85, 0x6a, 0x8a, 0x74, 0x47, 0x3f, 0x3a, 0xbb, 0x44,
	0xca, 0x8d, 0x6a, 0x6b, 0xa9, 0xf8, 0x4e, 0x9b, 0x54, 0x55, 0x20, 0x75, 0x16, 0xce, 0x17, 0x0b,
	0xcd, 0xea, 0xfb, 0xea, 0x68, 0x21, 0xe5, 0x2c, 0x05, 0x2e, 0x87, 0x46, 0xd3, 0x28, 0xce, 0x28,
	0x0f, 0x68, 0xdc, 0xcf, 0x1d, 0xd3, 0x41, 0x66, 0xa3, 0xa0, 0xb1, 0x34, 0x3a, 0xd4, 0x39, 0xcb,
	0xf5, 0xa8, 0xe8, 0x99, 0x01, 0xab, 0x33, 0xbe, 0x83, 0x16, 0x05, 0xe3, 0x12, 0x82, 0x4e, 0x16,
	0x82, 0xb0, 0x67, 0x95, 0xc3, 0x35, 0x9d, 0xdc, 0x57, 0x39, 0xe7, 0x9b, 0x85, 0x36, 0xcf, 0x19,
	0xe9, 0x53, 0xc6, 0x5f, 0x6a, 0xe7, 0xae, 0x32, 0x5c, 0x1b, 0xcd, 0x1b, 0xff, 0x0d, 0xd9, 0x3c,
	0x1c, 0x1b, 0xfb, 0xcc, 0xf8, 0xd8, 0x5b, 0x7f, 0xcb, 0x68, 0x29, 0x27, 0xf2, 0x5c, 0x6f, 0x02,
	0xfc, 0xdb, 0x42, 0x2b, 0xe7, 0x50, 0xc4, 0x9b, 0x85, 0xc3, 0x17, 0x3f, 0xca, 0xfa, 0x6a, 0x81,
	0x9a, 0xa8, 0x3b, 0x9f, 0xac, 0xd3, 0x9d, 0x57, 0xf5, 0x43, 0xfd, 0xa9, 0x20, 0x94, 0xc4, 0x91,
	0x90, 0x84, 0x75, 0x89, 0x59, 0x25, 0x44, 0x8f, 0x93, 0x74, 0x19, 0x27, 0xb2, 0x07, 0x44, 0xa4,
	0xe0, 0x67, 0x52, 0x03, 0xa2, 0xb9, 0x66, 0xd0, 0x2c, 0x9f, 0xb7, 0x27, 0x61, 0x34, 0x80, 0x84,
	0x1c, 0x0d, 0xc9, 0xc1, 0xde, 0xe7, 0x5f, 0x7f, 0xbe, 0x97, 0x6e, 0x3b, 0x6b, 0x6e, 0x5e, 0x74,
	0x4f, 0x0a, 0xab, 0x3e, 0xea, 0x85, 0xf4, 0xd8, 0xda, 0xc2, 0x5f, 0x4b, 0x68, 0xfd, 0x52, 0xf7,
	0x71, 0xf3, 0x52, 0x91, 0x53, 0x63, 0xba, 0x58, 0xee, 0x0f, 0xeb, 0x74, 0x27, 0xae, 0xbf, 0xbd,
	0xb2, 0x5c, 0xad, 0xd2, 0xcc, 0xf1, 0xbf, 0x1e, 0xdc, 0x77, 0xee, 0x5e, 0xe0, 0xc1, 0x89, 0x69,
	0x51, 0xb8, 0xb1, 0xbb, 0x85, 0x6a, 0x3e, 0x3b, 0x1e, 0x09, 0xd8, 0xad, 0x99, 0x17, 0xd0, 0xce,
	0x76, 0x71, 0xdb, 0x7a, 0x3d, 0xfa, 0xb3, 0xa7, 0x47, 0x47, 0x73, 0x6a, 0x41, 0x3f, 0xfa, 0x17,
	0x00, 0x00, 0xff, 0xff, 0x2e, 0x89, 0xab, 0x8b, 0x3a, 0x06, 0x00, 0x00,
}

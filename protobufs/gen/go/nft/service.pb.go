// Code generated by protoc-gen-go. DO NOT EDIT.
// source: nft/service.proto

package nftpb

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
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

type ResponseHeader struct {
	TransactionId        string   `protobuf:"bytes,5,opt,name=transaction_id,json=transactionId,proto3" json:"transaction_id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ResponseHeader) Reset()         { *m = ResponseHeader{} }
func (m *ResponseHeader) String() string { return proto.CompactTextString(m) }
func (*ResponseHeader) ProtoMessage()    {}
func (*ResponseHeader) Descriptor() ([]byte, []int) {
	return fileDescriptor_service_c824afb99e070ed1, []int{0}
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

func (m *ResponseHeader) GetTransactionId() string {
	if m != nil {
		return m.TransactionId
	}
	return ""
}

type NFTPaymentObligationRequest struct {
	// Invoice Document identifier
	Identifier string `protobuf:"bytes,1,opt,name=identifier,proto3" json:"identifier,omitempty"`
	// Deposit address for NFT Token created
	DepositAddress       string   `protobuf:"bytes,2,opt,name=deposit_address,json=depositAddress,proto3" json:"deposit_address,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *NFTPaymentObligationRequest) Reset()         { *m = NFTPaymentObligationRequest{} }
func (m *NFTPaymentObligationRequest) String() string { return proto.CompactTextString(m) }
func (*NFTPaymentObligationRequest) ProtoMessage()    {}
func (*NFTPaymentObligationRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_service_c824afb99e070ed1, []int{1}
}
func (m *NFTPaymentObligationRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_NFTPaymentObligationRequest.Unmarshal(m, b)
}
func (m *NFTPaymentObligationRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_NFTPaymentObligationRequest.Marshal(b, m, deterministic)
}
func (dst *NFTPaymentObligationRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_NFTPaymentObligationRequest.Merge(dst, src)
}
func (m *NFTPaymentObligationRequest) XXX_Size() int {
	return xxx_messageInfo_NFTPaymentObligationRequest.Size(m)
}
func (m *NFTPaymentObligationRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_NFTPaymentObligationRequest.DiscardUnknown(m)
}

var xxx_messageInfo_NFTPaymentObligationRequest proto.InternalMessageInfo

func (m *NFTPaymentObligationRequest) GetIdentifier() string {
	if m != nil {
		return m.Identifier
	}
	return ""
}

func (m *NFTPaymentObligationRequest) GetDepositAddress() string {
	if m != nil {
		return m.DepositAddress
	}
	return ""
}

type NFTMintRequest struct {
	// Document identifier
	Identifier string `protobuf:"bytes,1,opt,name=identifier,proto3" json:"identifier,omitempty"`
	// The contract address of the registry where the token should be minted
	RegistryAddress string   `protobuf:"bytes,2,opt,name=registry_address,json=registryAddress,proto3" json:"registry_address,omitempty"`
	DepositAddress  string   `protobuf:"bytes,3,opt,name=deposit_address,json=depositAddress,proto3" json:"deposit_address,omitempty"`
	ProofFields     []string `protobuf:"bytes,4,rep,name=proof_fields,json=proofFields,proto3" json:"proof_fields,omitempty"`
	// proof that nft is part of document
	SubmitTokenProof bool `protobuf:"varint,5,opt,name=submit_token_proof,json=submitTokenProof,proto3" json:"submit_token_proof,omitempty"`
	// proof that nft owner can access the document if nft_grant_access is true
	SubmitNftOwnerAccessProof bool `protobuf:"varint,7,opt,name=submit_nft_owner_access_proof,json=submitNftOwnerAccessProof,proto3" json:"submit_nft_owner_access_proof,omitempty"`
	// grant nft read access to the document
	GrantNftAccess       bool     `protobuf:"varint,8,opt,name=grant_nft_access,json=grantNftAccess,proto3" json:"grant_nft_access,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *NFTMintRequest) Reset()         { *m = NFTMintRequest{} }
func (m *NFTMintRequest) String() string { return proto.CompactTextString(m) }
func (*NFTMintRequest) ProtoMessage()    {}
func (*NFTMintRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_service_c824afb99e070ed1, []int{2}
}
func (m *NFTMintRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_NFTMintRequest.Unmarshal(m, b)
}
func (m *NFTMintRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_NFTMintRequest.Marshal(b, m, deterministic)
}
func (dst *NFTMintRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_NFTMintRequest.Merge(dst, src)
}
func (m *NFTMintRequest) XXX_Size() int {
	return xxx_messageInfo_NFTMintRequest.Size(m)
}
func (m *NFTMintRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_NFTMintRequest.DiscardUnknown(m)
}

var xxx_messageInfo_NFTMintRequest proto.InternalMessageInfo

func (m *NFTMintRequest) GetIdentifier() string {
	if m != nil {
		return m.Identifier
	}
	return ""
}

func (m *NFTMintRequest) GetRegistryAddress() string {
	if m != nil {
		return m.RegistryAddress
	}
	return ""
}

func (m *NFTMintRequest) GetDepositAddress() string {
	if m != nil {
		return m.DepositAddress
	}
	return ""
}

func (m *NFTMintRequest) GetProofFields() []string {
	if m != nil {
		return m.ProofFields
	}
	return nil
}

func (m *NFTMintRequest) GetSubmitTokenProof() bool {
	if m != nil {
		return m.SubmitTokenProof
	}
	return false
}

func (m *NFTMintRequest) GetSubmitNftOwnerAccessProof() bool {
	if m != nil {
		return m.SubmitNftOwnerAccessProof
	}
	return false
}

func (m *NFTMintRequest) GetGrantNftAccess() bool {
	if m != nil {
		return m.GrantNftAccess
	}
	return false
}

type NFTMintResponse struct {
	Header               *ResponseHeader `protobuf:"bytes,1,opt,name=header,proto3" json:"header,omitempty"`
	TokenId              string          `protobuf:"bytes,2,opt,name=token_id,json=tokenId,proto3" json:"token_id,omitempty"`
	XXX_NoUnkeyedLiteral struct{}        `json:"-"`
	XXX_unrecognized     []byte          `json:"-"`
	XXX_sizecache        int32           `json:"-"`
}

func (m *NFTMintResponse) Reset()         { *m = NFTMintResponse{} }
func (m *NFTMintResponse) String() string { return proto.CompactTextString(m) }
func (*NFTMintResponse) ProtoMessage()    {}
func (*NFTMintResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_service_c824afb99e070ed1, []int{3}
}
func (m *NFTMintResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_NFTMintResponse.Unmarshal(m, b)
}
func (m *NFTMintResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_NFTMintResponse.Marshal(b, m, deterministic)
}
func (dst *NFTMintResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_NFTMintResponse.Merge(dst, src)
}
func (m *NFTMintResponse) XXX_Size() int {
	return xxx_messageInfo_NFTMintResponse.Size(m)
}
func (m *NFTMintResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_NFTMintResponse.DiscardUnknown(m)
}

var xxx_messageInfo_NFTMintResponse proto.InternalMessageInfo

func (m *NFTMintResponse) GetHeader() *ResponseHeader {
	if m != nil {
		return m.Header
	}
	return nil
}

func (m *NFTMintResponse) GetTokenId() string {
	if m != nil {
		return m.TokenId
	}
	return ""
}

func init() {
	proto.RegisterType((*ResponseHeader)(nil), "nft.ResponseHeader")
	proto.RegisterType((*NFTPaymentObligationRequest)(nil), "nft.NFTPaymentObligationRequest")
	proto.RegisterType((*NFTMintRequest)(nil), "nft.NFTMintRequest")
	proto.RegisterType((*NFTMintResponse)(nil), "nft.NFTMintResponse")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// NFTServiceClient is the client API for NFTService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type NFTServiceClient interface {
	MintNFT(ctx context.Context, in *NFTMintRequest, opts ...grpc.CallOption) (*NFTMintResponse, error)
	MintPaymentObligationNFT(ctx context.Context, in *NFTPaymentObligationRequest, opts ...grpc.CallOption) (*NFTMintResponse, error)
}

type nFTServiceClient struct {
	cc *grpc.ClientConn
}

func NewNFTServiceClient(cc *grpc.ClientConn) NFTServiceClient {
	return &nFTServiceClient{cc}
}

func (c *nFTServiceClient) MintNFT(ctx context.Context, in *NFTMintRequest, opts ...grpc.CallOption) (*NFTMintResponse, error) {
	out := new(NFTMintResponse)
	err := c.cc.Invoke(ctx, "/nft.NFTService/MintNFT", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *nFTServiceClient) MintPaymentObligationNFT(ctx context.Context, in *NFTPaymentObligationRequest, opts ...grpc.CallOption) (*NFTMintResponse, error) {
	out := new(NFTMintResponse)
	err := c.cc.Invoke(ctx, "/nft.NFTService/MintPaymentObligationNFT", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// NFTServiceServer is the server API for NFTService service.
type NFTServiceServer interface {
	MintNFT(context.Context, *NFTMintRequest) (*NFTMintResponse, error)
	MintPaymentObligationNFT(context.Context, *NFTPaymentObligationRequest) (*NFTMintResponse, error)
}

func RegisterNFTServiceServer(s *grpc.Server, srv NFTServiceServer) {
	s.RegisterService(&_NFTService_serviceDesc, srv)
}

func _NFTService_MintNFT_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(NFTMintRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NFTServiceServer).MintNFT(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/nft.NFTService/MintNFT",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NFTServiceServer).MintNFT(ctx, req.(*NFTMintRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _NFTService_MintPaymentObligationNFT_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(NFTPaymentObligationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NFTServiceServer).MintPaymentObligationNFT(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/nft.NFTService/MintPaymentObligationNFT",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NFTServiceServer).MintPaymentObligationNFT(ctx, req.(*NFTPaymentObligationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _NFTService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "nft.NFTService",
	HandlerType: (*NFTServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "MintNFT",
			Handler:    _NFTService_MintNFT_Handler,
		},
		{
			MethodName: "MintPaymentObligationNFT",
			Handler:    _NFTService_MintPaymentObligationNFT_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "nft/service.proto",
}

func init() { proto.RegisterFile("nft/service.proto", fileDescriptor_service_c824afb99e070ed1) }

var fileDescriptor_service_c824afb99e070ed1 = []byte{
	// 576 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x53, 0xcb, 0x6e, 0x13, 0x3d,
	0x18, 0x55, 0xd2, 0xbf, 0x4d, 0xeb, 0xf6, 0x4f, 0x83, 0x61, 0x91, 0x86, 0x8b, 0xcc, 0x88, 0x4b,
	0x5a, 0xda, 0x8e, 0x54, 0x16, 0x15, 0x5d, 0x91, 0x82, 0x46, 0x74, 0xc1, 0x34, 0x1a, 0x66, 0x03,
	0x9b, 0x91, 0x33, 0xf3, 0x79, 0xb0, 0xda, 0xd8, 0x83, 0xed, 0xb4, 0xaa, 0x10, 0x1b, 0x1e, 0xa1,
	0x3c, 0x10, 0x62, 0xc5, 0x03, 0xf0, 0x0a, 0x3c, 0x08, 0xb2, 0x3d, 0xbd, 0xd1, 0x46, 0x62, 0x15,
	0xe5, 0xf8, 0x9c, 0xe3, 0x33, 0x9f, 0xcf, 0x87, 0x6e, 0x09, 0x66, 0x42, 0x0d, 0xea, 0x88, 0xe7,
	0xb0, 0x59, 0x29, 0x69, 0x24, 0x9e, 0x11, 0xcc, 0xf4, 0xee, 0x95, 0x52, 0x96, 0x87, 0x10, 0xd2,
	0x8a, 0x87, 0x54, 0x08, 0x69, 0xa8, 0xe1, 0x52, 0x68, 0x4f, 0xe9, 0xad, 0xbb, 0x9f, 0x7c, 0xa3,
	0x04, 0xb1, 0xa1, 0x8f, 0x69, 0x59, 0x82, 0x0a, 0x65, 0xe5, 0x18, 0xd7, 0xd9, 0xc1, 0x36, 0x6a,
	0x27, 0xa0, 0x2b, 0x29, 0x34, 0xbc, 0x01, 0x5a, 0x80, 0xc2, 0x8f, 0x51, 0xdb, 0x28, 0x2a, 0x34,
	0xcd, 0x2d, 0x2f, 0xe3, 0x45, 0x77, 0x96, 0x34, 0xfa, 0x0b, 0xc9, 0xff, 0x97, 0xd0, 0xbd, 0x22,
	0x60, 0xe8, 0x6e, 0x1c, 0xa5, 0x43, 0x7a, 0x32, 0x06, 0x61, 0xf6, 0x47, 0x87, 0xbc, 0x74, 0xbe,
	0x09, 0x7c, 0x9a, 0x80, 0x36, 0xf8, 0x01, 0x42, 0xbc, 0x00, 0x61, 0x38, 0xe3, 0xa0, 0xba, 0x0d,
	0xe7, 0x70, 0x09, 0xc1, 0x4f, 0xd1, 0x72, 0x01, 0x95, 0xd4, 0xdc, 0x64, 0xb4, 0x28, 0x14, 0x68,
	0xdd, 0x6d, 0x3a, 0x52, 0xbb, 0x86, 0x07, 0x1e, 0x0d, 0x7e, 0x34, 0x51, 0x3b, 0x8e, 0xd2, 0xb7,
	0x5c, 0x98, 0x7f, 0xf5, 0x5e, 0x45, 0x1d, 0x05, 0x25, 0xd7, 0x46, 0x9d, 0xfc, 0x65, 0xbe, 0x7c,
	0x86, 0xd7, 0xee, 0x37, 0xc5, 0x98, 0xb9, 0x29, 0x06, 0x7e, 0x88, 0x96, 0x2a, 0x25, 0x25, 0xcb,
	0x18, 0x87, 0xc3, 0x42, 0x77, 0xff, 0x23, 0x33, 0xfd, 0x85, 0x64, 0xd1, 0x61, 0x91, 0x83, 0xf0,
	0x3a, 0xc2, 0x7a, 0x32, 0x1a, 0x73, 0x93, 0x19, 0x79, 0x00, 0x22, 0x73, 0x67, 0x6e, 0x78, 0xf3,
	0x49, 0xc7, 0x9f, 0xa4, 0xf6, 0x60, 0x68, 0x71, 0xfc, 0x12, 0xdd, 0xaf, 0xd9, 0x82, 0x99, 0x4c,
	0x1e, 0x0b, 0x50, 0x19, 0xcd, 0x73, 0xd0, 0xba, 0x16, 0xb6, 0x9c, 0x70, 0xc5, 0x93, 0x62, 0x66,
	0xf6, 0x2d, 0x65, 0xe0, 0x18, 0xde, 0xa1, 0x8f, 0x3a, 0xa5, 0xa2, 0xc2, 0x1b, 0x78, 0x69, 0x77,
	0xde, 0x89, 0xda, 0x0e, 0x8f, 0x99, 0xf1, 0xf4, 0xe0, 0x3d, 0x5a, 0x3e, 0x1f, 0xa1, 0x7f, 0x6b,
	0xfc, 0x0c, 0xcd, 0x7d, 0x74, 0xef, 0xed, 0xe6, 0xb7, 0xb8, 0x75, 0x7b, 0x53, 0x30, 0xb3, 0x79,
	0xb5, 0x0a, 0x49, 0x4d, 0xc1, 0x2b, 0x68, 0xde, 0x7f, 0x12, 0x2f, 0xea, 0x41, 0xb6, 0xdc, 0xff,
	0xbd, 0x62, 0xeb, 0x67, 0x13, 0xa1, 0x38, 0x4a, 0xdf, 0xf9, 0x96, 0xe2, 0x63, 0xd4, 0xb2, 0xd7,
	0xc4, 0x51, 0x8a, 0xbd, 0xe3, 0xd5, 0xa7, 0xeb, 0xdd, 0xb9, 0x0a, 0xfa, 0xdb, 0x82, 0xc1, 0xe9,
	0xa0, 0xdf, 0x7b, 0x62, 0x21, 0x42, 0x05, 0x89, 0xa3, 0x94, 0x30, 0x25, 0xc7, 0x84, 0x92, 0x57,
	0x20, 0x8c, 0xe2, 0x6c, 0x52, 0x02, 0x79, 0x2d, 0xf3, 0x89, 0xad, 0xda, 0xd7, 0x5f, 0xbf, 0xbf,
	0x35, 0x3b, 0xc1, 0x62, 0xe8, 0x12, 0x84, 0x63, 0x2e, 0xcc, 0x4e, 0x63, 0x0d, 0x7f, 0x6f, 0xa0,
	0xae, 0x35, 0xb8, 0x56, 0x48, 0x1b, 0x85, 0x9c, 0xdd, 0x3a, 0xad, 0xae, 0x53, 0x72, 0x1d, 0x9c,
	0x0e, 0x5e, 0xf4, 0xb6, 0x7d, 0x2e, 0x52, 0x6b, 0xc9, 0x85, 0x78, 0x4a, 0xd2, 0x3d, 0x71, 0x24,
	0x79, 0x0e, 0x2e, 0xe8, 0x6a, 0xf0, 0x28, 0xb4, 0x5b, 0x5c, 0x79, 0x71, 0x26, 0xcf, 0xc5, 0xe1,
	0xe7, 0x8b, 0xc6, 0x7e, 0xd9, 0x69, 0xac, 0xed, 0x12, 0xd4, 0xca, 0xe5, 0xd8, 0xe6, 0xd8, 0x5d,
	0xaa, 0xc7, 0x39, 0xb4, 0x2b, 0x3a, 0x6c, 0x7c, 0x98, 0x15, 0xcc, 0x54, 0xa3, 0xd1, 0x9c, 0x5b,
	0xd9, 0xe7, 0x7f, 0x02, 0x00, 0x00, 0xff, 0xff, 0x71, 0xb3, 0x07, 0x46, 0x18, 0x04, 0x00, 0x00,
}

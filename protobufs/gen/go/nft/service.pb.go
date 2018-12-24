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

type NFTMintRequest struct {
	// Document identifier
	Identifier string `protobuf:"bytes,1,opt,name=identifier,proto3" json:"identifier,omitempty"`
	// The contract address of the registry where the token should be minted
	RegistryAddress      string   `protobuf:"bytes,2,opt,name=registry_address,json=registryAddress,proto3" json:"registry_address,omitempty"`
	DepositAddress       string   `protobuf:"bytes,3,opt,name=deposit_address,json=depositAddress,proto3" json:"deposit_address,omitempty"`
	ProofFields          []string `protobuf:"bytes,4,rep,name=proof_fields,json=proofFields,proto3" json:"proof_fields,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *NFTMintRequest) Reset()         { *m = NFTMintRequest{} }
func (m *NFTMintRequest) String() string { return proto.CompactTextString(m) }
func (*NFTMintRequest) ProtoMessage()    {}
func (*NFTMintRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_service_66d4e2152d82b027, []int{0}
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

type NFTMintResponse struct {
	TokenId              string   `protobuf:"bytes,1,opt,name=token_id,json=tokenId,proto3" json:"token_id,omitempty"`
	TransactionId        string   `protobuf:"bytes,2,opt,name=transaction_id,json=transactionId,proto3" json:"transaction_id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *NFTMintResponse) Reset()         { *m = NFTMintResponse{} }
func (m *NFTMintResponse) String() string { return proto.CompactTextString(m) }
func (*NFTMintResponse) ProtoMessage()    {}
func (*NFTMintResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_service_66d4e2152d82b027, []int{1}
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

func (m *NFTMintResponse) GetTokenId() string {
	if m != nil {
		return m.TokenId
	}
	return ""
}

func (m *NFTMintResponse) GetTransactionId() string {
	if m != nil {
		return m.TransactionId
	}
	return ""
}

func init() {
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

// NFTServiceServer is the server API for NFTService service.
type NFTServiceServer interface {
	MintNFT(context.Context, *NFTMintRequest) (*NFTMintResponse, error)
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

var _NFTService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "nft.NFTService",
	HandlerType: (*NFTServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "MintNFT",
			Handler:    _NFTService_MintNFT_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "nft/service.proto",
}

func init() { proto.RegisterFile("nft/service.proto", fileDescriptor_service_66d4e2152d82b027) }

var fileDescriptor_service_66d4e2152d82b027 = []byte{
	// 377 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x64, 0x91, 0x41, 0x8e, 0xd3, 0x30,
	0x14, 0x40, 0x95, 0x29, 0x50, 0xc6, 0x33, 0xb4, 0x83, 0x61, 0x51, 0x22, 0x84, 0x4c, 0x24, 0xa0,
	0x20, 0xa6, 0x91, 0x60, 0xc7, 0xae, 0x03, 0x8a, 0x34, 0x0b, 0xa2, 0x51, 0x9b, 0x15, 0x9b, 0xca,
	0x8d, 0xbf, 0x23, 0x8b, 0xe6, 0x3b, 0xd8, 0xbf, 0x54, 0x6c, 0x91, 0xb8, 0x00, 0x1c, 0x82, 0x03,
	0x71, 0x05, 0x0e, 0x82, 0xe2, 0xb4, 0x55, 0x2b, 0x56, 0x51, 0x9e, 0x9e, 0xec, 0xe7, 0xff, 0xd9,
	0x7d, 0xd4, 0x94, 0x7a, 0x70, 0x5f, 0x4d, 0x09, 0x93, 0xc6, 0x59, 0xb2, 0xbc, 0x87, 0x9a, 0xe2,
	0xc7, 0x95, 0xb5, 0xd5, 0x0a, 0x52, 0xd9, 0x98, 0x54, 0x22, 0x5a, 0x92, 0x64, 0x2c, 0xfa, 0x4e,
	0x89, 0x5f, 0x87, 0x4f, 0x79, 0x59, 0x01, 0x5e, 0xfa, 0x8d, 0xac, 0x2a, 0x70, 0xa9, 0x6d, 0x82,
	0xf1, 0xbf, 0x9d, 0xfc, 0x8e, 0xd8, 0x20, 0xcf, 0x8a, 0x8f, 0x06, 0x69, 0x06, 0x5f, 0xd6, 0xe0,
	0x89, 0x3f, 0x61, 0xcc, 0x28, 0x40, 0x32, 0xda, 0x80, 0x1b, 0x45, 0x22, 0x1a, 0x9f, 0xce, 0x0e,
	0x08, 0x7f, 0xc9, 0x2e, 0x1c, 0x54, 0xc6, 0x93, 0xfb, 0xb6, 0x90, 0x4a, 0x39, 0xf0, 0x7e, 0x74,
	0x12, 0xac, 0xe1, 0x8e, 0x4f, 0x3b, 0xcc, 0x5f, 0xb0, 0xa1, 0x82, 0xc6, 0x7a, 0x43, 0x7b, 0xb3,
	0x17, 0xcc, 0xc1, 0x16, 0xef, 0xc4, 0xa7, 0xec, 0xbc, 0x71, 0xd6, 0xea, 0x85, 0x36, 0xb0, 0x52,
	0x7e, 0x74, 0x4b, 0xf4, 0xc6, 0xa7, 0xb3, 0xb3, 0xc0, 0xb2, 0x80, 0x92, 0x39, 0x1b, 0xee, 0x43,
	0x7d, 0x63, 0xd1, 0x03, 0x7f, 0xc4, 0xee, 0x92, 0xfd, 0x0c, 0xb8, 0x30, 0x6a, 0xdb, 0xd9, 0x0f,
	0xff, 0xd7, 0x8a, 0x3f, 0x63, 0x03, 0x72, 0x12, 0xbd, 0x2c, 0xdb, 0xd7, 0xb6, 0x42, 0x97, 0x78,
	0xef, 0x80, 0x5e, 0xab, 0x37, 0x3f, 0x22, 0xc6, 0xf2, 0xac, 0x98, 0x77, 0x43, 0xe6, 0x1b, 0xd6,
	0x6f, 0x2f, 0xc8, 0xb3, 0x82, 0x3f, 0x98, 0xa0, 0xa6, 0xc9, 0xf1, 0x68, 0xe2, 0x87, 0xc7, 0xb0,
	0xcb, 0x48, 0xa6, 0x3f, 0xa7, 0xe3, 0xf8, 0x79, 0x8b, 0x84, 0x44, 0x91, 0x67, 0x85, 0xd0, 0xce,
	0xd6, 0x42, 0x8a, 0xf7, 0x80, 0xe4, 0x8c, 0x5e, 0x57, 0x20, 0x3e, 0xd8, 0x72, 0x5d, 0x03, 0xd2,
	0xf7, 0x3f, 0x7f, 0x7f, 0x9d, 0x5c, 0x24, 0x67, 0x69, 0x08, 0x4d, 0x6b, 0x83, 0xf4, 0x2e, 0x7a,
	0x75, 0x25, 0x58, 0xbf, 0xb4, 0x75, 0x7b, 0xfa, 0xd5, 0xf9, 0x36, 0xe6, 0xa6, 0xdd, 0xcf, 0x4d,
	0xf4, 0xe9, 0x36, 0x6a, 0x6a, 0x96, 0xcb, 0x3b, 0x61, 0x5f, 0x6f, 0xff, 0x05, 0x00, 0x00, 0xff,
	0xff, 0x3c, 0xa0, 0x02, 0xa5, 0x15, 0x02, 0x00, 0x00,
}

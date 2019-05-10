// Code generated by protoc-gen-go. DO NOT EDIT.
// source: nft/service.proto

package nftpb

import (
	context "context"
	fmt "fmt"
	math "math"

	proto "github.com/golang/protobuf/proto"
	_ "github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger/options"
	_ "google.golang.org/genproto/googleapis/api/annotations"
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
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type ResponseHeader struct {
	JobId                string   `protobuf:"bytes,5,opt,name=job_id,json=jobId,proto3" json:"job_id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ResponseHeader) Reset()         { *m = ResponseHeader{} }
func (m *ResponseHeader) String() string { return proto.CompactTextString(m) }
func (*ResponseHeader) ProtoMessage()    {}
func (*ResponseHeader) Descriptor() ([]byte, []int) {
	return fileDescriptor_a63c52875f346f52, []int{0}
}

func (m *ResponseHeader) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ResponseHeader.Unmarshal(m, b)
}
func (m *ResponseHeader) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ResponseHeader.Marshal(b, m, deterministic)
}
func (m *ResponseHeader) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ResponseHeader.Merge(m, src)
}
func (m *ResponseHeader) XXX_Size() int {
	return xxx_messageInfo_ResponseHeader.Size(m)
}
func (m *ResponseHeader) XXX_DiscardUnknown() {
	xxx_messageInfo_ResponseHeader.DiscardUnknown(m)
}

var xxx_messageInfo_ResponseHeader proto.InternalMessageInfo

func (m *ResponseHeader) GetJobId() string {
	if m != nil {
		return m.JobId
	}
	return ""
}

type NFTMintInvoiceUnpaidRequest struct {
	// Invoice Document identifier
	Identifier string `protobuf:"bytes,1,opt,name=identifier,proto3" json:"identifier,omitempty"`
	// Deposit address for NFT Token created
	DepositAddress       string   `protobuf:"bytes,2,opt,name=deposit_address,json=depositAddress,proto3" json:"deposit_address,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *NFTMintInvoiceUnpaidRequest) Reset()         { *m = NFTMintInvoiceUnpaidRequest{} }
func (m *NFTMintInvoiceUnpaidRequest) String() string { return proto.CompactTextString(m) }
func (*NFTMintInvoiceUnpaidRequest) ProtoMessage()    {}
func (*NFTMintInvoiceUnpaidRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_a63c52875f346f52, []int{1}
}

func (m *NFTMintInvoiceUnpaidRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_NFTMintInvoiceUnpaidRequest.Unmarshal(m, b)
}
func (m *NFTMintInvoiceUnpaidRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_NFTMintInvoiceUnpaidRequest.Marshal(b, m, deterministic)
}
func (m *NFTMintInvoiceUnpaidRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_NFTMintInvoiceUnpaidRequest.Merge(m, src)
}
func (m *NFTMintInvoiceUnpaidRequest) XXX_Size() int {
	return xxx_messageInfo_NFTMintInvoiceUnpaidRequest.Size(m)
}
func (m *NFTMintInvoiceUnpaidRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_NFTMintInvoiceUnpaidRequest.DiscardUnknown(m)
}

var xxx_messageInfo_NFTMintInvoiceUnpaidRequest proto.InternalMessageInfo

func (m *NFTMintInvoiceUnpaidRequest) GetIdentifier() string {
	if m != nil {
		return m.Identifier
	}
	return ""
}

func (m *NFTMintInvoiceUnpaidRequest) GetDepositAddress() string {
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
	return fileDescriptor_a63c52875f346f52, []int{2}
}

func (m *NFTMintRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_NFTMintRequest.Unmarshal(m, b)
}
func (m *NFTMintRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_NFTMintRequest.Marshal(b, m, deterministic)
}
func (m *NFTMintRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_NFTMintRequest.Merge(m, src)
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
	XXX_NoUnkeyedLiteral struct{}        `json:"-"`
	XXX_unrecognized     []byte          `json:"-"`
	XXX_sizecache        int32           `json:"-"`
}

func (m *NFTMintResponse) Reset()         { *m = NFTMintResponse{} }
func (m *NFTMintResponse) String() string { return proto.CompactTextString(m) }
func (*NFTMintResponse) ProtoMessage()    {}
func (*NFTMintResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_a63c52875f346f52, []int{3}
}

func (m *NFTMintResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_NFTMintResponse.Unmarshal(m, b)
}
func (m *NFTMintResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_NFTMintResponse.Marshal(b, m, deterministic)
}
func (m *NFTMintResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_NFTMintResponse.Merge(m, src)
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

func init() {
	proto.RegisterType((*ResponseHeader)(nil), "nft.ResponseHeader")
	proto.RegisterType((*NFTMintInvoiceUnpaidRequest)(nil), "nft.NFTMintInvoiceUnpaidRequest")
	proto.RegisterType((*NFTMintRequest)(nil), "nft.NFTMintRequest")
	proto.RegisterType((*NFTMintResponse)(nil), "nft.NFTMintResponse")
}

func init() { proto.RegisterFile("nft/service.proto", fileDescriptor_a63c52875f346f52) }

var fileDescriptor_a63c52875f346f52 = []byte{
	// 554 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x53, 0x4d, 0x6e, 0xd3, 0x40,
	0x18, 0x95, 0x53, 0x92, 0xb4, 0x93, 0x2a, 0x09, 0x43, 0x91, 0x42, 0xf8, 0x91, 0xf1, 0x82, 0x84,
	0x92, 0xc6, 0x28, 0xec, 0xba, 0x40, 0xa4, 0xa0, 0x88, 0x2e, 0x30, 0x91, 0x09, 0x1b, 0x36, 0xd6,
	0xc4, 0xfe, 0xc6, 0x4c, 0x69, 0x66, 0xcc, 0xcc, 0xb8, 0x11, 0x42, 0x6c, 0x38, 0x42, 0x39, 0x0c,
	0x4b, 0xc4, 0x19, 0xb8, 0x02, 0x07, 0x41, 0x9e, 0x71, 0x21, 0xa1, 0x41, 0x62, 0x65, 0xf9, 0x7d,
	0xef, 0xbd, 0x79, 0xfa, 0x7e, 0xd0, 0x55, 0x4e, 0xb5, 0xaf, 0x40, 0x9e, 0xb1, 0x18, 0x86, 0x99,
	0x14, 0x5a, 0xe0, 0x2d, 0x4e, 0x75, 0xf7, 0x56, 0x2a, 0x44, 0x7a, 0x0a, 0x3e, 0xc9, 0x98, 0x4f,
	0x38, 0x17, 0x9a, 0x68, 0x26, 0xb8, 0xb2, 0x94, 0xee, 0xc0, 0x7c, 0xe2, 0x83, 0x14, 0xf8, 0x81,
	0x5a, 0x92, 0x34, 0x05, 0xe9, 0x8b, 0xcc, 0x30, 0x2e, 0xb3, 0xbd, 0x1e, 0x6a, 0x86, 0xa0, 0x32,
	0xc1, 0x15, 0x3c, 0x07, 0x92, 0x80, 0xc4, 0xd7, 0x51, 0xed, 0x44, 0xcc, 0x23, 0x96, 0x74, 0xaa,
	0xae, 0xd3, 0xdf, 0x09, 0xab, 0x27, 0x62, 0x7e, 0x9c, 0x78, 0x14, 0xdd, 0x0c, 0x26, 0xb3, 0x17,
	0x8c, 0xeb, 0x63, 0x7e, 0x26, 0x58, 0x0c, 0xaf, 0x79, 0x46, 0x58, 0x12, 0xc2, 0xfb, 0x1c, 0x94,
	0xc6, 0x77, 0x10, 0x62, 0x09, 0x70, 0xcd, 0x28, 0x03, 0xd9, 0x71, 0x8c, 0x72, 0x05, 0xc1, 0x3d,
	0xd4, 0x4a, 0x20, 0x13, 0x8a, 0xe9, 0x88, 0x24, 0x89, 0x04, 0xa5, 0x3a, 0x15, 0x43, 0x6a, 0x96,
	0xf0, 0xd8, 0xa2, 0xde, 0xf7, 0x0a, 0x6a, 0x96, 0x0f, 0xfd, 0xaf, 0xf7, 0x7d, 0xd4, 0x96, 0x90,
	0x32, 0xa5, 0xe5, 0x87, 0xbf, 0xcc, 0x5b, 0x17, 0x78, 0xe9, 0xbe, 0x29, 0xc6, 0xd6, 0xa6, 0x18,
	0xf8, 0x2e, 0xda, 0xcd, 0xa4, 0x10, 0x34, 0xa2, 0x0c, 0x4e, 0x13, 0xd5, 0xb9, 0xe2, 0x6e, 0xf5,
	0x77, 0xc2, 0x86, 0xc1, 0x26, 0x06, 0xc2, 0x03, 0x84, 0x55, 0x3e, 0x5f, 0x30, 0x1d, 0x69, 0xf1,
	0x0e, 0x78, 0x64, 0x6a, 0xa6, 0x69, 0xdb, 0x61, 0xdb, 0x56, 0x66, 0x45, 0x61, 0x5a, 0xe0, 0xf8,
	0x09, 0xba, 0x5d, 0xb2, 0x39, 0xd5, 0x91, 0x58, 0x72, 0x90, 0x11, 0x89, 0x63, 0x50, 0xaa, 0x14,
	0xd6, 0x8d, 0xf0, 0x86, 0x25, 0x05, 0x54, 0xbf, 0x2c, 0x28, 0x63, 0xc3, 0xb0, 0x0e, 0x7d, 0xd4,
	0x4e, 0x25, 0xe1, 0xd6, 0xc0, 0x4a, 0x3b, 0xdb, 0x46, 0xd4, 0x34, 0x78, 0x40, 0xb5, 0xa5, 0x7b,
	0x8f, 0x51, 0xeb, 0x77, 0x0b, 0xed, 0x6c, 0xf1, 0x03, 0x54, 0x7b, 0x6b, 0xe6, 0x6b, 0xfa, 0xd7,
	0x18, 0x5d, 0x1b, 0x72, 0xaa, 0x87, 0xeb, 0xa3, 0x0f, 0x4b, 0xca, 0xe8, 0x5b, 0x05, 0xa1, 0x60,
	0x32, 0x7b, 0x65, 0x57, 0x0f, 0x2f, 0x51, 0xbd, 0xf0, 0x0a, 0x26, 0x33, 0x6c, 0x65, 0xeb, 0xf3,
	0xe9, 0xee, 0xad, 0x83, 0xd6, 0xd2, 0x1b, 0x9f, 0x8f, 0xfb, 0xdd, 0x7b, 0x05, 0xe4, 0x12, 0xee,
	0x06, 0x93, 0x99, 0x4b, 0xa5, 0x58, 0xb8, 0xc4, 0x7d, 0x0a, 0x5c, 0x4b, 0x46, 0xf3, 0x14, 0xdc,
	0x67, 0x22, 0xce, 0x17, 0xc0, 0xf5, 0xe7, 0x1f, 0x3f, 0xbf, 0x54, 0xda, 0x5e, 0xc3, 0x37, 0x9d,
	0xf4, 0x17, 0x8c, 0xeb, 0x43, 0x67, 0x1f, 0x7f, 0x75, 0xd0, 0xde, 0xa5, 0x8d, 0x2b, 0x62, 0xb8,
	0xab, 0x2f, 0x6e, 0xda, 0xc7, 0x7f, 0x64, 0x4a, 0xcf, 0xc7, 0xa3, 0xee, 0xc3, 0x02, 0x52, 0x17,
	0xa1, 0x44, 0xae, 0x5d, 0x41, 0x8b, 0x3f, 0x6b, 0xb0, 0x1a, 0xaf, 0x74, 0x36, 0xe9, 0x06, 0x5e,
	0x6f, 0x25, 0x9d, 0xcf, 0x6c, 0xc9, 0xcf, 0x8d, 0xc8, 0xff, 0xf8, 0x67, 0x1d, 0x3f, 0x1d, 0x3a,
	0xfb, 0x47, 0x2e, 0xaa, 0xc7, 0x62, 0x51, 0x64, 0x38, 0xda, 0x2d, 0xdb, 0x38, 0x2d, 0xee, 0x6d,
	0xea, 0xbc, 0xa9, 0x72, 0xaa, 0xb3, 0xf9, 0xbc, 0x66, 0xee, 0xef, 0xd1, 0xaf, 0x00, 0x00, 0x00,
	0xff, 0xff, 0x1a, 0x27, 0xde, 0xf0, 0xe5, 0x03, 0x00, 0x00,
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
	MintInvoiceUnpaidNFT(ctx context.Context, in *NFTMintInvoiceUnpaidRequest, opts ...grpc.CallOption) (*NFTMintResponse, error)
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

func (c *nFTServiceClient) MintInvoiceUnpaidNFT(ctx context.Context, in *NFTMintInvoiceUnpaidRequest, opts ...grpc.CallOption) (*NFTMintResponse, error) {
	out := new(NFTMintResponse)
	err := c.cc.Invoke(ctx, "/nft.NFTService/MintInvoiceUnpaidNFT", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// NFTServiceServer is the server API for NFTService service.
type NFTServiceServer interface {
	MintNFT(context.Context, *NFTMintRequest) (*NFTMintResponse, error)
	MintInvoiceUnpaidNFT(context.Context, *NFTMintInvoiceUnpaidRequest) (*NFTMintResponse, error)
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

func _NFTService_MintInvoiceUnpaidNFT_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(NFTMintInvoiceUnpaidRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NFTServiceServer).MintInvoiceUnpaidNFT(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/nft.NFTService/MintInvoiceUnpaidNFT",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NFTServiceServer).MintInvoiceUnpaidNFT(ctx, req.(*NFTMintInvoiceUnpaidRequest))
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
			MethodName: "MintInvoiceUnpaidNFT",
			Handler:    _NFTService_MintInvoiceUnpaidNFT_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "nft/service.proto",
}

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

type TokenTransferRequest struct {
	TokenId              string   `protobuf:"bytes,1,opt,name=token_id,json=tokenId,proto3" json:"token_id,omitempty"`
	RegistryAddress      string   `protobuf:"bytes,2,opt,name=registry_address,json=registryAddress,proto3" json:"registry_address,omitempty"`
	To                   string   `protobuf:"bytes,3,opt,name=to,proto3" json:"to,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *TokenTransferRequest) Reset()         { *m = TokenTransferRequest{} }
func (m *TokenTransferRequest) String() string { return proto.CompactTextString(m) }
func (*TokenTransferRequest) ProtoMessage()    {}
func (*TokenTransferRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_a63c52875f346f52, []int{1}
}

func (m *TokenTransferRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_TokenTransferRequest.Unmarshal(m, b)
}
func (m *TokenTransferRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_TokenTransferRequest.Marshal(b, m, deterministic)
}
func (m *TokenTransferRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TokenTransferRequest.Merge(m, src)
}
func (m *TokenTransferRequest) XXX_Size() int {
	return xxx_messageInfo_TokenTransferRequest.Size(m)
}
func (m *TokenTransferRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_TokenTransferRequest.DiscardUnknown(m)
}

var xxx_messageInfo_TokenTransferRequest proto.InternalMessageInfo

func (m *TokenTransferRequest) GetTokenId() string {
	if m != nil {
		return m.TokenId
	}
	return ""
}

func (m *TokenTransferRequest) GetRegistryAddress() string {
	if m != nil {
		return m.RegistryAddress
	}
	return ""
}

func (m *TokenTransferRequest) GetTo() string {
	if m != nil {
		return m.To
	}
	return ""
}

type TokenTransferResponse struct {
	Header               *ResponseHeader `protobuf:"bytes,1,opt,name=header,proto3" json:"header,omitempty"`
	XXX_NoUnkeyedLiteral struct{}        `json:"-"`
	XXX_unrecognized     []byte          `json:"-"`
	XXX_sizecache        int32           `json:"-"`
}

func (m *TokenTransferResponse) Reset()         { *m = TokenTransferResponse{} }
func (m *TokenTransferResponse) String() string { return proto.CompactTextString(m) }
func (*TokenTransferResponse) ProtoMessage()    {}
func (*TokenTransferResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_a63c52875f346f52, []int{2}
}

func (m *TokenTransferResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_TokenTransferResponse.Unmarshal(m, b)
}
func (m *TokenTransferResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_TokenTransferResponse.Marshal(b, m, deterministic)
}
func (m *TokenTransferResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TokenTransferResponse.Merge(m, src)
}
func (m *TokenTransferResponse) XXX_Size() int {
	return xxx_messageInfo_TokenTransferResponse.Size(m)
}
func (m *TokenTransferResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_TokenTransferResponse.DiscardUnknown(m)
}

var xxx_messageInfo_TokenTransferResponse proto.InternalMessageInfo

func (m *TokenTransferResponse) GetHeader() *ResponseHeader {
	if m != nil {
		return m.Header
	}
	return nil
}

type OwnerOfRequest struct {
	TokenId              string   `protobuf:"bytes,1,opt,name=token_id,json=tokenId,proto3" json:"token_id,omitempty"`
	RegistryAddress      string   `protobuf:"bytes,2,opt,name=registry_address,json=registryAddress,proto3" json:"registry_address,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *OwnerOfRequest) Reset()         { *m = OwnerOfRequest{} }
func (m *OwnerOfRequest) String() string { return proto.CompactTextString(m) }
func (*OwnerOfRequest) ProtoMessage()    {}
func (*OwnerOfRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_a63c52875f346f52, []int{3}
}

func (m *OwnerOfRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_OwnerOfRequest.Unmarshal(m, b)
}
func (m *OwnerOfRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_OwnerOfRequest.Marshal(b, m, deterministic)
}
func (m *OwnerOfRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_OwnerOfRequest.Merge(m, src)
}
func (m *OwnerOfRequest) XXX_Size() int {
	return xxx_messageInfo_OwnerOfRequest.Size(m)
}
func (m *OwnerOfRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_OwnerOfRequest.DiscardUnknown(m)
}

var xxx_messageInfo_OwnerOfRequest proto.InternalMessageInfo

func (m *OwnerOfRequest) GetTokenId() string {
	if m != nil {
		return m.TokenId
	}
	return ""
}

func (m *OwnerOfRequest) GetRegistryAddress() string {
	if m != nil {
		return m.RegistryAddress
	}
	return ""
}

type OwnerOfResponse struct {
	TokenId              string   `protobuf:"bytes,1,opt,name=token_id,json=tokenId,proto3" json:"token_id,omitempty"`
	RegistryAddress      string   `protobuf:"bytes,2,opt,name=registry_address,json=registryAddress,proto3" json:"registry_address,omitempty"`
	Owner                string   `protobuf:"bytes,3,opt,name=owner,proto3" json:"owner,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *OwnerOfResponse) Reset()         { *m = OwnerOfResponse{} }
func (m *OwnerOfResponse) String() string { return proto.CompactTextString(m) }
func (*OwnerOfResponse) ProtoMessage()    {}
func (*OwnerOfResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_a63c52875f346f52, []int{4}
}

func (m *OwnerOfResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_OwnerOfResponse.Unmarshal(m, b)
}
func (m *OwnerOfResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_OwnerOfResponse.Marshal(b, m, deterministic)
}
func (m *OwnerOfResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_OwnerOfResponse.Merge(m, src)
}
func (m *OwnerOfResponse) XXX_Size() int {
	return xxx_messageInfo_OwnerOfResponse.Size(m)
}
func (m *OwnerOfResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_OwnerOfResponse.DiscardUnknown(m)
}

var xxx_messageInfo_OwnerOfResponse proto.InternalMessageInfo

func (m *OwnerOfResponse) GetTokenId() string {
	if m != nil {
		return m.TokenId
	}
	return ""
}

func (m *OwnerOfResponse) GetRegistryAddress() string {
	if m != nil {
		return m.RegistryAddress
	}
	return ""
}

func (m *OwnerOfResponse) GetOwner() string {
	if m != nil {
		return m.Owner
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
	return fileDescriptor_a63c52875f346f52, []int{5}
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
	return fileDescriptor_a63c52875f346f52, []int{6}
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
	proto.RegisterType((*TokenTransferRequest)(nil), "nft.TokenTransferRequest")
	proto.RegisterType((*TokenTransferResponse)(nil), "nft.TokenTransferResponse")
	proto.RegisterType((*OwnerOfRequest)(nil), "nft.OwnerOfRequest")
	proto.RegisterType((*OwnerOfResponse)(nil), "nft.OwnerOfResponse")
	proto.RegisterType((*NFTMintInvoiceUnpaidRequest)(nil), "nft.NFTMintInvoiceUnpaidRequest")
	proto.RegisterType((*NFTMintResponse)(nil), "nft.NFTMintResponse")
}

func init() { proto.RegisterFile("nft/service.proto", fileDescriptor_a63c52875f346f52) }

var fileDescriptor_a63c52875f346f52 = []byte{
	// 588 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xac, 0x54, 0x4d, 0x4e, 0x14, 0x41,
	0x14, 0x4e, 0x0f, 0x19, 0x90, 0x52, 0x67, 0xb4, 0x1c, 0x22, 0xd3, 0x1a, 0x52, 0xf6, 0x42, 0xfc,
	0x63, 0x8a, 0x8c, 0xae, 0x5c, 0x98, 0x0c, 0x9a, 0x51, 0x16, 0x0e, 0x64, 0x6c, 0x31, 0x71, 0x43,
	0x7a, 0xba, 0x5f, 0x37, 0x85, 0x50, 0xd5, 0x56, 0xbd, 0x81, 0x10, 0xc2, 0xc6, 0xb5, 0x2b, 0x3c,
	0x80, 0x67, 0xf0, 0x2c, 0x5e, 0xc1, 0x0b, 0x78, 0x03, 0xd3, 0xd5, 0xd5, 0x08, 0x03, 0x71, 0x61,
	0x58, 0x75, 0xde, 0xeb, 0x57, 0xdf, 0xf7, 0xbd, 0xaf, 0xbe, 0x14, 0xb9, 0x29, 0x53, 0xe4, 0x06,
	0xf4, 0x9e, 0x88, 0xa1, 0x93, 0x6b, 0x85, 0x8a, 0x4e, 0xc9, 0x14, 0xfd, 0xbb, 0x99, 0x52, 0xd9,
	0x0e, 0xf0, 0x28, 0x17, 0x3c, 0x92, 0x52, 0x61, 0x84, 0x42, 0x49, 0x53, 0x8e, 0xf8, 0x4f, 0xec,
	0x27, 0x5e, 0xca, 0x40, 0x2e, 0x99, 0xfd, 0x28, 0xcb, 0x40, 0x73, 0x95, 0xdb, 0x89, 0xf3, 0xd3,
	0xc1, 0x22, 0x69, 0x0c, 0xc1, 0xe4, 0x4a, 0x1a, 0x78, 0x03, 0x51, 0x02, 0x9a, 0xce, 0x91, 0xe9,
	0x6d, 0x35, 0xda, 0x14, 0xc9, 0x7c, 0x9d, 0x79, 0x0f, 0x66, 0x87, 0xf5, 0x6d, 0x35, 0x5a, 0x4d,
	0x82, 0x1d, 0xd2, 0x0a, 0xd5, 0x27, 0x90, 0xa1, 0x8e, 0xa4, 0x49, 0x41, 0x0f, 0xe1, 0xf3, 0x18,
	0x0c, 0xd2, 0x36, 0xb9, 0x82, 0x45, 0xbf, 0x38, 0xe0, 0xd9, 0x03, 0x33, 0xb6, 0x5e, 0x4d, 0xe8,
	0x43, 0x72, 0x43, 0x43, 0x26, 0x0c, 0xea, 0x83, 0xcd, 0x28, 0x49, 0x34, 0x18, 0x33, 0x5f, 0xb3,
	0x23, 0xcd, 0xaa, 0xdf, 0x2b, 0xdb, 0xb4, 0x41, 0x6a, 0xa8, 0xe6, 0xa7, 0xec, 0xcf, 0x1a, 0xaa,
	0xe0, 0x15, 0x99, 0x9b, 0x60, 0x2b, 0x35, 0xd2, 0xc7, 0x64, 0x7a, 0xcb, 0xea, 0xb4, 0x64, 0x57,
	0xbb, 0xb7, 0x3a, 0x32, 0xc5, 0xce, 0xd9, 0x15, 0x86, 0x6e, 0x24, 0xd8, 0x20, 0x8d, 0xb5, 0x7d,
	0x09, 0x7a, 0x2d, 0xbd, 0x54, 0xb5, 0xc1, 0x2e, 0x69, 0x9e, 0xe0, 0x3a, 0x5d, 0x97, 0x63, 0x43,
	0x8b, 0xd4, 0x55, 0x01, 0xec, 0x9c, 0x28, 0x8b, 0x20, 0x25, 0x77, 0x06, 0xfd, 0xf0, 0xad, 0x90,
	0xb8, 0x2a, 0xf7, 0x94, 0x88, 0xe1, 0xbd, 0xcc, 0x23, 0x91, 0x54, 0x3b, 0x2d, 0x10, 0x22, 0x12,
	0x90, 0x28, 0x52, 0xe1, 0x6c, 0x99, 0x1d, 0x9e, 0xea, 0xd0, 0x45, 0xd2, 0x4c, 0x20, 0x57, 0x46,
	0xe0, 0x04, 0x7d, 0xc3, 0xb5, 0xab, 0xb5, 0x5e, 0x90, 0xa6, 0xe3, 0xf9, 0x2f, 0xbb, 0xbb, 0xbf,
	0xa7, 0x08, 0x19, 0xf4, 0xc3, 0x77, 0x65, 0x62, 0xe9, 0x0f, 0x8f, 0xb4, 0xce, 0x89, 0x1e, 0xf4,
	0x43, 0xca, 0x2c, 0xc8, 0x3f, 0x56, 0xf2, 0x5b, 0xa7, 0x27, 0x2a, 0xb6, 0x20, 0x3a, 0xee, 0x75,
	0xfd, 0xe5, 0xa2, 0x65, 0x58, 0x24, 0xd9, 0xa0, 0x1f, 0x32, 0x35, 0x46, 0xa6, 0xd2, 0xa2, 0x2a,
	0x01, 0xd8, 0x4b, 0x90, 0xa8, 0x45, 0x3a, 0xce, 0x80, 0x39, 0xe4, 0x2f, 0x3f, 0x7f, 0x7d, 0xab,
	0xdd, 0x0f, 0xee, 0x71, 0x51, 0xd6, 0xfc, 0xf0, 0xaf, 0x37, 0x47, 0x7c, 0x57, 0x48, 0xe4, 0x63,
	0x7b, 0xf6, 0xb9, 0xf7, 0x88, 0x7e, 0xf7, 0xc8, 0xf5, 0x33, 0xc1, 0xa3, 0x6d, 0x2b, 0xe5, 0xa2,
	0xe8, 0xfb, 0xfe, 0x45, 0xbf, 0x9c, 0xd6, 0x0f, 0xc7, 0xbd, 0x65, 0xbf, 0x53, 0xb5, 0x4f, 0xe4,
	0xee, 0x4b, 0x48, 0xd8, 0xe8, 0x80, 0xe1, 0x16, 0xb0, 0x52, 0x06, 0x1e, 0xb0, 0x58, 0x49, 0xd4,
	0x51, 0x8c, 0x56, 0xe9, 0x42, 0xd0, 0xe6, 0x32, 0x45, 0xc3, 0x0f, 0xab, 0x3c, 0x1d, 0x71, 0x74,
	0x30, 0x85, 0xc2, 0xaf, 0x1e, 0x99, 0x71, 0xe1, 0xa3, 0xe5, 0x6d, 0x9c, 0x8d, 0xb8, 0xf3, 0x6e,
	0x22, 0x9f, 0xc1, 0xc6, 0x71, 0xaf, 0xed, 0xdf, 0x7e, 0x0d, 0x68, 0xa9, 0x6d, 0xae, 0x9c, 0x71,
	0x83, 0x7e, 0x68, 0x89, 0x9f, 0xd1, 0xee, 0x39, 0xe2, 0x2a, 0x9f, 0xfc, 0x70, 0x32, 0xc1, 0x47,
	0xdc, 0x62, 0xac, 0x30, 0x32, 0x13, 0xab, 0xdd, 0x82, 0x72, 0xe5, 0x9a, 0xbb, 0xf8, 0xf5, 0xe2,
	0x61, 0x59, 0xf7, 0x3e, 0xd6, 0x65, 0x8a, 0xf9, 0x68, 0x34, 0x6d, 0x1f, 0x9a, 0xa7, 0x7f, 0x02,
	0x00, 0x00, 0xff, 0xff, 0x4e, 0xa3, 0xef, 0x50, 0xce, 0x04, 0x00, 0x00,
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
	MintInvoiceUnpaidNFT(ctx context.Context, in *NFTMintInvoiceUnpaidRequest, opts ...grpc.CallOption) (*NFTMintResponse, error)
	TokenTransfer(ctx context.Context, in *TokenTransferRequest, opts ...grpc.CallOption) (*TokenTransferResponse, error)
	OwnerOf(ctx context.Context, in *OwnerOfRequest, opts ...grpc.CallOption) (*OwnerOfResponse, error)
}

type nFTServiceClient struct {
	cc *grpc.ClientConn
}

func NewNFTServiceClient(cc *grpc.ClientConn) NFTServiceClient {
	return &nFTServiceClient{cc}
}

func (c *nFTServiceClient) MintInvoiceUnpaidNFT(ctx context.Context, in *NFTMintInvoiceUnpaidRequest, opts ...grpc.CallOption) (*NFTMintResponse, error) {
	out := new(NFTMintResponse)
	err := c.cc.Invoke(ctx, "/nft.NFTService/MintInvoiceUnpaidNFT", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *nFTServiceClient) TokenTransfer(ctx context.Context, in *TokenTransferRequest, opts ...grpc.CallOption) (*TokenTransferResponse, error) {
	out := new(TokenTransferResponse)
	err := c.cc.Invoke(ctx, "/nft.NFTService/TokenTransfer", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *nFTServiceClient) OwnerOf(ctx context.Context, in *OwnerOfRequest, opts ...grpc.CallOption) (*OwnerOfResponse, error) {
	out := new(OwnerOfResponse)
	err := c.cc.Invoke(ctx, "/nft.NFTService/OwnerOf", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// NFTServiceServer is the server API for NFTService service.
type NFTServiceServer interface {
	MintInvoiceUnpaidNFT(context.Context, *NFTMintInvoiceUnpaidRequest) (*NFTMintResponse, error)
	TokenTransfer(context.Context, *TokenTransferRequest) (*TokenTransferResponse, error)
	OwnerOf(context.Context, *OwnerOfRequest) (*OwnerOfResponse, error)
}

func RegisterNFTServiceServer(s *grpc.Server, srv NFTServiceServer) {
	s.RegisterService(&_NFTService_serviceDesc, srv)
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

func _NFTService_TokenTransfer_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TokenTransferRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NFTServiceServer).TokenTransfer(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/nft.NFTService/TokenTransfer",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NFTServiceServer).TokenTransfer(ctx, req.(*TokenTransferRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _NFTService_OwnerOf_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(OwnerOfRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NFTServiceServer).OwnerOf(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/nft.NFTService/OwnerOf",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(NFTServiceServer).OwnerOf(ctx, req.(*OwnerOfRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _NFTService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "nft.NFTService",
	HandlerType: (*NFTServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "MintInvoiceUnpaidNFT",
			Handler:    _NFTService_MintInvoiceUnpaidNFT_Handler,
		},
		{
			MethodName: "TokenTransfer",
			Handler:    _NFTService_TokenTransfer_Handler,
		},
		{
			MethodName: "OwnerOf",
			Handler:    _NFTService_OwnerOf_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "nft/service.proto",
}

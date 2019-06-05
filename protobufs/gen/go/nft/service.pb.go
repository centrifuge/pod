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

type OwnerOfResponse struct {
	TokenId              string   `protobuf:"bytes,1,opt,name=token_id,json=tokenId,proto3" json:"token_id,omitempty"`
	Owner                string   `protobuf:"bytes,2,opt,name=owner,proto3" json:"owner,omitempty"`
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
	return fileDescriptor_a63c52875f346f52, []int{6}
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
	return fileDescriptor_a63c52875f346f52, []int{7}
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
	proto.RegisterType((*NFTMintRequest)(nil), "nft.NFTMintRequest")
	proto.RegisterType((*NFTMintResponse)(nil), "nft.NFTMintResponse")
}

func init() { proto.RegisterFile("nft/service.proto", fileDescriptor_a63c52875f346f52) }

var fileDescriptor_a63c52875f346f52 = []byte{
	// 736 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x54, 0xcd, 0x6e, 0x13, 0x3d,
	0x14, 0x55, 0x92, 0xa6, 0x69, 0xdd, 0x7e, 0x49, 0x3e, 0x93, 0xaa, 0xc9, 0xf0, 0x67, 0x66, 0x41,
	0x03, 0x6d, 0x33, 0x55, 0xd9, 0x81, 0x84, 0x48, 0xa9, 0x02, 0x5d, 0x90, 0x56, 0x61, 0x90, 0x10,
	0x9b, 0x68, 0x32, 0x63, 0x4f, 0x5d, 0x1a, 0x7b, 0x6a, 0x3b, 0x8d, 0xaa, 0x8a, 0x0d, 0x8f, 0x10,
	0x5e, 0x80, 0xb7, 0x60, 0xcd, 0x33, 0x20, 0xde, 0x80, 0x07, 0x41, 0x63, 0x3b, 0x25, 0x69, 0xc3,
	0xef, 0x6a, 0x34, 0xe7, 0x9e, 0x7b, 0xcf, 0xf1, 0xf5, 0xf5, 0x05, 0xff, 0x33, 0xa2, 0x3c, 0x89,
	0xc5, 0x29, 0x0d, 0x71, 0x23, 0x11, 0x5c, 0x71, 0x98, 0x63, 0x44, 0x39, 0x37, 0x62, 0xce, 0xe3,
	0x63, 0xec, 0x05, 0x09, 0xf5, 0x02, 0xc6, 0xb8, 0x0a, 0x14, 0xe5, 0x4c, 0x1a, 0x8a, 0xb3, 0xa1,
	0x3f, 0xe1, 0x66, 0x8c, 0xd9, 0xa6, 0x1c, 0x06, 0x71, 0x8c, 0x85, 0xc7, 0x13, 0xcd, 0xb8, 0xca,
	0x76, 0xd7, 0x40, 0xb1, 0x83, 0x65, 0xc2, 0x99, 0xc4, 0xcf, 0x71, 0x10, 0x61, 0x01, 0x57, 0xc0,
	0xfc, 0x11, 0xef, 0x75, 0x69, 0x54, 0xcd, 0xa3, 0x4c, 0x7d, 0xb1, 0x93, 0x3f, 0xe2, 0xbd, 0xbd,
	0xc8, 0x3d, 0x06, 0x15, 0x9f, 0xbf, 0xc5, 0xcc, 0x17, 0x01, 0x93, 0x04, 0x8b, 0x0e, 0x3e, 0x19,
	0x60, 0xa9, 0x60, 0x0d, 0x2c, 0xa8, 0x14, 0x4f, 0x13, 0x32, 0x3a, 0xa1, 0xa0, 0xff, 0xf7, 0x22,
	0x78, 0x0f, 0x94, 0x05, 0x8e, 0xa9, 0x54, 0xe2, 0xac, 0x1b, 0x44, 0x91, 0xc0, 0x52, 0x56, 0xb3,
	0x9a, 0x52, 0x1a, 0xe3, 0x4d, 0x03, 0xc3, 0x22, 0xc8, 0x2a, 0x5e, 0xcd, 0xe9, 0x60, 0x56, 0x71,
	0x77, 0x17, 0xac, 0x5c, 0x52, 0x33, 0x1e, 0xe1, 0x3a, 0x98, 0x3f, 0xd4, 0x3e, 0xb5, 0xd8, 0xd2,
	0xf6, 0xb5, 0x06, 0x23, 0xaa, 0x31, 0x7d, 0x84, 0x8e, 0xa5, 0xb8, 0xeb, 0xa0, 0xb8, 0x3f, 0x64,
	0x58, 0xec, 0x93, 0xdf, 0xbb, 0x75, 0x77, 0x40, 0xe9, 0x82, 0x6c, 0xc5, 0x7e, 0x71, 0xb6, 0x0a,
	0xc8, 0xf3, 0x94, 0x6d, 0x0f, 0x64, 0x7e, 0x5c, 0x02, 0xae, 0xb7, 0x5b, 0xfe, 0x0b, 0xca, 0xd4,
	0x1e, 0x3b, 0xe5, 0x34, 0xc4, 0xaf, 0x58, 0x12, 0xd0, 0x68, 0xac, 0x7e, 0x0b, 0x00, 0x1a, 0x61,
	0xa6, 0x28, 0xa1, 0xf6, 0x00, 0x8b, 0x9d, 0x09, 0x04, 0xae, 0x81, 0x52, 0x84, 0x13, 0x2e, 0xa9,
	0xba, 0xd4, 0xaf, 0xa2, 0x85, 0x6d, 0xbb, 0xdc, 0xcf, 0x59, 0x50, 0xb4, 0x42, 0x7f, 0x5a, 0xfb,
	0x2f, 0x2e, 0x63, 0x86, 0x8d, 0xdc, 0x2c, 0x1b, 0xf0, 0x0e, 0x58, 0x4e, 0x04, 0xe7, 0xa4, 0x4b,
	0x28, 0x3e, 0x8e, 0x64, 0x75, 0x0e, 0xe5, 0xea, 0x8b, 0x9d, 0x25, 0x8d, 0xb5, 0x34, 0x04, 0x37,
	0x00, 0x94, 0x83, 0x5e, 0x9f, 0xaa, 0xae, 0xe9, 0xa4, 0x8e, 0xe9, 0xc9, 0x5a, 0xe8, 0x94, 0x4d,
	0x44, 0x5f, 0xf4, 0x41, 0x8a, 0xc3, 0x27, 0xe0, 0xa6, 0x65, 0x33, 0xa2, 0xba, 0xba, 0xa7, 0xdd,
	0x20, 0x0c, 0xb1, 0x94, 0x36, 0xb1, 0xa0, 0x13, 0x6b, 0x86, 0xd4, 0x26, 0x4a, 0xdf, 0x58, 0x53,
	0x33, 0x4c, 0x85, 0x3a, 0x28, 0xc7, 0x22, 0x60, 0xa6, 0x80, 0x49, 0xad, 0x2e, 0xe8, 0xa4, 0xa2,
	0xc6, 0xdb, 0x44, 0x19, 0xba, 0xfb, 0x18, 0x94, 0x2e, 0x5a, 0xf8, 0x0f, 0xc3, 0xb5, 0xfd, 0x75,
	0x0e, 0x80, 0x76, 0xcb, 0x7f, 0x69, 0xde, 0x27, 0x1c, 0x82, 0x42, 0x5a, 0xab, 0xdd, 0xf2, 0xa1,
	0x49, 0x9b, 0xbe, 0x1f, 0xa7, 0x32, 0x0d, 0x9a, 0x92, 0x6e, 0x73, 0xd4, 0xac, 0x3b, 0x77, 0x53,
	0x08, 0x05, 0x0c, 0xb5, 0x5b, 0x3e, 0x22, 0x82, 0xf7, 0x51, 0x80, 0x9e, 0x62, 0xa6, 0x04, 0x25,
	0x83, 0x18, 0xa3, 0x5d, 0x1e, 0x0e, 0xfa, 0x98, 0xa9, 0xf7, 0x5f, 0xbe, 0x7d, 0xc8, 0x96, 0xdd,
	0x25, 0x4f, 0x77, 0xd2, 0xeb, 0x53, 0xa6, 0x1e, 0x66, 0xee, 0xc3, 0x4f, 0x19, 0x50, 0xb9, 0x32,
	0x71, 0xa9, 0x0d, 0x34, 0xa9, 0x38, 0x6b, 0x1e, 0x7f, 0xe2, 0x29, 0x1e, 0x35, 0xb7, 0x9d, 0xad,
	0x14, 0x92, 0x63, 0x53, 0x7c, 0xa0, 0x10, 0x27, 0xe9, 0x9f, 0x29, 0x30, 0x69, 0xcf, 0x56, 0xd6,
	0xee, 0x36, 0xdc, 0xb5, 0x09, 0x77, 0x1e, 0x35, 0x21, 0x6f, 0xa0, 0x93, 0xbc, 0xf3, 0x1f, 0xe3,
	0xf8, 0x2e, 0x75, 0xfe, 0x31, 0x03, 0xfe, 0x9b, 0x7a, 0xe5, 0xb0, 0xa6, 0x0d, 0xcd, 0xda, 0x33,
	0x8e, 0x33, 0x2b, 0x64, 0x1d, 0xbf, 0x1e, 0x35, 0xb7, 0x9c, 0xc6, 0x18, 0xbe, 0x30, 0x3d, 0x64,
	0x38, 0x42, 0xbd, 0x33, 0xa4, 0x0e, 0x31, 0x32, 0xd2, 0xea, 0x0c, 0x85, 0x9c, 0x29, 0x11, 0x84,
	0xa6, 0x9b, 0xb7, 0x5d, 0xc7, 0xfa, 0x55, 0x36, 0xd9, 0x3b, 0x1f, 0xbf, 0x78, 0x6d, 0xf1, 0x04,
	0x14, 0xec, 0x52, 0xb0, 0xb7, 0x3a, 0xbd, 0x4f, 0x6c, 0x07, 0x2f, 0xed, 0x0d, 0xf7, 0xd1, 0xa8,
	0x59, 0x73, 0x56, 0x9f, 0x61, 0xa5, 0xa5, 0xf5, 0x18, 0xdb, 0xf6, 0xb5, 0x5b, 0xbe, 0x16, 0xae,
	0xc1, 0x55, 0x2b, 0xac, 0xa3, 0x13, 0xaa, 0x3b, 0x08, 0x14, 0x42, 0xde, 0x4f, 0xeb, 0xee, 0x2c,
	0xdb, 0xe1, 0x3a, 0x48, 0x57, 0xf5, 0x41, 0xe6, 0x4d, 0x9e, 0x11, 0x95, 0xf4, 0x7a, 0xf3, 0x7a,
	0x75, 0x3f, 0xf8, 0x1e, 0x00, 0x00, 0xff, 0xff, 0x78, 0xfc, 0xcd, 0x1b, 0x20, 0x06, 0x00, 0x00,
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
	TokenTransfer(ctx context.Context, in *TokenTransferRequest, opts ...grpc.CallOption) (*TokenTransferResponse, error)
	OwnerOf(ctx context.Context, in *OwnerOfRequest, opts ...grpc.CallOption) (*OwnerOfResponse, error)
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
	MintNFT(context.Context, *NFTMintRequest) (*NFTMintResponse, error)
	MintInvoiceUnpaidNFT(context.Context, *NFTMintInvoiceUnpaidRequest) (*NFTMintResponse, error)
	TokenTransfer(context.Context, *TokenTransferRequest) (*TokenTransferResponse, error)
	OwnerOf(context.Context, *OwnerOfRequest) (*OwnerOfResponse, error)
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
			MethodName: "MintNFT",
			Handler:    _NFTService_MintNFT_Handler,
		},
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

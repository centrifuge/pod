// Code generated by protoc-gen-go. DO NOT EDIT.
// source: entity/service.proto

package entitypb

import (
	context "context"
	fmt "fmt"
	entity "github.com/centrifuge/centrifuge-protobufs/gen/go/entity"
	proto "github.com/golang/protobuf/proto"
	_ "github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger/options"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	grpc "google.golang.org/grpc"
	math "math"
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

type GetRequest struct {
	Identifier           string   `protobuf:"bytes,1,opt,name=identifier,proto3" json:"identifier,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetRequest) Reset()         { *m = GetRequest{} }
func (m *GetRequest) String() string { return proto.CompactTextString(m) }
func (*GetRequest) ProtoMessage()    {}
func (*GetRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_c1b437217b9e14a2, []int{0}
}

func (m *GetRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetRequest.Unmarshal(m, b)
}
func (m *GetRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetRequest.Marshal(b, m, deterministic)
}
func (m *GetRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetRequest.Merge(m, src)
}
func (m *GetRequest) XXX_Size() int {
	return xxx_messageInfo_GetRequest.Size(m)
}
func (m *GetRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GetRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GetRequest proto.InternalMessageInfo

func (m *GetRequest) GetIdentifier() string {
	if m != nil {
		return m.Identifier
	}
	return ""
}

type GetVersionRequest struct {
	Identifier           string   `protobuf:"bytes,1,opt,name=identifier,proto3" json:"identifier,omitempty"`
	Version              string   `protobuf:"bytes,2,opt,name=version,proto3" json:"version,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetVersionRequest) Reset()         { *m = GetVersionRequest{} }
func (m *GetVersionRequest) String() string { return proto.CompactTextString(m) }
func (*GetVersionRequest) ProtoMessage()    {}
func (*GetVersionRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_c1b437217b9e14a2, []int{1}
}

func (m *GetVersionRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetVersionRequest.Unmarshal(m, b)
}
func (m *GetVersionRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetVersionRequest.Marshal(b, m, deterministic)
}
func (m *GetVersionRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetVersionRequest.Merge(m, src)
}
func (m *GetVersionRequest) XXX_Size() int {
	return xxx_messageInfo_GetVersionRequest.Size(m)
}
func (m *GetVersionRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GetVersionRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GetVersionRequest proto.InternalMessageInfo

func (m *GetVersionRequest) GetIdentifier() string {
	if m != nil {
		return m.Identifier
	}
	return ""
}

func (m *GetVersionRequest) GetVersion() string {
	if m != nil {
		return m.Version
	}
	return ""
}

type EntityCreatePayload struct {
	Collaborators        []string    `protobuf:"bytes,1,rep,name=collaborators,proto3" json:"collaborators,omitempty"`
	Data                 *EntityData `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *EntityCreatePayload) Reset()         { *m = EntityCreatePayload{} }
func (m *EntityCreatePayload) String() string { return proto.CompactTextString(m) }
func (*EntityCreatePayload) ProtoMessage()    {}
func (*EntityCreatePayload) Descriptor() ([]byte, []int) {
	return fileDescriptor_c1b437217b9e14a2, []int{2}
}

func (m *EntityCreatePayload) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_EntityCreatePayload.Unmarshal(m, b)
}
func (m *EntityCreatePayload) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_EntityCreatePayload.Marshal(b, m, deterministic)
}
func (m *EntityCreatePayload) XXX_Merge(src proto.Message) {
	xxx_messageInfo_EntityCreatePayload.Merge(m, src)
}
func (m *EntityCreatePayload) XXX_Size() int {
	return xxx_messageInfo_EntityCreatePayload.Size(m)
}
func (m *EntityCreatePayload) XXX_DiscardUnknown() {
	xxx_messageInfo_EntityCreatePayload.DiscardUnknown(m)
}

var xxx_messageInfo_EntityCreatePayload proto.InternalMessageInfo

func (m *EntityCreatePayload) GetCollaborators() []string {
	if m != nil {
		return m.Collaborators
	}
	return nil
}

func (m *EntityCreatePayload) GetData() *EntityData {
	if m != nil {
		return m.Data
	}
	return nil
}

type EntityUpdatePayload struct {
	Identifier           string      `protobuf:"bytes,1,opt,name=identifier,proto3" json:"identifier,omitempty"`
	Collaborators        []string    `protobuf:"bytes,2,rep,name=collaborators,proto3" json:"collaborators,omitempty"`
	Data                 *EntityData `protobuf:"bytes,3,opt,name=data,proto3" json:"data,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *EntityUpdatePayload) Reset()         { *m = EntityUpdatePayload{} }
func (m *EntityUpdatePayload) String() string { return proto.CompactTextString(m) }
func (*EntityUpdatePayload) ProtoMessage()    {}
func (*EntityUpdatePayload) Descriptor() ([]byte, []int) {
	return fileDescriptor_c1b437217b9e14a2, []int{3}
}

func (m *EntityUpdatePayload) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_EntityUpdatePayload.Unmarshal(m, b)
}
func (m *EntityUpdatePayload) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_EntityUpdatePayload.Marshal(b, m, deterministic)
}
func (m *EntityUpdatePayload) XXX_Merge(src proto.Message) {
	xxx_messageInfo_EntityUpdatePayload.Merge(m, src)
}
func (m *EntityUpdatePayload) XXX_Size() int {
	return xxx_messageInfo_EntityUpdatePayload.Size(m)
}
func (m *EntityUpdatePayload) XXX_DiscardUnknown() {
	xxx_messageInfo_EntityUpdatePayload.DiscardUnknown(m)
}

var xxx_messageInfo_EntityUpdatePayload proto.InternalMessageInfo

func (m *EntityUpdatePayload) GetIdentifier() string {
	if m != nil {
		return m.Identifier
	}
	return ""
}

func (m *EntityUpdatePayload) GetCollaborators() []string {
	if m != nil {
		return m.Collaborators
	}
	return nil
}

func (m *EntityUpdatePayload) GetData() *EntityData {
	if m != nil {
		return m.Data
	}
	return nil
}

type EntityResponse struct {
	Header               *ResponseHeader `protobuf:"bytes,1,opt,name=header,proto3" json:"header,omitempty"`
	Data                 *EntityData     `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`
	XXX_NoUnkeyedLiteral struct{}        `json:"-"`
	XXX_unrecognized     []byte          `json:"-"`
	XXX_sizecache        int32           `json:"-"`
}

func (m *EntityResponse) Reset()         { *m = EntityResponse{} }
func (m *EntityResponse) String() string { return proto.CompactTextString(m) }
func (*EntityResponse) ProtoMessage()    {}
func (*EntityResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_c1b437217b9e14a2, []int{4}
}

func (m *EntityResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_EntityResponse.Unmarshal(m, b)
}
func (m *EntityResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_EntityResponse.Marshal(b, m, deterministic)
}
func (m *EntityResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_EntityResponse.Merge(m, src)
}
func (m *EntityResponse) XXX_Size() int {
	return xxx_messageInfo_EntityResponse.Size(m)
}
func (m *EntityResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_EntityResponse.DiscardUnknown(m)
}

var xxx_messageInfo_EntityResponse proto.InternalMessageInfo

func (m *EntityResponse) GetHeader() *ResponseHeader {
	if m != nil {
		return m.Header
	}
	return nil
}

func (m *EntityResponse) GetData() *EntityData {
	if m != nil {
		return m.Data
	}
	return nil
}

// ResponseHeader contains a set of common fields for most document
type ResponseHeader struct {
	DocumentId           string   `protobuf:"bytes,1,opt,name=document_id,json=documentId,proto3" json:"document_id,omitempty"`
	VersionId            string   `protobuf:"bytes,2,opt,name=version_id,json=versionId,proto3" json:"version_id,omitempty"`
	State                string   `protobuf:"bytes,3,opt,name=state,proto3" json:"state,omitempty"`
	Collaborators        []string `protobuf:"bytes,4,rep,name=collaborators,proto3" json:"collaborators,omitempty"`
	TransactionId        string   `protobuf:"bytes,5,opt,name=transaction_id,json=transactionId,proto3" json:"transaction_id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ResponseHeader) Reset()         { *m = ResponseHeader{} }
func (m *ResponseHeader) String() string { return proto.CompactTextString(m) }
func (*ResponseHeader) ProtoMessage()    {}
func (*ResponseHeader) Descriptor() ([]byte, []int) {
	return fileDescriptor_c1b437217b9e14a2, []int{5}
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

func (m *ResponseHeader) GetCollaborators() []string {
	if m != nil {
		return m.Collaborators
	}
	return nil
}

func (m *ResponseHeader) GetTransactionId() string {
	if m != nil {
		return m.TransactionId
	}
	return ""
}

// EntityData is the default entity schema
type EntityData struct {
	Identity  string `protobuf:"bytes,1,opt,name=identity,proto3" json:"identity,omitempty"`
	LegalName string `protobuf:"bytes,2,opt,name=legal_name,json=legalName,proto3" json:"legal_name,omitempty"`
	// address
	Addresses []*entity.Address `protobuf:"bytes,3,rep,name=addresses,proto3" json:"addresses,omitempty"`
	// tax information
	PaymentDetails []*entity.PaymentDetail `protobuf:"bytes,4,rep,name=payment_details,json=paymentDetails,proto3" json:"payment_details,omitempty"`
	// Entity contact list
	Contacts             []*entity.Contact `protobuf:"bytes,5,rep,name=contacts,proto3" json:"contacts,omitempty"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (m *EntityData) Reset()         { *m = EntityData{} }
func (m *EntityData) String() string { return proto.CompactTextString(m) }
func (*EntityData) ProtoMessage()    {}
func (*EntityData) Descriptor() ([]byte, []int) {
	return fileDescriptor_c1b437217b9e14a2, []int{6}
}

func (m *EntityData) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_EntityData.Unmarshal(m, b)
}
func (m *EntityData) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_EntityData.Marshal(b, m, deterministic)
}
func (m *EntityData) XXX_Merge(src proto.Message) {
	xxx_messageInfo_EntityData.Merge(m, src)
}
func (m *EntityData) XXX_Size() int {
	return xxx_messageInfo_EntityData.Size(m)
}
func (m *EntityData) XXX_DiscardUnknown() {
	xxx_messageInfo_EntityData.DiscardUnknown(m)
}

var xxx_messageInfo_EntityData proto.InternalMessageInfo

func (m *EntityData) GetIdentity() string {
	if m != nil {
		return m.Identity
	}
	return ""
}

func (m *EntityData) GetLegalName() string {
	if m != nil {
		return m.LegalName
	}
	return ""
}

func (m *EntityData) GetAddresses() []*entity.Address {
	if m != nil {
		return m.Addresses
	}
	return nil
}

func (m *EntityData) GetPaymentDetails() []*entity.PaymentDetail {
	if m != nil {
		return m.PaymentDetails
	}
	return nil
}

func (m *EntityData) GetContacts() []*entity.Contact {
	if m != nil {
		return m.Contacts
	}
	return nil
}

type EntityRelationshipData struct {
	OwnerIdentity        string   `protobuf:"bytes,1,opt,name=owner_identity,json=ownerIdentity,proto3" json:"owner_identity,omitempty"`
	Label                string   `protobuf:"bytes,2,opt,name=label,proto3" json:"label,omitempty"`
	TargetIdentity       string   `protobuf:"bytes,3,opt,name=target_identity,json=targetIdentity,proto3" json:"target_identity,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *EntityRelationshipData) Reset()         { *m = EntityRelationshipData{} }
func (m *EntityRelationshipData) String() string { return proto.CompactTextString(m) }
func (*EntityRelationshipData) ProtoMessage()    {}
func (*EntityRelationshipData) Descriptor() ([]byte, []int) {
	return fileDescriptor_c1b437217b9e14a2, []int{7}
}

func (m *EntityRelationshipData) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_EntityRelationshipData.Unmarshal(m, b)
}
func (m *EntityRelationshipData) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_EntityRelationshipData.Marshal(b, m, deterministic)
}
func (m *EntityRelationshipData) XXX_Merge(src proto.Message) {
	xxx_messageInfo_EntityRelationshipData.Merge(m, src)
}
func (m *EntityRelationshipData) XXX_Size() int {
	return xxx_messageInfo_EntityRelationshipData.Size(m)
}
func (m *EntityRelationshipData) XXX_DiscardUnknown() {
	xxx_messageInfo_EntityRelationshipData.DiscardUnknown(m)
}

var xxx_messageInfo_EntityRelationshipData proto.InternalMessageInfo

func (m *EntityRelationshipData) GetOwnerIdentity() string {
	if m != nil {
		return m.OwnerIdentity
	}
	return ""
}

func (m *EntityRelationshipData) GetLabel() string {
	if m != nil {
		return m.Label
	}
	return ""
}

func (m *EntityRelationshipData) GetTargetIdentity() string {
	if m != nil {
		return m.TargetIdentity
	}
	return ""
}

type EntityRelationshipCreatePayload struct {
	Data                 *EntityRelationshipData `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                `json:"-"`
	XXX_unrecognized     []byte                  `json:"-"`
	XXX_sizecache        int32                   `json:"-"`
}

func (m *EntityRelationshipCreatePayload) Reset()         { *m = EntityRelationshipCreatePayload{} }
func (m *EntityRelationshipCreatePayload) String() string { return proto.CompactTextString(m) }
func (*EntityRelationshipCreatePayload) ProtoMessage()    {}
func (*EntityRelationshipCreatePayload) Descriptor() ([]byte, []int) {
	return fileDescriptor_c1b437217b9e14a2, []int{8}
}

func (m *EntityRelationshipCreatePayload) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_EntityRelationshipCreatePayload.Unmarshal(m, b)
}
func (m *EntityRelationshipCreatePayload) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_EntityRelationshipCreatePayload.Marshal(b, m, deterministic)
}
func (m *EntityRelationshipCreatePayload) XXX_Merge(src proto.Message) {
	xxx_messageInfo_EntityRelationshipCreatePayload.Merge(m, src)
}
func (m *EntityRelationshipCreatePayload) XXX_Size() int {
	return xxx_messageInfo_EntityRelationshipCreatePayload.Size(m)
}
func (m *EntityRelationshipCreatePayload) XXX_DiscardUnknown() {
	xxx_messageInfo_EntityRelationshipCreatePayload.DiscardUnknown(m)
}

var xxx_messageInfo_EntityRelationshipCreatePayload proto.InternalMessageInfo

func (m *EntityRelationshipCreatePayload) GetData() *EntityRelationshipData {
	if m != nil {
		return m.Data
	}
	return nil
}

type EntityRelationshipUpdatePayload struct {
	Data                 *EntityRelationshipData `protobuf:"bytes,3,opt,name=data,proto3" json:"data,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                `json:"-"`
	XXX_unrecognized     []byte                  `json:"-"`
	XXX_sizecache        int32                   `json:"-"`
}

func (m *EntityRelationshipUpdatePayload) Reset()         { *m = EntityRelationshipUpdatePayload{} }
func (m *EntityRelationshipUpdatePayload) String() string { return proto.CompactTextString(m) }
func (*EntityRelationshipUpdatePayload) ProtoMessage()    {}
func (*EntityRelationshipUpdatePayload) Descriptor() ([]byte, []int) {
	return fileDescriptor_c1b437217b9e14a2, []int{9}
}

func (m *EntityRelationshipUpdatePayload) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_EntityRelationshipUpdatePayload.Unmarshal(m, b)
}
func (m *EntityRelationshipUpdatePayload) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_EntityRelationshipUpdatePayload.Marshal(b, m, deterministic)
}
func (m *EntityRelationshipUpdatePayload) XXX_Merge(src proto.Message) {
	xxx_messageInfo_EntityRelationshipUpdatePayload.Merge(m, src)
}
func (m *EntityRelationshipUpdatePayload) XXX_Size() int {
	return xxx_messageInfo_EntityRelationshipUpdatePayload.Size(m)
}
func (m *EntityRelationshipUpdatePayload) XXX_DiscardUnknown() {
	xxx_messageInfo_EntityRelationshipUpdatePayload.DiscardUnknown(m)
}

var xxx_messageInfo_EntityRelationshipUpdatePayload proto.InternalMessageInfo

func (m *EntityRelationshipUpdatePayload) GetData() *EntityRelationshipData {
	if m != nil {
		return m.Data
	}
	return nil
}

type EntityRelationshipResponse struct {
	Header               *ResponseHeader         `protobuf:"bytes,1,opt,name=header,proto3" json:"header,omitempty"`
	Data                 *EntityRelationshipData `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                `json:"-"`
	XXX_unrecognized     []byte                  `json:"-"`
	XXX_sizecache        int32                   `json:"-"`
}

func (m *EntityRelationshipResponse) Reset()         { *m = EntityRelationshipResponse{} }
func (m *EntityRelationshipResponse) String() string { return proto.CompactTextString(m) }
func (*EntityRelationshipResponse) ProtoMessage()    {}
func (*EntityRelationshipResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_c1b437217b9e14a2, []int{10}
}

func (m *EntityRelationshipResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_EntityRelationshipResponse.Unmarshal(m, b)
}
func (m *EntityRelationshipResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_EntityRelationshipResponse.Marshal(b, m, deterministic)
}
func (m *EntityRelationshipResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_EntityRelationshipResponse.Merge(m, src)
}
func (m *EntityRelationshipResponse) XXX_Size() int {
	return xxx_messageInfo_EntityRelationshipResponse.Size(m)
}
func (m *EntityRelationshipResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_EntityRelationshipResponse.DiscardUnknown(m)
}

var xxx_messageInfo_EntityRelationshipResponse proto.InternalMessageInfo

func (m *EntityRelationshipResponse) GetHeader() *ResponseHeader {
	if m != nil {
		return m.Header
	}
	return nil
}

func (m *EntityRelationshipResponse) GetData() *EntityRelationshipData {
	if m != nil {
		return m.Data
	}
	return nil
}

func init() {
	proto.RegisterType((*GetRequest)(nil), "entity.GetRequest")
	proto.RegisterType((*GetVersionRequest)(nil), "entity.GetVersionRequest")
	proto.RegisterType((*EntityCreatePayload)(nil), "entity.EntityCreatePayload")
	proto.RegisterType((*EntityUpdatePayload)(nil), "entity.EntityUpdatePayload")
	proto.RegisterType((*EntityResponse)(nil), "entity.EntityResponse")
	proto.RegisterType((*ResponseHeader)(nil), "entity.ResponseHeader")
	proto.RegisterType((*EntityData)(nil), "entity.EntityData")
	proto.RegisterType((*EntityRelationshipData)(nil), "entity.EntityRelationshipData")
	proto.RegisterType((*EntityRelationshipCreatePayload)(nil), "entity.EntityRelationshipCreatePayload")
	proto.RegisterType((*EntityRelationshipUpdatePayload)(nil), "entity.EntityRelationshipUpdatePayload")
	proto.RegisterType((*EntityRelationshipResponse)(nil), "entity.EntityRelationshipResponse")
}

func init() { proto.RegisterFile("entity/service.proto", fileDescriptor_c1b437217b9e14a2) }

var fileDescriptor_c1b437217b9e14a2 = []byte{
	// 778 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x55, 0xdd, 0x6e, 0xeb, 0x44,
	0x10, 0x96, 0x9b, 0x26, 0x6d, 0x26, 0x6d, 0xa2, 0x6e, 0x43, 0x94, 0x1a, 0x68, 0x2d, 0x43, 0x4b,
	0x04, 0x6d, 0x23, 0x19, 0x21, 0x24, 0x2e, 0x90, 0xd2, 0x16, 0x85, 0x5c, 0x50, 0x45, 0x46, 0xe5,
	0x82, 0x9b, 0x6a, 0x63, 0x4f, 0x1d, 0x23, 0xc7, 0x6b, 0xbc, 0xdb, 0x56, 0x51, 0x55, 0x84, 0xe0,
	0x01, 0x90, 0xca, 0xa3, 0xf0, 0x28, 0x3c, 0x02, 0x5c, 0x9f, 0x67, 0x38, 0xca, 0xee, 0xba, 0x89,
	0x93, 0xf4, 0x34, 0x3a, 0x57, 0xd6, 0x7e, 0xf3, 0xf3, 0x7d, 0x33, 0xbb, 0x33, 0x86, 0x3a, 0xc6,
	0x22, 0x14, 0xe3, 0x36, 0xc7, 0xf4, 0x2e, 0xf4, 0xf0, 0x34, 0x49, 0x99, 0x60, 0xa4, 0xa4, 0x50,
	0x73, 0x57, 0x5b, 0xd5, 0x47, 0x19, 0xcd, 0x8f, 0x02, 0xc6, 0x82, 0x08, 0xdb, 0x34, 0x09, 0xdb,
	0x34, 0x8e, 0x99, 0xa0, 0x22, 0x64, 0x31, 0xd7, 0xd6, 0x63, 0xf9, 0xf1, 0x4e, 0x02, 0x8c, 0x4f,
	0xf8, 0x3d, 0x0d, 0x02, 0x4c, 0xdb, 0x2c, 0x91, 0x1e, 0x8b, 0xde, 0xf6, 0x31, 0x40, 0x17, 0x85,
	0x8b, 0xbf, 0xde, 0x22, 0x17, 0x64, 0x1f, 0x20, 0xf4, 0x27, 0x5c, 0x37, 0x21, 0xa6, 0x4d, 0xc3,
	0x32, 0x5a, 0x65, 0x77, 0x06, 0xb1, 0x7f, 0x80, 0x9d, 0x2e, 0x8a, 0x9f, 0x30, 0xe5, 0x21, 0x8b,
	0x57, 0x0c, 0x22, 0x4d, 0xd8, 0xb8, 0x53, 0x11, 0xcd, 0x35, 0x69, 0xcc, 0x8e, 0xb6, 0x07, 0xbb,
	0xdf, 0xc9, 0xc2, 0xce, 0x53, 0xa4, 0x02, 0xfb, 0x74, 0x1c, 0x31, 0xea, 0x93, 0x4f, 0x61, 0xdb,
	0x63, 0x51, 0x44, 0x07, 0x2c, 0xa5, 0x82, 0xa5, 0xbc, 0x69, 0x58, 0x85, 0x56, 0xd9, 0xcd, 0x83,
	0xe4, 0x08, 0xd6, 0x7d, 0x2a, 0xa8, 0xcc, 0x59, 0x71, 0xc8, 0xa9, 0x6e, 0x91, 0x4a, 0x78, 0x41,
	0x05, 0x75, 0xa5, 0xdd, 0xfe, 0xd3, 0xc8, 0x58, 0xae, 0x12, 0x7f, 0x86, 0xe5, 0x35, 0xd9, 0x0b,
	0x2a, 0xd6, 0xde, 0xa5, 0xa2, 0xf0, 0x8a, 0x8a, 0x21, 0x54, 0x15, 0xe6, 0x22, 0x4f, 0x58, 0xcc,
	0x91, 0x9c, 0x42, 0x69, 0x88, 0xd4, 0xd7, 0xdc, 0x15, 0xa7, 0x91, 0xc5, 0x66, 0x1e, 0xdf, 0x4b,
	0xab, 0xab, 0xbd, 0x56, 0xae, 0xf7, 0x1f, 0x03, 0xaa, 0xf9, 0x14, 0xe4, 0x00, 0x2a, 0x3e, 0xf3,
	0x6e, 0x47, 0x18, 0x8b, 0xeb, 0xd0, 0xcf, 0x6a, 0xcd, 0xa0, 0x9e, 0x4f, 0x3e, 0x06, 0xd0, 0x77,
	0x32, 0xb1, 0xab, 0x5b, 0x2a, 0x6b, 0xa4, 0xe7, 0x93, 0x3a, 0x14, 0xb9, 0xa0, 0x02, 0x65, 0x95,
	0x65, 0x57, 0x1d, 0x16, 0x1b, 0xb4, 0xbe, 0xac, 0x41, 0x87, 0x50, 0x15, 0x29, 0x8d, 0x39, 0xf5,
	0x84, 0x4e, 0x5f, 0x94, 0x49, 0xb6, 0x67, 0xd0, 0x9e, 0x6f, 0xff, 0x67, 0x00, 0x4c, 0x4b, 0x21,
	0x26, 0x6c, 0xaa, 0xab, 0x10, 0x63, 0x2d, 0xf7, 0xf9, 0x3c, 0x11, 0x1b, 0x61, 0x40, 0xa3, 0xeb,
	0x98, 0x8e, 0x30, 0x13, 0x2b, 0x91, 0x4b, 0x3a, 0x42, 0x72, 0x02, 0x65, 0xea, 0xfb, 0x29, 0x72,
	0x8e, 0xbc, 0x59, 0xb0, 0x0a, 0xad, 0x8a, 0x53, 0xcb, 0x9a, 0xd5, 0x51, 0x06, 0x77, 0xea, 0x41,
	0xbe, 0x85, 0x5a, 0x42, 0xc7, 0xb2, 0x35, 0x3e, 0x0a, 0x1a, 0x46, 0xaa, 0x8e, 0x8a, 0xf3, 0x41,
	0x16, 0xd4, 0x57, 0xe6, 0x0b, 0x69, 0x75, 0xab, 0xc9, 0xec, 0x91, 0x93, 0x2f, 0x60, 0xd3, 0x63,
	0xb1, 0xa0, 0x9e, 0xe0, 0xcd, 0x62, 0x9e, 0xed, 0x5c, 0xe1, 0xee, 0xb3, 0x83, 0xfd, 0x1b, 0x34,
	0xb2, 0x57, 0x10, 0xa9, 0x31, 0x1c, 0x86, 0x89, 0x2c, 0xf8, 0x10, 0xaa, 0xec, 0x3e, 0xc6, 0xf4,
	0x7a, 0xae, 0xec, 0x6d, 0x89, 0xf6, 0xb2, 0xda, 0xeb, 0x50, 0x8c, 0xe8, 0x00, 0x23, 0x5d, 0xb6,
	0x3a, 0x90, 0xcf, 0xa0, 0x26, 0x68, 0x1a, 0xa0, 0x98, 0x46, 0xab, 0x9b, 0xaa, 0x2a, 0x38, 0x0b,
	0xb7, 0xaf, 0xe0, 0x60, 0x91, 0x3f, 0x3f, 0x7c, 0x8e, 0x7e, 0x66, 0xea, 0x51, 0xee, 0xe7, 0x9f,
	0xd9, 0xbc, 0x6c, 0xfd, 0xe4, 0x96, 0xa6, 0xcd, 0x4f, 0x9b, 0x93, 0x9b, 0x93, 0xd5, 0xd2, 0xfe,
	0x6e, 0x80, 0xb9, 0xe8, 0xf0, 0xde, 0x03, 0xe4, 0xe4, 0x06, 0x68, 0x25, 0x09, 0xce, 0x9b, 0x02,
	0xd4, 0x2e, 0xf4, 0x9c, 0xfc, 0xa8, 0x36, 0x34, 0x09, 0xa0, 0xa4, 0x5a, 0x46, 0x3e, 0xcc, 0xe7,
	0xc8, 0x35, 0xd2, 0x6c, 0xcc, 0x13, 0x28, 0x51, 0x76, 0xeb, 0xa9, 0xb3, 0x6b, 0xee, 0x28, 0x5f,
	0x6e, 0xd1, 0xd8, 0x52, 0x6e, 0x7f, 0xfc, 0xfb, 0xff, 0xdf, 0x6b, 0x5b, 0xf6, 0x86, 0x5e, 0xf5,
	0xdf, 0x18, 0x9f, 0x13, 0x01, 0x25, 0xd5, 0xc4, 0x79, 0xa2, 0x5c, 0x6b, 0x5f, 0x24, 0xfa, 0x4a,
	0x12, 0x29, 0xdf, 0x79, 0xa2, 0x3d, 0xb3, 0xae, 0x89, 0xda, 0x0f, 0xd3, 0x9d, 0xf7, 0x38, 0x61,
	0xfd, 0xcb, 0x90, 0xbf, 0x04, 0xbd, 0xe4, 0xc9, 0x5e, 0x96, 0x7d, 0x61, 0xf1, 0xbf, 0x48, 0x7c,
	0xf9, 0xd4, 0x39, 0x34, 0x3f, 0xe9, 0xa2, 0xb0, 0xa8, 0xc5, 0x13, 0xf4, 0xc2, 0x9b, 0xd0, 0xb3,
	0xf4, 0x36, 0xb1, 0xd8, 0xcd, 0x9c, 0x14, 0x8b, 0xec, 0x2f, 0x93, 0xd2, 0x7e, 0xd0, 0x11, 0x8f,
	0xe4, 0x17, 0x28, 0x74, 0x51, 0x10, 0x32, 0xa3, 0xe4, 0x35, 0x09, 0x5f, 0x3f, 0x75, 0x9a, 0x66,
	0x63, 0x22, 0x41, 0x0c, 0xd1, 0xf2, 0x6e, 0xd3, 0x14, 0x63, 0x31, 0xcb, 0xda, 0x20, 0x4b, 0x1b,
	0x70, 0x76, 0x04, 0xe0, 0xb1, 0x91, 0xce, 0x7a, 0xb6, 0xa5, 0xef, 0xbc, 0x3f, 0xf9, 0x57, 0xf6,
	0x8d, 0x9f, 0x37, 0x15, 0x9e, 0x0c, 0x06, 0x25, 0xf9, 0xfb, 0xfc, 0xf2, 0x6d, 0x00, 0x00, 0x00,
	0xff, 0xff, 0x8e, 0x2c, 0x7a, 0xaa, 0xbf, 0x07, 0x00, 0x00,
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
	Create(ctx context.Context, in *EntityCreatePayload, opts ...grpc.CallOption) (*EntityResponse, error)
	Update(ctx context.Context, in *EntityUpdatePayload, opts ...grpc.CallOption) (*EntityResponse, error)
	GetVersion(ctx context.Context, in *GetVersionRequest, opts ...grpc.CallOption) (*EntityResponse, error)
	Get(ctx context.Context, in *GetRequest, opts ...grpc.CallOption) (*EntityResponse, error)
}

type documentServiceClient struct {
	cc *grpc.ClientConn
}

func NewDocumentServiceClient(cc *grpc.ClientConn) DocumentServiceClient {
	return &documentServiceClient{cc}
}

func (c *documentServiceClient) Create(ctx context.Context, in *EntityCreatePayload, opts ...grpc.CallOption) (*EntityResponse, error) {
	out := new(EntityResponse)
	err := c.cc.Invoke(ctx, "/entity.DocumentService/Create", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *documentServiceClient) Update(ctx context.Context, in *EntityUpdatePayload, opts ...grpc.CallOption) (*EntityResponse, error) {
	out := new(EntityResponse)
	err := c.cc.Invoke(ctx, "/entity.DocumentService/Update", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *documentServiceClient) GetVersion(ctx context.Context, in *GetVersionRequest, opts ...grpc.CallOption) (*EntityResponse, error) {
	out := new(EntityResponse)
	err := c.cc.Invoke(ctx, "/entity.DocumentService/GetVersion", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *documentServiceClient) Get(ctx context.Context, in *GetRequest, opts ...grpc.CallOption) (*EntityResponse, error) {
	out := new(EntityResponse)
	err := c.cc.Invoke(ctx, "/entity.DocumentService/Get", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// DocumentServiceServer is the server API for DocumentService service.
type DocumentServiceServer interface {
	Create(context.Context, *EntityCreatePayload) (*EntityResponse, error)
	Update(context.Context, *EntityUpdatePayload) (*EntityResponse, error)
	GetVersion(context.Context, *GetVersionRequest) (*EntityResponse, error)
	Get(context.Context, *GetRequest) (*EntityResponse, error)
}

func RegisterDocumentServiceServer(s *grpc.Server, srv DocumentServiceServer) {
	s.RegisterService(&_DocumentService_serviceDesc, srv)
}

func _DocumentService_Create_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EntityCreatePayload)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DocumentServiceServer).Create(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/entity.DocumentService/Create",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DocumentServiceServer).Create(ctx, req.(*EntityCreatePayload))
	}
	return interceptor(ctx, in, info, handler)
}

func _DocumentService_Update_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(EntityUpdatePayload)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DocumentServiceServer).Update(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/entity.DocumentService/Update",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DocumentServiceServer).Update(ctx, req.(*EntityUpdatePayload))
	}
	return interceptor(ctx, in, info, handler)
}

func _DocumentService_GetVersion_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetVersionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DocumentServiceServer).GetVersion(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/entity.DocumentService/GetVersion",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DocumentServiceServer).GetVersion(ctx, req.(*GetVersionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _DocumentService_Get_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DocumentServiceServer).Get(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/entity.DocumentService/Get",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DocumentServiceServer).Get(ctx, req.(*GetRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _DocumentService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "entity.DocumentService",
	HandlerType: (*DocumentServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Create",
			Handler:    _DocumentService_Create_Handler,
		},
		{
			MethodName: "Update",
			Handler:    _DocumentService_Update_Handler,
		},
		{
			MethodName: "GetVersion",
			Handler:    _DocumentService_GetVersion_Handler,
		},
		{
			MethodName: "Get",
			Handler:    _DocumentService_Get_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "entity/service.proto",
}

// Code generated by protoc-gen-go. DO NOT EDIT.
// source: config/service.proto

package configpb

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import account "github.com/centrifuge/go-centrifuge/protobufs/gen/go/account"
import duration "github.com/golang/protobuf/ptypes/duration"
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

type ConfigData struct {
	StoragePath               string               `protobuf:"bytes,1,opt,name=storage_path,json=storagePath" json:"storage_path,omitempty"`
	P2PPort                   int32                `protobuf:"varint,2,opt,name=p2p_port,json=p2pPort" json:"p2p_port,omitempty"`
	P2PExternalIp             string               `protobuf:"bytes,3,opt,name=p2p_external_ip,json=p2pExternalIp" json:"p2p_external_ip,omitempty"`
	P2PConnectionTimeout      *duration.Duration   `protobuf:"bytes,4,opt,name=p2p_connection_timeout,json=p2pConnectionTimeout" json:"p2p_connection_timeout,omitempty"`
	ServerPort                int32                `protobuf:"varint,5,opt,name=server_port,json=serverPort" json:"server_port,omitempty"`
	ServerAddress             string               `protobuf:"bytes,6,opt,name=server_address,json=serverAddress" json:"server_address,omitempty"`
	NumWorkers                int32                `protobuf:"varint,7,opt,name=num_workers,json=numWorkers" json:"num_workers,omitempty"`
	WorkerWaitTimeMs          int32                `protobuf:"varint,8,opt,name=worker_wait_time_ms,json=workerWaitTimeMs" json:"worker_wait_time_ms,omitempty"`
	EthNodeUrl                string               `protobuf:"bytes,9,opt,name=eth_node_url,json=ethNodeUrl" json:"eth_node_url,omitempty"`
	EthContextReadWaitTimeout *duration.Duration   `protobuf:"bytes,10,opt,name=eth_context_read_wait_timeout,json=ethContextReadWaitTimeout" json:"eth_context_read_wait_timeout,omitempty"`
	EthContextWaitTimeout     *duration.Duration   `protobuf:"bytes,11,opt,name=eth_context_wait_timeout,json=ethContextWaitTimeout" json:"eth_context_wait_timeout,omitempty"`
	EthIntervalRetry          *duration.Duration   `protobuf:"bytes,12,opt,name=eth_interval_retry,json=ethIntervalRetry" json:"eth_interval_retry,omitempty"`
	EthMaxRetries             uint32               `protobuf:"varint,13,opt,name=eth_max_retries,json=ethMaxRetries" json:"eth_max_retries,omitempty"`
	EthGasPrice               uint64               `protobuf:"varint,14,opt,name=eth_gas_price,json=ethGasPrice" json:"eth_gas_price,omitempty"`
	EthGasLimit               uint64               `protobuf:"varint,15,opt,name=eth_gas_limit,json=ethGasLimit" json:"eth_gas_limit,omitempty"`
	TxPoolEnabled             bool                 `protobuf:"varint,16,opt,name=tx_pool_enabled,json=txPoolEnabled" json:"tx_pool_enabled,omitempty"`
	Network                   string               `protobuf:"bytes,17,opt,name=network" json:"network,omitempty"`
	BootstrapPeers            []string             `protobuf:"bytes,18,rep,name=bootstrap_peers,json=bootstrapPeers" json:"bootstrap_peers,omitempty"`
	NetworkId                 uint32               `protobuf:"varint,19,opt,name=network_id,json=networkId" json:"network_id,omitempty"`
	MainIdentity              *account.AccountData `protobuf:"bytes,20,opt,name=main_identity,json=mainIdentity" json:"main_identity,omitempty"`
	SmartContractAddresses    map[string]string    `protobuf:"bytes,21,rep,name=smart_contract_addresses,json=smartContractAddresses" json:"smart_contract_addresses,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	SmartContractBytecode     map[string]string    `protobuf:"bytes,23,rep,name=smart_contract_bytecode,json=smartContractBytecode" json:"smart_contract_bytecode,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	PprofEnabled              bool                 `protobuf:"varint,22,opt,name=pprof_enabled,json=pprofEnabled" json:"pprof_enabled,omitempty"`
	XXX_NoUnkeyedLiteral      struct{}             `json:"-"`
	XXX_unrecognized          []byte               `json:"-"`
	XXX_sizecache             int32                `json:"-"`
}

func (m *ConfigData) Reset()         { *m = ConfigData{} }
func (m *ConfigData) String() string { return proto.CompactTextString(m) }
func (*ConfigData) ProtoMessage()    {}
func (*ConfigData) Descriptor() ([]byte, []int) {
	return fileDescriptor_service_3fe01735a8c62984, []int{0}
}
func (m *ConfigData) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ConfigData.Unmarshal(m, b)
}
func (m *ConfigData) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ConfigData.Marshal(b, m, deterministic)
}
func (dst *ConfigData) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ConfigData.Merge(dst, src)
}
func (m *ConfigData) XXX_Size() int {
	return xxx_messageInfo_ConfigData.Size(m)
}
func (m *ConfigData) XXX_DiscardUnknown() {
	xxx_messageInfo_ConfigData.DiscardUnknown(m)
}

var xxx_messageInfo_ConfigData proto.InternalMessageInfo

func (m *ConfigData) GetStoragePath() string {
	if m != nil {
		return m.StoragePath
	}
	return ""
}

func (m *ConfigData) GetP2PPort() int32 {
	if m != nil {
		return m.P2PPort
	}
	return 0
}

func (m *ConfigData) GetP2PExternalIp() string {
	if m != nil {
		return m.P2PExternalIp
	}
	return ""
}

func (m *ConfigData) GetP2PConnectionTimeout() *duration.Duration {
	if m != nil {
		return m.P2PConnectionTimeout
	}
	return nil
}

func (m *ConfigData) GetServerPort() int32 {
	if m != nil {
		return m.ServerPort
	}
	return 0
}

func (m *ConfigData) GetServerAddress() string {
	if m != nil {
		return m.ServerAddress
	}
	return ""
}

func (m *ConfigData) GetNumWorkers() int32 {
	if m != nil {
		return m.NumWorkers
	}
	return 0
}

func (m *ConfigData) GetWorkerWaitTimeMs() int32 {
	if m != nil {
		return m.WorkerWaitTimeMs
	}
	return 0
}

func (m *ConfigData) GetEthNodeUrl() string {
	if m != nil {
		return m.EthNodeUrl
	}
	return ""
}

func (m *ConfigData) GetEthContextReadWaitTimeout() *duration.Duration {
	if m != nil {
		return m.EthContextReadWaitTimeout
	}
	return nil
}

func (m *ConfigData) GetEthContextWaitTimeout() *duration.Duration {
	if m != nil {
		return m.EthContextWaitTimeout
	}
	return nil
}

func (m *ConfigData) GetEthIntervalRetry() *duration.Duration {
	if m != nil {
		return m.EthIntervalRetry
	}
	return nil
}

func (m *ConfigData) GetEthMaxRetries() uint32 {
	if m != nil {
		return m.EthMaxRetries
	}
	return 0
}

func (m *ConfigData) GetEthGasPrice() uint64 {
	if m != nil {
		return m.EthGasPrice
	}
	return 0
}

func (m *ConfigData) GetEthGasLimit() uint64 {
	if m != nil {
		return m.EthGasLimit
	}
	return 0
}

func (m *ConfigData) GetTxPoolEnabled() bool {
	if m != nil {
		return m.TxPoolEnabled
	}
	return false
}

func (m *ConfigData) GetNetwork() string {
	if m != nil {
		return m.Network
	}
	return ""
}

func (m *ConfigData) GetBootstrapPeers() []string {
	if m != nil {
		return m.BootstrapPeers
	}
	return nil
}

func (m *ConfigData) GetNetworkId() uint32 {
	if m != nil {
		return m.NetworkId
	}
	return 0
}

func (m *ConfigData) GetMainIdentity() *account.AccountData {
	if m != nil {
		return m.MainIdentity
	}
	return nil
}

func (m *ConfigData) GetSmartContractAddresses() map[string]string {
	if m != nil {
		return m.SmartContractAddresses
	}
	return nil
}

func (m *ConfigData) GetSmartContractBytecode() map[string]string {
	if m != nil {
		return m.SmartContractBytecode
	}
	return nil
}

func (m *ConfigData) GetPprofEnabled() bool {
	if m != nil {
		return m.PprofEnabled
	}
	return false
}

func init() {
	proto.RegisterType((*ConfigData)(nil), "config.ConfigData")
	proto.RegisterMapType((map[string]string)(nil), "config.ConfigData.SmartContractAddressesEntry")
	proto.RegisterMapType((map[string]string)(nil), "config.ConfigData.SmartContractBytecodeEntry")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for ConfigService service

type ConfigServiceClient interface {
	GetConfig(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*ConfigData, error)
}

type configServiceClient struct {
	cc *grpc.ClientConn
}

func NewConfigServiceClient(cc *grpc.ClientConn) ConfigServiceClient {
	return &configServiceClient{cc}
}

func (c *configServiceClient) GetConfig(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*ConfigData, error) {
	out := new(ConfigData)
	err := grpc.Invoke(ctx, "/config.ConfigService/GetConfig", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for ConfigService service

type ConfigServiceServer interface {
	GetConfig(context.Context, *empty.Empty) (*ConfigData, error)
}

func RegisterConfigServiceServer(s *grpc.Server, srv ConfigServiceServer) {
	s.RegisterService(&_ConfigService_serviceDesc, srv)
}

func _ConfigService_GetConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(empty.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ConfigServiceServer).GetConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/config.ConfigService/GetConfig",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ConfigServiceServer).GetConfig(ctx, req.(*empty.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

var _ConfigService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "config.ConfigService",
	HandlerType: (*ConfigServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetConfig",
			Handler:    _ConfigService_GetConfig_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "config/service.proto",
}

func init() { proto.RegisterFile("config/service.proto", fileDescriptor_service_3fe01735a8c62984) }

var fileDescriptor_service_3fe01735a8c62984 = []byte{
	// 838 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x54, 0xdd, 0x6e, 0x1b, 0x45,
	0x14, 0xd6, 0x36, 0xcd, 0x8f, 0xc7, 0x76, 0x92, 0x4e, 0x9d, 0x74, 0xe2, 0x52, 0x58, 0x52, 0x51,
	0x7c, 0x41, 0xd6, 0x92, 0xb9, 0x01, 0xee, 0x92, 0x34, 0x0a, 0x96, 0x28, 0x58, 0x5b, 0x50, 0x25,
	0x40, 0x1a, 0x8d, 0x77, 0x4f, 0xbc, 0xab, 0xee, 0xce, 0x8c, 0x66, 0x8f, 0x13, 0xfb, 0x96, 0x47,
	0x80, 0x87, 0xe1, 0x41, 0x78, 0x05, 0x1e, 0x04, 0xcd, 0x8f, 0x93, 0xb4, 0x29, 0x44, 0x5c, 0xd9,
	0xfb, 0x7d, 0xdf, 0xf9, 0xce, 0x99, 0x73, 0x66, 0x0e, 0xe9, 0x65, 0x4a, 0x5e, 0x94, 0xb3, 0x61,
	0x03, 0xe6, 0xb2, 0xcc, 0x20, 0xd1, 0x46, 0xa1, 0xa2, 0x1b, 0x1e, 0xed, 0xef, 0x89, 0x2c, 0x53,
	0x73, 0x89, 0xef, 0xd2, 0xfd, 0x8f, 0x66, 0x4a, 0xcd, 0x2a, 0x18, 0x0a, 0x5d, 0x0e, 0x85, 0x94,
	0x0a, 0x05, 0x96, 0x4a, 0x36, 0x81, 0xfd, 0x38, 0xb0, 0xee, 0x6b, 0x3a, 0xbf, 0x18, 0xe6, 0x73,
	0xe3, 0x04, 0x81, 0x7f, 0xfa, 0x3e, 0x0f, 0xb5, 0xc6, 0x65, 0x20, 0xbf, 0x70, 0x3f, 0xd9, 0xd1,
	0x0c, 0xe4, 0x51, 0x73, 0x25, 0x66, 0x33, 0x30, 0x43, 0xa5, 0x9d, 0xfd, 0xdd, 0x54, 0x87, 0x7f,
	0x12, 0x42, 0x4e, 0x5d, 0xa9, 0x2f, 0x05, 0x0a, 0xfa, 0x29, 0xe9, 0x34, 0xa8, 0x8c, 0x98, 0x01,
	0xd7, 0x02, 0x0b, 0x16, 0xc5, 0xd1, 0xa0, 0x95, 0xb6, 0x03, 0x36, 0x11, 0x58, 0xd0, 0x03, 0xb2,
	0xa5, 0x47, 0x9a, 0x6b, 0x65, 0x90, 0x3d, 0x88, 0xa3, 0xc1, 0x7a, 0xba, 0xa9, 0x47, 0x7a, 0xa2,
	0x0c, 0xd2, 0x17, 0x64, 0xc7, 0x52, 0xb0, 0x40, 0x30, 0x52, 0x54, 0xbc, 0xd4, 0x6c, 0xcd, 0x19,
	0x74, 0xf5, 0x48, 0x9f, 0x05, 0x74, 0xac, 0xe9, 0x0f, 0x64, 0xdf, 0xea, 0x32, 0x25, 0x25, 0x64,
	0xb6, 0x1a, 0x8e, 0x65, 0x0d, 0x6a, 0x8e, 0xec, 0x61, 0x1c, 0x0d, 0xda, 0xa3, 0x83, 0xc4, 0x1f,
	0x30, 0x59, 0x1d, 0x30, 0x79, 0x19, 0x1a, 0x90, 0xf6, 0xf4, 0x48, 0x9f, 0x5e, 0xc7, 0xfd, 0xe8,
	0xc3, 0xe8, 0x27, 0xa4, 0x6d, 0xfb, 0x0b, 0xc6, 0x97, 0xb5, 0xee, 0xca, 0x22, 0x1e, 0x72, 0x95,
	0x7d, 0x46, 0xb6, 0x83, 0x40, 0xe4, 0xb9, 0x81, 0xa6, 0x61, 0x1b, 0xbe, 0x30, 0x8f, 0x1e, 0x7b,
	0xd0, 0xfa, 0xc8, 0x79, 0xcd, 0xaf, 0x94, 0x79, 0x0b, 0xa6, 0x61, 0x9b, 0xde, 0x47, 0xce, 0xeb,
	0x37, 0x1e, 0xa1, 0x47, 0xe4, 0xb1, 0x27, 0xf9, 0x95, 0x28, 0xd1, 0x95, 0xcd, 0xeb, 0x86, 0x6d,
	0x39, 0xe1, 0xae, 0xa7, 0xde, 0x88, 0x12, 0x6d, 0x61, 0xaf, 0x1a, 0x1a, 0x93, 0x0e, 0x60, 0xc1,
	0xa5, 0xca, 0x81, 0xcf, 0x4d, 0xc5, 0x5a, 0x2e, 0x29, 0x01, 0x2c, 0xbe, 0x57, 0x39, 0xfc, 0x64,
	0x2a, 0xfa, 0x0b, 0x79, 0x66, 0x15, 0x99, 0x92, 0x08, 0x0b, 0xe4, 0x06, 0x44, 0x7e, 0x63, 0x6d,
	0x3b, 0x42, 0xee, 0xeb, 0xc8, 0x01, 0x60, 0x71, 0xea, 0xc3, 0x53, 0x10, 0xf9, 0x2a, 0xbb, 0x6d,
	0x4b, 0x4a, 0xd8, 0x6d, 0xf3, 0x77, 0x7c, 0xdb, 0xf7, 0xf9, 0xee, 0xdd, 0xf8, 0xde, 0xf6, 0x3c,
	0x27, 0xd4, 0x7a, 0x96, 0x12, 0xc1, 0x5c, 0x8a, 0x8a, 0x1b, 0x40, 0xb3, 0x64, 0x9d, 0xfb, 0xdc,
	0x76, 0x01, 0x8b, 0x71, 0x88, 0x49, 0x6d, 0x88, 0xbd, 0x2c, 0xd6, 0xa8, 0x16, 0x0b, 0xe7, 0x51,
	0x42, 0xc3, 0xba, 0x71, 0x34, 0xe8, 0xa6, 0x5d, 0xc0, 0xe2, 0x95, 0x58, 0xa4, 0x1e, 0xa4, 0x87,
	0xc4, 0x02, 0x7c, 0x26, 0x1a, 0xae, 0x4d, 0x99, 0x01, 0xdb, 0x8e, 0xa3, 0xc1, 0xc3, 0xb4, 0x0d,
	0x58, 0x9c, 0x8b, 0x66, 0x62, 0xa1, 0xdb, 0x9a, 0xaa, 0xac, 0x4b, 0x64, 0x3b, 0xb7, 0x35, 0xdf,
	0x59, 0xc8, 0xe6, 0xc3, 0x05, 0xd7, 0x4a, 0x55, 0x1c, 0xa4, 0x98, 0x56, 0x90, 0xb3, 0xdd, 0x38,
	0x1a, 0x6c, 0xa5, 0x5d, 0x5c, 0x4c, 0x94, 0xaa, 0xce, 0x3c, 0x48, 0x19, 0xd9, 0x94, 0x80, 0x76,
	0x94, 0xec, 0x91, 0x1b, 0xd7, 0xea, 0x93, 0x7e, 0x4e, 0x76, 0xa6, 0x4a, 0x61, 0x83, 0x46, 0x68,
	0xae, 0xc1, 0xde, 0x10, 0x1a, 0xaf, 0x0d, 0x5a, 0xe9, 0xf6, 0x35, 0x3c, 0xb1, 0x28, 0x7d, 0x46,
	0x48, 0x88, 0xe1, 0x65, 0xce, 0x1e, 0xbb, 0x53, 0xb5, 0x02, 0x32, 0xce, 0xe9, 0xd7, 0xa4, 0x5b,
	0x8b, 0x52, 0xf2, 0x32, 0x07, 0x89, 0x25, 0x2e, 0x59, 0xcf, 0x75, 0xaf, 0x97, 0x84, 0x5d, 0x91,
	0x1c, 0xfb, 0x5f, 0xfb, 0x22, 0xd3, 0x8e, 0x95, 0x8e, 0x83, 0x92, 0x16, 0x84, 0x35, 0xb5, 0x30,
	0xe8, 0x66, 0x6a, 0x44, 0x86, 0xab, 0xfb, 0x0c, 0x0d, 0xdb, 0x8b, 0xd7, 0x06, 0xed, 0x51, 0x92,
	0xf8, 0xcd, 0x93, 0xdc, 0xbc, 0xea, 0xe4, 0xb5, 0x0d, 0x39, 0x0d, 0x11, 0xc7, 0xab, 0x80, 0x33,
	0x89, 0x66, 0x99, 0xee, 0x37, 0x1f, 0x24, 0x29, 0x90, 0x27, 0xef, 0x65, 0x9a, 0x2e, 0x11, 0x32,
	0x95, 0x03, 0x7b, 0xe2, 0x12, 0x1d, 0xdd, 0x97, 0xe8, 0x24, 0xe8, 0x7d, 0x9e, 0xbd, 0xe6, 0x43,
	0x1c, 0x7d, 0x4e, 0xba, 0x5a, 0x1b, 0x75, 0x71, 0x3d, 0x93, 0x7d, 0x37, 0x93, 0x8e, 0x03, 0xc3,
	0x48, 0xfa, 0x63, 0xf2, 0xf4, 0x3f, 0x8e, 0x40, 0x77, 0xc9, 0xda, 0x5b, 0x58, 0x86, 0x5d, 0x65,
	0xff, 0xd2, 0x1e, 0x59, 0xbf, 0x14, 0xd5, 0x1c, 0xdc, 0x82, 0x6a, 0xa5, 0xfe, 0xe3, 0x9b, 0x07,
	0x5f, 0x45, 0xfd, 0x6f, 0x49, 0xff, 0xdf, 0x8b, 0xfc, 0x3f, 0x4e, 0xa3, 0x9a, 0x74, 0xfd, 0xc9,
	0x5f, 0xfb, 0xcd, 0x4e, 0x7f, 0x25, 0xad, 0x73, 0x40, 0x8f, 0xd1, 0xfd, 0x3b, 0x4f, 0xe1, 0xcc,
	0xee, 0xe8, 0x3e, 0xbd, 0xdb, 0xb5, 0xc3, 0xe7, 0xbf, 0x1f, 0x3f, 0xea, 0xef, 0x9c, 0x03, 0xc6,
	0x76, 0x2b, 0xc4, 0x9e, 0xf9, 0xed, 0xaf, 0xbf, 0xff, 0x78, 0xd0, 0xa2, 0x9b, 0x43, 0xaf, 0x3f,
	0x79, 0x41, 0x48, 0xa6, 0xea, 0x10, 0x7d, 0xd2, 0x09, 0x49, 0x27, 0xd6, 0x7d, 0x12, 0xfd, 0xbc,
	0xe5, 0x71, 0x3d, 0x9d, 0x6e, 0xb8, 0x84, 0x5f, 0xfe, 0x13, 0x00, 0x00, 0xff, 0xff, 0xc3, 0x6d,
	0x0b, 0x1a, 0x97, 0x06, 0x00, 0x00,
}

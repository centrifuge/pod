package configstore

import (
	"encoding/json"
	"math/big"
	"reflect"
	"time"

	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/accounts"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/config"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/golang/protobuf/ptypes/duration"
)

// KeyPair represents a key pair config
type KeyPair struct {
	Pub, Priv string
}

// NewKeyPair creates a KeyPair
func NewKeyPair(pub, priv string) KeyPair {
	return KeyPair{Pub: pub, Priv: priv}
}

// NodeConfig exposes configs specific to the node
type NodeConfig struct {
	MainIdentity                   TenantConfig
	StoragePath                    string
	TenantsKeystore                string
	P2PPort                        int
	P2PExternalIP                  string
	P2PConnectionTimeout           time.Duration
	ServerPort                     int
	ServerAddress                  string
	NumWorkers                     int
	WorkerWaitTimeMS               int
	EthereumNodeURL                string
	EthereumContextReadWaitTimeout time.Duration
	EthereumContextWaitTimeout     time.Duration
	EthereumIntervalRetry          time.Duration
	EthereumMaxRetries             int
	EthereumGasPrice               *big.Int
	EthereumGasLimit               uint64
	TxPoolAccessEnabled            bool
	NetworkString                  string
	BootstrapPeers                 []string
	NetworkID                      uint32
	SmartContractAddresses         map[config.ContractName]common.Address
	PprofEnabled                   bool
}

// IsSet refer the interface
func (nc *NodeConfig) IsSet(key string) bool {
	panic("irrelevant, NodeConfig#IsSet must not be used")
}

// Set refer the interface
func (nc *NodeConfig) Set(key string, value interface{}) {
	panic("irrelevant, NodeConfig#Set must not be used")
}

// SetDefault refer the interface
func (nc *NodeConfig) SetDefault(key string, value interface{}) {
	panic("irrelevant, NodeConfig#SetDefault must not be used")
}

// SetupSmartContractAddresses refer the interface
func (nc *NodeConfig) SetupSmartContractAddresses(network string, smartContractAddresses *config.SmartContractAddresses) {
	panic("irrelevant, NodeConfig#SetupSmartContractAddresses must not be used")
}

// Get refer the interface
func (nc *NodeConfig) Get(key string) interface{} {
	panic("irrelevant, NodeConfig#Get must not be used")
}

// GetString refer the interface
func (nc *NodeConfig) GetString(key string) string {
	panic("irrelevant, NodeConfig#GetString must not be used")
}

// GetBool refer the interface
func (nc *NodeConfig) GetBool(key string) bool {
	panic("irrelevant, NodeConfig#GetBool must not be used")
}

// GetInt refer the interface
func (nc *NodeConfig) GetInt(key string) int {
	panic("irrelevant, NodeConfig#GetInt must not be used")
}

// GetDuration refer the interface
func (nc *NodeConfig) GetDuration(key string) time.Duration {
	panic("irrelevant, NodeConfig#GetDuration must not be used")
}

// GetStoragePath refer the interface
func (nc *NodeConfig) GetStoragePath() string {
	return nc.StoragePath
}

// GetConfigStoragePath refer the interface
func (nc *NodeConfig) GetConfigStoragePath() string {
	panic("irrelevant, NodeConfig#GetConfigStoragePath must not be used")
}

// GetTenantsKeystore returns the tenant keystore path.
func (nc *NodeConfig) GetTenantsKeystore() string {
	return nc.TenantsKeystore
}

// GetP2PPort refer the interface
func (nc *NodeConfig) GetP2PPort() int {
	return nc.P2PPort
}

// GetP2PExternalIP refer the interface
func (nc *NodeConfig) GetP2PExternalIP() string {
	return nc.P2PExternalIP
}

// GetP2PConnectionTimeout refer the interface
func (nc *NodeConfig) GetP2PConnectionTimeout() time.Duration {
	return nc.P2PConnectionTimeout
}

// GetServerPort refer the interface
func (nc *NodeConfig) GetServerPort() int {
	return nc.ServerPort
}

// GetServerAddress refer the interface
func (nc *NodeConfig) GetServerAddress() string {
	return nc.ServerAddress
}

// GetNumWorkers refer the interface
func (nc *NodeConfig) GetNumWorkers() int {
	return nc.NumWorkers
}

// GetWorkerWaitTimeMS refer the interface
func (nc *NodeConfig) GetWorkerWaitTimeMS() int {
	return nc.WorkerWaitTimeMS
}

// GetEthereumNodeURL refer the interface
func (nc *NodeConfig) GetEthereumNodeURL() string {
	return nc.EthereumNodeURL
}

// GetEthereumContextReadWaitTimeout refer the interface
func (nc *NodeConfig) GetEthereumContextReadWaitTimeout() time.Duration {
	return nc.EthereumContextReadWaitTimeout
}

// GetEthereumContextWaitTimeout refer the interface
func (nc *NodeConfig) GetEthereumContextWaitTimeout() time.Duration {
	return nc.EthereumContextWaitTimeout
}

// GetEthereumIntervalRetry refer the interface
func (nc *NodeConfig) GetEthereumIntervalRetry() time.Duration {
	return nc.EthereumIntervalRetry
}

// GetEthereumMaxRetries refer the interface
func (nc *NodeConfig) GetEthereumMaxRetries() int {
	return nc.EthereumMaxRetries
}

// GetEthereumGasPrice refer the interface
func (nc *NodeConfig) GetEthereumGasPrice() *big.Int {
	return nc.EthereumGasPrice
}

// GetEthereumGasLimit refer the interface
func (nc *NodeConfig) GetEthereumGasLimit() uint64 {
	return nc.EthereumGasLimit
}

// GetTxPoolAccessEnabled refer the interface
func (nc *NodeConfig) GetTxPoolAccessEnabled() bool {
	return nc.TxPoolAccessEnabled
}

// GetNetworkString refer the interface
func (nc *NodeConfig) GetNetworkString() string {
	return nc.NetworkString
}

// GetNetworkKey refer the interface
func (nc *NodeConfig) GetNetworkKey(k string) string {
	panic("irrelevant, NodeConfig#GetNetworkKey must not be used")
}

// GetContractAddressString refer the interface
func (nc *NodeConfig) GetContractAddressString(address string) string {
	panic("irrelevant, NodeConfig#GetContractAddressString must not be used")
}

// GetContractAddress refer the interface
func (nc *NodeConfig) GetContractAddress(contractName config.ContractName) common.Address {
	return nc.SmartContractAddresses[contractName]
}

// GetBootstrapPeers refer the interface
func (nc *NodeConfig) GetBootstrapPeers() []string {
	return nc.BootstrapPeers
}

// GetNetworkID refer the interface
func (nc *NodeConfig) GetNetworkID() uint32 {
	return nc.NetworkID
}

// GetEthereumAccount refer the interface
func (nc *NodeConfig) GetEthereumAccount(accountName string) (account *config.AccountConfig, err error) {
	return nc.MainIdentity.EthereumAccount, nil
}

// GetEthereumDefaultAccountName refer the interface
func (nc *NodeConfig) GetEthereumDefaultAccountName() string {
	return nc.MainIdentity.EthereumDefaultAccountName
}

// GetReceiveEventNotificationEndpoint refer the interface
func (nc *NodeConfig) GetReceiveEventNotificationEndpoint() string {
	return nc.MainIdentity.ReceiveEventNotificationEndpoint
}

// GetIdentityID refer the interface
func (nc *NodeConfig) GetIdentityID() ([]byte, error) {
	return nc.MainIdentity.IdentityID, nil
}

// GetSigningKeyPair refer the interface
func (nc *NodeConfig) GetSigningKeyPair() (pub, priv string) {
	return nc.MainIdentity.SigningKeyPair.Pub, nc.MainIdentity.SigningKeyPair.Priv
}

// GetEthAuthKeyPair refer the interface
func (nc *NodeConfig) GetEthAuthKeyPair() (pub, priv string) {
	return nc.MainIdentity.EthAuthKeyPair.Pub, nc.MainIdentity.EthAuthKeyPair.Priv
}

// IsPProfEnabled refer the interface
func (nc *NodeConfig) IsPProfEnabled() bool {
	return nc.PprofEnabled
}

// ID Gets the ID of the document represented by this model
func (nc *NodeConfig) ID() ([]byte, error) {
	return []byte{}, nil
}

// Type Returns the underlying type of the Model
func (nc *NodeConfig) Type() reflect.Type {
	return reflect.TypeOf(nc)
}

// JSON return the json representation of the model
func (nc *NodeConfig) JSON() ([]byte, error) {
	return json.Marshal(nc)
}

// FromJSON initialize the model with a json
func (nc *NodeConfig) FromJSON(data []byte) error {
	return json.Unmarshal(data, nc)
}

// CreateProtobuf creates protobuf for config
func (nc *NodeConfig) CreateProtobuf() *configpb.ConfigData {
	return &configpb.ConfigData{
		MainIdentity: &accountpb.AccountData{
			EthAccount: &accountpb.EthereumAccount{
				Address:  nc.MainIdentity.EthereumAccount.Address,
				Key:      nc.MainIdentity.EthereumAccount.Key,
				Password: nc.MainIdentity.EthereumAccount.Password,
			},
			EthDefaultAccountName:            nc.MainIdentity.EthereumDefaultAccountName,
			IdentityId:                       hexutil.Encode(nc.MainIdentity.IdentityID),
			ReceiveEventNotificationEndpoint: nc.MainIdentity.ReceiveEventNotificationEndpoint,
			EthauthKeyPair: &accountpb.KeyPair{
				Pub: nc.MainIdentity.EthAuthKeyPair.Pub,
				Pvt: nc.MainIdentity.EthAuthKeyPair.Priv,
			},
			SigningKeyPair: &accountpb.KeyPair{
				Pub: nc.MainIdentity.SigningKeyPair.Pub,
				Pvt: nc.MainIdentity.SigningKeyPair.Priv,
			},
		},
		StoragePath:               nc.StoragePath,
		P2PPort:                   int32(nc.P2PPort),
		P2PExternalIp:             nc.P2PExternalIP,
		P2PConnectionTimeout:      &duration.Duration{Seconds: int64(nc.P2PConnectionTimeout.Seconds())},
		ServerPort:                int32(nc.ServerPort),
		ServerAddress:             nc.ServerAddress,
		NumWorkers:                int32(nc.NumWorkers),
		WorkerWaitTimeMs:          int32(nc.WorkerWaitTimeMS),
		EthContextReadWaitTimeout: &duration.Duration{Seconds: int64(nc.EthereumContextReadWaitTimeout.Seconds())},
		EthContextWaitTimeout:     &duration.Duration{Seconds: int64(nc.EthereumContextWaitTimeout.Seconds())},
		EthIntervalRetry:          &duration.Duration{Seconds: int64(nc.EthereumIntervalRetry.Seconds())},
		EthGasPrice:               nc.EthereumGasPrice.Uint64(),
		EthGasLimit:               nc.EthereumGasLimit,
		TxPoolEnabled:             nc.TxPoolAccessEnabled,
		Network:                   nc.NetworkString,
		NetworkId:                 nc.NetworkID,
		PprofEnabled:              nc.PprofEnabled,
		SmartContractAddresses:    convertAddressesToStringMap(nc.SmartContractAddresses),
	}
}

func convertAddressesToStringMap(addresses map[config.ContractName]common.Address) map[string]string {
	m := make(map[string]string)
	for k, v := range addresses {
		m[string(k)] = v.String()
	}
	return m
}

func (nc *NodeConfig) loadFromProtobuf(data *configpb.ConfigData) error {
	identityID, _ := hexutil.Decode(data.MainIdentity.IdentityId)

	nc.MainIdentity = TenantConfig{
		EthereumAccount: &config.AccountConfig{
			Address:  data.MainIdentity.EthAccount.Address,
			Key:      data.MainIdentity.EthAccount.Key,
			Password: data.MainIdentity.EthAccount.Password,
		},
		EthereumDefaultAccountName:       data.MainIdentity.EthDefaultAccountName,
		IdentityID:                       identityID,
		ReceiveEventNotificationEndpoint: data.MainIdentity.ReceiveEventNotificationEndpoint,
		SigningKeyPair: KeyPair{
			Pub:  data.MainIdentity.SigningKeyPair.Pub,
			Priv: data.MainIdentity.SigningKeyPair.Pvt,
		},
		EthAuthKeyPair: KeyPair{
			Pub:  data.MainIdentity.EthauthKeyPair.Pub,
			Priv: data.MainIdentity.EthauthKeyPair.Pvt,
		},
	}
	nc.StoragePath = data.StoragePath
	nc.P2PPort = int(data.P2PPort)
	nc.P2PExternalIP = data.P2PExternalIp
	nc.P2PConnectionTimeout = time.Duration(data.P2PConnectionTimeout.Seconds)
	nc.ServerPort = int(data.ServerPort)
	nc.ServerAddress = data.ServerAddress
	nc.NumWorkers = int(data.NumWorkers)
	nc.WorkerWaitTimeMS = int(data.WorkerWaitTimeMs)
	nc.EthereumNodeURL = data.EthNodeUrl
	nc.EthereumContextReadWaitTimeout = time.Duration(data.EthContextReadWaitTimeout.Seconds)
	nc.EthereumContextWaitTimeout = time.Duration(data.EthContextWaitTimeout.Seconds)
	nc.EthereumIntervalRetry = time.Duration(data.EthIntervalRetry.Seconds)
	nc.EthereumMaxRetries = int(data.EthMaxRetries)
	nc.EthereumGasPrice = big.NewInt(int64(data.EthGasPrice))
	nc.EthereumGasLimit = data.EthGasLimit
	nc.TxPoolAccessEnabled = data.TxPoolEnabled
	nc.NetworkString = data.Network
	nc.BootstrapPeers = data.BootstrapPeers
	nc.NetworkID = data.NetworkId
	var err error
	nc.SmartContractAddresses, err = convertStringMapToSmartContractAddresses(data.SmartContractAddresses)
	if err != nil {
		return err
	}
	nc.PprofEnabled = data.PprofEnabled
	return nil
}

func convertStringMapToSmartContractAddresses(addrs map[string]string) (map[config.ContractName]common.Address, error) {
	m := make(map[config.ContractName]common.Address)
	for k, v := range addrs {
		if common.IsHexAddress(v) {
			m[config.ContractName(k)] = common.HexToAddress(v)
		} else {
			return nil, errors.New("provided smart contract address %s is invalid", v)
		}
	}
	return m, nil
}

// NewNodeConfig creates a new NodeConfig instance with configs
func NewNodeConfig(c config.Configuration) config.Configuration {
	mainAccount, _ := c.GetEthereumAccount(c.GetEthereumDefaultAccountName())
	mainIdentity, _ := c.GetIdentityID()
	signPub, signPriv := c.GetSigningKeyPair()
	ethAuthPub, ethAuthPriv := c.GetEthAuthKeyPair()

	return &NodeConfig{
		MainIdentity: TenantConfig{
			EthereumAccount: &config.AccountConfig{
				Address:  mainAccount.Address,
				Key:      mainAccount.Key,
				Password: mainAccount.Password,
			},
			EthereumDefaultAccountName:       c.GetEthereumDefaultAccountName(),
			IdentityID:                       mainIdentity,
			ReceiveEventNotificationEndpoint: c.GetReceiveEventNotificationEndpoint(),
			SigningKeyPair: KeyPair{
				Pub:  signPub,
				Priv: signPriv,
			},
			EthAuthKeyPair: KeyPair{
				Pub:  ethAuthPub,
				Priv: ethAuthPriv,
			},
		},
		StoragePath:                    c.GetStoragePath(),
		TenantsKeystore:                c.GetTenantsKeystore(),
		P2PPort:                        c.GetP2PPort(),
		P2PExternalIP:                  c.GetP2PExternalIP(),
		P2PConnectionTimeout:           c.GetP2PConnectionTimeout(),
		ServerPort:                     c.GetServerPort(),
		ServerAddress:                  c.GetServerAddress(),
		NumWorkers:                     c.GetNumWorkers(),
		WorkerWaitTimeMS:               c.GetWorkerWaitTimeMS(),
		EthereumNodeURL:                c.GetEthereumNodeURL(),
		EthereumContextReadWaitTimeout: c.GetEthereumContextReadWaitTimeout(),
		EthereumContextWaitTimeout:     c.GetEthereumContextWaitTimeout(),
		EthereumIntervalRetry:          c.GetEthereumIntervalRetry(),
		EthereumMaxRetries:             c.GetEthereumMaxRetries(),
		EthereumGasPrice:               c.GetEthereumGasPrice(),
		EthereumGasLimit:               c.GetEthereumGasLimit(),
		TxPoolAccessEnabled:            c.GetTxPoolAccessEnabled(),
		NetworkString:                  c.GetNetworkString(),
		BootstrapPeers:                 c.GetBootstrapPeers(),
		NetworkID:                      c.GetNetworkID(),
		SmartContractAddresses:         extractSmartContractAddresses(c),
		PprofEnabled:                   c.IsPProfEnabled(),
	}
}

func extractSmartContractAddresses(c config.Configuration) map[config.ContractName]common.Address {
	sms := make(map[config.ContractName]common.Address)
	names := config.ContractNames()
	for _, n := range names {
		sms[n] = c.GetContractAddress(n)
	}
	return sms
}

// TenantConfig exposes configs specific to a tenant in the node
type TenantConfig struct {
	EthereumAccount                  *config.AccountConfig
	EthereumDefaultAccountName       string
	EthereumContextWaitTimeout       time.Duration
	ReceiveEventNotificationEndpoint string
	IdentityID                       []byte
	SigningKeyPair                   KeyPair
	EthAuthKeyPair                   KeyPair
	P2PKeyPair                       KeyPair
}

// GetEthereumAccount gets EthereumAccount
func (tc *TenantConfig) GetEthereumAccount() *config.AccountConfig {
	return tc.EthereumAccount
}

// GetEthereumDefaultAccountName gets EthereumDefaultAccountName
func (tc *TenantConfig) GetEthereumDefaultAccountName() string {
	return tc.EthereumDefaultAccountName
}

// GetReceiveEventNotificationEndpoint gets ReceiveEventNotificationEndpoint
func (tc *TenantConfig) GetReceiveEventNotificationEndpoint() string {
	return tc.ReceiveEventNotificationEndpoint
}

// GetIdentityID gets IdentityID
func (tc *TenantConfig) GetIdentityID() ([]byte, error) {
	return tc.IdentityID, nil
}

// GetSigningKeyPair gets SigningKeyPair
func (tc *TenantConfig) GetSigningKeyPair() (pub, priv string) {
	return tc.SigningKeyPair.Pub, tc.SigningKeyPair.Priv
}

// GetEthAuthKeyPair gets EthAuthKeyPair
func (tc *TenantConfig) GetEthAuthKeyPair() (pub, priv string) {
	return tc.EthAuthKeyPair.Pub, tc.EthAuthKeyPair.Priv
}

// GetEthereumContextWaitTimeout gets EthereumContextWaitTimeout
func (tc *TenantConfig) GetEthereumContextWaitTimeout() time.Duration {
	return tc.EthereumContextWaitTimeout
}

// ID Get the ID of the document represented by this model
func (tc *TenantConfig) ID() []byte {
	return tc.IdentityID
}

// Type Returns the underlying type of the Model
func (tc *TenantConfig) Type() reflect.Type {
	return reflect.TypeOf(tc)
}

// JSON return the json representation of the model
func (tc *TenantConfig) JSON() ([]byte, error) {
	return json.Marshal(tc)
}

// FromJSON initialize the model with a json
func (tc *TenantConfig) FromJSON(data []byte) error {
	return json.Unmarshal(data, tc)
}

// CreateProtobuf creates protobuf for config
func (tc *TenantConfig) CreateProtobuf() *accountpb.AccountData {
	return &accountpb.AccountData{
		EthAccount: &accountpb.EthereumAccount{
			Address:  tc.EthereumAccount.Address,
			Key:      tc.EthereumAccount.Key,
			Password: tc.EthereumAccount.Password,
		},
		EthDefaultAccountName:            tc.EthereumDefaultAccountName,
		ReceiveEventNotificationEndpoint: tc.ReceiveEventNotificationEndpoint,
		IdentityId:                       hexutil.Encode(tc.IdentityID),
		SigningKeyPair: &accountpb.KeyPair{
			Pub: tc.SigningKeyPair.Pub,
			Pvt: tc.SigningKeyPair.Priv,
		},
		EthauthKeyPair: &accountpb.KeyPair{
			Pub: tc.EthAuthKeyPair.Pub,
			Pvt: tc.EthAuthKeyPair.Priv,
		},
	}
}

func (tc *TenantConfig) loadFromProtobuf(data *accountpb.AccountData) {
	tc.EthereumAccount = &config.AccountConfig{
		Address:  data.EthAccount.Address,
		Key:      data.EthAccount.Key,
		Password: data.EthAccount.Password,
	}
	tc.EthereumDefaultAccountName = data.EthDefaultAccountName
	tc.IdentityID, _ = hexutil.Decode(data.IdentityId)
	tc.ReceiveEventNotificationEndpoint = data.ReceiveEventNotificationEndpoint
	tc.SigningKeyPair = KeyPair{
		Pub:  data.SigningKeyPair.Pub,
		Priv: data.SigningKeyPair.Pvt,
	}
	tc.EthAuthKeyPair = KeyPair{
		Pub:  data.EthauthKeyPair.Pub,
		Priv: data.EthauthKeyPair.Pvt,
	}
}

// NewTenantConfig creates a new TenantConfig instance with configs
func NewTenantConfig(ethAccountName string, c config.Configuration) (config.TenantConfiguration, error) {
	id, err := c.GetIdentityID()
	if err != nil {
		return nil, err
	}
	acc, err := c.GetEthereumAccount(ethAccountName)
	if err != nil && ethAccountName != "" {
		return nil, err
	}
	return &TenantConfig{
		EthereumAccount:                  acc,
		EthereumDefaultAccountName:       c.GetEthereumDefaultAccountName(),
		EthereumContextWaitTimeout:       c.GetEthereumContextWaitTimeout(),
		IdentityID:                       id,
		ReceiveEventNotificationEndpoint: c.GetReceiveEventNotificationEndpoint(),
		SigningKeyPair:                   NewKeyPair(c.GetSigningKeyPair()),
		EthAuthKeyPair:                   NewKeyPair(c.GetEthAuthKeyPair()),
	}, nil
}

// TempTenantConfig creates a new TenantConfig without id validation, Must only be used for tenant creation.
func TempTenantConfig(ethAccountName string, c config.Configuration) (config.TenantConfiguration, error) {
	acc, err := c.GetEthereumAccount(ethAccountName)
	if err != nil && ethAccountName != "" {
		return nil, err
	}
	return &TenantConfig{
		EthereumAccount:                  acc,
		EthereumDefaultAccountName:       c.GetEthereumDefaultAccountName(),
		IdentityID:                       []byte{},
		ReceiveEventNotificationEndpoint: c.GetReceiveEventNotificationEndpoint(),
		SigningKeyPair:                   NewKeyPair(c.GetSigningKeyPair()),
		EthAuthKeyPair:                   NewKeyPair(c.GetEthAuthKeyPair()),
	}, nil
}

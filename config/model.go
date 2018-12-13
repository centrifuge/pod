package config

import (
	"encoding/json"
	"math/big"
	"reflect"
	"time"
)

// Model is an interface for both tenant and node config models
type Model interface {
	// Get the ID of the document represented by this model
	ID() ([]byte, error)

	//Returns the underlying type of the Model
	Type() reflect.Type

	// JSON return the json representation of the model
	JSON() ([]byte, error)

	// FromJSON initialize the model with a json
	FromJSON(json []byte) error
}

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
	StoragePath                    string
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

	// TODO what to do about contract addresses?
}

// Get the ID of the document represented by this model
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

// NewNodeConfig creates a new NodeConfig instance with configs
func NewNodeConfig(config Configuration) *NodeConfig {
	return &NodeConfig{
		StoragePath:                    config.GetStoragePath(),
		P2PPort:                        config.GetP2PPort(),
		P2PExternalIP:                  config.GetP2PExternalIP(),
		P2PConnectionTimeout:           config.GetP2PConnectionTimeout(),
		ServerPort:                     config.GetServerPort(),
		ServerAddress:                  config.GetServerAddress(),
		NumWorkers:                     config.GetNumWorkers(),
		WorkerWaitTimeMS:               config.GetWorkerWaitTimeMS(),
		EthereumNodeURL:                config.GetEthereumNodeURL(),
		EthereumContextReadWaitTimeout: config.GetEthereumContextReadWaitTimeout(),
		EthereumContextWaitTimeout:     config.GetEthereumContextWaitTimeout(),
		EthereumIntervalRetry:          config.GetEthereumIntervalRetry(),
		EthereumMaxRetries:             config.GetEthereumMaxRetries(),
		EthereumGasPrice:               config.GetEthereumGasPrice(),
		EthereumGasLimit:               config.GetEthereumGasLimit(),
		TxPoolAccessEnabled:            config.GetTxPoolAccessEnabled(),
		NetworkString:                  config.GetNetworkString(),
		BootstrapPeers:                 config.GetBootstrapPeers(),
		NetworkID:                      config.GetNetworkID(),
	}
}

// TenantConfig exposes configs specific to a tenant in the node
type TenantConfig struct {
	EthereumAccount                  *AccountConfig
	EthereumDefaultAccountName       string
	ReceiveEventNotificationEndpoint string
	IdentityID                       []byte
	SigningKeyPair                   KeyPair
	EthAuthKeyPair                   KeyPair
}

// Get the ID of the document represented by this model
func (tc *TenantConfig) ID() ([]byte, error) {
	return tc.IdentityID, nil
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

// NewTenantConfig creates a new TenantConfig instance with configs
func NewTenantConfig(ethAccountName string, config Configuration) (*TenantConfig, error) {
	id, err := config.GetIdentityID()
	if err != nil {
		return nil, err
	}
	acc, err := config.GetEthereumAccount(ethAccountName)
	if err != nil {
		return nil, err
	}
	return &TenantConfig{
		EthereumAccount:                  acc,
		EthereumDefaultAccountName:       config.GetEthereumDefaultAccountName(),
		IdentityID:                       id,
		ReceiveEventNotificationEndpoint: config.GetReceiveEventNotificationEndpoint(),
		SigningKeyPair:                   NewKeyPair(config.GetSigningKeyPair()),
		EthAuthKeyPair:                   NewKeyPair(config.GetEthAuthKeyPair()),
	}, nil
}

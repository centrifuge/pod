package config

import (
	"math/big"
	"time"
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
	StoragePath                      string
	P2PPort                          int
	P2PExternalIP                    string
	P2PConnectionTimeout             time.Duration
	ReceiveEventNotificationEndpoint string
	ServerPort                       int
	ServerAddress                    string
	NumWorkers                       int
	WorkerWaitTimeMS                 int
	EthereumNodeURL                  string
	EthereumContextReadWaitTimeout   time.Duration
	EthereumContextWaitTimeout       time.Duration
	EthereumIntervalRetry            time.Duration
	EthereumMaxRetries               int
	EthereumGasPrice                 *big.Int
	EthereumGasLimit                 uint64
	EthereumDefaultAccountName       string
	TxPoolAccessEnabled              bool
	NetworkString                    string
	BootstrapPeers                   []string
	NetworkID                        uint32

	// TODO what to do about contract addresses?
}

// NewNodeConfig creates a new NodeConfig instance with configs
func NewNodeConfig(config Config) *NodeConfig {
	return &NodeConfig{
		StoragePath:                      config.GetStoragePath(),
		P2PPort:                          config.GetP2PPort(),
		P2PExternalIP:                    config.GetP2PExternalIP(),
		P2PConnectionTimeout:             config.GetP2PConnectionTimeout(),
		ReceiveEventNotificationEndpoint: config.GetReceiveEventNotificationEndpoint(),
		ServerPort:                       config.GetServerPort(),
		ServerAddress:                    config.GetServerAddress(),
		NumWorkers:                       config.GetNumWorkers(),
		WorkerWaitTimeMS:                 config.GetWorkerWaitTimeMS(),
		EthereumNodeURL:                  config.GetEthereumNodeURL(),
		EthereumContextReadWaitTimeout:   config.GetEthereumContextReadWaitTimeout(),
		EthereumContextWaitTimeout:       config.GetEthereumContextWaitTimeout(),
		EthereumIntervalRetry:            config.GetEthereumIntervalRetry(),
		EthereumMaxRetries:               config.GetEthereumMaxRetries(),
		EthereumGasPrice:                 config.GetEthereumGasPrice(),
		EthereumGasLimit:                 config.GetEthereumGasLimit(),
		EthereumDefaultAccountName:       config.GetEthereumDefaultAccountName(),
		TxPoolAccessEnabled:              config.GetTxPoolAccessEnabled(),
		NetworkString:                    config.GetNetworkString(),
		BootstrapPeers:                   config.GetBootstrapPeers(),
		NetworkID:                        config.GetNetworkID(),
	}
}

// TenantConfig exposes configs specific to a tenant in the node
type TenantConfig struct {
	EthereumAccount *AccountConfig
	IdentityID      []byte
	SigningKeyPair  KeyPair
	EthAuthKeyPair  KeyPair
}

// NewTenantConfig creates a new TenantConfig instance with configs
func NewTenantConfig(name string, config Config) (*TenantConfig, error) {
	id, err := config.GetIdentityID()
	if err != nil {
		return nil, err
	}
	acc, err := config.GetEthereumAccount(name)
	if err != nil {
		return nil, err
	}
	return &TenantConfig{
		EthereumAccount: acc,
		IdentityID:      id,
		SigningKeyPair:  NewKeyPair(config.GetSigningKeyPair()),
		EthAuthKeyPair:  NewKeyPair(config.GetEthAuthKeyPair()),
	}, nil
}

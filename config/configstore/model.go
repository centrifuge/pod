package configstore

import (
	"encoding/json"
	"math/big"
	"reflect"
	"time"

	"github.com/centrifuge/go-centrifuge/config"

	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/config"
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

func (nc *NodeConfig) createProtobuf() *configpb.ConfigData {
	return &configpb.ConfigData{
		MainIdentity: &configpb.TenantData{
			EthAccount: &configpb.EthereumAccount{
				Address:  nc.MainIdentity.EthereumAccount.Address,
				Key:      nc.MainIdentity.EthereumAccount.Key,
				Password: nc.MainIdentity.EthereumAccount.Password,
			},
			EthDefaultAccountName:            nc.MainIdentity.EthereumDefaultAccountName,
			IdentityId:                       hexutil.Encode(nc.MainIdentity.IdentityID),
			ReceiveEventNotificationEndpoint: nc.MainIdentity.ReceiveEventNotificationEndpoint,
			EthauthKeyPair: &configpb.KeyPair{
				Pub: nc.MainIdentity.EthAuthKeyPair.Pub,
				Pvt: nc.MainIdentity.EthAuthKeyPair.Priv,
			},
			SigningKeyPair: &configpb.KeyPair{
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
	}
}

func (nc *NodeConfig) loadFromProtobuf(data *configpb.ConfigData) {
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
}

// NewNodeConfig creates a new NodeConfig instance with configs
func NewNodeConfig(c config.Configuration) *NodeConfig {
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
	}
}

// TenantConfig exposes configs specific to a tenant in the node
type TenantConfig struct {
	EthereumAccount                  *config.AccountConfig
	EthereumDefaultAccountName       string
	ReceiveEventNotificationEndpoint string
	IdentityID                       []byte
	SigningKeyPair                   KeyPair
	EthAuthKeyPair                   KeyPair
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

func (tc *TenantConfig) createProtobuf() *configpb.TenantData {
	return &configpb.TenantData{
		EthAccount: &configpb.EthereumAccount{
			Address:  tc.EthereumAccount.Address,
			Key:      tc.EthereumAccount.Key,
			Password: tc.EthereumAccount.Password,
		},
		EthDefaultAccountName:            tc.EthereumDefaultAccountName,
		ReceiveEventNotificationEndpoint: tc.ReceiveEventNotificationEndpoint,
		IdentityId:                       hexutil.Encode(tc.IdentityID),
		SigningKeyPair: &configpb.KeyPair{
			Pub: tc.SigningKeyPair.Pub,
			Pvt: tc.SigningKeyPair.Priv,
		},
		EthauthKeyPair: &configpb.KeyPair{
			Pub: tc.EthAuthKeyPair.Pub,
			Pvt: tc.EthAuthKeyPair.Priv,
		},
	}
}

func (tc *TenantConfig) loadFromProtobuf(data *configpb.TenantData) {
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
func NewTenantConfig(ethAccountName string, c config.Configuration) (*TenantConfig, error) {
	id, err := c.GetIdentityID()
	if err != nil {
		return nil, err
	}
	acc, err := c.GetEthereumAccount(ethAccountName)
	if err != nil {
		return nil, err
	}
	return &TenantConfig{
		EthereumAccount:                  acc,
		EthereumDefaultAccountName:       c.GetEthereumDefaultAccountName(),
		IdentityID:                       id,
		ReceiveEventNotificationEndpoint: c.GetReceiveEventNotificationEndpoint(),
		SigningKeyPair:                   NewKeyPair(c.GetSigningKeyPair()),
		EthAuthKeyPair:                   NewKeyPair(c.GetEthAuthKeyPair()),
	}, nil
}

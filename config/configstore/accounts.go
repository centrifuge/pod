package configstore

import (
	"encoding/json"
	"math/big"
	"reflect"
	"time"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/crypto"
	"github.com/centrifuge/go-centrifuge/crypto/ed25519"
	"github.com/centrifuge/go-centrifuge/crypto/secp256k1"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common"
)

// KeyPair represents a key pair config
type KeyPair struct {
	Pub string `json:"Pub"`
	Pvt string `json:"Priv"`
}

// NewKeyPair creates a KeyPair
func NewKeyPair(pub, priv string) KeyPair {
	return KeyPair{Pub: pub, Pvt: priv}
}

// NodeConfig exposes configs specific to the node
type NodeConfig struct {
	MainIdentity                   Account
	StoragePath                    string
	AccountsKeystore               string
	P2PPort                        int
	P2PExternalIP                  string
	P2PConnectionTimeout           time.Duration
	P2PResponseDelay               time.Duration
	ServerPort                     int
	ServerAddress                  string
	NumWorkers                     int
	WorkerWaitTimeMS               int
	TaskValidDuration              time.Duration
	EthereumNodeURL                string
	EthereumContextReadWaitTimeout time.Duration
	EthereumContextWaitTimeout     time.Duration
	EthereumIntervalRetry          time.Duration
	EthereumMaxRetries             int
	EthereumMaxGasPrice            *big.Int
	EthereumGasLimits              map[config.ContractOp]uint64
	EthereumGasMultiplier          float64
	NetworkString                  string
	BootstrapPeers                 []string
	NetworkID                      uint32
	SmartContractAddresses         map[config.ContractName]common.Address
	SmartContractBytecode          map[config.ContractName]string
	PprofEnabled                   bool
	LowEntropyNFTTokenEnabled      bool
	DebugLogEnabled                bool
	CentChainNodeURL               string
	CentChainIntervalRetry         time.Duration
	CentChainMaxRetries            int
	CentChainAnchorLifespan        time.Duration
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

// GetFloat refer the interface
func (nc *NodeConfig) GetFloat(key string) float64 {
	panic("irrelevant, NodeConfig#GetFloat32 must not be used")
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

// GetAccountsKeystore returns the accounts keystore path.
func (nc *NodeConfig) GetAccountsKeystore() string {
	return nc.AccountsKeystore
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

// GetP2PResponseDelay refer the interface
func (nc *NodeConfig) GetP2PResponseDelay() time.Duration {
	return nc.P2PResponseDelay
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

// GetTaskValidDuration returns the time duration until which task is valid
func (nc *NodeConfig) GetTaskValidDuration() time.Duration {
	return nc.TaskValidDuration
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

// GetEthereumMaxGasPrice refer the interface
func (nc *NodeConfig) GetEthereumMaxGasPrice() *big.Int {
	return nc.EthereumMaxGasPrice
}

// GetEthereumGasLimit refer the interface
func (nc *NodeConfig) GetEthereumGasLimit(op config.ContractOp) uint64 {
	return nc.EthereumGasLimits[op]
}

// GetEthereumGasMultiplier refer the interface
func (nc *NodeConfig) GetEthereumGasMultiplier() float64 {
	return nc.EthereumGasMultiplier
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

// GetCentChainAccount returns the Cent Chain account of the current node.
func (nc *NodeConfig) GetCentChainAccount() (config.CentChainAccount, error) {
	return nc.MainIdentity.CentChainAccount, nil
}

// GetCentChainNodeURL returns the URL of the CentChain Node.
func (nc *NodeConfig) GetCentChainNodeURL() string {
	return nc.CentChainNodeURL
}

// GetCentChainIntervalRetry returns duration to wait between retries.
func (nc *NodeConfig) GetCentChainIntervalRetry() time.Duration {
	return nc.CentChainIntervalRetry
}

// GetCentChainMaxRetries returns the max acceptable retries.
func (nc *NodeConfig) GetCentChainMaxRetries() int {
	return nc.CentChainMaxRetries
}

// GetCentChainAnchorLifespan returns the default lifespan of an anchor.
func (nc *NodeConfig) GetCentChainAnchorLifespan() time.Duration {
	return nc.CentChainAnchorLifespan
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

// GetP2PKeyPair refer the interface
func (nc *NodeConfig) GetP2PKeyPair() (pub, priv string) {
	return nc.MainIdentity.P2PKeyPair.Pub, nc.MainIdentity.P2PKeyPair.Pvt
}

// GetSigningKeyPair refer the interface
func (nc *NodeConfig) GetSigningKeyPair() (pub, priv string) {
	return nc.MainIdentity.SigningKeyPair.Pub, nc.MainIdentity.SigningKeyPair.Pvt
}

// GetPrecommitEnabled refer the interface
func (nc *NodeConfig) GetPrecommitEnabled() bool {
	return nc.MainIdentity.PrecommitEnabled
}

// GetLowEntropyNFTTokenEnabled refer the interface
func (nc *NodeConfig) GetLowEntropyNFTTokenEnabled() bool {
	return nc.LowEntropyNFTTokenEnabled
}

// IsPProfEnabled refer the interface
func (nc *NodeConfig) IsPProfEnabled() bool {
	return nc.PprofEnabled
}

// IsDebugLogEnabled refer the interface
func (nc *NodeConfig) IsDebugLogEnabled() bool {
	return nc.DebugLogEnabled
}

// ID Gets the ID of the document represented by this model
func (nc *NodeConfig) ID() ([]byte, error) {
	return []byte{}, nil
}

// Type Returns the underlying type of the NodeConfig
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
func NewNodeConfig(c config.Configuration) config.Configuration {
	mainAccount, _ := c.GetEthereumAccount(c.GetEthereumDefaultAccountName())
	mainIdentity, _ := c.GetIdentityID()
	p2pPub, p2pPriv := c.GetP2PKeyPair()
	signPub, signPriv := c.GetSigningKeyPair()
	centChainAccount, _ := c.GetCentChainAccount()

	return &NodeConfig{
		MainIdentity: Account{
			EthereumAccount: &config.AccountConfig{
				Address:  mainAccount.Address,
				Key:      mainAccount.Key,
				Password: mainAccount.Password,
			},
			EthereumDefaultAccountName:       c.GetEthereumDefaultAccountName(),
			IdentityID:                       mainIdentity,
			ReceiveEventNotificationEndpoint: c.GetReceiveEventNotificationEndpoint(),
			P2PKeyPair: KeyPair{
				Pub: p2pPub,
				Pvt: p2pPriv,
			},
			SigningKeyPair: KeyPair{
				Pub: signPub,
				Pvt: signPriv,
			},
			CentChainAccount: centChainAccount,
		},
		StoragePath:                    c.GetStoragePath(),
		AccountsKeystore:               c.GetAccountsKeystore(),
		P2PPort:                        c.GetP2PPort(),
		P2PExternalIP:                  c.GetP2PExternalIP(),
		P2PConnectionTimeout:           c.GetP2PConnectionTimeout(),
		P2PResponseDelay:               c.GetP2PResponseDelay(),
		ServerPort:                     c.GetServerPort(),
		ServerAddress:                  c.GetServerAddress(),
		NumWorkers:                     c.GetNumWorkers(),
		WorkerWaitTimeMS:               c.GetWorkerWaitTimeMS(),
		TaskValidDuration:              c.GetTaskValidDuration(),
		EthereumNodeURL:                c.GetEthereumNodeURL(),
		EthereumContextReadWaitTimeout: c.GetEthereumContextReadWaitTimeout(),
		EthereumContextWaitTimeout:     c.GetEthereumContextWaitTimeout(),
		EthereumIntervalRetry:          c.GetEthereumIntervalRetry(),
		EthereumMaxRetries:             c.GetEthereumMaxRetries(),
		EthereumMaxGasPrice:            c.GetEthereumMaxGasPrice(),
		EthereumGasLimits:              extractGasLimits(c),
		NetworkString:                  c.GetNetworkString(),
		BootstrapPeers:                 c.GetBootstrapPeers(),
		NetworkID:                      c.GetNetworkID(),
		SmartContractAddresses:         extractSmartContractAddresses(c),
		PprofEnabled:                   c.IsPProfEnabled(),
		DebugLogEnabled:                c.IsDebugLogEnabled(),
		LowEntropyNFTTokenEnabled:      c.GetLowEntropyNFTTokenEnabled(),
		CentChainMaxRetries:            c.GetCentChainMaxRetries(),
		CentChainIntervalRetry:         c.GetCentChainIntervalRetry(),
		CentChainAnchorLifespan:        c.GetCentChainAnchorLifespan(),
		CentChainNodeURL:               c.GetCentChainNodeURL(),
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

func extractGasLimits(c config.Configuration) map[config.ContractOp]uint64 {
	sms := make(map[config.ContractOp]uint64)
	names := config.ContractOps()
	for _, n := range names {
		sms[n] = c.GetEthereumGasLimit(n)
	}
	return sms
}

// Account exposes options specific to an account in the node
type Account struct {
	EthereumAccount                  *config.AccountConfig
	EthereumDefaultAccountName       string
	EthereumContextWaitTimeout       time.Duration
	ReceiveEventNotificationEndpoint string
	IdentityID                       []byte
	SigningKeyPair                   KeyPair
	P2PKeyPair                       KeyPair
	keys                             map[string]config.IDKey
	PrecommitEnabled                 bool
	CentChainAccount                 config.CentChainAccount
}

// GetPrecommitEnabled gets the enable pre commit value
func (acc *Account) GetPrecommitEnabled() bool {
	return acc.PrecommitEnabled
}

// GetEthereumAccount gets EthereumAccount
func (acc *Account) GetEthereumAccount() *config.AccountConfig {
	return acc.EthereumAccount
}

// GetCentChainAccount gets CentChainAccount
func (acc *Account) GetCentChainAccount() config.CentChainAccount {
	return acc.CentChainAccount
}

// GetEthereumDefaultAccountName gets EthereumDefaultAccountName
func (acc *Account) GetEthereumDefaultAccountName() string {
	return acc.EthereumDefaultAccountName
}

// GetReceiveEventNotificationEndpoint gets ReceiveEventNotificationEndpoint
func (acc *Account) GetReceiveEventNotificationEndpoint() string {
	return acc.ReceiveEventNotificationEndpoint
}

// GetIdentityID gets IdentityID
func (acc *Account) GetIdentityID() []byte {
	return acc.IdentityID
}

// GetP2PKeyPair gets P2PKeyPair
func (acc *Account) GetP2PKeyPair() (pub, priv string) {
	return acc.P2PKeyPair.Pub, acc.P2PKeyPair.Pvt
}

// GetSigningKeyPair gets SigningKeyPair
func (acc *Account) GetSigningKeyPair() (pub, priv string) {
	return acc.SigningKeyPair.Pub, acc.SigningKeyPair.Pvt
}

// GetEthereumContextWaitTimeout gets EthereumContextWaitTimeout
func (acc *Account) GetEthereumContextWaitTimeout() time.Duration {
	return acc.EthereumContextWaitTimeout
}

// SignMsg signs a message with the signing key
func (acc *Account) SignMsg(msg []byte) (*coredocumentpb.Signature, error) {
	keys, err := acc.GetKeys()
	if err != nil {
		return nil, err
	}
	signingKeyPair := keys[identity.KeyPurposeSigning.Name]
	signature, err := crypto.SignMessage(signingKeyPair.PrivateKey, msg, crypto.CurveSecp256K1)
	if err != nil {
		return nil, err
	}

	did := acc.GetIdentityID()

	return &coredocumentpb.Signature{
		SignatureId: append(did, signingKeyPair.PublicKey...),
		SignerId:    did,
		PublicKey:   signingKeyPair.PublicKey,
		Signature:   signature,
	}, nil
}

// GetKeys returns the keys of an account
// TODO remove GetKeys and add signing methods to account
func (acc *Account) GetKeys() (idKeys map[string]config.IDKey, err error) {
	if acc.keys == nil {
		acc.keys = map[string]config.IDKey{}
	}

	// KeyPurposeAction
	if _, ok := acc.keys[identity.KeyPurposeAction.Name]; !ok {
		pk := common.HexToAddress(acc.GetEthereumAccount().Address)
		address32Bytes, err := utils.ByteArrayTo32BytesLeftPadded(pk.Bytes())
		if err != nil {
			return idKeys, err
		}
		acc.keys[identity.KeyPurposeAction.Name] = config.IDKey{
			PublicKey: address32Bytes[:],
		}
	}

	// KeyPurposeP2PDiscovery
	if _, ok := acc.keys[identity.KeyPurposeP2PDiscovery.Name]; !ok {
		pk, sk, err := ed25519.GetSigningKeyPair(acc.GetP2PKeyPair())
		if err != nil {
			return idKeys, err
		}

		acc.keys[identity.KeyPurposeP2PDiscovery.Name] = config.IDKey{
			PublicKey:  pk,
			PrivateKey: sk}
	}

	// KeyPurposeSigning
	if _, ok := acc.keys[identity.KeyPurposeSigning.Name]; !ok {
		pk, sk, err := secp256k1.GetSigningKeyPair(acc.GetSigningKeyPair())
		if err != nil {
			return idKeys, err
		}
		address32Bytes := utils.AddressTo32Bytes(common.HexToAddress(secp256k1.GetAddress(pk)))

		acc.keys[identity.KeyPurposeSigning.Name] = config.IDKey{
			PublicKey:  address32Bytes[:],
			PrivateKey: sk}
	}
	acc.IdentityID = acc.GetIdentityID()
	return acc.keys, nil
}

// ID Get the ID of the document represented by this model
func (acc *Account) ID() []byte {
	return acc.IdentityID
}

// Type Returns the underlying type of the Account
func (acc *Account) Type() reflect.Type {
	return reflect.TypeOf(acc)
}

// JSON return the json representation of the model
func (acc *Account) JSON() ([]byte, error) {
	return json.Marshal(acc)
}

// FromJSON initialize the model with a json
func (acc *Account) FromJSON(data []byte) error {
	return json.Unmarshal(data, acc)
}

// NewAccount creates a new Account instance with configs
// TODO(ved): why do we need temp and this function ?
func NewAccount(ethAccountName string, c config.Configuration) (config.Account, error) {
	if ethAccountName == "" {
		return nil, errors.New("ethAccountName not provided")
	}
	id, err := c.GetIdentityID()
	if err != nil {
		return nil, err
	}
	acc, err := c.GetEthereumAccount(ethAccountName)
	if err != nil {
		return nil, err
	}

	cacc, err := c.GetCentChainAccount()
	if err != nil {
		return nil, err
	}

	return &Account{
		EthereumAccount:                  acc,
		EthereumDefaultAccountName:       c.GetEthereumDefaultAccountName(),
		EthereumContextWaitTimeout:       c.GetEthereumContextWaitTimeout(),
		IdentityID:                       id,
		ReceiveEventNotificationEndpoint: c.GetReceiveEventNotificationEndpoint(),
		P2PKeyPair:                       NewKeyPair(c.GetP2PKeyPair()),
		SigningKeyPair:                   NewKeyPair(c.GetSigningKeyPair()),
		PrecommitEnabled:                 c.GetPrecommitEnabled(),
		CentChainAccount:                 cacc,
	}, nil
}

// TempAccount creates a new Account without id validation, Must only be used for account creation.
func TempAccount(ethAccountName string, c config.Configuration) (config.Account, error) {
	if ethAccountName == "" {
		return nil, errors.New("ethAccountName not provided")
	}
	acc, err := c.GetEthereumAccount(ethAccountName)
	if err != nil {
		return nil, err
	}
	cacc, err := c.GetCentChainAccount()
	if err != nil {
		return nil, err
	}

	return &Account{
		EthereumAccount:                  acc,
		EthereumDefaultAccountName:       c.GetEthereumDefaultAccountName(),
		EthereumContextWaitTimeout:       c.GetEthereumContextWaitTimeout(),
		IdentityID:                       []byte{},
		ReceiveEventNotificationEndpoint: c.GetReceiveEventNotificationEndpoint(),
		P2PKeyPair:                       NewKeyPair(c.GetP2PKeyPair()),
		SigningKeyPair:                   NewKeyPair(c.GetSigningKeyPair()),
		PrecommitEnabled:                 c.GetPrecommitEnabled(),
		CentChainAccount:                 cacc,
	}, nil
}

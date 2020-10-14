package config

// Package the default resources into binary data that is embedded in centrifuge
// executable
//
//go:generate go-bindata -pkg resources -prefix "../../" -o ../resources/data.go ../build/configs/...

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/resources"
	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/centrifuge/go-substrate-rpc-client/signature"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	logging "github.com/ipfs/go-log"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

var log = logging.Logger("config")

// AccountHeaderKey is used as key for the account identity in the context.ContextWithValue.
var AccountHeaderKey struct{}

// ContractName is a type to indicate a contract name parameter
type ContractName string

// ContractOp is a type to indicate a contract operation name parameter
type ContractOp string

const (
	// AnchorRepo is the contract name for AnchorRepo
	AnchorRepo ContractName = "anchorRepository"

	// Identity is the contract name for Identity
	Identity ContractName = "identity"

	// IdentityFactory is the contract name for IdentityFactory
	IdentityFactory ContractName = "identityFactory"

	// IdentityRegistry is the contract name for IdentityRegistry
	IdentityRegistry ContractName = "identityRegistry"

	// InvoiceUnpaidNFT is the contract name for InvoiceUnpaidNFT
	InvoiceUnpaidNFT ContractName = "invoiceUnpaid"

	// IDCreate identity creation operation
	IDCreate ContractOp = "idCreate"

	// IDAddKey identity add key operation
	IDAddKey ContractOp = "idAddKey"

	// IDRevokeKey identity key revocation operation
	IDRevokeKey ContractOp = "idRevokeKey"

	// AnchorCommit anchor commit operation
	AnchorCommit ContractOp = "anchorCommit"

	// AnchorPreCommit anchor pre-commit operation
	AnchorPreCommit ContractOp = "anchorPreCommit"

	// NftMint nft minting operation
	NftMint ContractOp = "nftMint"

	// NftTransferFrom nft transferFrom operation
	NftTransferFrom ContractOp = "nftTransferFrom"

	// AssetStore is the operation name to store asset on chain
	AssetStore ContractOp = "assetStore"

	// PushToOracle for pushing data to oracle
	PushToOracle ContractOp = "pushToOracle"
)

// ContractNames returns the list of smart contract names currently used in the system, please update this when adding new contracts
func ContractNames() [5]ContractName {
	return [5]ContractName{AnchorRepo, IdentityFactory, Identity, IdentityRegistry, InvoiceUnpaidNFT}
}

// ContractOps returns the list of smart contract ops currently used in the system, please update this when adding new ops
func ContractOps() [8]ContractOp {
	return [8]ContractOp{IDCreate, IDAddKey, IDRevokeKey, AnchorCommit, AnchorPreCommit, NftMint, NftTransferFrom, AssetStore}
}

// Configuration defines the methods that a config type should implement.
type Configuration interface {
	storage.Model

	// generic methods
	IsSet(key string) bool
	Set(key string, value interface{})
	SetDefault(key string, value interface{})
	SetupSmartContractAddresses(network string, smartContractAddresses *SmartContractAddresses)
	Get(key string) interface{}
	GetString(key string) string
	GetBool(key string) bool
	GetInt(key string) int
	GetFloat(key string) float64
	GetDuration(key string) time.Duration

	GetStoragePath() string
	GetConfigStoragePath() string
	GetAccountsKeystore() string
	GetP2PPort() int
	GetP2PExternalIP() string
	GetP2PConnectionTimeout() time.Duration
	GetP2PResponseDelay() time.Duration
	GetServerPort() int
	GetServerAddress() string
	GetNumWorkers() int
	GetWorkerWaitTimeMS() int
	GetTaskValidDuration() time.Duration
	GetEthereumNodeURL() string
	GetEthereumContextReadWaitTimeout() time.Duration
	GetEthereumContextWaitTimeout() time.Duration
	GetEthereumIntervalRetry() time.Duration
	GetEthereumMaxRetries() int
	GetEthereumMaxGasPrice() *big.Int
	GetEthereumGasLimit(op ContractOp) uint64
	GetEthereumGasMultiplier() float64
	GetNetworkString() string
	GetNetworkKey(k string) string
	GetContractAddressString(address string) string
	GetContractAddress(contractName ContractName) common.Address
	GetBootstrapPeers() []string
	GetNetworkID() uint32

	// CentID specific configs (eg: for multi tenancy)
	GetEthereumAccount(accountName string) (account *AccountConfig, err error)
	GetEthereumDefaultAccountName() string
	GetReceiveEventNotificationEndpoint() string
	GetIdentityID() ([]byte, error)
	GetP2PKeyPair() (pub, priv string)
	GetSigningKeyPair() (pub, priv string)
	GetPrecommitEnabled() bool

	// GetLowEntropyNFTTokenEnabled enables low entropy token IDs.
	// The Dharma NFT Collateralizer and other contracts require tokenIds that are shorter than
	// the ERC721 standard bytes32. This option reduces the maximum value of the tokenId.
	// There are security implications of doing this. Specifically the risk of two users picking the
	// same token id and minting it at the same time goes up and it theoretically could lead to a loss of an
	// NFT with large enough NFTRegistries (>100'000 tokens). It is not recommended to use this option.
	GetLowEntropyNFTTokenEnabled() bool

	// debug specific methods
	IsPProfEnabled() bool
	IsDebugLogEnabled() bool

	// CentChain specific details.
	GetCentChainAccount() (CentChainAccount, error)
	GetCentChainIntervalRetry() time.Duration
	GetCentChainMaxRetries() int
	GetCentChainNodeURL() string
	GetCentChainAnchorLifespan() time.Duration
}

// Account exposes account options
type Account interface {
	storage.Model
	GetKeys() (map[string]IDKey, error)
	SignMsg(msg []byte) (*coredocumentpb.Signature, error)
	GetEthereumAccount() *AccountConfig
	GetEthereumDefaultAccountName() string
	GetReceiveEventNotificationEndpoint() string
	GetIdentityID() []byte
	GetP2PKeyPair() (pub, priv string)
	GetSigningKeyPair() (pub, priv string)
	GetEthereumContextWaitTimeout() time.Duration
	GetPrecommitEnabled() bool
	GetCentChainAccount() CentChainAccount
}

// Service exposes functions over the config objects
type Service interface {
	GetConfig() (Configuration, error)
	GetAccount(identifier []byte) (Account, error)
	GetAccounts() ([]Account, error)
	CreateConfig(data Configuration) (Configuration, error)
	CreateAccount(data Account) (Account, error)
	GenerateAccount(CentChainAccount) (Account, error)
	UpdateAccount(data Account) (Account, error)
	DeleteAccount(identifier []byte) error
	Sign(account, payload []byte) (*coredocumentpb.Signature, error)
}

// IDKey represents a key pair
type IDKey struct {
	PublicKey  []byte
	PrivateKey []byte
}

// configuration holds the configuration details for the node.
type configuration struct {
	mu         sync.RWMutex
	configFile string
	v          *viper.Viper
}

func (c *configuration) Type() reflect.Type {
	panic("irrelevant, configuration#Type must not be used")
}

func (c *configuration) JSON() ([]byte, error) {
	panic("irrelevant, configuration#JSON must not be used")
}

func (c *configuration) FromJSON(json []byte) error {
	panic("irrelevant, configuration#FromJSON must not be used")
}

// AccountConfig holds the account details.
type AccountConfig struct {
	Address  string
	Key      string
	Password string
}

// CentChainAccount holds the cent chain account details.
type CentChainAccount struct {
	ID       string `json:"id"`
	Secret   string `json:"secret,omitempty"`
	SS58Addr string `json:"ss_58_address"`
}

// KeyRingPair returns the keyring pair for the given account.
func (cacc CentChainAccount) KeyRingPair() (signature.KeyringPair, error) {
	pubKey, err := hexutil.Decode(cacc.ID)
	return signature.KeyringPair{
		URI:       cacc.Secret,
		Address:   cacc.SS58Addr,
		PublicKey: pubKey,
	}, err
}

// IsSet check if the key is set in the config.
func (c *configuration) IsSet(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.v.IsSet(key)
}

// Set update the key and the value it holds in the configuration.
func (c *configuration) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.v.Set(key, value)
}

// SetDefault sets the default value for the given key.
func (c *configuration) SetDefault(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.v.SetDefault(key, value)
}

// Get returns associated value for the key.
func (c *configuration) Get(key string) interface{} {
	return c.get(key)
}

// GetString returns value string associated with key.
func (c *configuration) GetString(key string) string {
	return cast.ToString(c.get(key))
}

// GetInt returns value int associated with key.
func (c *configuration) GetInt(key string) int {
	return cast.ToInt(c.get(key))
}

// GetFloat returns value float associated with key.
func (c *configuration) GetFloat(key string) float64 {
	return cast.ToFloat64(c.get(key))
}

// GetBool returns value bool associated with key.
func (c *configuration) GetBool(key string) bool {
	return cast.ToBool(c.get(key))
}

// GetDuration returns value duration associated with key.
func (c *configuration) GetDuration(key string) time.Duration {
	return cast.ToDuration(c.get(key))
}

func (c *configuration) get(key string) interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.v.Get(key)
}

// GetStoragePath returns the data storage backend.
func (c *configuration) GetStoragePath() string {
	return c.GetString("storage.path")
}

// GetConfigStoragePath returns the config storage backend.
func (c *configuration) GetConfigStoragePath() string {
	return c.GetString("configStorage.path")
}

// GetAccountsKeystore returns the accounts keystore location.
func (c *configuration) GetAccountsKeystore() string {
	return c.GetString("accounts.keystore")
}

// GetP2PPort returns P2P Port.
func (c *configuration) GetP2PPort() int {
	return c.GetInt("p2p.port")
}

// GetP2PExternalIP returns P2P External IP.
func (c *configuration) GetP2PExternalIP() string {
	return c.GetString("p2p.externalIP")
}

// GetP2PConnectionTimeout returns P2P Connect Timeout.
func (c *configuration) GetP2PConnectionTimeout() time.Duration {
	return c.GetDuration("p2p.connectTimeout")
}

// GetP2PResponseDelay returns P2P Response Delay.
func (c *configuration) GetP2PResponseDelay() time.Duration {
	return c.GetDuration("p2p.responseDelay")
}

// GetReceiveEventNotificationEndpoint returns the webhook endpoint defined in the config.
func (c *configuration) GetReceiveEventNotificationEndpoint() string {
	return c.GetString("notifications.endpoint")
}

// GetServerPort returns the defined server port in the config.
func (c *configuration) GetServerPort() int {
	return c.GetInt("nodePort")
}

// GetServerAddress returns the defined server address of form host:port in the config.
func (c *configuration) GetServerAddress() string {
	return fmt.Sprintf("%s:%s", c.GetString("nodeHostname"), c.GetString("nodePort"))
}

// GetNumWorkers returns number of queue workers defined in the config.
func (c *configuration) GetNumWorkers() int {
	return c.GetInt("queue.numWorkers")
}

// GetWorkerWaitTimeMS returns the queue worker sleep time between cycles.
func (c *configuration) GetWorkerWaitTimeMS() int {
	return c.GetInt("queue.workerWaitTimeMS")
}

func (c *configuration) GetTaskValidDuration() time.Duration {
	return c.GetDuration("queue.ValidFor")
}

// GetEthereumNodeURL returns the URL of the Ethereum Node.
func (c *configuration) GetEthereumNodeURL() string {
	return c.GetString("ethereum.nodeURL")
}

// GetEthereumContextReadWaitTimeout returns the read duration to pass for context.Deadline.
func (c *configuration) GetEthereumContextReadWaitTimeout() time.Duration {
	return c.GetDuration("ethereum.contextReadWaitTimeout")
}

// GetEthereumContextWaitTimeout returns the commit duration to pass for context.Deadline.
func (c *configuration) GetEthereumContextWaitTimeout() time.Duration {
	return c.GetDuration("ethereum.contextWaitTimeout")
}

// GetEthereumIntervalRetry returns duration to wait between retries.
func (c *configuration) GetEthereumIntervalRetry() time.Duration {
	return c.GetDuration("ethereum.intervalRetry")
}

// GetEthereumMaxRetries returns the max acceptable retries.
func (c *configuration) GetEthereumMaxRetries() int {
	return c.GetInt("ethereum.maxRetries")
}

// GetEthereumMaxGasPrice returns the gas price to use for a ethereum transaction.
func (c *configuration) GetEthereumMaxGasPrice() *big.Int {
	n := new(big.Int)
	n, ok := n.SetString(c.GetString("ethereum.maxGasPrice"), 10)
	if !ok {
		// node must not continue to run
		log.Panic("could not read ethereum.maxGasPrice")
	}
	return n
}

// GetEthereumGasLimit returns the gas limit to use for a ethereum transaction.
func (c *configuration) GetEthereumGasLimit(op ContractOp) uint64 {
	return cast.ToUint64(c.get(fmt.Sprintf("ethereum.gasLimits.%s", string(op))))
}

// GetEthereumGasMultiplier returns the gas multiplier to use for a ethereum transaction.
func (c *configuration) GetEthereumGasMultiplier() float64 {
	return c.GetFloat("ethereum.gasMultiplier")
}

// GetEthereumDefaultAccountName returns the default account to use for the transaction.
func (c *configuration) GetEthereumDefaultAccountName() string {
	return c.GetString("ethereum.defaultAccountName")
}

// GetEthereumAccount returns the account details associated with the account name.
func (c *configuration) GetEthereumAccount(accountName string) (account *AccountConfig, err error) {
	k := fmt.Sprintf("ethereum.accounts.%s", accountName)

	if !c.IsSet(k) {
		return nil, errors.New("no account found with account name %s", accountName)
	}

	// Workaround for bug https://github.com/spf13/viper/issues/309 && https://github.com/spf13/viper/issues/513
	account = &AccountConfig{
		Address:  c.GetString(fmt.Sprintf("%s.address", k)),
		Key:      c.GetString(fmt.Sprintf("%s.key", k)),
		Password: c.GetString(fmt.Sprintf("%s.password", k)),
	}

	return account, nil
}

// GetCentChainAccount returns Cent chain account from YAMl.
func (c *configuration) GetCentChainAccount() (acc CentChainAccount, err error) {
	k := "centChain.account"

	if !c.IsSet(k) {
		return acc, errors.New("Cent Chain Account not set")
	}

	return CentChainAccount{
		ID:       c.GetString(fmt.Sprintf("%s.id", k)),
		Secret:   c.GetString(fmt.Sprintf("%s.secret", k)),
		SS58Addr: c.GetString(fmt.Sprintf("%s.address", k)),
	}, nil
}

// GetCentChainNodeURL returns the URL of the CentChain Node.
func (c *configuration) GetCentChainNodeURL() string {
	return c.GetString("centChain.nodeURL")
}

// GetCentChainIntervalRetry returns duration to wait between retries.
func (c *configuration) GetCentChainIntervalRetry() time.Duration {
	return c.GetDuration("centChain.intervalRetry")
}

// GetCentChainMaxRetries returns the max acceptable retries.
func (c *configuration) GetCentChainMaxRetries() int {
	return c.GetInt("centChain.maxRetries")
}

// GetCentChainAnchorLifespan returns the default lifespan of an anchor.
func (c *configuration) GetCentChainAnchorLifespan() time.Duration {
	return c.GetDuration("centChain.anchorLifespan")
}

// GetNetworkString returns defined network the node is connected to.
func (c *configuration) GetNetworkString() string {
	return c.GetString("centrifugeNetwork")
}

// GetNetworkKey returns the specific key(k) value defined in the default network.
func (c *configuration) GetNetworkKey(k string) string {
	return fmt.Sprintf("networks.%s.%s", c.GetNetworkString(), k)
}

// GetContractAddressString returns the deployed contract address for a given contract.
func (c *configuration) GetContractAddressString(contract string) (address string) {
	return c.GetString(c.GetNetworkKey(fmt.Sprintf("contractAddresses.%s", contract)))
}

// GetContractAddress returns the deployed contract address for a given contract.
func (c *configuration) GetContractAddress(contractName ContractName) common.Address {
	return common.HexToAddress(c.GetContractAddressString(string(contractName)))
}

// GetBootstrapPeers returns the list of configured bootstrap nodes for the given network.
func (c *configuration) GetBootstrapPeers() []string {
	return cast.ToStringSlice(c.get(c.GetNetworkKey("bootstrapPeers")))
}

// GetNetworkID returns the numerical network id.
func (c *configuration) GetNetworkID() uint32 {
	return uint32(c.GetInt(c.GetNetworkKey("id")))
}

// GetIdentityID returns the self centID in bytes.
func (c *configuration) GetIdentityID() ([]byte, error) {
	id, err := hexutil.Decode(c.GetString("identityId"))
	if err != nil {
		return nil, errors.New("can't read identityId from config %v", err)
	}
	return id, err
}

// GetP2PKeyPair returns the P2P key pair.
func (c *configuration) GetP2PKeyPair() (pub, priv string) {
	return c.GetString("keys.p2p.publicKey"), c.GetString("keys.p2p.privateKey")
}

// GetSigningKeyPair returns the signing key pair.
func (c *configuration) GetSigningKeyPair() (pub, priv string) {
	return c.GetString("keys.signing.publicKey"), c.GetString("keys.signing.privateKey")
}

// IsPProfEnabled returns true if the pprof is enabled
func (c *configuration) IsPProfEnabled() bool {
	return c.GetBool("debug.pprof")
}

// IsDebugLogEnabled returns true if the debug logging is enabled
func (c *configuration) IsDebugLogEnabled() bool {
	return c.GetBool("debug.log")
}

// GetPrecommitEnabled returns true if precommit for anchors is enabled
func (c *configuration) GetPrecommitEnabled() bool {
	return c.GetBool("anchoring.precommit")
}

// GetLowEntropyNFTTokenEnabled returns true if low entropy nft token IDs are not enabled
func (c *configuration) GetLowEntropyNFTTokenEnabled() bool {
	return c.GetBool("nft.lowEntropyTokenIDEnabled")
}

// LoadConfiguration loads the configuration from the given file.
func LoadConfiguration(configFile string) Configuration {
	cfg := &configuration{configFile: configFile, mu: sync.RWMutex{}}
	cfg.initializeViper()
	return cfg
}

func (c *configuration) readConfigFile(path string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	err = c.v.MergeConfig(file)
	return err
}

// initializeViper loads viper if not loaded already.
// This method should not have any effects if Viper is already initialized.
func (c *configuration) initializeViper() {
	if c.v != nil {
		return
	}

	c.v = viper.New()
	c.v.SetConfigType("yaml")

	// Load defaults
	data, err := resources.Asset("go-centrifuge/build/configs/default_config.yaml")
	if err != nil {
		log.Panicf("failed to load (go-centrifuge/build/configs/default_config.yaml): %s", err)
	}

	err = c.v.ReadConfig(bytes.NewReader(data))
	if err != nil {
		log.Panicf("Error reading from default configuration (go-centrifuge/build/configs/default_config.yaml): %s", err)
	}
	// Load user specified config
	if c.configFile != "" {
		log.Infof("Loading user specified config from %s", c.configFile)
		err = c.readConfigFile(c.configFile)
		if err != nil {
			log.Panicf("Error reading config %s, %s", c.configFile, err)
		}
	} else {
		log.Info("No user config specified")
	}
	c.v.AutomaticEnv()
	c.v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	c.v.SetEnvPrefix("CENT")
}

// SmartContractAddresses encapsulates the smart contract addresses
type SmartContractAddresses struct {
	IdentityFactoryAddr string
}

// CreateConfigFile creates minimum config file with arguments
func CreateConfigFile(args map[string]interface{}) (*viper.Viper, error) {
	targetDataDir := args["targetDataDir"].(string)
	accountKeyPath := args["accountKeyPath"].(string)
	accountPassword := args["accountPassword"].(string)
	network := args["network"].(string)
	ethNodeURL := args["ethNodeURL"].(string)
	bootstraps := args["bootstraps"].([]string)
	apiPort := args["apiPort"].(int64)
	p2pPort := args["p2pPort"].(int64)
	p2pConnectTimeout := args["p2pConnectTimeout"].(string)
	preCommitEnabled := args["preCommitEnabled"].(bool)
	apiHost := args["apiHost"].(string)
	webhookURL, _ := args["webhookURL"].(string)
	centChainURL, _ := args["centChainURL"].(string)
	centChainID, _ := args["centChainID"].(string)
	centChainSecret, _ := args["centChainSecret"].(string)
	centChainAddr, _ := args["centChainAddr"].(string)

	if targetDataDir == "" {
		return nil, errors.New("targetDataDir not provided")
	}
	if _, err := os.Stat(targetDataDir); os.IsNotExist(err) {
		err := os.MkdirAll(targetDataDir, os.ModePerm)
		if err != nil {
			return nil, err
		}
	}

	if _, err := os.Stat(accountKeyPath); os.IsNotExist(err) {
		return nil, errors.New("account Key Path [%s] does not exist", accountKeyPath)
	}

	bfile, err := ioutil.ReadFile(accountKeyPath)
	if err != nil {
		return nil, err
	}

	err = os.Setenv("CENT_ETHEREUM_ACCOUNTS_MAIN_KEY", string(bfile))
	if err != nil {
		return nil, err
	}

	if accountPassword == "" {
		log.Warningf("Account Password not provided")
	}

	err = os.Setenv("CENT_ETHEREUM_ACCOUNTS_MAIN_PASSWORD", accountPassword)
	if err != nil {
		return nil, err
	}

	if centChainAddr == "" || centChainSecret == "" || centChainID == "" {
		return nil, errors.New("Centrifuge chain ID, Secret, and Address are required")
	}

	v := viper.New()
	v.SetConfigType("yaml")
	v.Set("storage.path", targetDataDir+"/db/centrifuge_data.leveldb")
	v.Set("configStorage.path", targetDataDir+"/db/centrifuge_config_data.leveldb")
	v.Set("accounts.keystore", targetDataDir+"/accounts")
	v.Set("anchoring.precommit", preCommitEnabled)
	v.Set("identityId", "")
	v.Set("centrifugeNetwork", network)
	v.Set("nodeHostname", apiHost)
	v.Set("nodePort", apiPort)
	v.Set("p2p.port", p2pPort)
	v.Set("notifications.endpoint", webhookURL)
	if p2pConnectTimeout != "" {
		v.Set("p2p.connectTimeout", p2pConnectTimeout)
	}
	v.Set("ethereum.nodeURL", ethNodeURL)
	v.Set("ethereum.accounts.main.key", "")
	v.Set("ethereum.accounts.main.password", "")
	v.Set("centChain.nodeURL", centChainURL)
	v.Set("centChain.account.id", centChainID)
	v.Set("centChain.account.secret", centChainSecret)
	v.Set("centChain.account.address", centChainAddr)
	v.Set("keys.p2p.privateKey", targetDataDir+"/p2p.key.pem")
	v.Set("keys.p2p.publicKey", targetDataDir+"/p2p.pub.pem")
	v.Set("keys.signing.privateKey", targetDataDir+"/signing.key.pem")
	v.Set("keys.signing.publicKey", targetDataDir+"/signing.pub.pem")

	if bootstraps != nil {
		v.Set("networks."+network+".bootstrapPeers", bootstraps)
	}

	if smartContractAddresses, ok := args["smartContractAddresses"].(*SmartContractAddresses); ok {
		v.Set("networks."+network+".contractAddresses.identityFactory", smartContractAddresses.IdentityFactoryAddr)
	}

	v.SetConfigFile(targetDataDir + "/config.yaml")

	err = v.WriteConfig()
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	return v, nil
}

func (c *configuration) SetupSmartContractAddresses(network string, smartContractAddresses *SmartContractAddresses) {
	c.v.Set("networks."+network+".contractAddresses.identityFactory", smartContractAddresses.IdentityFactoryAddr)
}

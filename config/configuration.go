package config

// Package the default resources into binary data that is embedded in centrifuge
// executable
//
//go:generate go-bindata -pkg resources -prefix "../../" -o ../resources/data.go ../build/configs/...

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/centrifuge/go-centrifuge/bootstrap"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/resources"
	"github.com/centrifuge/go-centrifuge/storage"
	logging "github.com/ipfs/go-log"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

var log = logging.Logger("config")

var allowedURLScheme = map[string]struct{}{
	"http":  {},
	"https": {},
	"ws":    {},
	"wss":   {},
}

// AccountHeaderKey is used as key for the account identity in the context.ContextWithValue.
var AccountHeaderKey struct{}

// ContractName is a type to indicate a contract name parameter
type ContractName string

// ContractOp is a type to indicate a contract operation name parameter
type ContractOp string

const (
	defaultURLScheme = "https"

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
	GetNetworkString() string
	GetNetworkKey(k string) string
	GetBootstrapPeers() []string
	GetNetworkID() uint32

	GetP2PKeyPair() (string, string)
	GetSigningKeyPair() (string, string)

	// CentID specific configs (eg: for multi tenancy)
	//GetReceiveEventNotificationEndpoint() string
	GetPrecommitEnabled() bool

	// debug specific methods
	IsPProfEnabled() bool
	IsDebugLogEnabled() bool

	// CentChain specific details.
	GetCentChainIntervalRetry() time.Duration
	GetCentChainMaxRetries() int
	GetCentChainNodeURL() string
	GetCentChainAnchorLifespan() time.Duration
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

// GetP2PKeyPair returns the P2P key pair.
func (c *configuration) GetP2PKeyPair() (pub, priv string) {
	return c.GetString("keys.p2p.publicKey"), c.GetString("keys.p2p.privateKey")
}

// GetSigningKeyPair returns the signing key pair.
func (c *configuration) GetSigningKeyPair() (pub, priv string) {
	return c.GetString("keys.signing.publicKey"), c.GetString("keys.signing.privateKey")
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

// GetBootstrapPeers returns the list of configured bootstrap nodes for the given network.
func (c *configuration) GetBootstrapPeers() []string {
	return cast.ToStringSlice(c.get(c.GetNetworkKey("bootstrapPeers")))
}

// GetNetworkID returns the numerical network id.
func (c *configuration) GetNetworkID() uint32 {
	return uint32(c.GetInt(c.GetNetworkKey("id")))
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

	err = c.validateURLs([]string{"ethNodeURL", "centChainURL"})
	if err != nil {
		log.Panicf("error: %v", err)
	}
}

func (c *configuration) validateURLs(keys []string) error {
	for _, key := range keys {
		value, _ := c.v.Get(key).(string)
		value, err := validateURL(value)
		if err != nil {
			return err
		}
		c.v.Set(key, value)
	}
	return nil
}

// CreateConfigFile creates minimum config file with arguments
func CreateConfigFile(args map[string]interface{}) (*viper.Viper, error) {
	targetDataDir := args["targetDataDir"].(string)
	network := args["network"].(string)
	bootstraps := args["bootstraps"].([]string)
	apiPort := args["apiPort"].(int64)
	p2pPort := args["p2pPort"].(int64)
	p2pConnectTimeout := args["p2pConnectTimeout"].(string)
	apiHost := args["apiHost"].(string)

	centChainURL, err := validateURL(args["centChainURL"].(string))

	if err != nil {
		return nil, err
	}

	if targetDataDir == "" {
		return nil, errors.New("targetDataDir not provided")
	}

	if _, err := os.Stat(targetDataDir); os.IsNotExist(err) {
		err := os.MkdirAll(targetDataDir, os.ModePerm)
		if err != nil {
			return nil, err
		}
	}

	v := viper.New()
	v.SetConfigType("yaml")
	v.Set("storage.path", targetDataDir+"/db/centrifuge_data.leveldb")
	v.Set("configStorage.path", targetDataDir+"/db/centrifuge_config_data.leveldb")
	v.Set("accounts.keystore", targetDataDir+"/accounts")
	v.Set("centrifugeNetwork", network)
	v.Set("nodeHostname", apiHost)
	v.Set("nodePort", apiPort)
	v.Set("p2p.port", p2pPort)
	v.Set("keys.p2p.privateKey", targetDataDir+"/p2p.key.pem")
	v.Set("keys.p2p.publicKey", targetDataDir+"/p2p.pub.pem")
	v.Set("keys.signing.privateKey", targetDataDir+"/signing.key.pem")
	v.Set("keys.signing.publicKey", targetDataDir+"/signing.pub.pem")
	if p2pConnectTimeout != "" {
		v.Set("p2p.connectTimeout", p2pConnectTimeout)
	}
	v.Set("centChain.nodeURL", centChainURL)

	if bootstraps != nil {
		v.Set("networks."+network+".bootstrapPeers", bootstraps)
	}

	v.SetConfigFile(targetDataDir + "/config.yaml")

	err = v.WriteConfig()
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	return v, nil
}

const (
	// ErrConfigRetrieve must be returned when there is an error while retrieving config
	ErrConfigRetrieve = errors.Error("error when retrieving config")
)

// RetrieveConfig retrieves system config giving priority to db stored config
func RetrieveConfig(dbOnly bool, ctx map[string]interface{}) (Configuration, error) {
	var cfg Configuration
	var err error
	if cfgService, ok := ctx[BootstrappedConfigStorage].(Service); ok {
		// may be we need a way to detect a corrupted db here
		cfg, err = cfgService.GetConfig()
		if err != nil {
			return nil, err
		}
		return cfg, nil
	}

	// we have to allow loading from file in case this is coming from create config cmd where we don't add configs to db
	if _, ok := ctx[bootstrap.BootstrappedConfig]; ok && !dbOnly {
		cfg = ctx[bootstrap.BootstrappedConfig].(Configuration)
	} else {
		return nil, errors.NewTypedError(ErrConfigRetrieve, err)
	}
	return cfg, nil
}

func validateURL(u string) (string, error) {
	parsedURL, err := url.Parse(u)
	if err != nil {
		return "", err
	}

	if parsedURL.Scheme == "" {
		parsedURL.Scheme = defaultURLScheme
	}

	if _, ok := allowedURLScheme[parsedURL.Scheme]; !ok {
		return "", errors.New("url scheme %s is not allowed", parsedURL.Scheme)
	}

	return parsedURL.String(), nil
}

// Account exposes account options
type Account interface {
	storage.Model
	GetIdentity() identity.DID

	GetP2PPublicKey() []byte
	GetSigningPublicKey() []byte

	SignMsg(msg []byte) (*coredocumentpb.Signature, error)

	GetWebhookURL() string
	GetPrecommitEnabled() bool

	GetAccountProxies() AccountProxies
}

type AccountProxy struct {
	Default      bool             `json:"default"`
	AccountID    identity.DID     `json:"account_id"`
	ChainAccount CentChainAccount `json:"centrifuge_chain_account"`
	ProxyType    string           `json:"proxy_type" enums:"any,non_transfer,governance,staking,non_proxy,borrow,price,invest,proxy_management,keystore_management,nft_mint,nft_transfer,nft_management"`
}

type AccountProxies []AccountProxy

const (
	ErrDefaultAccountProxyNotFound = errors.Error("default account proxy not found")
)

func (a *AccountProxies) GetDefault() (*AccountProxy, error) {
	for _, accountProxy := range *a {
		if accountProxy.Default {
			return &accountProxy, nil
		}
	}

	return nil, ErrDefaultAccountProxyNotFound
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

// Service exposes functions over the config objects
type Service interface {
	GetConfig() (Configuration, error)
	GetAccount(identifier []byte) (Account, error)
	GetAccounts() ([]Account, error)
	CreateConfig(data Configuration) (Configuration, error)
	CreateAccount(data Account) (Account, error)
	UpdateAccount(data Account) (Account, error)
	DeleteAccount(identifier []byte) error
	Sign(account, payload []byte) (*coredocumentpb.Signature, error)
	GenerateAccountAsync(account CentChainAccount) (did []byte, jobID []byte, err error)
}

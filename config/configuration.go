package config

// Package the default resources into binary data that is embedded in centrifuge
// executable
//
//go:generate go-bindata -pkg resources -prefix "../../" -o ../resources/data.go ../build/configs/...

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/resources"
	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	"github.com/ethereum/go-ethereum/common/hexutil"
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
)

// Configuration defines the methods that a config type should implement.
type Configuration interface {
	storage.Model

	GetStoragePath() string
	GetConfigStoragePath() string
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
	GetBootstrapPeers() []string
	GetNetworkID() uint32

	GetP2PKeyPair() (string, string)
	GetSigningKeyPair() (string, string)
	GetNodeAdminKeyPair() (string, string)

	// debug specific methods
	IsPProfEnabled() bool
	IsDebugLogEnabled() bool
	IsAuthenticationEnabled() bool

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
	return reflect.TypeOf(c)
}

func (c *configuration) JSON() ([]byte, error) {
	return json.Marshal(c)
}

func (c *configuration) FromJSON(data []byte) error {
	return json.Unmarshal(data, c)
}

// GetStoragePath returns the data storage backend.
func (c *configuration) GetStoragePath() string {
	return c.getString("storage.path")
}

// GetConfigStoragePath returns the config storage backend.
func (c *configuration) GetConfigStoragePath() string {
	return c.getString("configStorage.path")
}

// GetP2PPort returns P2P Port.
func (c *configuration) GetP2PPort() int {
	return c.getInt("p2p.port")
}

// GetP2PExternalIP returns P2P External IP.
func (c *configuration) GetP2PExternalIP() string {
	return c.getString("p2p.externalIP")
}

// GetP2PConnectionTimeout returns P2P Connect Timeout.
func (c *configuration) GetP2PConnectionTimeout() time.Duration {
	return c.getDuration("p2p.connectTimeout")
}

// GetP2PResponseDelay returns P2P Response Delay.
func (c *configuration) GetP2PResponseDelay() time.Duration {
	return c.getDuration("p2p.responseDelay")
}

// GetP2PKeyPair returns the P2P key pair.
func (c *configuration) GetP2PKeyPair() (pub, priv string) {
	return c.getString("keys.p2p.publicKey"), c.getString("keys.p2p.privateKey")
}

// GetSigningKeyPair returns the signing key pair.
func (c *configuration) GetSigningKeyPair() (pub, priv string) {
	return c.getString("keys.signing.publicKey"), c.getString("keys.signing.privateKey")
}

// GetNodeAdminKeyPair returns the node admin key pair.
func (c *configuration) GetNodeAdminKeyPair() (pub, priv string) {
	return c.getString("keys.nodeAdmin.publicKey"), c.getString("keys.nodeAdmin.privateKey")
}

// GetServerPort returns the defined server port in the config.
func (c *configuration) GetServerPort() int {
	return c.getInt("nodePort")
}

// GetServerAddress returns the defined server address of form host:port in the config.
func (c *configuration) GetServerAddress() string {
	return fmt.Sprintf("%s:%s", c.getString("nodeHostname"), c.getString("nodePort"))
}

// GetNumWorkers returns number of queue workers defined in the config.
func (c *configuration) GetNumWorkers() int {
	return c.getInt("queue.numWorkers")
}

// GetWorkerWaitTimeMS returns the queue worker sleep time between cycles.
func (c *configuration) GetWorkerWaitTimeMS() int {
	return c.getInt("queue.workerWaitTimeMS")
}

func (c *configuration) GetTaskValidDuration() time.Duration {
	return c.getDuration("queue.ValidFor")
}

// GetCentChainNodeURL returns the URL of the CentChain Node.
func (c *configuration) GetCentChainNodeURL() string {
	return c.getString("centChain.nodeURL")
}

// GetCentChainIntervalRetry returns duration to wait between retries.
func (c *configuration) GetCentChainIntervalRetry() time.Duration {
	return c.getDuration("centChain.intervalRetry")
}

// GetCentChainMaxRetries returns the max acceptable retries.
func (c *configuration) GetCentChainMaxRetries() int {
	return c.getInt("centChain.maxRetries")
}

// GetCentChainAnchorLifespan returns the default lifespan of an anchor.
func (c *configuration) GetCentChainAnchorLifespan() time.Duration {
	return c.getDuration("centChain.anchorLifespan")
}

// GetNetworkString returns defined network the node is connected to.
func (c *configuration) GetNetworkString() string {
	return c.getString("centrifugeNetwork")
}

// getNetworkKey returns the specific key(k) value defined in the default network.
func (c *configuration) getNetworkKey(k string) string {
	return fmt.Sprintf("networks.%s.%s", c.GetNetworkString(), k)
}

// GetBootstrapPeers returns the list of configured bootstrap nodes for the given network.
func (c *configuration) GetBootstrapPeers() []string {
	return cast.ToStringSlice(c.get(c.getNetworkKey("bootstrapPeers")))
}

// GetNetworkID returns the numerical network id.
func (c *configuration) GetNetworkID() uint32 {
	return uint32(c.getInt(c.getNetworkKey("id")))
}

// IsPProfEnabled returns true if the pprof is enabled
func (c *configuration) IsPProfEnabled() bool {
	return c.getBool("debug.pprof")
}

// IsDebugLogEnabled returns true if the debug logging is enabled
func (c *configuration) IsDebugLogEnabled() bool {
	return c.getBool("debug.log")
}

// IsAuthenticationEnabled returns true if the authentication is enabled
func (c *configuration) IsAuthenticationEnabled() bool {
	return c.getBool("authentication.enabled")
}

func (c *configuration) get(key string) interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.v.Get(key)
}

// getString returns value string associated with key.
func (c *configuration) getString(key string) string {
	return cast.ToString(c.get(key))
}

// getInt returns value int associated with key.
func (c *configuration) getInt(key string) int {
	return cast.ToInt(c.get(key))
}

// getFloat returns value float associated with key.
func (c *configuration) getFloat(key string) float64 {
	return cast.ToFloat64(c.get(key))
}

// getBool returns value bool associated with key.
func (c *configuration) getBool(key string) bool {
	return cast.ToBool(c.get(key))
}

// getDuration returns value duration associated with key.
func (c *configuration) getDuration(key string) time.Duration {
	return cast.ToDuration(c.get(key))
}

// AccountConfig holds the account details.
type AccountConfig struct {
	Address  string
	Key      string
	Password string
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
	authenticationEnabled := args["authenticationEnabled"].(bool)

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
	v.Set("centrifugeNetwork", network)
	v.Set("nodeHostname", apiHost)
	v.Set("nodePort", apiPort)
	v.Set("p2p.port", p2pPort)
	v.Set("keys.p2p.privateKey", targetDataDir+"/p2p.key.pem")
	v.Set("keys.p2p.publicKey", targetDataDir+"/p2p.pub.pem")
	v.Set("keys.signing.privateKey", targetDataDir+"/signing.key.pem")
	v.Set("keys.signing.publicKey", targetDataDir+"/signing.pub.pem")
	v.Set("keys.nodeAdmin.privateKey", targetDataDir+"/node_admin.key.pem")
	v.Set("keys.nodeAdmin.publicKey", targetDataDir+"/node_admin.pub.pem")
	v.Set("authentication.enabled", authenticationEnabled)
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

type NodeAdmin interface {
	storage.Model

	AccountID() *types.AccountID
}

// Account exposes account options
type Account interface {
	storage.Model

	GetIdentity() *types.AccountID

	GetP2PPublicKey() []byte
	GetSigningPublicKey() []byte

	SignMsg(msg []byte) (*coredocumentpb.Signature, error)

	GetWebhookURL() string
	GetPrecommitEnabled() bool

	GetAccountProxies() AccountProxies
}

type AccountProxy struct {
	Default     bool             `json:"default"`
	AccountID   *types.AccountID `json:"account_id"`
	Secret      string           `json:"secret"`
	SS58Address string           `json:"ss58_address"`
	ProxyType   types.ProxyType  `json:"proxy_type"`
}

const (
	ErrNilAccountProxy = errors.Error("nil account proxy")
)

func (a *AccountProxy) ToKeyringPair() (*signature.KeyringPair, error) {
	if a == nil {
		return nil, ErrNilAccountProxy
	}

	return &signature.KeyringPair{
		URI:       a.Secret,
		Address:   a.SS58Address,
		PublicKey: a.AccountID[:],
	}, nil
}

type AccountProxies []*AccountProxy

const (
	ErrDefaultAccountProxyNotFound = errors.Error("default account proxy not found")
)

func (a AccountProxies) GetDefault() (*AccountProxy, error) {
	for _, accountProxy := range a {
		if accountProxy.Default {
			return accountProxy, nil
		}
	}

	return nil, ErrDefaultAccountProxyNotFound
}

func (a AccountProxies) WithProxyType(proxyType types.ProxyType) (*AccountProxy, error) {
	for _, accountProxy := range a {
		if accountProxy.ProxyType == proxyType {
			return accountProxy, nil
		}
	}

	return a.GetDefault()
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
	GetNodeAdmin() (NodeAdmin, error)
	GetAccount(identifier []byte) (Account, error)
	GetAccounts() ([]Account, error)
	CreateConfig(data Configuration) (Configuration, error)
	CreateNodeAdmin(nodeAdmin NodeAdmin) (NodeAdmin, error)
	CreateAccount(data Account) (Account, error)
	UpdateAccount(data Account) (Account, error)
	DeleteAccount(identifier []byte) error
	Sign(account, payload []byte) (*coredocumentpb.Signature, error)
}

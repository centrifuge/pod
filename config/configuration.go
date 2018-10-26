package config

// Package the default resources into binary data that is embedded in centrifuge
// executable
//
//go:generate go-bindata -pkg resources -prefix "../../" -o ../resources/data.go ../build/configs/...

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/centrifuge/go-centrifuge/centerrors"
	"github.com/centrifuge/go-centrifuge/resources"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	logging "github.com/ipfs/go-log"
	"github.com/spf13/viper"
)

var log = logging.Logger("config")
var Config *Configuration

type Configuration struct {
	configFile string
	V          *viper.Viper
}

type AccountConfig struct {
	Address  string
	Key      string
	Password string
}

// IdentityConfig holds ID, public and private key of a single entity
type IdentityConfig struct {
	ID         []byte
	PublicKey  []byte
	PrivateKey []byte
}

// Storage backend
func (c *Configuration) GetStoragePath() string {
	return c.V.GetString("storage.Path")
}

// P2P Port
func (c *Configuration) GetP2PPort() int {
	return c.V.GetInt("p2p.port")
}

// P2P External IP
func (c *Configuration) GetP2PExternalIP() string {
	return c.V.GetString("p2p.externalIP")
}

////////////////////////////////////////////////////////////////////////////////
// Notifications
////////////////////////////////////////////////////////////////////////////////
func (c *Configuration) GetReceiveEventNotificationEndpoint() string {
	return c.V.GetString("notifications.endpoint")
}

////////////////////////////////////////////////////////////////////////////////
// Server
////////////////////////////////////////////////////////////////////////////////

func (c *Configuration) GetServerPort() int {
	return c.V.GetInt("nodePort")
}

func (c *Configuration) GetServerAddress() string {
	return fmt.Sprintf("%s:%s", c.V.GetString("nodeHostname"), c.V.GetString("nodePort"))
}

////////////////////////////////////////////////////////////////////////////////
// Queuing
////////////////////////////////////////////////////////////////////////////////

func (c *Configuration) GetNumWorkers() int {
	return c.V.GetInt("queue.numWorkers")
}

func (c *Configuration) GetWorkerWaitTimeMS() int {
	return c.V.GetInt("queue.workerWaitTimeMS")
}

////////////////////////////////////////////////////////////////////////////////
// Ethereum
////////////////////////////////////////////////////////////////////////////////
func (c *Configuration) GetEthereumNodeURL() string {
	return c.V.GetString("ethereum.nodeURL")
}

func (c *Configuration) GetEthereumContextWaitTimeout() time.Duration {
	return c.V.GetDuration("ethereum.contextWaitTimeout")
}

func (c *Configuration) GetEthereumIntervalRetry() time.Duration {
	return c.V.GetDuration("ethereum.intervalRetry")
}

func (c *Configuration) GetEthereumMaxRetries() int {
	return c.V.GetInt("ethereum.maxRetries")
}

func (c *Configuration) GetEthereumGasPrice() *big.Int {
	return big.NewInt(c.V.GetInt64("ethereum.gasPrice"))
}

func (c *Configuration) GetEthereumGasLimit() uint64 {
	return uint64(c.V.GetInt64("ethereum.gasLimit"))
}

func (c *Configuration) GetEthereumDefaultAccountName() string {
	return c.V.GetString("ethereum.defaultAccountName")
}

func (c *Configuration) GetEthereumAccount(accountName string) (account *AccountConfig, err error) {
	k := fmt.Sprintf("ethereum.accounts.%s", accountName)

	if !c.V.IsSet(k) {
		return nil, fmt.Errorf("no account found with account name %s", accountName)
	}

	// Workaround for bug https://github.com/spf13/viper/issues/309 && https://github.com/spf13/viper/issues/513
	account = &AccountConfig{
		Address:  c.V.GetString(fmt.Sprintf("%s.address", k)),
		Key:      c.V.GetString(fmt.Sprintf("%s.key", k)),
		Password: c.V.GetString(fmt.Sprintf("%s.password", k)),
	}

	return account, nil
}

// Important flag for concurrency handling. Disable if Ethereum client doesn't support txpool API (INFURA)
func (c *Configuration) GetTxPoolAccessEnabled() bool {
	return c.V.GetBool("ethereum.txPoolAccessEnabled")
}

////////////////////////////////////////////////////////////////////////////////
// Network Configuration
////////////////////////////////////////////////////////////////////////////////
func (c *Configuration) GetNetworkString() string {
	return c.V.GetString("centrifugeNetwork")
}

func (c *Configuration) GetNetworkKey(k string) string {
	return fmt.Sprintf("networks.%s.%s", c.GetNetworkString(), k)
}

// GetContractAddressString returns the deployed contract address for a given contract.
func (c *Configuration) GetContractAddressString(contract string) (address string) {
	return c.V.GetString(c.GetNetworkKey(fmt.Sprintf("contractAddresses.%s", contract)))
}

// GetContractAddress returns the deployed contract address for a given contract.
func (c *Configuration) GetContractAddress(contract string) (address common.Address) {
	return common.HexToAddress(c.GetContractAddressString(contract))
}

// GetBootstrapPeers returns the list of configured bootstrap nodes for the given network.
func (c *Configuration) GetBootstrapPeers() []string {
	return c.V.GetStringSlice(c.GetNetworkKey("bootstrapPeers"))
}

// GetNetworkID returns the numerical network id.
func (c *Configuration) GetNetworkID() uint32 {
	return uint32(c.V.GetInt(c.GetNetworkKey("id")))
}

// GetIdentityID returns the self centID
func (c *Configuration) GetIdentityID() ([]byte, error) {
	id, err := hexutil.Decode(c.V.GetString("identityId"))
	if err != nil {
		return nil, centerrors.Wrap(err, "can't read identityId from config")
	}
	return id, err
}

func (c *Configuration) GetSigningKeyPair() (pub, priv string) {
	return c.V.GetString("keys.signing.publicKey"), c.V.GetString("keys.signing.privateKey")
}

func (c *Configuration) GetEthAuthKeyPair() (pub, priv string) {
	return c.V.GetString("keys.ethauth.publicKey"), c.V.GetString("keys.ethauth.privateKey")
}

// Configuration Implementation
func NewConfiguration(configFile string) *Configuration {
	c := Configuration{configFile: configFile}
	return &c
}

// SetConfigFile returns an error if viper was already initialized.
func (c *Configuration) SetConfigFile(path string) error {
	if c.V != nil {
		return errors.New("viper already initialized. Can't set config file")
	}
	c.configFile = path
	return nil
}

func (c *Configuration) ReadConfigFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	err = c.V.MergeConfig(file)
	return err
}

func (c *Configuration) InitializeViper() {
	// This method should not have any effects if Viper is already initialized.
	if c.V != nil {
		return
	}

	c.V = viper.New()
	c.V.SetConfigType("yaml")

	// Load defaults
	data, err := resources.Asset("go-centrifuge/build/configs/default_config.yaml")
	if err != nil {
		log.Panicf("failed to load (go-centrifuge/build/configs/default_config.yaml): %s", err)
	}

	err = c.V.ReadConfig(bytes.NewReader(data))
	if err != nil {
		log.Panicf("Error reading from default configuration (go-centrifuge/build/configs/default_config.yaml): %s", err)
	}
	// Load user specified config
	if c.configFile != "" {
		log.Infof("Loading user specified config from %s", c.configFile)
		err = c.ReadConfigFile(c.configFile)
		if err != nil {
			log.Panicf("Error reading config %s, %s", c.configFile, err)
		}
	} else {
		log.Info("No user config specified")
	}
	c.V.AutomaticEnv()
	c.V.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	c.V.SetEnvPrefix("CENT")
}

func Bootstrap(configFile string) {
	Config = NewConfiguration(configFile)
	Config.InitializeViper()
}

// CreateConfigFile creates minimum config file with arguments
func CreateConfigFile(args map[string]interface{}) (*viper.Viper, error) {
	targetDataDir := args["targetDataDir"].(string)
	accountKeyPath := args["accountKeyPath"].(string)
	accountPassword := args["accountPassword"].(string)
	network := args["network"].(string)
	ethNodeUrl := args["ethNodeUrl"].(string)
	bootstraps := args["bootstraps"].([]string)
	apiPort := args["apiPort"].(int64)
	p2pPort := args["p2pPort"].(int64)

	if targetDataDir == "" {
		return nil, errors.New("targetDataDir not provided")
	}
	if _, err := os.Stat(targetDataDir); os.IsNotExist(err) {
		os.Mkdir(targetDataDir, os.ModePerm)
	}

	if _, err := os.Stat(accountKeyPath); os.IsNotExist(err) {
		return nil, errors.New("Account Key Path does not exist")
	}

	bfile, err := ioutil.ReadFile(accountKeyPath)
	if err != nil {
		return nil, err
	}

	if accountPassword == "" {
		log.Warningf("Account Password not provided")
	}

	v := viper.New()
	v.SetConfigType("yaml")
	v.Set("storage.path", targetDataDir+"/db/centrifuge_data.leveldb")
	v.Set("identityId", "")
	v.Set("centrifugeNetwork", network)
	v.Set("nodeHostname", "0.0.0.0")
	v.Set("nodePort", apiPort)
	v.Set("p2p.port", p2pPort)
	v.Set("ethereum.nodeURL", ethNodeUrl)
	v.Set("ethereum.accounts.main.key", string(bfile))
	v.Set("ethereum.accounts.main.password", accountPassword)
	v.Set("keys.p2p.privateKey", targetDataDir+"/p2p.key.pem")
	v.Set("keys.p2p.publicKey", targetDataDir+"/p2p.pub.pem")
	v.Set("keys.ethauth.privateKey", targetDataDir+"/ethauth.key.pem")
	v.Set("keys.ethauth.publicKey", targetDataDir+"/ethauth.pub.pem")
	v.Set("keys.signing.privateKey", targetDataDir+"/signing.key.pem")
	v.Set("keys.signing.publicKey", targetDataDir+"/signing.pub.pem")

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

//go:build integration || unit
// +build integration unit

package config

import (
	"fmt"
	"math/big"
	"os"
	"path"
	"time"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/mock"
)

func (*Bootstrapper) TestBootstrap(context map[string]interface{}) error {
	if _, ok := context[bootstrap.BootstrappedConfig]; ok {
		return nil
	}
	// To get the config location, we need to traverse the path to find the `go-centrifuge` folder
	gp := os.Getenv("BASE_PATH")
	projDir := path.Join(gp, "centrifuge", "go-centrifuge")
	context[bootstrap.BootstrappedConfig] = LoadConfiguration(fmt.Sprintf("%s/build/configs/testing_config.yaml", projDir))
	return nil
}

func (b *Bootstrapper) TestTearDown() error {
	return nil
}

type MockConfig struct {
	Configuration
	mock.Mock
}

func (m *MockConfig) GetStoragePath() string {
	args := m.Called()
	return args.Get(0).(string)
}

func (m *MockConfig) GetP2PPort() int {
	args := m.Called()
	return args.Get(0).(int)
}

func (m *MockConfig) GetP2PExternalIP() string {
	args := m.Called()
	return args.Get(0).(string)
}

func (m *MockConfig) GetP2PConnectionTimeout() time.Duration {
	args := m.Called()
	return args.Get(0).(time.Duration)
}

func (m *MockConfig) GetP2PResponseDelay() time.Duration {
	args := m.Called()
	return args.Get(0).(time.Duration)
}

func (m *MockConfig) GetReceiveEventNotificationEndpoint() string {
	args := m.Called()
	return args.Get(0).(string)
}

func (m *MockConfig) GetServerPort() int {
	args := m.Called()
	return args.Get(0).(int)
}

func (m *MockConfig) GetServerAddress() string {
	args := m.Called()
	return args.Get(0).(string)
}

func (m *MockConfig) GetNumWorkers() int {
	args := m.Called()
	return args.Get(0).(int)
}

func (m *MockConfig) GetWorkerWaitTimeMS() int {
	args := m.Called()
	return args.Get(0).(int)
}

func (m *MockConfig) GetEthereumNodeURL() string {
	args := m.Called()
	return args.Get(0).(string)
}

func (m *MockConfig) GetEthereumContextReadWaitTimeout() time.Duration {
	args := m.Called()
	return args.Get(0).(time.Duration)
}

func (m *MockConfig) GetEthereumContextWaitTimeout() time.Duration {
	args := m.Called()
	return args.Get(0).(time.Duration)
}

func (m *MockConfig) GetEthereumIntervalRetry() time.Duration {
	args := m.Called()
	return args.Get(0).(time.Duration)
}

func (m *MockConfig) GetEthereumMaxRetries() int {
	args := m.Called()
	return args.Get(0).(int)
}

func (m *MockConfig) GetEthereumMaxGasPrice() *big.Int {
	args := m.Called()
	return args.Get(0).(*big.Int)
}

func (m *MockConfig) GetEthereumGasMultiplier() float64 {
	args := m.Called()
	return args.Get(0).(float64)
}

func (m *MockConfig) GetEthereumGasLimit(op ContractOp) uint64 {
	args := m.Called(op)
	return args.Get(0).(uint64)
}

func (m *MockConfig) GetEthereumDefaultAccountName() string {
	args := m.Called()
	return args.Get(0).(string)
}

func (m *MockConfig) GetEthereumAccount(accountName string) (account *AccountConfig, err error) {
	args := m.Called(accountName)
	return args.Get(0).(*AccountConfig), args.Error(1)
}

func (m *MockConfig) GetNetworkString() string {
	args := m.Called()
	return args.Get(0).(string)
}

func (m *MockConfig) GetNetworkKey(k string) string {
	args := m.Called(k)
	return args.Get(0).(string)
}

func (m *MockConfig) GetContractAddressString(address string) string {
	args := m.Called(address)
	return args.Get(0).(string)
}

func (m *MockConfig) GetContractAddress(contractName ContractName) common.Address {
	args := m.Called(contractName)
	return args.Get(0).(common.Address)
}

func (m *MockConfig) GetBootstrapPeers() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *MockConfig) GetNetworkID() uint32 {
	args := m.Called()
	return args.Get(0).(uint32)
}

func (m *MockConfig) GetIdentityID() ([]byte, error) {
	args := m.Called()
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockConfig) GetP2PKeyPair() (pub, priv string) {
	args := m.Called()
	return args.Get(0).(string), args.Get(1).(string)
}

func (m *MockConfig) GetSigningKeyPair() (pub, priv string) {
	args := m.Called()
	return args.Get(0).(string), args.Get(1).(string)
}

func (m *MockConfig) GetPrecommitEnabled() bool {
	args := m.Called()
	return args.Get(0).(bool)
}

func (m *MockConfig) IsDebugLogEnabled() bool {
	args := m.Called()
	return args.Get(0).(bool)
}

func (m *MockConfig) GetCentChainAccount() (CentChainAccount, error) {
	args := m.Called()
	return args.Get(0).(CentChainAccount), args.Error(1)
}

func (m *MockConfig) GetCentChainIntervalRetry() time.Duration {
	args := m.Called()
	return args.Get(0).(time.Duration)
}

func (m *MockConfig) GetCentChainMaxRetries() int {
	args := m.Called()
	return args.Get(0).(int)
}

func (m *MockConfig) GetCentChainNodeURL() string {
	args := m.Called()
	return args.Get(0).(string)
}

type MockAccount struct {
	Account
	mock.Mock
}

func (m *MockAccount) GetReceiveEventNotificationEndpoint() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockAccount) SignMsg(msg []byte) (*coredocumentpb.Signature, error) {
	args := m.Called(msg)
	sig, _ := args.Get(0).(*coredocumentpb.Signature)
	return sig, args.Error(1)
}

func (m *MockAccount) GetCentChainAccount() CentChainAccount {
	args := m.Called()
	acc, _ := args.Get(0).(CentChainAccount)
	return acc
}

type MockService struct {
	mock.Mock
	Service
}

func (m *MockService) GenerateAccount(cacc CentChainAccount) (Account, error) {
	args := m.Called(cacc)
	acc, _ := args.Get(0).(Account)
	return acc, args.Error(1)
}

func (m *MockService) GetConfig() (Configuration, error) {
	args := m.Called()
	return args.Get(0).(Configuration), args.Error(1)
}

func (m *MockService) GetAccount(identifier []byte) (Account, error) {
	args := m.Called(identifier)
	acc, _ := args.Get(0).(Account)
	return acc, args.Error(1)
}

func (m *MockService) GetAccounts() ([]Account, error) {
	args := m.Called()
	v, _ := args.Get(0).([]Account)
	return v, args.Error(1)
}

func (m *MockService) CreateConfig(data Configuration) (Configuration, error) {
	args := m.Called(data)
	return args.Get(0).(Configuration), args.Error(0)
}

func (m *MockService) CreateAccount(data Account) (Account, error) {
	args := m.Called(data)
	acc, _ := args.Get(0).(Account)
	return acc, args.Error(1)
}

func (m *MockService) UpdateAccount(data Account) (Account, error) {
	args := m.Called(data)
	acc, _ := args.Get(0).(Account)
	return acc, args.Error(1)
}

func (m *MockService) DeleteAccount(identifier []byte) error {
	args := m.Called(identifier)
	return args.Error(0)
}

func (m *MockService) Sign(accountID, payload []byte) (*coredocumentpb.Signature, error) {
	args := m.Called(accountID, payload)
	sig, _ := args.Get(0).(*coredocumentpb.Signature)
	return sig, args.Error(1)
}

func (m *MockService) GenerateAccountAsync(cacc CentChainAccount) (did []byte, jobID []byte, err error) {
	args := m.Called(cacc)
	did, _ = args.Get(0).([]byte)
	jobID, _ = args.Get(1).([]byte)
	return did, jobID, args.Error(2)
}

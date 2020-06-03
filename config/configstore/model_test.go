// +build unit

package configstore

import (
	"math/big"
	"reflect"
	"testing"
	"time"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockConfig struct {
	config.Configuration
	mock.Mock
}

func (m *mockConfig) IsDebugLogEnabled() bool {
	args := m.Called()
	return args.Get(0).(bool)
}

func (m *mockConfig) GetLowEntropyNFTTokenEnabled() bool {
	args := m.Called()
	return args.Get(0).(bool)
}

func (m *mockConfig) GetPrecommitEnabled() bool {
	args := m.Called()
	return args.Get(0).(bool)
}

func (m *mockConfig) Type() reflect.Type {
	args := m.Called()
	return args.Get(0).(reflect.Type)
}

func (m *mockConfig) JSON() ([]byte, error) {
	args := m.Called()
	return args.Get(0).([]byte), args.Error(0)
}

func (m *mockConfig) FromJSON(json []byte) error {
	args := m.Called(json)
	return args.Error(0)
}

func (m *mockConfig) IsSet(key string) bool {
	args := m.Called(key)
	return args.Get(0).(bool)
}

func (m *mockConfig) Set(key string, value interface{}) {
	m.Called(key, value)
}

func (m *mockConfig) SetDefault(key string, value interface{}) {
	m.Called(key, value)
}

func (m *mockConfig) SetupSmartContractAddresses(network string, smartContractAddresses *config.SmartContractAddresses) {
	m.Called(network, smartContractAddresses)
}

func (m *mockConfig) Get(key string) interface{} {
	args := m.Called(key)
	return args.Get(0)
}

func (m *mockConfig) GetString(key string) string {
	args := m.Called(key)
	return args.Get(0).(string)
}

func (m *mockConfig) GetBool(key string) bool {
	args := m.Called(key)
	return args.Get(0).(bool)
}

func (m *mockConfig) GetInt(key string) int {
	args := m.Called(key)
	return args.Get(0).(int)
}

func (m *mockConfig) GetDuration(key string) time.Duration {
	args := m.Called(key)
	return args.Get(0).(time.Duration)
}

func (m *mockConfig) IsPProfEnabled() bool {
	args := m.Called()
	return args.Get(0).(bool)
}

func (m *mockConfig) GetStoragePath() string {
	args := m.Called()
	return args.Get(0).(string)
}

func (m *mockConfig) GetConfigStoragePath() string {
	args := m.Called()
	return args.Get(0).(string)
}

func (m *mockConfig) GetAccountsKeystore() string {
	args := m.Called()
	return args.Get(0).(string)
}

func (m *mockConfig) GetP2PPort() int {
	args := m.Called()
	return args.Get(0).(int)
}

func (m *mockConfig) GetP2PExternalIP() string {
	args := m.Called()
	return args.Get(0).(string)
}

func (m *mockConfig) GetP2PConnectionTimeout() time.Duration {
	args := m.Called()
	return args.Get(0).(time.Duration)
}

func (m *mockConfig) GetP2PResponseDelay() time.Duration {
	args := m.Called()
	return args.Get(0).(time.Duration)
}

func (m *mockConfig) GetReceiveEventNotificationEndpoint() string {
	args := m.Called()
	return args.Get(0).(string)
}

func (m *mockConfig) GetServerPort() int {
	args := m.Called()
	return args.Get(0).(int)
}

func (m *mockConfig) GetServerAddress() string {
	args := m.Called()
	return args.Get(0).(string)
}

func (m *mockConfig) GetNumWorkers() int {
	args := m.Called()
	return args.Get(0).(int)
}

func (m *mockConfig) GetWorkerWaitTimeMS() int {
	args := m.Called()
	return args.Get(0).(int)
}

func (m *mockConfig) GetEthereumNodeURL() string {
	args := m.Called()
	return args.Get(0).(string)
}

func (m *mockConfig) GetEthereumContextReadWaitTimeout() time.Duration {
	args := m.Called()
	return args.Get(0).(time.Duration)
}

func (m *mockConfig) GetEthereumContextWaitTimeout() time.Duration {
	args := m.Called()
	return args.Get(0).(time.Duration)
}

func (m *mockConfig) GetEthereumIntervalRetry() time.Duration {
	args := m.Called()
	return args.Get(0).(time.Duration)
}

func (m *mockConfig) GetEthereumMaxRetries() int {
	args := m.Called()
	return args.Get(0).(int)
}

func (m *mockConfig) GetEthereumMaxGasPrice() *big.Int {
	args := m.Called()
	return args.Get(0).(*big.Int)
}

func (m *mockConfig) GetEthereumGasLimit(op config.ContractOp) uint64 {
	args := m.Called(op)
	return args.Get(0).(uint64)
}

func (m *mockConfig) GetEthereumDefaultAccountName() string {
	args := m.Called()
	return args.Get(0).(string)
}

func (m *mockConfig) GetEthereumAccount(accountName string) (account *config.AccountConfig, err error) {
	args := m.Called(accountName)
	return args.Get(0).(*config.AccountConfig), args.Error(1)
}

func (m *mockConfig) GetNetworkString() string {
	args := m.Called()
	return args.Get(0).(string)
}

func (m *mockConfig) GetNetworkKey(k string) string {
	args := m.Called()
	return args.Get(0).(string)
}

func (m *mockConfig) GetContractAddressString(address string) string {
	args := m.Called()
	return args.Get(0).(string)
}

func (m *mockConfig) GetContractAddress(contractName config.ContractName) common.Address {
	args := m.Called()
	return args.Get(0).(common.Address)
}

func (m *mockConfig) GetBootstrapPeers() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *mockConfig) GetNetworkID() uint32 {
	args := m.Called()
	return args.Get(0).(uint32)
}

func (m *mockConfig) GetIdentityID() ([]byte, error) {
	args := m.Called()
	return args.Get(0).([]byte), args.Error(1)
}

func (m *mockConfig) GetP2PKeyPair() (pub, priv string) {
	args := m.Called()
	return args.Get(0).(string), args.Get(1).(string)
}

func (m *mockConfig) GetSigningKeyPair() (pub, priv string) {
	args := m.Called()
	return args.Get(0).(string), args.Get(1).(string)
}

func (m *mockConfig) GetCentChainAccount() (config.CentChainAccount, error) {
	args := m.Called()
	return args.Get(0).(config.CentChainAccount), args.Error(1)
}

func (m *mockConfig) GetCentChainIntervalRetry() time.Duration {
	args := m.Called()
	return args.Get(0).(time.Duration)
}

func (m *mockConfig) GetCentChainMaxRetries() int {
	args := m.Called()
	return args.Get(0).(int)
}

func (m *mockConfig) GetCentChainAnchorLifespan() time.Duration {
	args := m.Called()
	return args.Get(0).(time.Duration)
}

func (m *mockConfig) GetCentChainNodeURL() string {
	args := m.Called()
	return args.Get(0).(string)
}

func (m *mockConfig) GetTaskValidDuration() time.Duration {
	args := m.Called()
	return args.Get(0).(time.Duration)
}

func TestNewNodeConfig(t *testing.T) {
	c := createMockConfig()
	NewNodeConfig(c)

	c.AssertExpectations(t)
}

func TestNewAccountConfig(t *testing.T) {
	c := &mockConfig{}
	c.On("GetEthereumAccount", "name").Return(&config.AccountConfig{}, nil).Once()
	c.On("GetEthereumDefaultAccountName").Return("dummyAcc").Once()
	c.On("GetReceiveEventNotificationEndpoint").Return("dummyNotifier").Once()
	c.On("GetIdentityID").Return(utils.RandomSlice(identity.DIDLength), nil).Once()
	c.On("GetP2PKeyPair").Return("pub", "priv").Once()
	c.On("GetSigningKeyPair").Return("pub", "priv").Once()
	c.On("GetEthereumContextWaitTimeout").Return(time.Second).Once()
	c.On("GetPrecommitEnabled").Return(true).Once()
	c.On("GetCentChainAccount").Return(config.CentChainAccount{}, nil).Once()
	_, err := NewAccount("name", c)
	assert.NoError(t, err)
	c.AssertExpectations(t)
}

func createMockConfig() *mockConfig {
	c := &mockConfig{}
	c.On("GetStoragePath").Return("dummyStorage").Once()
	c.On("GetAccountsKeystore").Return("dummyKeyStorage").Once()
	c.On("GetP2PPort").Return(30000).Once()
	c.On("GetP2PExternalIP").Return("ip").Once()
	c.On("GetP2PConnectionTimeout").Return(time.Second).Once()
	c.On("GetP2PResponseDelay").Return(time.Millisecond).Once()
	c.On("GetServerPort").Return(8080).Once()
	c.On("GetServerAddress").Return("dummyServer").Once()
	c.On("GetNumWorkers").Return(2).Once()
	c.On("GetWorkerWaitTimeMS").Return(1).Once()
	c.On("GetEthereumNodeURL").Return("dummyNode").Once()
	c.On("GetIdentityID").Return(utils.RandomSlice(identity.DIDLength), nil).Once()
	c.On("GetP2PKeyPair").Return("pub", "priv").Once()
	c.On("GetSigningKeyPair").Return("pub", "priv").Once()
	c.On("GetReceiveEventNotificationEndpoint").Return("dummyNotifier").Once()
	c.On("GetEthereumAccount", "dummyAcc").Return(&config.AccountConfig{}, nil).Once()
	c.On("GetEthereumDefaultAccountName").Return("dummyAcc").Twice()
	c.On("GetEthereumContextReadWaitTimeout").Return(time.Second).Once()
	c.On("GetEthereumContextWaitTimeout").Return(time.Second).Once()
	c.On("GetEthereumIntervalRetry").Return(time.Second).Once()
	c.On("GetEthereumMaxRetries").Return(1).Once()
	c.On("GetEthereumMaxGasPrice").Return(big.NewInt(1)).Once()
	c.On("GetEthereumGasLimit", mock.Anything).Return(uint64(100))
	c.On("GetNetworkString").Return("somehill").Once()
	c.On("GetBootstrapPeers").Return([]string{"p1", "p2"}).Once()
	c.On("GetNetworkID").Return(uint32(1)).Once()
	c.On("GetContractAddress", mock.Anything).Return(common.Address{})
	c.On("IsPProfEnabled", mock.Anything).Return(true)
	c.On("IsDebugLogEnabled", mock.Anything).Return(true)
	c.On("GetLowEntropyNFTTokenEnabled", mock.Anything).Return(true)
	c.On("GetCentChainAccount").Return(config.CentChainAccount{}, nil).Once()
	c.On("GetCentChainIntervalRetry").Return(time.Second).Once()
	c.On("GetCentChainAnchorLifespan").Return(time.Second).Once()
	c.On("GetCentChainMaxRetries").Return(1).Once()
	c.On("GetCentChainNodeURL").Return("dummyNode").Once()
	c.On("GetTaskValidDuration").Return(time.Minute).Once()
	return c
}

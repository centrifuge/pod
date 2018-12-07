package config

import (
	"testing"

	"math/big"
	"time"

	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/mock"
)

type MockConfig struct {
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

func (m *MockConfig) GetEthereumGasPrice() *big.Int {
	args := m.Called()
	return args.Get(0).(*big.Int)
}

func (m *MockConfig) GetEthereumGasLimit() uint64 {
	args := m.Called()
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

func (m *MockConfig) GetTxPoolAccessEnabled() bool {
	args := m.Called()
	return args.Get(0).(bool)
}

func (m *MockConfig) GetNetworkString() string {
	args := m.Called()
	return args.Get(0).(string)
}

func (m *MockConfig) GetNetworkKey(k string) string {
	args := m.Called()
	return args.Get(0).(string)
}

func (m *MockConfig) GetContractAddressString(address string) string {
	args := m.Called()
	return args.Get(0).(string)
}

func (m *MockConfig) GetContractAddress(address string) common.Address {
	args := m.Called()
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

func (m *MockConfig) GetSigningKeyPair() (pub, priv string) {
	args := m.Called()
	return args.Get(0).(string), args.Get(1).(string)
}

func (m *MockConfig) GetEthAuthKeyPair() (pub, priv string) {
	args := m.Called()
	return args.Get(0).(string), args.Get(1).(string)
}

func TestNewNodeConfig(t *testing.T) {
	c := &MockConfig{}
	c.On("GetStoragePath").Return("dummyStorage").Once()
	c.On("GetP2PPort").Return(30000).Once()
	c.On("GetP2PExternalIP").Return("ip").Once()
	c.On("GetP2PConnectionTimeout").Return(time.Second).Once()
	c.On("GetReceiveEventNotificationEndpoint").Return("dummyNotifier").Once()

	c.On("GetServerPort").Return(8080).Once()
	c.On("GetServerAddress").Return("dummyServer").Once()
	c.On("GetNumWorkers").Return(2).Once()
	c.On("GetWorkerWaitTimeMS").Return(1).Once()
	c.On("GetEthereumNodeURL").Return("dummyNode").Once()

	c.On("GetEthereumContextReadWaitTimeout").Return(time.Second).Once()
	c.On("GetEthereumContextWaitTimeout").Return(time.Second).Once()
	c.On("GetEthereumIntervalRetry").Return(time.Second).Once()
	c.On("GetEthereumMaxRetries").Return(1).Once()
	c.On("GetEthereumGasPrice").Return(big.NewInt(1)).Once()

	c.On("GetEthereumGasLimit").Return(uint64(100)).Once()
	c.On("GetEthereumDefaultAccountName").Return("dummyAcc").Once()
	c.On("GetTxPoolAccessEnabled").Return(true).Once()
	c.On("GetNetworkString").Return("somehill").Once()
	c.On("GetBootstrapPeers").Return([]string{"p1", "p2"}).Once()

	c.On("GetNetworkID").Return(uint32(1)).Once()
	NewNodeConfig(c)

	c.AssertExpectations(t)
}

func TestNewTenantConfig(t *testing.T) {
	c := &MockConfig{}
	c.On("GetEthereumAccount", "name").Return(&AccountConfig{}, nil).Once()
	c.On("GetIdentityID").Return(utils.RandomSlice(6), nil).Once()
	c.On("GetSigningKeyPair").Return("pub", "priv").Once()
	c.On("GetEthAuthKeyPair").Return("pub", "priv").Once()
	NewTenantConfig("name", c)
	c.AssertExpectations(t)
}

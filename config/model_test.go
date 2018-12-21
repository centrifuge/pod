// +build unit

package config

import (
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"

	"math/big"
	"time"

	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockConfig struct {
	Configuration
	mock.Mock
}

func (m *mockConfig) GetStoragePath() string {
	args := m.Called()
	return args.Get(0).(string)
}

func (m *mockConfig) GetConfigStoragePath() string {
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

func (m *mockConfig) GetEthereumGasPrice() *big.Int {
	args := m.Called()
	return args.Get(0).(*big.Int)
}

func (m *mockConfig) GetEthereumGasLimit() uint64 {
	args := m.Called()
	return args.Get(0).(uint64)
}

func (m *mockConfig) GetEthereumDefaultAccountName() string {
	args := m.Called()
	return args.Get(0).(string)
}

func (m *mockConfig) GetEthereumAccount(accountName string) (account *AccountConfig, err error) {
	args := m.Called(accountName)
	return args.Get(0).(*AccountConfig), args.Error(1)
}

func (m *mockConfig) GetTxPoolAccessEnabled() bool {
	args := m.Called()
	return args.Get(0).(bool)
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

func (m *mockConfig) GetContractAddress(address string) common.Address {
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

func (m *mockConfig) GetSigningKeyPair() (pub, priv string) {
	args := m.Called()
	return args.Get(0).(string), args.Get(1).(string)
}

func (m *mockConfig) GetEthAuthKeyPair() (pub, priv string) {
	args := m.Called()
	return args.Get(0).(string), args.Get(1).(string)
}

func TestNewNodeConfig(t *testing.T) {
	c := &mockConfig{}
	c.On("GetStoragePath").Return("dummyStorage").Once()
	c.On("GetP2PPort").Return(30000).Once()
	c.On("GetP2PExternalIP").Return("ip").Once()
	c.On("GetP2PConnectionTimeout").Return(time.Second).Once()

	c.On("GetServerPort").Return(8080).Once()
	c.On("GetServerAddress").Return("dummyServer").Once()
	c.On("GetNumWorkers").Return(2).Once()
	c.On("GetWorkerWaitTimeMS").Return(1).Once()
	c.On("GetEthereumNodeURL").Return("dummyNode").Once()

	c.On("GetIdentityID").Return(utils.RandomSlice(6), nil).Once()
	c.On("GetSigningKeyPair").Return("pub", "priv").Once()
	c.On("GetEthAuthKeyPair").Return("pub", "priv").Once()
	c.On("GetReceiveEventNotificationEndpoint").Return("dummyNotifier").Once()
	c.On("GetEthereumAccount", "dummyAcc").Return(&AccountConfig{}, nil).Once()
	c.On("GetEthereumDefaultAccountName").Return("dummyAcc").Twice()

	c.On("GetEthereumContextReadWaitTimeout").Return(time.Second).Once()
	c.On("GetEthereumContextWaitTimeout").Return(time.Second).Once()
	c.On("GetEthereumIntervalRetry").Return(time.Second).Once()
	c.On("GetEthereumMaxRetries").Return(1).Once()
	c.On("GetEthereumGasPrice").Return(big.NewInt(1)).Once()

	c.On("GetEthereumGasLimit").Return(uint64(100)).Once()
	c.On("GetTxPoolAccessEnabled").Return(true).Once()
	c.On("GetNetworkString").Return("somehill").Once()
	c.On("GetBootstrapPeers").Return([]string{"p1", "p2"}).Once()

	c.On("GetNetworkID").Return(uint32(1)).Once()
	NewNodeConfig(c)

	c.AssertExpectations(t)
}

func TestNewTenantConfig(t *testing.T) {
	c := &mockConfig{}
	c.On("GetEthereumAccount", "name").Return(&AccountConfig{}, nil).Once()
	c.On("GetEthereumDefaultAccountName").Return("dummyAcc").Once()
	c.On("GetReceiveEventNotificationEndpoint").Return("dummyNotifier").Once()
	c.On("GetIdentityID").Return(utils.RandomSlice(6), nil).Once()
	c.On("GetSigningKeyPair").Return("pub", "priv").Once()
	c.On("GetEthAuthKeyPair").Return("pub", "priv").Once()
	NewTenantConfig("name", c)
	c.AssertExpectations(t)
}

func TestNodeConfigProtobuf(t *testing.T) {
	c := &mockConfig{}
	c.On("GetStoragePath").Return("dummyStorage").Once()
	c.On("GetP2PPort").Return(30000).Once()
	c.On("GetP2PExternalIP").Return("ip").Once()
	c.On("GetP2PConnectionTimeout").Return(time.Second).Once()

	c.On("GetServerPort").Return(8080).Once()
	c.On("GetServerAddress").Return("dummyServer").Once()
	c.On("GetNumWorkers").Return(2).Once()
	c.On("GetWorkerWaitTimeMS").Return(1).Once()
	c.On("GetEthereumNodeURL").Return("dummyNode").Once()

	c.On("GetIdentityID").Return(utils.RandomSlice(6), nil).Once()
	c.On("GetSigningKeyPair").Return("pub", "priv").Once()
	c.On("GetEthAuthKeyPair").Return("pub", "priv").Once()
	c.On("GetReceiveEventNotificationEndpoint").Return("dummyNotifier").Once()
	c.On("GetEthereumAccount", "dummyAcc").Return(&AccountConfig{}, nil).Once()
	c.On("GetEthereumDefaultAccountName").Return("dummyAcc").Twice()

	c.On("GetEthereumContextReadWaitTimeout").Return(time.Second).Once()
	c.On("GetEthereumContextWaitTimeout").Return(time.Second).Once()
	c.On("GetEthereumIntervalRetry").Return(time.Second).Once()
	c.On("GetEthereumMaxRetries").Return(1).Once()
	c.On("GetEthereumGasPrice").Return(big.NewInt(1)).Once()

	c.On("GetEthereumGasLimit").Return(uint64(100)).Once()
	c.On("GetTxPoolAccessEnabled").Return(true).Once()
	c.On("GetNetworkString").Return("somehill").Once()
	c.On("GetBootstrapPeers").Return([]string{"p1", "p2"}).Once()

	c.On("GetNetworkID").Return(uint32(1)).Once()
	nc := NewNodeConfig(c)
	c.AssertExpectations(t)

	ncpb := nc.createProtobuf()
	assert.Equal(t, nc.StoragePath, ncpb.StoragePath)
	assert.Equal(t, nc.ServerPort, int(ncpb.ServerPort))
	assert.Equal(t, hexutil.Encode(nc.MainIdentity.IdentityID), ncpb.MainIdentity.IdentityId)

	ncCopy := new(NodeConfig)
	ncCopy.loadFromProtobuf(ncpb)
	assert.Equal(t, ncpb.StoragePath, ncCopy.StoragePath)
	assert.Equal(t, int(ncpb.ServerPort), ncCopy.ServerPort)
	assert.Equal(t, ncpb.MainIdentity.IdentityId, hexutil.Encode(ncCopy.MainIdentity.IdentityID))
}

func TestTenantConfigProtobuf(t *testing.T) {
	c := &mockConfig{}
	c.On("GetEthereumAccount", "name").Return(&AccountConfig{}, nil).Once()
	c.On("GetEthereumDefaultAccountName").Return("dummyAcc").Once()
	c.On("GetReceiveEventNotificationEndpoint").Return("dummyNotifier").Once()
	c.On("GetIdentityID").Return(utils.RandomSlice(6), nil).Once()
	c.On("GetSigningKeyPair").Return("pub", "priv").Once()
	c.On("GetEthAuthKeyPair").Return("pub", "priv").Once()
	tc, err := NewTenantConfig("name", c)
	assert.Nil(t, err)
	c.AssertExpectations(t)

	tcpb := tc.createProtobuf()
	assert.Equal(t, tc.ReceiveEventNotificationEndpoint, tcpb.ReceiveEventNotificationEndpoint)
	assert.Equal(t, hexutil.Encode(tc.IdentityID), tcpb.IdentityId)
	assert.Equal(t, tc.SigningKeyPair.Priv, tcpb.SigningKeyPair.Pvt)

	tcCopy := new(TenantConfig)
	tcCopy.loadFromProtobuf(tcpb)
	assert.Equal(t, tcpb.ReceiveEventNotificationEndpoint, tcCopy.ReceiveEventNotificationEndpoint)
	assert.Equal(t, tcpb.IdentityId, hexutil.Encode(tcCopy.IdentityID))
	assert.Equal(t, tcpb.SigningKeyPair.Pvt, tcCopy.SigningKeyPair.Priv)
}

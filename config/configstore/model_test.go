// +build unit

package configstore

import (
	"math/big"
	"reflect"
	"testing"
	"time"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/account"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockConfig struct {
	mock.Mock
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

// GetTaskRetries returns the number of retries allowed for a queued task
func (m *mockConfig) GetTaskRetries() int {
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
	_, err := NewAccount("name", c)
	assert.NoError(t, err)
	c.AssertExpectations(t)
}

func TestAccountProtobuf_validationFailures(t *testing.T) {
	c := &mockConfig{}
	c.On("GetEthereumAccount", "name").Return(&config.AccountConfig{}, nil)
	c.On("GetEthereumDefaultAccountName").Return("dummyAcc")
	c.On("GetReceiveEventNotificationEndpoint").Return("dummyNotifier")
	c.On("GetIdentityID").Return(utils.RandomSlice(identity.DIDLength), nil)
	c.On("GetP2PKeyPair").Return("pub", "priv")
	c.On("GetSigningKeyPair").Return("pub", "priv")
	c.On("GetEthereumContextWaitTimeout").Return(time.Second)
	c.On("GetPrecommitEnabled").Return(true)
	tc, err := NewAccount("name", c)
	assert.Nil(t, err)
	c.AssertExpectations(t)

	// Nil EthAccount
	tco := tc.(*Account)
	tco.EthereumAccount = nil
	accpb, err := tco.CreateProtobuf()
	assert.Error(t, err)
	assert.Nil(t, accpb)

	// Nil payload
	tc, err = NewAccount("name", c)
	assert.Nil(t, err)
	accpb, err = tc.CreateProtobuf()
	assert.NoError(t, err)
	tco = tc.(*Account)
	err = tco.loadFromProtobuf(nil)
	assert.Error(t, err)

	// Nil EthAccount
	ethacc := proto.Clone(accpb.EthAccount)
	accpb.EthAccount = nil
	err = tco.loadFromProtobuf(accpb)
	assert.Error(t, err)
	accpb.EthAccount = ethacc.(*accountpb.EthereumAccount)

	// Nil P2PKeyPair
	p2pKey := proto.Clone(accpb.P2PKeyPair)
	accpb.P2PKeyPair = nil
	err = tco.loadFromProtobuf(accpb)
	assert.Error(t, err)
	accpb.P2PKeyPair = p2pKey.(*accountpb.KeyPair)

	// Nil SigningKeyPair
	signKey := proto.Clone(accpb.SigningKeyPair)
	accpb.SigningKeyPair = nil
	err = tco.loadFromProtobuf(accpb)
	assert.Error(t, err)
	accpb.SigningKeyPair = signKey.(*accountpb.KeyPair)

}

func TestAccountConfigProtobuf(t *testing.T) {
	c := &mockConfig{}
	c.On("GetEthereumAccount", "name").Return(&config.AccountConfig{}, nil).Once()
	c.On("GetEthereumDefaultAccountName").Return("dummyAcc").Once()
	c.On("GetReceiveEventNotificationEndpoint").Return("dummyNotifier").Once()
	c.On("GetIdentityID").Return(utils.RandomSlice(identity.DIDLength), nil).Once()
	c.On("GetP2PKeyPair").Return("pub", "priv").Once()
	c.On("GetSigningKeyPair").Return("pub", "priv").Once()
	c.On("GetEthereumContextWaitTimeout").Return(time.Second).Once()
	c.On("GetPrecommitEnabled").Return(true).Once()
	tc, err := NewAccount("name", c)
	assert.Nil(t, err)
	c.AssertExpectations(t)

	accpb, err := tc.CreateProtobuf()
	assert.NoError(t, err)
	assert.Equal(t, tc.GetReceiveEventNotificationEndpoint(), accpb.ReceiveEventNotificationEndpoint)
	i, err := tc.GetIdentityID()
	assert.Nil(t, err)

	assert.Equal(t, common.BytesToAddress(i).Hex(), common.HexToAddress(accpb.IdentityId).Hex())
	_, priv := tc.GetSigningKeyPair()
	assert.Equal(t, priv, accpb.SigningKeyPair.Pvt)

	tcCopy := new(Account)
	err = tcCopy.loadFromProtobuf(accpb)
	assert.NoError(t, err)
	assert.Equal(t, accpb.ReceiveEventNotificationEndpoint, tcCopy.ReceiveEventNotificationEndpoint)
	assert.Equal(t, common.HexToAddress(accpb.IdentityId).Hex(), common.BytesToAddress(tcCopy.IdentityID).Hex())
	assert.Equal(t, accpb.SigningKeyPair.Pvt, tcCopy.SigningKeyPair.Priv)
}

func createMockConfig() *mockConfig {
	c := &mockConfig{}
	c.On("GetStoragePath").Return("dummyStorage").Once()
	c.On("GetAccountsKeystore").Return("dummyKeyStorage").Once()
	c.On("GetP2PPort").Return(30000).Once()
	c.On("GetP2PExternalIP").Return("ip").Once()
	c.On("GetP2PConnectionTimeout").Return(time.Second).Once()
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
	c.On("GetTxPoolAccessEnabled").Return(true).Once()
	c.On("GetNetworkString").Return("somehill").Once()
	c.On("GetBootstrapPeers").Return([]string{"p1", "p2"}).Once()
	c.On("GetNetworkID").Return(uint32(1)).Once()
	c.On("GetContractAddress", mock.Anything).Return(common.Address{})
	c.On("IsPProfEnabled", mock.Anything).Return(true)
	return c
}

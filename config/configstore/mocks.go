//go:build unit || integration || testworld
// +build unit integration testworld

package configstore

import (
	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/stretchr/testify/mock"
)

// TODO(ved): this should be gone as well
type MockService struct {
	mock.Mock
	config.Service
}

func (m *MockService) GenerateAccount(cacc config.CentChainAccount) (config.Account, error) {
	args := m.Called(cacc)
	acc, _ := args.Get(0).(config.Account)
	return acc, args.Error(1)
}

func (m *MockService) GetConfig() (config.Configuration, error) {
	args := m.Called()
	return args.Get(0).(config.Configuration), args.Error(1)
}

func (m *MockService) GetAccount(identifier []byte) (config.Account, error) {
	args := m.Called(identifier)
	acc, _ := args.Get(0).(config.Account)
	return acc, args.Error(1)
}

func (m *MockService) GetAccounts() ([]config.Account, error) {
	args := m.Called()
	v, _ := args.Get(0).([]config.Account)
	return v, args.Error(1)
}

func (m *MockService) CreateConfig(data config.Configuration) (config.Configuration, error) {
	args := m.Called(data)
	return args.Get(0).(*config.NodeConfig), args.Error(0)
}

func (m *MockService) CreateAccount(data config.Account) (config.Account, error) {
	args := m.Called(data)
	acc, _ := args.Get(0).(*Account)
	return acc, args.Error(1)
}

func (m *MockService) UpdateAccount(data config.Account) (config.Account, error) {
	args := m.Called(data)
	acc, _ := args.Get(0).(*Account)
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

func (m *MockService) GenerateAccountAsync(cacc config.CentChainAccount) (did []byte, jobID []byte, err error) {
	args := m.Called(cacc)
	did, _ = args.Get(0).([]byte)
	jobID, _ = args.Get(1).([]byte)
	return did, jobID, args.Error(2)
}

func (b *Bootstrapper) TestBootstrap(context map[string]interface{}) error {
	return b.Bootstrap(context)
}

func (b *Bootstrapper) TestTearDown() error {
	return nil
}

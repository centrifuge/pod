//go:build unit

package configstore

import (
	"testing"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/commons"
	"github.com/centrifuge/go-centrifuge/utils"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/stretchr/testify/assert"
)

var (
	repoErr = errors.New("error")
)

func TestService_CreateConfig(t *testing.T) {
	repoMock := NewRepositoryMock(t)
	service := NewService(repoMock)

	cfg := &config.NodeConfig{}

	repoMock.On("GetConfig").
		Once().
		Return(nil, repoErr)

	repoMock.On("CreateConfig", cfg).
		Once().
		Return(nil)

	err := service.CreateConfig(cfg)
	assert.NoError(t, err)

	repoMock.On("GetConfig").
		Once().
		Return(nil, nil)

	repoMock.On("UpdateConfig", cfg).
		Once().
		Return(nil)

	err = service.CreateConfig(cfg)
	assert.NoError(t, err)
}

func TestService_CreateConfig_RepoErrors(t *testing.T) {
	repoMock := NewRepositoryMock(t)
	service := NewService(repoMock)

	cfg := &config.NodeConfig{}

	repoMock.On("GetConfig").
		Once().
		Return(nil, repoErr)

	repoMock.On("CreateConfig", cfg).
		Once().
		Return(repoErr)

	err := service.CreateConfig(cfg)
	assert.ErrorIs(t, err, repoErr)

	repoMock.On("GetConfig").
		Once().
		Return(nil, nil)

	repoMock.On("UpdateConfig", cfg).
		Once().
		Return(repoErr)

	err = service.CreateConfig(cfg)
	assert.ErrorIs(t, err, repoErr)
}

func TestService_CreateNodeAdmin(t *testing.T) {
	repoMock := NewRepositoryMock(t)
	service := NewService(repoMock)

	nodeAdmin := &NodeAdmin{}

	repoMock.On("CreateNodeAdmin", nodeAdmin).
		Once().
		Return(nil)

	err := service.CreateNodeAdmin(nodeAdmin)
	assert.NoError(t, err)

	repoMock.On("CreateNodeAdmin", nodeAdmin).
		Once().
		Return(repoErr)

	err = service.CreateNodeAdmin(nodeAdmin)
	assert.ErrorIs(t, err, repoErr)
}

func TestService_CreateAccount(t *testing.T) {
	repoMock := NewRepositoryMock(t)
	service := NewService(repoMock)

	account, err := getRandomAccount()
	assert.NoError(t, err)

	repoMock.On("CreateAccount", account).
		Once().
		Return(nil)

	err = service.CreateAccount(account)
	assert.NoError(t, err)

	repoMock.On("CreateAccount", account).
		Once().
		Return(repoErr)

	err = service.CreateAccount(account)
	assert.ErrorIs(t, err, repoErr)
}

func TestService_GetConfig(t *testing.T) {
	repoMock := NewRepositoryMock(t)
	service := NewService(repoMock)

	cfg := &config.NodeConfig{}

	repoMock.On("GetConfig").
		Once().
		Return(cfg, nil)

	res, err := service.GetConfig()
	assert.NoError(t, err)
	assert.Equal(t, cfg, res)

	repoMock.On("GetConfig").
		Once().
		Return(nil, repoErr)

	res, err = service.GetConfig()
	assert.ErrorIs(t, err, repoErr)
	assert.Nil(t, res)
}

func TestService_GetNodeAdmin(t *testing.T) {
	repoMock := NewRepositoryMock(t)
	service := NewService(repoMock)

	nodeAdmin := &NodeAdmin{}

	repoMock.On("GetNodeAdmin").
		Once().
		Return(nodeAdmin, nil)

	res, err := service.GetNodeAdmin()
	assert.NoError(t, err)
	assert.Equal(t, nodeAdmin, res)

	repoMock.On("GetNodeAdmin").
		Once().
		Return(nil, repoErr)

	res, err = service.GetNodeAdmin()
	assert.ErrorIs(t, err, repoErr)
	assert.Nil(t, res)
}

func TestService_GetAccount(t *testing.T) {
	repoMock := NewRepositoryMock(t)
	service := NewService(repoMock)

	account, err := getRandomAccount()
	assert.NoError(t, err)

	repoMock.On("GetAccount", account.GetIdentity().ToBytes()).
		Once().
		Return(account, nil)

	res, err := service.GetAccount(account.GetIdentity().ToBytes())
	assert.NoError(t, err)
	assert.Equal(t, account, res)

	repoMock.On("GetAccount", account.GetIdentity().ToBytes()).
		Once().
		Return(nil, repoErr)

	res, err = service.GetAccount(account.GetIdentity().ToBytes())
	assert.ErrorIs(t, err, repoErr)
	assert.Nil(t, res)
}

func TestService_GetAccounts(t *testing.T) {
	repoMock := NewRepositoryMock(t)
	service := NewService(repoMock)

	accounts, err := getRandomAccounts(3)
	assert.NoError(t, err)

	repoMock.On("GetAllAccounts").
		Once().
		Return(accounts, nil)

	res, err := service.GetAccounts()
	assert.NoError(t, err)
	assert.Equal(t, accounts, res)

	repoMock.On("GetAllAccounts").
		Once().
		Return(nil, repoErr)

	res, err = service.GetAccounts()
	assert.ErrorIs(t, err, repoErr)
	assert.Nil(t, res)
}

func TestService_UpdateNodeAdmin(t *testing.T) {
	repoMock := NewRepositoryMock(t)
	service := NewService(repoMock)

	nodeAdmin := &NodeAdmin{}

	repoMock.On("UpdateNodeAdmin", nodeAdmin).
		Once().
		Return(nil)

	err := service.UpdateNodeAdmin(nodeAdmin)
	assert.NoError(t, err)

	repoMock.On("UpdateNodeAdmin", nodeAdmin).
		Once().
		Return(repoErr)

	err = service.UpdateNodeAdmin(nodeAdmin)
	assert.ErrorIs(t, err, repoErr)
}

func TestService_UpdateAccount(t *testing.T) {
	repoMock := NewRepositoryMock(t)
	service := NewService(repoMock)

	account, err := getRandomAccount()
	assert.NoError(t, err)

	repoMock.On("UpdateAccount", account).
		Once().
		Return(nil)

	err = service.UpdateAccount(account)
	assert.NoError(t, err)

	repoMock.On("UpdateAccount", account).
		Once().
		Return(repoErr)

	err = service.UpdateAccount(account)
	assert.ErrorIs(t, err, repoErr)
}

func TestService_DeleteAccount(t *testing.T) {
	repoMock := NewRepositoryMock(t)
	service := NewService(repoMock)

	account, err := getRandomAccount()
	assert.NoError(t, err)

	repoMock.On("DeleteAccount", account.GetIdentity().ToBytes()).
		Once().
		Return(nil)

	err = service.DeleteAccount(account.GetIdentity().ToBytes())
	assert.NoError(t, err)

	repoMock.On("DeleteAccount", account.GetIdentity().ToBytes()).
		Once().
		Return(repoErr)

	err = service.DeleteAccount(account.GetIdentity().ToBytes())
	assert.ErrorIs(t, err, repoErr)
}

func TestService_Sign(t *testing.T) {
	repoMock := NewRepositoryMock(t)
	service := NewService(repoMock)

	payload := utils.RandomSlice(32)
	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)

	repoMock.On("GetAccount", accountID.ToBytes()).
		Once().
		Return(accountMock, nil)

	signature := &coredocumentpb.Signature{}

	accountMock.On("SignMsg", payload).
		Once().
		Return(signature, nil)

	res, err := service.Sign(accountID.ToBytes(), payload)
	assert.NoError(t, err)
	assert.Equal(t, signature, res)
}

func TestService_Sign_RepoError(t *testing.T) {
	repoMock := NewRepositoryMock(t)
	service := NewService(repoMock)

	payload := utils.RandomSlice(32)
	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	repoMock.On("GetAccount", accountID.ToBytes()).
		Once().
		Return(nil, repoErr)

	res, err := service.Sign(accountID.ToBytes(), payload)
	assert.ErrorIs(t, err, repoErr)
	assert.Nil(t, res)
}

func TestService_Sign_AccountSignError(t *testing.T) {
	repoMock := NewRepositoryMock(t)
	service := NewService(repoMock)

	payload := utils.RandomSlice(32)
	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)

	repoMock.On("GetAccount", accountID.ToBytes()).
		Once().
		Return(accountMock, nil)

	accountSignErr := errors.New("error")

	accountMock.On("SignMsg", payload).
		Once().
		Return(nil, accountSignErr)

	res, err := service.Sign(accountID.ToBytes(), payload)
	assert.ErrorIs(t, err, accountSignErr)
	assert.Nil(t, res)
}

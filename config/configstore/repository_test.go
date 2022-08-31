//go:build unit || integration

package configstore

import (
	"crypto/rand"
	"testing"

	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/storage"
	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/commons"
	"github.com/stretchr/testify/assert"
)

var (
	storageRepoErr = errors.New("error")
)

func TestRepository_Register(t *testing.T) {
	storageRepo := storage.NewRepositoryMock(t)

	repo := NewDBRepository(storageRepo)

	nodeAdmin := &NodeAdmin{}

	storageRepo.On("Register", nodeAdmin).Once()

	repo.RegisterNodeAdmin(nodeAdmin)

	acc := &Account{}

	storageRepo.On("Register", acc).Once()

	repo.RegisterAccount(acc)

	cfg := &NodeConfig{}

	storageRepo.On("Register", cfg).Once()

	repo.RegisterConfig(cfg)
}

func TestRepository_GetNodeAdmin(t *testing.T) {
	storageRepo := storage.NewRepositoryMock(t)

	repo := NewDBRepository(storageRepo)

	nodeAdmin := &NodeAdmin{}

	storageRepo.On("Get", getNodeAdminKey()).
		Once().
		Return(nodeAdmin, nil)

	res, err := repo.GetNodeAdmin()
	assert.NoError(t, err)
	assert.Equal(t, nodeAdmin, res)
}

func TestRepository_GetNodeAdmin_Error(t *testing.T) {
	storageRepo := storage.NewRepositoryMock(t)

	repo := NewDBRepository(storageRepo)

	storageRepo.On("Get", getNodeAdminKey()).
		Once().
		Return(nil, storageRepoErr)

	res, err := repo.GetNodeAdmin()
	assert.ErrorIs(t, err, storageRepoErr)
	assert.Nil(t, res)
}

func TestRepository_GetAccount(t *testing.T) {
	storageRepo := storage.NewRepositoryMock(t)

	repo := NewDBRepository(storageRepo)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	account := &Account{}

	accountKey := getAccountKey(accountID.ToBytes())

	storageRepo.On("Get", accountKey).
		Once().
		Return(account, nil)

	res, err := repo.GetAccount(accountID.ToBytes())
	assert.NoError(t, err)
	assert.Equal(t, account, res)
}

func TestRepository_GetAccount_Error(t *testing.T) {
	storageRepo := storage.NewRepositoryMock(t)

	repo := NewDBRepository(storageRepo)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountKey := getAccountKey(accountID.ToBytes())

	storageRepo.On("Get", accountKey).
		Once().
		Return(nil, storageRepoErr)

	res, err := repo.GetAccount(accountID.ToBytes())
	assert.ErrorIs(t, err, storageRepoErr)
	assert.Nil(t, res)
}

func TestRepository_GetConfig(t *testing.T) {
	storageRepo := storage.NewRepositoryMock(t)

	repo := NewDBRepository(storageRepo)

	nodeConfig := &NodeConfig{}

	storageRepo.On("Get", getConfigKey()).
		Once().
		Return(nodeConfig, nil)

	res, err := repo.GetConfig()
	assert.NoError(t, err)
	assert.Equal(t, nodeConfig, res)
}

func TestRepository_GetConfig_Error(t *testing.T) {
	storageRepo := storage.NewRepositoryMock(t)

	repo := NewDBRepository(storageRepo)

	storageRepo.On("Get", getConfigKey()).
		Once().
		Return(nil, storageRepoErr)

	res, err := repo.GetConfig()
	assert.ErrorIs(t, err, storageRepoErr)
	assert.Nil(t, res)
}

func TestRepository_GetAllAccounts(t *testing.T) {
	storageRepo := storage.NewRepositoryMock(t)

	repo := NewDBRepository(storageRepo)

	accounts, err := getRandomAccounts(2)
	assert.NoError(t, err)

	models := make([]storage.Model, 0, len(accounts))

	for _, account := range accounts {
		models = append(models, account)
	}

	storageRepo.On("GetAllByPrefix", accountPrefix).
		Once().
		Return(models, nil)

	res, err := repo.GetAllAccounts()
	assert.NoError(t, err)
	assert.Equal(t, accounts, res)
}

func TestRepository_GetAllAccounts_Error(t *testing.T) {
	storageRepo := storage.NewRepositoryMock(t)

	repo := NewDBRepository(storageRepo)

	storageRepo.On("GetAllByPrefix", accountPrefix).
		Once().
		Return(nil, storageRepoErr)

	res, err := repo.GetAllAccounts()
	assert.ErrorIs(t, err, storageRepoErr)
	assert.Nil(t, res)
}

func TestRepository_CreateNodeAdmin(t *testing.T) {
	storageRepo := storage.NewRepositoryMock(t)

	repo := NewDBRepository(storageRepo)

	nodeAdmin := &NodeAdmin{}

	storageRepo.On("Create", getNodeAdminKey(), nodeAdmin).
		Once().
		Return(nil)

	err := repo.CreateNodeAdmin(nodeAdmin)
	assert.NoError(t, err)
}

func TestRepository_CreateNodeAdmin_Error(t *testing.T) {
	storageRepo := storage.NewRepositoryMock(t)

	repo := NewDBRepository(storageRepo)

	nodeAdmin := &NodeAdmin{}

	storageRepo.On("Create", getNodeAdminKey(), nodeAdmin).
		Once().
		Return(storageRepoErr)

	err := repo.CreateNodeAdmin(nodeAdmin)
	assert.ErrorIs(t, err, storageRepoErr)
}

func TestRepository_CreateAccount(t *testing.T) {
	storageRepo := storage.NewRepositoryMock(t)

	repo := NewDBRepository(storageRepo)

	acc, err := getRandomAccount()
	assert.NoError(t, err)

	storageRepo.On("Create", getAccountKey(acc.GetIdentity().ToBytes()), acc).
		Once().
		Return(nil)

	err = repo.CreateAccount(acc)
	assert.NoError(t, err)
}

func TestRepository_CreateAccount_Error(t *testing.T) {
	storageRepo := storage.NewRepositoryMock(t)

	repo := NewDBRepository(storageRepo)

	acc, err := getRandomAccount()
	assert.NoError(t, err)

	storageRepo.On("Create", getAccountKey(acc.GetIdentity().ToBytes()), acc).
		Once().
		Return(storageRepoErr)

	err = repo.CreateAccount(acc)
	assert.ErrorIs(t, err, storageRepoErr)
}

func TestRepository_CreateConfig(t *testing.T) {
	storageRepo := storage.NewRepositoryMock(t)

	repo := NewDBRepository(storageRepo)

	cfg := &NodeConfig{}

	storageRepo.On("Create", getConfigKey(), cfg).
		Once().
		Return(nil)

	err := repo.CreateConfig(cfg)
	assert.NoError(t, err)
}

func TestRepository_CreateConfig_Error(t *testing.T) {
	storageRepo := storage.NewRepositoryMock(t)

	repo := NewDBRepository(storageRepo)

	cfg := &NodeConfig{}

	storageRepo.On("Create", getConfigKey(), cfg).
		Once().
		Return(storageRepoErr)

	err := repo.CreateConfig(cfg)
	assert.ErrorIs(t, err, storageRepoErr)
}

func TestRepository_UpdateNodeAdmin(t *testing.T) {
	storageRepo := storage.NewRepositoryMock(t)

	repo := NewDBRepository(storageRepo)

	nodeAdmin := &NodeAdmin{}

	storageRepo.On("Update", getNodeAdminKey(), nodeAdmin).
		Once().
		Return(nil)

	err := repo.UpdateNodeAdmin(nodeAdmin)
	assert.NoError(t, err)
}

func TestRepository_UpdateNodeAdmin_Error(t *testing.T) {
	storageRepo := storage.NewRepositoryMock(t)

	repo := NewDBRepository(storageRepo)

	nodeAdmin := &NodeAdmin{}

	storageRepo.On("Update", getNodeAdminKey(), nodeAdmin).
		Once().
		Return(storageRepoErr)

	err := repo.UpdateNodeAdmin(nodeAdmin)
	assert.ErrorIs(t, err, storageRepoErr)
}

func TestRepository_UpdateAccount(t *testing.T) {
	storageRepo := storage.NewRepositoryMock(t)

	repo := NewDBRepository(storageRepo)

	acc, err := getRandomAccount()
	assert.NoError(t, err)

	storageRepo.On("Update", getAccountKey(acc.GetIdentity().ToBytes()), acc).
		Once().
		Return(nil)

	err = repo.UpdateAccount(acc)
	assert.NoError(t, err)
}

func TestRepository_UpdateAccount_Error(t *testing.T) {
	storageRepo := storage.NewRepositoryMock(t)

	repo := NewDBRepository(storageRepo)

	acc, err := getRandomAccount()
	assert.NoError(t, err)

	storageRepo.On("Update", getAccountKey(acc.GetIdentity().ToBytes()), acc).
		Once().
		Return(storageRepoErr)

	err = repo.UpdateAccount(acc)
	assert.ErrorIs(t, err, storageRepoErr)
}

func TestRepository_UpdateConfig(t *testing.T) {
	storageRepo := storage.NewRepositoryMock(t)

	repo := NewDBRepository(storageRepo)

	cfg := &NodeConfig{}

	storageRepo.On("Update", getConfigKey(), cfg).
		Once().
		Return(nil)

	err := repo.UpdateConfig(cfg)
	assert.NoError(t, err)
}

func TestRepository_UpdateConfig_Error(t *testing.T) {
	storageRepo := storage.NewRepositoryMock(t)

	repo := NewDBRepository(storageRepo)

	cfg := &NodeConfig{}

	storageRepo.On("Update", getConfigKey(), cfg).
		Once().
		Return(storageRepoErr)

	err := repo.UpdateConfig(cfg)
	assert.ErrorIs(t, err, storageRepoErr)
}

func TestRepository_DeleteAccount(t *testing.T) {
	storageRepo := storage.NewRepositoryMock(t)

	repo := NewDBRepository(storageRepo)

	acc, err := getRandomAccount()
	assert.NoError(t, err)

	storageRepo.On("Delete", getAccountKey(acc.GetIdentity().ToBytes())).
		Once().
		Return(nil)

	err = repo.DeleteAccount(acc.GetIdentity().ToBytes())
	assert.NoError(t, err)
}

func TestRepository_DeleteAccountError(t *testing.T) {
	storageRepo := storage.NewRepositoryMock(t)

	repo := NewDBRepository(storageRepo)

	acc, err := getRandomAccount()
	assert.NoError(t, err)

	storageRepo.On("Delete", getAccountKey(acc.GetIdentity().ToBytes())).
		Once().
		Return(storageRepoErr)

	err = repo.DeleteAccount(acc.GetIdentity().ToBytes())
	assert.ErrorIs(t, err, storageRepoErr)
}

func TestRepository_DeleteConfig(t *testing.T) {
	storageRepo := storage.NewRepositoryMock(t)

	repo := NewDBRepository(storageRepo)

	storageRepo.On("Delete", getConfigKey()).
		Once().
		Return(nil)

	err := repo.DeleteConfig()
	assert.NoError(t, err)
}

func TestRepository_DeleteConfig_Error(t *testing.T) {
	storageRepo := storage.NewRepositoryMock(t)

	repo := NewDBRepository(storageRepo)

	storageRepo.On("Delete", getConfigKey()).
		Once().
		Return(storageRepoErr)

	err := repo.DeleteConfig()
	assert.ErrorIs(t, err, storageRepoErr)
}

func getRandomAccount() (config.Account, error) {
	accountID, err := testingcommons.GetRandomAccountID()

	if err != nil {
		return nil, err
	}

	accountProxies, err := getRandomAccountProxies(2)

	if err != nil {
		return nil, err
	}

	account := &Account{
		Identity:          accountID,
		P2PPublicKey:      utils.RandomSlice(32),
		P2PPrivateKey:     utils.RandomSlice(32),
		SigningPublicKey:  utils.RandomSlice(32),
		SigningPrivateKey: utils.RandomSlice(32),
		WebhookURL:        "https://centrifuge.io",
		PrecommitEnabled:  false,
		AccountProxies:    accountProxies,
	}

	return account, nil
}

func getRandomAccounts(count int) ([]config.Account, error) {
	var accounts []config.Account

	for i := 0; i < count; i++ {
		account, err := getRandomAccount()

		if err != nil {
			return nil, err
		}

		accounts = append(accounts, account)
	}

	return accounts, nil
}

func getRandomAccountProxies(count int) (config.AccountProxies, error) {
	var accountProxies config.AccountProxies

	for i := 0; i < count; i++ {
		b := make([]byte, 32)

		if _, err := rand.Read(b); err != nil {
			return nil, err
		}

		accountID, err := types.NewAccountID(b)

		if err != nil {
			return nil, err
		}

		accountProxy := &config.AccountProxy{
			Default:     false,
			AccountID:   accountID,
			Secret:      "some_secret",
			SS58Address: "some_address",
			ProxyType:   testingcommons.GetRandomProxyType(),
		}

		accountProxies = append(accountProxies, accountProxy)
	}

	return accountProxies, nil
}

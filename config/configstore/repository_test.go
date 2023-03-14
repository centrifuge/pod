//go:build unit || integration

package configstore

import (
	"testing"

	"github.com/centrifuge/pod/config"
	"github.com/centrifuge/pod/errors"
	"github.com/centrifuge/pod/storage"
	testingcommons "github.com/centrifuge/pod/testingutils/common"
	"github.com/centrifuge/pod/utils"
	"github.com/stretchr/testify/assert"
)

var (
	storageRepoErr = errors.New("error")
)

func TestRepository_Register(t *testing.T) {
	storageRepo := storage.NewRepositoryMock(t)

	repo := NewDBRepository(storageRepo)

	nodeAdmin := &PodAdmin{}

	storageRepo.On("Register", nodeAdmin).Once()

	repo.RegisterNodeAdmin(nodeAdmin)

	acc := &Account{}

	storageRepo.On("Register", acc).Once()

	repo.RegisterAccount(acc)

	cfg := &NodeConfig{}

	storageRepo.On("Register", cfg).Once()

	repo.RegisterConfig(cfg)

	podOperator := &PodOperator{}

	storageRepo.On("Register", podOperator).Once()

	repo.RegisterPodOperator(podOperator)
}

func TestRepository_GetNodeAdmin(t *testing.T) {
	storageRepo := storage.NewRepositoryMock(t)

	repo := NewDBRepository(storageRepo)

	nodeAdmin := &PodAdmin{}

	storageRepo.On("Get", getNodeAdminKey()).
		Once().
		Return(nodeAdmin, nil)

	res, err := repo.GetPodAdmin()
	assert.NoError(t, err)
	assert.Equal(t, nodeAdmin, res)
}

func TestRepository_GetNodeAdmin_Error(t *testing.T) {
	storageRepo := storage.NewRepositoryMock(t)

	repo := NewDBRepository(storageRepo)

	storageRepo.On("Get", getNodeAdminKey()).
		Once().
		Return(nil, storageRepoErr)

	res, err := repo.GetPodAdmin()
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

func TestRepository_GetPodOperator(t *testing.T) {
	storageRepo := storage.NewRepositoryMock(t)

	repo := NewDBRepository(storageRepo)

	podOperator := &PodOperator{}

	storageRepo.On("Get", getPodOperatorKey()).
		Once().
		Return(podOperator, nil)

	res, err := repo.GetPodOperator()
	assert.NoError(t, err)
	assert.Equal(t, podOperator, res)
}

func TestRepository_GetPodOperator_Error(t *testing.T) {
	storageRepo := storage.NewRepositoryMock(t)

	repo := NewDBRepository(storageRepo)

	storageRepo.On("Get", getPodOperatorKey()).
		Once().
		Return(nil, storageRepoErr)

	res, err := repo.GetPodOperator()
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

	nodeAdmin := &PodAdmin{}

	storageRepo.On("Create", getNodeAdminKey(), nodeAdmin).
		Once().
		Return(nil)

	err := repo.CreatePodAdmin(nodeAdmin)
	assert.NoError(t, err)
}

func TestRepository_CreateNodeAdmin_Error(t *testing.T) {
	storageRepo := storage.NewRepositoryMock(t)

	repo := NewDBRepository(storageRepo)

	nodeAdmin := &PodAdmin{}

	storageRepo.On("Create", getNodeAdminKey(), nodeAdmin).
		Once().
		Return(storageRepoErr)

	err := repo.CreatePodAdmin(nodeAdmin)
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

func TestRepository_CreatePodOperator(t *testing.T) {
	storageRepo := storage.NewRepositoryMock(t)

	repo := NewDBRepository(storageRepo)

	podOperator := &PodOperator{}

	storageRepo.On("Create", getPodOperatorKey(), podOperator).
		Once().
		Return(nil)

	err := repo.CreatePodOperator(podOperator)
	assert.NoError(t, err)
}

func TestRepository_CreatePodOperator_Error(t *testing.T) {
	storageRepo := storage.NewRepositoryMock(t)

	repo := NewDBRepository(storageRepo)

	podOperator := &PodOperator{}

	storageRepo.On("Create", getPodOperatorKey(), podOperator).
		Once().
		Return(storageRepoErr)

	err := repo.CreatePodOperator(podOperator)
	assert.ErrorIs(t, err, storageRepoErr)
}

func TestRepository_UpdateNodeAdmin(t *testing.T) {
	storageRepo := storage.NewRepositoryMock(t)

	repo := NewDBRepository(storageRepo)

	nodeAdmin := &PodAdmin{}

	storageRepo.On("Update", getNodeAdminKey(), nodeAdmin).
		Once().
		Return(nil)

	err := repo.UpdatePodAdmin(nodeAdmin)
	assert.NoError(t, err)
}

func TestRepository_UpdateNodeAdmin_Error(t *testing.T) {
	storageRepo := storage.NewRepositoryMock(t)

	repo := NewDBRepository(storageRepo)

	nodeAdmin := &PodAdmin{}

	storageRepo.On("Update", getNodeAdminKey(), nodeAdmin).
		Once().
		Return(storageRepoErr)

	err := repo.UpdatePodAdmin(nodeAdmin)
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

func TestRepository_UpdatePodOperator(t *testing.T) {
	storageRepo := storage.NewRepositoryMock(t)

	repo := NewDBRepository(storageRepo)

	podOperator := &PodOperator{}

	storageRepo.On("Update", getPodOperatorKey(), podOperator).
		Once().
		Return(nil)

	err := repo.UpdatePodOperator(podOperator)
	assert.NoError(t, err)
}

func TestRepository_UpdatePodOperator_Error(t *testing.T) {
	storageRepo := storage.NewRepositoryMock(t)

	repo := NewDBRepository(storageRepo)

	podOperator := &PodOperator{}

	storageRepo.On("Update", getPodOperatorKey(), podOperator).
		Once().
		Return(storageRepoErr)

	err := repo.UpdatePodOperator(podOperator)
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

	account := &Account{
		Identity:          accountID,
		SigningPublicKey:  utils.RandomSlice(32),
		SigningPrivateKey: utils.RandomSlice(32),
		WebhookURL:        "https://centrifuge.io",
		PrecommitEnabled:  false,
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

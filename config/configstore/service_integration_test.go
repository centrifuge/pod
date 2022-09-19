//go:build integration

package configstore

import (
	"os"
	"testing"

	storage "github.com/centrifuge/go-centrifuge/storage/leveldb"
	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/common"
	"github.com/stretchr/testify/assert"
)

const (
	serviceTestDirPattern = "configstore-service-integration-test"
)

func TestService_ConfigOperations(t *testing.T) {
	randomPath, err := testingcommons.GetRandomTestStoragePath(serviceTestDirPattern)
	assert.NoError(t, err)

	defer func() {
		err = os.RemoveAll(randomPath)
		assert.NoError(t, err)
	}()

	db, err := storage.NewLevelDBStorage(randomPath)
	assert.NoError(t, err)

	repo := NewDBRepository(storage.NewLevelDBRepository(db))
	assert.NotNil(t, repo)

	service := NewService(repo)

	// Config not present.
	res, err := service.GetConfig()
	assert.NotNil(t, err)
	assert.Nil(t, res)

	cfg := &NodeConfig{}

	err = service.CreateConfig(cfg)
	assert.NoError(t, err)

	// Config not registered.
	res, err = service.GetConfig()
	assert.NotNil(t, err)
	assert.Nil(t, res)

	repo.RegisterConfig(cfg)

	res, err = service.GetConfig()
	assert.NoError(t, err)
	assert.Equal(t, cfg, res)
}

func TestService_NodeAdminOperations(t *testing.T) {
	randomPath, err := testingcommons.GetRandomTestStoragePath(serviceTestDirPattern)
	assert.NoError(t, err)

	defer func() {
		err = os.RemoveAll(randomPath)
		assert.NoError(t, err)
	}()

	db, err := storage.NewLevelDBStorage(randomPath)
	assert.NoError(t, err)

	repo := NewDBRepository(storage.NewLevelDBRepository(db))
	assert.NotNil(t, repo)

	service := NewService(repo)

	// Node admin not present.
	res, err := service.GetPodAdmin()
	assert.NotNil(t, err)
	assert.Nil(t, res)

	nodeAdmin := &PodAdmin{}

	err = service.CreateNodeAdmin(nodeAdmin)
	assert.NoError(t, err)

	// Node admin not registered.
	res, err = service.GetPodAdmin()
	assert.NotNil(t, err)
	assert.Nil(t, res)

	repo.RegisterNodeAdmin(nodeAdmin)

	res, err = service.GetPodAdmin()
	assert.NoError(t, err)
	assert.Equal(t, nodeAdmin, res)
}

func TestService_PodOperatorOperations(t *testing.T) {
	randomPath, err := testingcommons.GetRandomTestStoragePath(serviceTestDirPattern)
	assert.NoError(t, err)

	defer func() {
		err = os.RemoveAll(randomPath)
		assert.NoError(t, err)
	}()

	db, err := storage.NewLevelDBStorage(randomPath)
	assert.NoError(t, err)

	repo := NewDBRepository(storage.NewLevelDBRepository(db))
	assert.NotNil(t, repo)

	service := NewService(repo)

	// Pod operator not present.
	res, err := service.GetPodOperator()
	assert.NotNil(t, err)
	assert.Nil(t, res)

	podOperator := &PodOperator{}

	err = service.CreatePodOperator(podOperator)
	assert.NoError(t, err)

	// Pod operator not registered.
	res, err = service.GetPodOperator()
	assert.NotNil(t, err)
	assert.Nil(t, res)

	repo.RegisterPodOperator(podOperator)

	res, err = service.GetPodOperator()
	assert.NoError(t, err)
	assert.Equal(t, podOperator, res)
}

func TestService_AccountOperations(t *testing.T) {
	randomPath, err := testingcommons.GetRandomTestStoragePath(serviceTestDirPattern)
	assert.NoError(t, err)

	defer func() {
		err = os.RemoveAll(randomPath)
		assert.NoError(t, err)
	}()

	db, err := storage.NewLevelDBStorage(randomPath)
	assert.NoError(t, err)

	repo := NewDBRepository(storage.NewLevelDBRepository(db))
	assert.NotNil(t, repo)

	service := NewService(repo)

	// Account not present.
	account, err := getRandomAccount()
	assert.NoError(t, err)

	res, err := service.GetAccount(account.GetIdentity().ToBytes())
	assert.NotNil(t, err)
	assert.Nil(t, res)

	err = service.CreateAccount(account)
	assert.NoError(t, err)

	// Account not registered.
	res, err = service.GetAccount(account.GetIdentity().ToBytes())
	assert.NotNil(t, err)
	assert.Nil(t, res)

	repo.RegisterAccount(account)

	// Account present.
	res, err = service.GetAccount(account.GetIdentity().ToBytes())
	assert.NoError(t, err)
	assert.Equal(t, account, res)

	acc := account.(*Account)
	acc.WebhookURL = acc.WebhookURL + "/path"

	// Update valid account.
	err = service.UpdateAccount(account)
	assert.NoError(t, err)

	res, err = service.GetAccount(account.GetIdentity().ToBytes())
	assert.NoError(t, err)
	assert.Equal(t, account, res)

	// Delete account.
	err = service.DeleteAccount(account.GetIdentity().ToBytes())
	assert.NoError(t, err)

	// Get, update non-existing account.
	res, err = service.GetAccount(account.GetIdentity().ToBytes())
	assert.NotNil(t, err)
	assert.Nil(t, res)

	err = service.UpdateAccount(account)
	assert.NotNil(t, err)
}

func TestService_Accounts(t *testing.T) {
	randomPath, err := testingcommons.GetRandomTestStoragePath(serviceTestDirPattern)
	assert.NoError(t, err)

	defer func() {
		err = os.RemoveAll(randomPath)
		assert.NoError(t, err)
	}()

	db, err := storage.NewLevelDBStorage(randomPath)
	assert.NoError(t, err)

	repo := NewDBRepository(storage.NewLevelDBRepository(db))
	assert.NotNil(t, repo)

	service := NewService(repo)

	accs, err := service.GetAccounts()
	assert.Nil(t, err)
	assert.Nil(t, accs)

	accounts, err := getRandomAccounts(3)
	assert.NoError(t, err)

	repo.RegisterAccount(accounts[0])

	for _, account := range accounts {
		err = service.CreateAccount(account)
		assert.NoError(t, err)
	}

	accs, err = service.GetAccounts()
	assert.NoError(t, err)
	assert.Len(t, accs, 3)
	assert.Contains(t, accs, accounts[0])
	assert.Contains(t, accs, accounts[1])
	assert.Contains(t, accs, accounts[2])
}

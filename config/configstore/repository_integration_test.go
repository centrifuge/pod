//go:build integration

package configstore

import (
	"os"
	"testing"

	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/commons"

	storage "github.com/centrifuge/go-centrifuge/storage/leveldb"
	"github.com/stretchr/testify/assert"
)

func TestNewLevelDBRepository(t *testing.T) {
	randomPath, err := testingcommons.GetRandomTestStoragePath()
	assert.NoError(t, err)

	defer func() {
		err = os.RemoveAll(randomPath)
		assert.NoError(t, err)
	}()

	db, err := storage.NewLevelDBStorage(randomPath)
	assert.NoError(t, err)

	repo := NewDBRepository(storage.NewLevelDBRepository(db))
	assert.NotNil(t, repo)
}

func TestAccountOperations(t *testing.T) {
	randomPath, err := testingcommons.GetRandomTestStoragePath()
	assert.NoError(t, err)

	defer func() {
		err = os.RemoveAll(randomPath)
		assert.NoError(t, err)
	}()

	db, err := storage.NewLevelDBStorage(randomPath)
	assert.NoError(t, err)

	repo := NewDBRepository(storage.NewLevelDBRepository(db))
	assert.NotNil(t, repo)

	account, err := getRandomAccount()
	assert.NoError(t, err)

	err = repo.CreateAccount(account)
	assert.NoError(t, err)

	// Account not registered.
	acc, err := repo.GetAccount(account.GetIdentity().ToBytes())
	assert.NotNil(t, err)
	assert.Nil(t, acc)

	repo.RegisterAccount(account)

	acc, err = repo.GetAccount(account.GetIdentity().ToBytes())
	assert.NoError(t, err)
	assert.Equal(t, acc, account)

	// Account already exists.
	err = repo.CreateAccount(account)
	assert.NotNil(t, err)

	// Update account.
	testAcc := account.(*Account)
	testAcc.WebhookURL = "https://some.url"

	err = repo.UpdateAccount(account)
	assert.NoError(t, err)

	acc, err = repo.GetAccount(testAcc.GetIdentity().ToBytes())
	assert.NoError(t, err)
	assert.Equal(t, acc, account)

	// Non-existent account update.
	secondAccount, err := getRandomAccount()
	assert.NoError(t, err)

	err = repo.UpdateAccount(secondAccount)
	assert.NotNil(t, err)

	// Account deletion.
	err = repo.DeleteAccount(account.GetIdentity().ToBytes())
	assert.NoError(t, err)
}

func TestConfigOperations(t *testing.T) {
	randomPath, err := testingcommons.GetRandomTestStoragePath()
	assert.NoError(t, err)

	defer func() {
		err = os.RemoveAll(randomPath)
		assert.NoError(t, err)
	}()

	db, err := storage.NewLevelDBStorage(randomPath)
	assert.NoError(t, err)

	repo := NewDBRepository(storage.NewLevelDBRepository(db))
	assert.NotNil(t, repo)

	config := &NodeConfig{
		NetworkID: 32,
	}

	err = repo.CreateConfig(config)
	assert.NoError(t, err)

	// Config not registered.
	cfg, err := repo.GetConfig()
	assert.NotNil(t, err)
	assert.Nil(t, cfg)

	repo.RegisterConfig(config)

	cfg, err = repo.GetConfig()
	assert.NoError(t, err)
	assert.Equal(t, config, cfg)

	// Config already exists.
	err = repo.CreateConfig(config)
	assert.NotNil(t, err)

	// Update config.
	config.NetworkID = 123

	err = repo.UpdateConfig(config)
	assert.NoError(t, err)

	cfg, err = repo.GetConfig()
	assert.NoError(t, err)
	assert.Equal(t, config, cfg)

	// Delete config.
	err = repo.DeleteConfig()
	assert.NoError(t, err)

	// Update non-existent config.
	err = repo.UpdateConfig(config)
	assert.NotNil(t, err)

	cfg, err = repo.GetConfig()
	assert.NotNil(t, err)
	assert.Nil(t, cfg)
}

func TestNodeAdminOperations(t *testing.T) {
	randomPath, err := testingcommons.GetRandomTestStoragePath()
	assert.NoError(t, err)

	defer func() {
		err = os.RemoveAll(randomPath)
		assert.NoError(t, err)
	}()

	db, err := storage.NewLevelDBStorage(randomPath)
	assert.NoError(t, err)

	repo := NewDBRepository(storage.NewLevelDBRepository(db))
	assert.NotNil(t, repo)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	nodeAdmin := &NodeAdmin{AccountID: accountID}

	err = repo.CreateNodeAdmin(nodeAdmin)
	assert.NoError(t, err)

	// NodeAdmin not registered.
	res, err := repo.GetNodeAdmin()
	assert.NotNil(t, err)
	assert.Nil(t, res)

	repo.RegisterNodeAdmin(nodeAdmin)

	res, err = repo.GetNodeAdmin()
	assert.NoError(t, err)
	assert.Equal(t, nodeAdmin, res)

	// Node admin already exists.
	err = repo.CreateNodeAdmin(nodeAdmin)
	assert.NotNil(t, err)

	// Update node admin.
	newAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	nodeAdmin.AccountID = newAccountID

	err = repo.UpdateNodeAdmin(nodeAdmin)
	assert.NoError(t, err)

	res, err = repo.GetNodeAdmin()
	assert.NoError(t, err)
	assert.Equal(t, nodeAdmin, res)
}

func TestLevelDBRepo_GetAllAccounts(t *testing.T) {
	randomPath, err := testingcommons.GetRandomTestStoragePath()
	assert.NoError(t, err)

	defer func() {
		err = os.RemoveAll(randomPath)
		assert.NoError(t, err)
	}()

	db, err := storage.NewLevelDBStorage(randomPath)
	assert.NoError(t, err)

	repo := NewDBRepository(storage.NewLevelDBRepository(db))
	assert.NotNil(t, repo)

	accounts, err := getRandomAccounts(2)
	assert.NoError(t, err)

	repo.RegisterAccount(accounts[0])

	for _, account := range accounts {
		err = repo.CreateAccount(account)
		assert.NoError(t, err)
	}

	accs, err := repo.GetAllAccounts()
	assert.NoError(t, err)
	assert.Contains(t, accs, accounts[0])
	assert.Contains(t, accs, accounts[1])
}

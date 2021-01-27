// +build unit

package configstore

import (
	"os"
	"reflect"
	"testing"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testlogging"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs/jobsv2"
	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/commons"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/stretchr/testify/assert"
)

var dbFiles []string
var ctx = map[string]interface{}{}
var cfg config.Configuration

func TestMain(m *testing.M) {
	ibootstappers := []bootstrap.TestBootstrapper{
		&testlogging.TestLoggingBootstrapper{},
		&config.Bootstrapper{},
		&leveldb.Bootstrapper{},
		jobsv2.Bootstrapper{},
	}
	ctx[identity.BootstrappedDIDService] = &testingcommons.MockIdentityService{}
	ctx[identity.BootstrappedDIDFactory] = &testingcommons.MockIdentityFactory{}
	ctx[identity.BootstrappedDIDFactoryV2] = &identity.MockFactory{}
	ctx[ethereum.BootstrappedEthereumClient] = new(ethereum.MockEthClient)
	bootstrap.RunTestBootstrappers(ibootstappers, ctx)
	configdb := ctx[storage.BootstrappedConfigDB].(storage.Repository)
	cfg = ctx[bootstrap.BootstrappedConfig].(config.Configuration)
	nc := NewNodeConfig(cfg)
	// clean db
	_ = configdb.Delete(getConfigKey())
	i, _ := nc.GetIdentityID()
	_ = configdb.Delete(getAccountKey(i))
	result := m.Run()
	cleanupDBFiles()
	os.Exit(result)
}

func TestNewLevelDBRepository(t *testing.T) {
	repo, _, err := getRandomStorage()
	assert.NoError(t, err)
	assert.NotNil(t, repo)
}

func TestUnregisteredModel(t *testing.T) {
	repo, _, err := getRandomStorage()
	assert.NoError(t, err)
	assert.NotNil(t, repo)
	id := utils.RandomSlice(32)
	newaccount := &Account{
		IdentityID:                 id,
		EthereumDefaultAccountName: "main",
	}
	err = repo.CreateAccount(id, newaccount)
	assert.Nil(t, err)

	// Error on non registered model
	_, err = repo.GetAccount(id)
	assert.NotNil(t, err)

	repo.RegisterAccount(&Account{})

	_, err = repo.GetAccount(id)
	assert.Nil(t, err)
}

func TestAccountOperations(t *testing.T) {
	repo, _, err := getRandomStorage()
	assert.NoError(t, err)
	assert.NotNil(t, repo)
	id := utils.RandomSlice(32)
	newaccount := &Account{
		IdentityID:                 id,
		EthereumDefaultAccountName: "main",
	}
	repo.RegisterAccount(&Account{})
	err = repo.CreateAccount(id, newaccount)
	assert.Nil(t, err)

	// Create account already exist
	err = repo.CreateAccount(id, newaccount)
	assert.NotNil(t, err)

	readaccount, err := repo.GetAccount(id)
	assert.Nil(t, err)
	assert.Equal(t, reflect.TypeOf(newaccount), readaccount.Type())
	i := readaccount.GetIdentityID()
	assert.Equal(t, newaccount.IdentityID, i)

	// Update account
	newaccount.EthereumDefaultAccountName = "secondary"
	err = repo.UpdateAccount(id, newaccount)
	assert.Nil(t, err)

	// Update account does not exist
	newId := utils.RandomSlice(32)
	err = repo.UpdateAccount(newId, newaccount)
	assert.NotNil(t, err)

	// Delete account
	err = repo.DeleteAccount(id)
	assert.Nil(t, err)
	_, err = repo.GetAccount(id)
	assert.NotNil(t, err)
}

func TestConfigOperations(t *testing.T) {
	repo, _, _ := getRandomStorage()
	assert.NotNil(t, repo)
	newConfig := &NodeConfig{
		NetworkID: 4,
	}
	repo.RegisterConfig(&NodeConfig{})
	err := repo.CreateConfig(newConfig)
	assert.Nil(t, err)

	// Create config already exist
	err = repo.CreateConfig(newConfig)
	assert.NotNil(t, err)

	readDoc, err := repo.GetConfig()
	assert.Nil(t, err)
	assert.Equal(t, reflect.TypeOf(newConfig), readDoc.Type())
	assert.Equal(t, newConfig.NetworkID, readDoc.GetNetworkID())

	// Update config
	newConfig.NetworkID = 42
	err = repo.UpdateConfig(newConfig)
	assert.Nil(t, err)

	// Delete config
	err = repo.DeleteConfig()
	assert.Nil(t, err)
	_, err = repo.GetConfig()
	assert.NotNil(t, err)

	// Update config does not exist
	err = repo.UpdateConfig(newConfig)
	assert.NotNil(t, err)
}

func TestLevelDBRepo_GetAllAccounts(t *testing.T) {
	repo, _, _ := getRandomStorage()
	assert.NotNil(t, repo)
	repo.RegisterAccount(&Account{})
	ids := [][]byte{utils.RandomSlice(32), utils.RandomSlice(32), utils.RandomSlice(32)}
	ten1 := &Account{
		IdentityID:                 ids[0],
		EthereumDefaultAccountName: "main",
	}
	ten2 := &Account{
		IdentityID:                 ids[1],
		EthereumDefaultAccountName: "main",
	}
	ten3 := &Account{
		IdentityID:                 ids[2],
		EthereumDefaultAccountName: "main",
	}

	err := repo.CreateAccount(ids[0], ten1)
	assert.Nil(t, err)
	err = repo.CreateAccount(ids[1], ten2)
	assert.Nil(t, err)
	err = repo.CreateAccount(ids[2], ten3)
	assert.Nil(t, err)

	accounts, err := repo.GetAllAccounts()
	assert.Nil(t, err)
	assert.Equal(t, 3, len(accounts))
	t0Id := accounts[0].GetIdentityID()
	t1Id := accounts[1].GetIdentityID()
	t2Id := accounts[2].GetIdentityID()
	assert.Contains(t, ids, t0Id)
	assert.Contains(t, ids, t1Id)
	assert.Contains(t, ids, t2Id)
}

func cleanupDBFiles() {
	for _, db := range dbFiles {
		err := os.RemoveAll(db)
		if err != nil {
			accLog.Warningf("Cleanup warn: %v", err)
		}
	}
}

func getRandomStorage() (Repository, string, error) {
	randomPath := leveldb.GetRandomTestStoragePath()
	db, err := leveldb.NewLevelDBStorage(randomPath)
	if err != nil {
		return nil, "", err
	}
	dbFiles = append(dbFiles, randomPath)
	return NewDBRepository(leveldb.NewLevelDBRepository(db)), randomPath, nil
}

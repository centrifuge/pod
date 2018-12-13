// +build unit

package config

import (
	"os"
	"reflect"
	"testing"

	"github.com/syndtr/goleveldb/leveldb"

	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/stretchr/testify/assert"
)

func getRandomStorage() (*leveldb.DB, error) {
	return storage.NewLevelDBStorage(storage.GetRandomTestStoragePath())
}

func TestMain(m *testing.M) {
	result := m.Run()
	os.Exit(result)
}

func TestNewLevelDBRepository(t *testing.T) {
	testStorage, _ := getRandomStorage()
	repo := NewLevelDBRepository(testStorage)
	assert.NotNil(t, repo)
}

func TestUnregisteredModel(t *testing.T) {
	testStorage, _ := getRandomStorage()
	repo := NewLevelDBRepository(testStorage)
	assert.NotNil(t, repo)
	id := utils.RandomSlice(32)
	newTenant := &TenantConfig{
		IdentityID:                 id,
		EthereumDefaultAccountName: "main",
	}
	err := repo.CreateTenant(id, newTenant)
	assert.Nil(t, err)

	// Error on non registered model
	_, err = repo.GetTenant(id)
	assert.NotNil(t, err)

	repo.Register(&TenantConfig{})

	_, err = repo.GetTenant(id)
	assert.Nil(t, err)
}

func TestTenantOperations(t *testing.T) {
	testStorage, _ := getRandomStorage()
	repo := NewLevelDBRepository(testStorage)
	assert.NotNil(t, repo)
	id := utils.RandomSlice(32)
	newTenant := &TenantConfig{
		IdentityID:                 id,
		EthereumDefaultAccountName: "main",
	}
	repo.Register(&TenantConfig{})
	err := repo.CreateTenant(id, newTenant)
	assert.Nil(t, err)

	// Create tenant already exist
	err = repo.CreateTenant(id, newTenant)
	assert.NotNil(t, err)

	readModel, err := repo.GetTenant(id)
	assert.Nil(t, err)
	readTenant := readModel.(*TenantConfig)
	assert.Equal(t, reflect.TypeOf(newTenant), readTenant.Type())
	assert.Equal(t, newTenant.IdentityID, readTenant.IdentityID)

	// Update tenant
	newTenant.EthereumDefaultAccountName = "secondary"
	err = repo.UpdateTenant(id, newTenant)
	assert.Nil(t, err)

	// Update tenant does not exist
	newId := utils.RandomSlice(32)
	err = repo.UpdateTenant(newId, newTenant)
	assert.NotNil(t, err)

	// Delete tenant
	err = repo.DeleteTenant(id)
	assert.Nil(t, err)
	_, err = repo.GetTenant(id)
	assert.NotNil(t, err)
}

func TestConfigOperations(t *testing.T) {
	testStorage, _ := getRandomStorage()
	repo := NewLevelDBRepository(testStorage)
	assert.NotNil(t, repo)
	newConfig := &NodeConfig{
		NetworkID: 4,
	}
	repo.Register(&NodeConfig{})
	err := repo.CreateConfig(newConfig)
	assert.Nil(t, err)

	// Create config already exist
	err = repo.CreateConfig(newConfig)
	assert.NotNil(t, err)

	readModel, err := repo.GetConfig()
	readDoc := readModel.(*NodeConfig)
	assert.Nil(t, err)
	assert.Equal(t, reflect.TypeOf(newConfig), readDoc.Type())
	assert.Equal(t, newConfig.NetworkID, readDoc.NetworkID)

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

func TestLevelDBRepo_GetAllTenants(t *testing.T) {
	testStorage, _ := getRandomStorage()
	repo := NewLevelDBRepository(testStorage)
	assert.NotNil(t, repo)
	repo.Register(&TenantConfig{})
	ids := [][]byte{utils.RandomSlice(32), utils.RandomSlice(32), utils.RandomSlice(32)}
	ten1 := &TenantConfig{
		IdentityID:                 ids[0],
		EthereumDefaultAccountName: "main",
	}
	ten2 := &TenantConfig{
		IdentityID:                 ids[1],
		EthereumDefaultAccountName: "main",
	}
	ten3 := &TenantConfig{
		IdentityID:                 ids[2],
		EthereumDefaultAccountName: "main",
	}

	err := repo.CreateTenant(ids[0], ten1)
	assert.Nil(t, err)
	err = repo.CreateTenant(ids[1], ten2)
	assert.Nil(t, err)
	err = repo.CreateTenant(ids[2], ten3)
	assert.Nil(t, err)

	tenants, err := repo.GetAllTenants()
	assert.Nil(t, err)
	assert.Equal(t, 3, len(tenants))
	t0Id, _ := tenants[0].ID()
	t1Id, _ := tenants[1].ID()
	t2Id, _ := tenants[2].ID()
	assert.Contains(t, ids, t0Id)
	assert.Contains(t, ids, t1Id)
	assert.Contains(t, ids, t2Id)
}

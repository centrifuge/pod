// +build unit

package accounts

import (
	"os"
	"testing"

	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/stretchr/testify/assert"
)

var repo repository

func TestMain(m *testing.M) {
	path := leveldb.GetRandomTestStoragePath()
	db, err := leveldb.NewLevelDBStorage(path)
	if err != nil {
		panic(err)
	}

	repo = newRepository(leveldb.NewLevelDBRepository(db))
	result := m.Run()
	err = os.RemoveAll(path)
	if err != nil {
		accLog.Warningf("Cleanup warn: %v", err)
	}
	os.Exit(result)
}

func TestRepository_GetAccount_missing(t *testing.T) {
	acc, err := repo.GetAccount(utils.RandomSlice(32))
	assert.Error(t, err)
	assert.Nil(t, acc)
}

func TestRepository_CreateAccount(t *testing.T) {
	acc := NewAccount(utils.RandomSlice(32), utils.RandomSlice(32), "")
	err := repo.CreateAccount(acc.AccountID(), acc)
	assert.NoError(t, err)
	gacc, err := repo.GetAccount(acc.AccountID())
	assert.NoError(t, err)
	assert.Equal(t, acc, gacc)
}

func TestRepository_UpdateAccount(t *testing.T) {
	acc := NewAccount(utils.RandomSlice(32), utils.RandomSlice(32), "")
	err := repo.UpdateAccount(acc.AccountID(), acc)
	assert.Error(t, err)
	err = repo.CreateAccount(acc.AccountID(), acc)
	assert.NoError(t, err)

	acc.(*account).SS58Addr = "randomaddress"
	err = repo.UpdateAccount(acc.AccountID(), acc)

	gacc, err := repo.GetAccount(acc.AccountID())
	assert.NoError(t, err)
	assert.Equal(t, acc, gacc)
}

func TestRepository_DeleteAccount(t *testing.T) {
	acc := NewAccount(utils.RandomSlice(32), utils.RandomSlice(32), "")
	err := repo.CreateAccount(acc.AccountID(), acc)
	assert.NoError(t, err)

	assert.NoError(t, repo.DeleteAccount(acc.AccountID()))
	gacc, err := repo.GetAccount(acc.AccountID())
	assert.Error(t, err)
	assert.Nil(t, gacc)
}

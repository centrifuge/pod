package migration

import (
	"encoding/json"
	"fmt"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-errors/errors"
	"github.com/stretchr/testify/assert"
	"github.com/syndtr/goleveldb/leveldb"
	"os"
	"path/filepath"
	"testing"
)

type content struct {
	Name string `json:"name"`
}

func fakeCalculateMigrationHash(name string) (string, error) {
	out, err := sha256Hash([]byte(name))
	if err != nil {
		return "", err
	}
	return hexutil.Encode(out), nil
}

// Test migration items
func Migration0(db *leveldb.DB) error {
	err := db.Put([]byte("new"), []byte("sample"), nil)
	if err != nil {
		return err
	}
	log.Infof("Migration 0 Run successfully")
	return nil
}

func Migration1(db *leveldb.DB) error {
	err := db.Put([]byte("revert"), []byte("shouldbe"), nil)
	if err != nil {
		return err
	}
	log.Errorf("Migration 1 Run failed: %s", "Something failed")
	return errors.New("Something failed")
}
//

func TestHashContent_Migration0(t *testing.T) {
	data, err := Asset("migration/files/0_job_key_to_hex.go")
	assert.NoError(t, err)
	hd, err := sha256Hash(data)
	assert.NoError(t, err)
	assert.Equal(t, "0xe98afe2587cecbf2726e85cb3c6e6eb50e3b435e7eb63c4653b5b5aa2e97b28c", hexutil.Encode(hd))
	fmt.Println()
}

func TestRunMigrations_singleSuccess(t *testing.T) {
	prefix := fmt.Sprintf("/tmp/datadir_%x", utils.RandomByte32())
	targetDir := fmt.Sprintf("%s.leveldb", prefix)

	// Cleanup after test
	defer cleanupDBFiles(prefix)

	// Create test leveldb with some random data with non hex bytes as keys
	db, err := leveldb.OpenFile(targetDir, nil)
	assert.NoError(t, err)
	sampleKey := utils.RandomSlice(52)
	data, err := json.Marshal(&content{"john"})
	assert.NoError(t, err)
	err = db.Put(sampleKey, data, nil)
	assert.NoError(t, err)
	assert.NoError(t, db.Close())

	// Override migrations for testing purposes
	migrations = map[string]func(*leveldb.DB) error{
		"0_success_migration": Migration0,
	}
	runner := NewMigrationRunnerWithHashFunction(fakeCalculateMigrationHash)
	// Run migration to convert binary key to hex
	err = runner.RunMigrations(targetDir)
	assert.NoError(t, err)

	db, err = leveldb.OpenFile(targetDir, nil)
	assert.NoError(t, err)
	has, err := db.Has(sampleKey, nil)
	assert.NoError(t, err)
	assert.True(t, has)
	has, err = db.Has([]byte("new"), nil)
	assert.NoError(t, err)
	assert.True(t, has)
	db.Close()

	// Check that migration success status is stored
	repo, err := NewMigrationRepository(targetDir)
	assert.NoError(t, err)
	mi, err := repo.GetMigrationByID("0_success_migration")
	assert.NoError(t, err)
	assert.Equal(t, "0_success_migration", mi.ID)
	assert.NotNil(t, mi.DateRun)
	hs, err := sha256Hash([]byte("0_success_migration"))
	assert.NoError(t, err)
	assert.Equal(t, hexutil.Encode(hs), mi.Hash)
	repo.db.Close()

	// Try running again, and run should be skipped
	dRun := mi.DateRun
	err = runner.RunMigrations(targetDir)
	assert.NoError(t, err)
	err = repo.RefreshDB()
	assert.NoError(t, err)
	mi, err = repo.GetMigrationByID("0_success_migration")
	assert.NoError(t, err)
	assert.Equal(t, "0_success_migration", mi.ID)
	assert.Equal(t, dRun, mi.DateRun)
	repo.db.Close()

}

func TestRunMigrations_Failure(t *testing.T) {
	prefix := fmt.Sprintf("/tmp/datadir_%x", utils.RandomByte32())
	targetDir := fmt.Sprintf("%s.leveldb", prefix)

	// Cleanup after test
	defer cleanupDBFiles(prefix)

	// Create test leveldb with some random data with non hex bytes as keys
	db, err := leveldb.OpenFile(targetDir, nil)
	assert.NoError(t, err)
	sampleKey := utils.RandomSlice(52)
	data, err := json.Marshal(&content{"john"})
	assert.NoError(t, err)
	err = db.Put(sampleKey, data, nil)
	assert.NoError(t, err)
	assert.NoError(t, db.Close())

	// Override migrations for testing purposes
	migrations = map[string]func(*leveldb.DB) error{
		"1_failed_migration": Migration1,
	}
	// Run migration to convert binary key to hex
	runner := NewMigrationRunnerWithHashFunction(fakeCalculateMigrationHash)
	err = runner.RunMigrations(targetDir)
	assert.Error(t, err)

	db, err = leveldb.OpenFile(targetDir, nil)
	assert.NoError(t, err)
	has, err := db.Has(sampleKey, nil)
	assert.NoError(t, err)
	assert.True(t, has)
	has, err = db.Has([]byte("revert"), nil)
	assert.NoError(t, err)
	assert.False(t, has)
	db.Close()
}


// util functions
func cleanupDBFiles(prefix string) {
	files, err := filepath.Glob(prefix+"*")
	if err != nil {
		panic(err)
	}
	for _, f := range files {
		if err := os.RemoveAll(f); err != nil {
			panic(err)
		}
	}
}
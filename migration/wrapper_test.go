package migration

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-errors/errors"
	"github.com/stretchr/testify/assert"
	"github.com/syndtr/goleveldb/leveldb"
)

type content struct {
	Name string `json:"name"`
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

func TestNewMigrationRunner(t *testing.T) {
	assert.NotNil(t, NewMigrationRunner())
}

func TestRunner_RunMigrations_AlreadyOpenError(t *testing.T) {
	prefix := fmt.Sprintf("/tmp/datadir_%x", utils.RandomByte32())
	targetDir := fmt.Sprintf("%s.leveldb", prefix)

	// Cleanup after test
	defer cleanupDBFiles(prefix)

	_, err := leveldb.OpenFile(targetDir, nil)
	assert.NoError(t, err)

	runner := NewMigrationRunner()
	err = runner.RunMigrations(targetDir)
	assert.Error(t, err)
}

func TestBackupDB(t *testing.T) {
	prefix := fmt.Sprintf("/tmp/datadir_%x", utils.RandomByte32())
	targetDir := fmt.Sprintf("%s.leveldb", prefix)

	// Cleanup after test
	defer cleanupDBFiles(prefix)

	repo, err := NewMigrationRepository(targetDir)
	assert.NoError(t, err)

	// Force DB close error
	err = repo.Close()
	assert.NoError(t, err)
	_, err = backupDB(repo, "SomeID")
	assert.Error(t, err)

	err = repo.Open()
	assert.NoError(t, err)

	bkp, err := backupDB(repo, "SomeID")
	assert.NoError(t, err)
	assert.NotNil(t, bkp)
	_, err = os.Stat(bkp.dbPath)
	assert.NoError(t, err)
}

func TestRevertToBackupDB(t *testing.T) {
	prefix := fmt.Sprintf("/tmp/datadir_%x", utils.RandomByte32())
	targetDir := fmt.Sprintf("%s.leveldb", prefix)

	// Cleanup after test
	defer cleanupDBFiles(prefix)

	repo, err := NewMigrationRepository(targetDir)
	assert.NoError(t, err)

	bkp, err := backupDB(repo, "SomeID")
	assert.NoError(t, err)

	// Force already closed error
	assert.NoError(t, bkp.Close())
	assert.Error(t, revertDBToBackup(repo, bkp))

	// Force wrong src path
	assert.NoError(t, bkp.Open())
	repoPath := repo.dbPath
	repo.dbPath = repo.dbPath + "_wrong"
	assert.Error(t, revertDBToBackup(repo, bkp))
	repo.dbPath = repoPath

	// Force wrong bkp path
	bkpPath := bkp.dbPath
	bkp.dbPath = bkp.dbPath + "_wrong"
	assert.Error(t, revertDBToBackup(repo, bkp))
	bkp.dbPath = bkpPath

	// Backup revert succeeds
	assert.NoError(t, repo.Open())
	assert.NoError(t, revertDBToBackup(repo, bkp))

	// Bkp doesnt exist anymore
	_, err = os.Stat(bkp.dbPath)
	assert.Error(t, err)
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
		"0SuccessMigration": Migration0,
	}
	runner := NewMigrationRunner()
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
	assert.NoError(t, db.Close())

	// Check that migration success status is stored
	repo, err := NewMigrationRepository(targetDir)
	assert.NoError(t, err)
	mi, err := repo.GetMigrationByID("0SuccessMigration")
	assert.NoError(t, err)
	assert.Equal(t, "0SuccessMigration", mi.ID)
	assert.NotNil(t, mi.DateRun)
	hs, err := sha256Hash([]byte("0SuccessMigration"))
	assert.NoError(t, err)
	assert.Equal(t, hexutil.Encode(hs), mi.Hash)
	assert.NoError(t, repo.Close())

	// Try running again, and run should be skipped
	dRun := mi.DateRun
	err = runner.RunMigrations(targetDir)
	assert.NoError(t, err)
	err = repo.Open()
	assert.NoError(t, err)
	mi, err = repo.GetMigrationByID("0SuccessMigration")
	assert.NoError(t, err)
	assert.Equal(t, "0SuccessMigration", mi.ID)
	assert.Equal(t, dRun, mi.DateRun)
	assert.NoError(t, repo.Close())

}

func TestRunner_RunMigrations_FailureHashFunction(t *testing.T) {
	prefix := fmt.Sprintf("/tmp/datadir_%x", utils.RandomByte32())
	targetDir := fmt.Sprintf("%s.leveldb", prefix)

	// Cleanup after test
	defer cleanupDBFiles(prefix)

	// Override migrations for testing purposes
	migrations = map[string]func(*leveldb.DB) error{
		"0SuccessMigration": Migration0,
	}
	runner := NewMigrationRunner()
	// Run migration and expect error in calculateHash function
	err := runner.RunMigrations(targetDir)
	assert.Error(t, err)
	assert.EqualError(t, err, "0SuccessMigration")
}

func TestRunner_RunMigrations_Failure(t *testing.T) {
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
		"1FailedMigration": Migration1,
	}
	// Run migration to convert binary key to hex
	runner := NewMigrationRunner()
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
	assert.NoError(t, db.Close())
}

// util functions
func cleanupDBFiles(prefix string) {
	files, err := filepath.Glob(prefix + "*")
	if err != nil {
		panic(err)
	}
	for _, f := range files {
		if err := os.RemoveAll(f); err != nil {
			panic(err)
		}
	}
}

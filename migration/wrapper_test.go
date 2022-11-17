//go:build unit

package migration

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"testing"

	migrationutils "github.com/centrifuge/go-centrifuge/migration/utils"
	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/common"
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

func TestNewMigrationRunner(t *testing.T) {
	assert.NotNil(t, NewMigrationRunner())
}

func TestRunner_RunMigrations_AlreadyOpenError(t *testing.T) {
	testDir, err := testingcommons.GetRandomTestStoragePath(migrationTestDirPattern)
	assert.NoError(t, err)

	defer migrationutils.CleanupDBFiles(testDir)

	dbFileName := fmt.Sprintf("%x.leveldb", migrationutils.RandomByte32())

	targetDir := path.Join(testDir, dbFileName)

	_, err = leveldb.OpenFile(targetDir, nil)
	assert.NoError(t, err)

	runner := NewMigrationRunner()
	err = runner.RunMigrations(targetDir)
	assert.Error(t, err)
}

func TestBackupDB(t *testing.T) {
	testDir, err := testingcommons.GetRandomTestStoragePath(migrationTestDirPattern)
	assert.NoError(t, err)

	defer migrationutils.CleanupDBFiles(testDir)

	dbFileName := fmt.Sprintf("%x.leveldb", migrationutils.RandomByte32())

	targetDir := path.Join(testDir, dbFileName)

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
	testDir, err := testingcommons.GetRandomTestStoragePath(migrationTestDirPattern)
	assert.NoError(t, err)

	defer migrationutils.CleanupDBFiles(testDir)

	dbFileName := fmt.Sprintf("%x.leveldb", migrationutils.RandomByte32())

	targetDir := path.Join(testDir, dbFileName)

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
	testDir, err := testingcommons.GetRandomTestStoragePath(migrationTestDirPattern)
	assert.NoError(t, err)

	defer migrationutils.CleanupDBFiles(testDir)

	dbFileName := fmt.Sprintf("%x.leveldb", migrationutils.RandomByte32())

	targetDir := path.Join(testDir, dbFileName)

	// Create test leveldb with some random data with non hex bytes as keys
	db, err := leveldb.OpenFile(targetDir, nil)
	assert.NoError(t, err)
	sampleKey := migrationutils.RandomSlice(52)
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
	assert.Equal(t, "0x", mi.Hash)
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

func TestRunner_RunMigrations_Failure(t *testing.T) {
	testDir, err := testingcommons.GetRandomTestStoragePath(migrationTestDirPattern)
	assert.NoError(t, err)

	defer migrationutils.CleanupDBFiles(testDir)

	dbFileName := fmt.Sprintf("%x.leveldb", migrationutils.RandomByte32())

	targetDir := path.Join(testDir, dbFileName)

	// Create test leveldb with some random data with non hex bytes as keys
	db, err := leveldb.OpenFile(targetDir, nil)
	assert.NoError(t, err)
	sampleKey := migrationutils.RandomSlice(52)
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

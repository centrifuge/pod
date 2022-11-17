//go:build unit

package migration

import (
	"fmt"
	"path"
	"testing"
	"time"

	migrationutils "github.com/centrifuge/go-centrifuge/migration/utils"
	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/common"
	"github.com/stretchr/testify/assert"
)

const (
	migrationTestDirPattern = "migration-test-*"
)

func TestNewMigrationRepository(t *testing.T) {
	testDir, err := testingcommons.GetRandomTestStoragePath(migrationTestDirPattern)
	assert.NoError(t, err)

	defer migrationutils.CleanupDBFiles(testDir)

	dbFileName := fmt.Sprintf("%x.leveldb", migrationutils.RandomByte32())

	targetDir := path.Join(testDir, dbFileName)

	// Succeeds on opening a new DB
	repo, err := NewMigrationRepository(targetDir)
	assert.NoError(t, err)

	defer repo.Close()
	// Fails opening on an already open DB
	_, err = NewMigrationRepository(targetDir)
	assert.Error(t, err)
}

func TestMigrationRepo_Exists_DBClosed(t *testing.T) {
	testDir, err := testingcommons.GetRandomTestStoragePath(migrationTestDirPattern)
	assert.NoError(t, err)

	defer migrationutils.CleanupDBFiles(testDir)

	dbFileName := fmt.Sprintf("%x.leveldb", migrationutils.RandomByte32())

	targetDir := path.Join(testDir, dbFileName)

	repo, err := NewMigrationRepository(targetDir)
	assert.NoError(t, err)
	// Forces error
	err = repo.Close()
	assert.NoError(t, err)
	assert.False(t, repo.Exists("blabla"))
}

func TestMigrationRepo_CreateMigration(t *testing.T) {
	testDir, err := testingcommons.GetRandomTestStoragePath(migrationTestDirPattern)
	assert.NoError(t, err)

	defer migrationutils.CleanupDBFiles(testDir)

	dbFileName := fmt.Sprintf("%x.leveldb", migrationutils.RandomByte32())

	targetDir := path.Join(testDir, dbFileName)

	repo, err := NewMigrationRepository(targetDir)
	assert.NoError(t, err)

	defer repo.Close()

	// Successfully creates migration
	mi := &Item{
		Hash:     "Smthng",
		DateRun:  time.Now().UTC(),
		Duration: 3 * time.Second,
		ID:       "else",
	}
	err = repo.CreateMigration(mi)
	assert.NoError(t, err)
	// Migration exists
	mir, err := repo.GetMigrationByID(mi.ID)
	assert.NoError(t, err)
	assert.Equal(t, mi.Hash, mir.Hash)

	// Fails creating same migration
	err = repo.CreateMigration(mi)
	assert.Error(t, err)

	// Fails when nil item provided
	err = repo.CreateMigration(nil)
	assert.Error(t, err)

	// Migration doesnt exist
	_, err = repo.GetMigrationByID("blabla")
	assert.Error(t, err)

	// Wrong migration type stored
	err = repo.db.Put([]byte("migration_blabla"), []byte{0, 1, 2, 3, 4}, nil)
	assert.NoError(t, err)
	_, err = repo.GetMigrationByID("blabla")
	assert.Error(t, err)
}

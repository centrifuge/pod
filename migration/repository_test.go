package migration

import (
	"fmt"
	"testing"
	"time"

	"github.com/centrifuge/go-centrifuge/migration/utils"
	"github.com/stretchr/testify/assert"
)

func TestNewMigrationRepository(t *testing.T) {
	prefix := fmt.Sprintf("/tmp/datadir_%x", migrationutils.RandomByte32())
	targetDir := fmt.Sprintf("%s.leveldb", prefix)

	defer migrationutils.CleanupDBFiles(prefix)

	// Succeeds on opening a new DB
	repo, err := NewMigrationRepository(targetDir)
	assert.NoError(t, err)

	defer repo.Close()
	// Fails opening on an already open DB
	_, err = NewMigrationRepository(targetDir)
	assert.Error(t, err)
}

func TestMigrationRepo_Exists_DBClosed(t *testing.T) {
	prefix := fmt.Sprintf("/tmp/datadir_%x", migrationutils.RandomByte32())
	targetDir := fmt.Sprintf("%s.leveldb", prefix)

	defer migrationutils.CleanupDBFiles(prefix)

	repo, err := NewMigrationRepository(targetDir)
	assert.NoError(t, err)
	// Forces error
	err = repo.Close()
	assert.NoError(t, err)
	assert.False(t, repo.Exists("blabla"))
}

func TestMigrationRepo_CreateMigration(t *testing.T) {
	prefix := fmt.Sprintf("/tmp/datadir_%x", migrationutils.RandomByte32())
	targetDir := fmt.Sprintf("%s.leveldb", prefix)

	defer migrationutils.CleanupDBFiles(prefix)

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

//go:build unit

package migrationfiles

import (
	"fmt"
	"testing"

	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/generic"
	migrationutils "github.com/centrifuge/go-centrifuge/migration/utils"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	testingidentity "github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/stretchr/testify/assert"
	ldb "github.com/syndtr/goleveldb/leveldb"
)

func TestAddStatusToDocuments04(t *testing.T) {
	prefix := fmt.Sprintf("/tmp/datadir_%x", migrationutils.RandomByte32())
	targetDir := fmt.Sprintf("%s.leveldb", prefix)

	// Cleanup after test
	defer migrationutils.CleanupDBFiles(prefix)

	db, err := ldb.OpenFile(targetDir, nil)
	assert.NoError(t, err)
	strRepo := leveldb.NewLevelDBRepository(db)
	repo := documents.NewDBRepository(strRepo)
	repo.Register(new(generic.Generic))
	did := testingidentity.GenerateRandomDID()

	// successful change
	g := generic.InitGeneric(t, did, generic.CreateGenericPayload(t, nil))
	assert.Equal(t, g.GetStatus(), documents.Pending)
	assert.NoError(t, repo.Create(did[:], g.CurrentVersion(), g))
	assert.NoError(t, AddStatusToDocuments04(db))
	m, err := repo.Get(did[:], g.CurrentVersion())
	assert.NoError(t, err)
	g, ok := m.(*generic.Generic)
	assert.True(t, ok)
	assert.Equal(t, g.GetStatus(), documents.Committed)
}

//go:build unit

package migrationfiles

import (
	"fmt"
	"testing"

	"github.com/centrifuge/go-centrifuge/utils"

	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/generic"
	migrationutils "github.com/centrifuge/go-centrifuge/migration/utils"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/common"
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

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	cd, err := documents.NewCoreDocument(utils.RandomSlice(32), documents.CollaboratorsAccess{}, nil)
	assert.NoError(t, err)

	doc := &generic.Generic{CoreDocument: cd}

	// successful change
	assert.Equal(t, doc.GetStatus(), documents.Pending)

	err = repo.Create(accountID.ToBytes(), doc.CurrentVersion(), doc)
	assert.NoError(t, err)

	err = AddStatusToDocuments04(db)
	assert.NoError(t, err)

	res, err := repo.Get(accountID.ToBytes(), doc.CurrentVersion())
	assert.NoError(t, err)

	g, ok := res.(*generic.Generic)
	assert.True(t, ok)
	assert.Equal(t, g.GetStatus(), documents.Committed)
}

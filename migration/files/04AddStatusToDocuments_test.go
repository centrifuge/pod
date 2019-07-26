package migrationfiles

import (
	"fmt"
	"testing"

	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/invoice"
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
	repo.Register(new(invoice.Invoice))
	did := testingidentity.GenerateRandomDID()

	// successful change
	inv := invoice.InitInvoice(t, did, invoice.CreateInvoicePayload(t, nil))
	assert.Equal(t, inv.GetStatus(), documents.Pending)
	assert.NoError(t, repo.Create(did[:], inv.CurrentVersion(), inv))
	assert.NoError(t, AddStatusToDocuments04(db))
	m, err := repo.Get(did[:], inv.CurrentVersion())
	assert.NoError(t, err)
	inv, ok := m.(*invoice.Invoice)
	assert.True(t, ok)
	assert.Equal(t, inv.GetStatus(), documents.Committed)
}

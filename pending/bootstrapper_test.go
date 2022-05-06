//go:build unit
// +build unit

package pending

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	"github.com/stretchr/testify/assert"
)

func TestBootstrapper_Bootstrap(t *testing.T) {
	ctx := make(map[string]interface{})
	randomPath := leveldb.GetRandomTestStoragePath()
	db, err := leveldb.NewLevelDBStorage(randomPath)
	assert.Nil(t, err)
	repo := leveldb.NewLevelDBRepository(db)

	// missing doc srv
	b := Bootstrapper{}
	assert.Error(t, b.Bootstrap(ctx))

	// missing repo
	ctx[documents.BootstrappedDocumentService] = documents.NewServiceMock(t)
	assert.Error(t, b.Bootstrap(ctx))

	// success
	ctx[storage.BootstrappedDB] = repo
	assert.NoError(t, b.Bootstrap(ctx))
}

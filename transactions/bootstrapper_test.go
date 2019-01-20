// +build unit

package transactions

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/storage"

	"github.com/centrifuge/go-centrifuge/storage/leveldb"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/stretchr/testify/assert"
)

func TestBootstrapper_Bootstrap(t *testing.T) {
	b := Bootstrapper{}
	ctx := make(map[string]interface{})
	err := b.Bootstrap(ctx)
	assert.True(t, errors.IsOfType(ErrTransactionBootstrap, err))

	randomPath := leveldb.GetRandomTestStoragePath()
	db, err := leveldb.NewLevelDBStorage(randomPath)
	assert.Nil(t, err)
	ctx[storage.BootstrappedDB] = leveldb.NewLevelDBRepository(db)
	err = b.Bootstrap(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, ctx[BootstrappedRepo])
	assert.NotNil(t, ctx[BootstrappedService])
}

// +build unit

package transactions

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/stretchr/testify/assert"
)

func TestBootstrapper_Bootstrap(t *testing.T) {
	b := Bootstrapper{}
	ctx := make(map[string]interface{})
	err := b.Bootstrap(ctx)
	assert.True(t, errors.IsOfType(ErrTransactionBootstrap, err))

	randomPath := storage.GetRandomTestStoragePath()
	db, err := storage.NewLevelDBStorage(randomPath)
	assert.Nil(t, err)
	ctx[storage.BootstrappedDB] = storage.NewLevelDBRepository(db)
	err = b.Bootstrap(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, ctx[BootstrappedRepo])
}

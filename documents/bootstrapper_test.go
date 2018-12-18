// +build unit

package documents

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/stretchr/testify/assert"
)

func TestBootstrapper_Bootstrap(t *testing.T) {
	ctx := make(map[string]interface{})
	randomPath := storage.GetRandomTestStoragePath()
	db, err := storage.NewLevelDBStorage(randomPath)
	assert.Nil(t, err)
	ctx[storage.BootstrappedDB] = storage.NewLevelDBRepository(db)
	err = Bootstrapper{}.Bootstrap(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, ctx[BootstrappedRegistry])
	_, ok := ctx[BootstrappedRegistry].(*ServiceRegistry)
	assert.True(t, ok)
}

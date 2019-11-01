// +build unit

package accounts

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	"github.com/stretchr/testify/assert"
)

func TestBootstrapper_Bootstrap(t *testing.T) {
	ctx := make(map[string]interface{})
	b := Bootstrapper{}

	// missing storage db
	assert.Error(t, b.Bootstrap(ctx))

	path := leveldb.GetRandomTestStoragePath()
	db, err := leveldb.NewLevelDBStorage(path)
	assert.NoError(t, err)
	ctx[storage.BootstrappedDB] = leveldb.NewLevelDBRepository(db)
	assert.NoError(t, b.Bootstrap(ctx))
	assert.NotNil(t, ctx[BootstrappedAccountSrv])
}

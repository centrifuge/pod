// +build unit

package documents

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/storage"

	"github.com/centrifuge/go-centrifuge/storage/leveldb"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/transactions"
	"github.com/stretchr/testify/assert"
)

func TestBootstrapper_Bootstrap(t *testing.T) {
	ctx := make(map[string]interface{})
	randomPath := leveldb.GetRandomTestStoragePath()
	db, err := leveldb.NewLevelDBStorage(randomPath)
	assert.Nil(t, err)
	repo := leveldb.NewLevelDBRepository(db)
	ctx[bootstrap.BootstrappedConfig] = &testingconfig.MockConfig{}
	ctx[storage.BootstrappedDB] = repo
	ctx[bootstrap.BootstrappedQueueServer] = new(queue.Server)
	ctx[transactions.BootstrappedService] = transactions.NewService(transactions.NewRepository(repo))
	err = Bootstrapper{}.Bootstrap(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, ctx[BootstrappedRegistry])
	_, ok := ctx[BootstrappedRegistry].(*ServiceRegistry)
	assert.True(t, ok)
}

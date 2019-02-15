// +build unit

package documents

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	"github.com/centrifuge/go-centrifuge/testingutils/anchors"
	"github.com/centrifuge/go-centrifuge/testingutils/commons"
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
	ctx[transactions.BootstrappedService] = transactions.NewManager(&testingconfig.MockConfig{}, transactions.NewRepository(repo))
	ctx[identity.BootstrappedIDService] = new(testingcommons.MockIDService)
	ctx[anchors.BootstrappedAnchorRepo] = new(testinganchors.MockAnchorRepo)

	err = Bootstrapper{}.Bootstrap(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, ctx[BootstrappedRegistry])
	_, ok := ctx[BootstrappedRegistry].(*ServiceRegistry)
	assert.True(t, ok)
}

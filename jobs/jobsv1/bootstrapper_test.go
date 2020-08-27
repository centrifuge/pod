// +build unit

package jobsv1

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/stretchr/testify/assert"
)

func TestBootstrapper_Bootstrap(t *testing.T) {
	b := Bootstrapper{}
	ctx := make(map[string]interface{})
	err := b.Bootstrap(ctx)
	assert.True(t, errors.IsOfType(config.ErrConfigRetrieve, err))

	randomPath := leveldb.GetRandomTestStoragePath()
	db, err := leveldb.NewLevelDBStorage(randomPath)
	assert.Nil(t, err)
	ctx[bootstrap.BootstrappedConfig] = &testingconfig.MockConfig{}
	ctx[storage.BootstrappedDB] = leveldb.NewLevelDBRepository(db)
	err = b.Bootstrap(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, ctx[jobs.BootstrappedRepo])
	assert.NotNil(t, ctx[jobs.BootstrappedService])
}

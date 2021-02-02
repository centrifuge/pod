// +build unit

package documents

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/jobs/jobsv1"
	"github.com/centrifuge/go-centrifuge/jobs/jobsv2"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	testinganchors "github.com/centrifuge/go-centrifuge/testingutils/anchors"
	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/commons"
	testingconfig "github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/testingutils/testingjobs"
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
	ctx[jobs.BootstrappedService] = jobsv1.NewManager(&testingconfig.MockConfig{}, jobsv1.NewRepository(repo))
	ctx[anchors.BootstrappedAnchorService] = new(testinganchors.MockAnchorService)
	ctx[identity.BootstrappedDIDService] = new(testingcommons.MockIdentityService)
	ctx[jobs.BootstrappedService] = new(testingjobs.MockJobManager)
	ctx[jobsv2.BootstrappedDispatcher] = new(jobsv2.MockDispatcher)
	ctx[bootstrap.BootstrappedQueueServer] = new(queue.Server)

	err = Bootstrapper{}.Bootstrap(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, ctx[BootstrappedRegistry])
	_, ok := ctx[BootstrappedRegistry].(*ServiceRegistry)
	assert.True(t, ok)
}

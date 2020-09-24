// +build unit

package oracle

import (
	"os"
	"testing"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testlogging"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/jobs/jobsv1"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	"github.com/centrifuge/go-centrifuge/testingutils"
	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/commons"
	testingdocuments "github.com/centrifuge/go-centrifuge/testingutils/documents"
	"github.com/centrifuge/go-centrifuge/testingutils/testingjobs"
	"github.com/stretchr/testify/assert"
)

var ctx = map[string]interface{}{}

func TestMain(m *testing.M) {
	ibootstappers := []bootstrap.TestBootstrapper{
		&testlogging.TestLoggingBootstrapper{},
		&config.Bootstrapper{},
		&leveldb.Bootstrapper{},
		jobsv1.Bootstrapper{},
	}
	bootstrap.RunTestBootstrappers(ibootstappers, ctx)
	result := m.Run()
	bootstrap.RunTestTeardown(ibootstappers)
	os.Exit(result)
}

func TestBootstrapper_Bootstrap(t *testing.T) {
	b := Bootstrapper{}
	ctx := make(map[string]interface{})
	assert.Error(t, b.Bootstrap(ctx))

	ctx[documents.BootstrappedDocumentService] = new(testingdocuments.MockService)
	assert.Error(t, b.Bootstrap(ctx))

	ctx[identity.BootstrappedDIDService] = new(testingcommons.MockIdentityService)
	assert.Error(t, b.Bootstrap(ctx))

	ctx[bootstrap.BootstrappedQueueServer] = new(testingutils.MockQueue)
	assert.Error(t, b.Bootstrap(ctx))

	ctx[jobs.BootstrappedService] = new(testingjobs.MockJobManager)
	assert.NoError(t, b.Bootstrap(ctx))
	assert.NotNil(t, ctx[BootstrappedOracleService])
}

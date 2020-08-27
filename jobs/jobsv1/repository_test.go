// +build unit

package jobsv1

import (
	"os"
	"testing"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testlogging"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	"github.com/centrifuge/go-centrifuge/testingutils/commons"
	"github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
)

var ctx = map[string]interface{}{}

func TestMain(m *testing.M) {
	ibootstappers := []bootstrap.TestBootstrapper{
		&testlogging.TestLoggingBootstrapper{},
		&config.Bootstrapper{},
		&leveldb.Bootstrapper{},
		&configstore.Bootstrapper{},
		Bootstrapper{},
	}
	ctx[identity.BootstrappedDIDFactory] = &testingcommons.MockIdentityFactory{}
	ctx[identity.BootstrappedDIDService] = &testingcommons.MockIdentityService{}
	bootstrap.RunTestBootstrappers(ibootstappers, ctx)
	result := m.Run()
	bootstrap.RunTestTeardown(ibootstappers)
	os.Exit(result)
}

func Test_getKey(t *testing.T) {
	did := testingidentity.GenerateRandomDID()
	id := jobs.NilJobID()

	// empty id
	key, err := getKey(did, id)
	assert.Nil(t, key)
	assert.Error(t, err)
	assert.Equal(t, "job ID is not valid", err.Error())

	id = jobs.NewJobID()
	key, err = getKey(did, id)
	assert.Nil(t, err)
	assert.Equal(t, append([]byte(jobPrefix), []byte(hexutil.Encode(append(did[:], id.Bytes()...)))...), key)
}

func TestRepository(t *testing.T) {
	did := testingidentity.GenerateRandomDID()
	bytes := utils.RandomSlice(identity.DIDLength)
	assert.Equal(t, identity.DIDLength, copy(did[:], bytes))

	repo := ctx[jobs.BootstrappedRepo].(jobs.Repository)
	job := jobs.NewJob(did, "Some transaction")
	assert.NotNil(t, job.ID)
	assert.NotNil(t, job.DID)
	assert.Equal(t, jobs.Pending, job.Status)

	// get job from repo
	_, err := repo.Get(did, job.ID)
	assert.True(t, errors.IsOfType(jobs.ErrJobsMissing, err))

	// save job into repo
	job.Status = jobs.Success
	err = repo.Save(job)
	assert.Nil(t, err)

	// get job back
	job, err = repo.Get(did, job.ID)
	assert.Nil(t, err)
	assert.NotNil(t, job)
	assert.Equal(t, did, job.DID)
	assert.Equal(t, jobs.Success, job.Status)
}

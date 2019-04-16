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
	cid := testingidentity.GenerateRandomDID()
	id := jobs.NilJobID()

	// empty id
	key, err := getKey(cid, id)
	assert.Nil(t, key)
	assert.Equal(t, "job ID is not valid", err.Error())

	id = jobs.NewJobID()
	key, err = getKey(cid, id)
	assert.Nil(t, err)
	assert.Equal(t, append(cid[:], id.Bytes()...), key)
}

func TestRepository(t *testing.T) {
	cid := testingidentity.GenerateRandomDID()
	bytes := utils.RandomSlice(identity.DIDLength)
	assert.Equal(t, identity.DIDLength, copy(cid[:], bytes))

	repo := ctx[jobs.BootstrappedRepo].(jobs.Repository)
	tx := jobs.NewJob(cid, "Some transaction")
	assert.NotNil(t, tx.ID)
	assert.NotNil(t, tx.DID)
	assert.Equal(t, jobs.Pending, tx.Status)

	// get tx from repo
	_, err := repo.Get(cid, tx.ID)
	assert.True(t, errors.IsOfType(jobs.ErrJobsMissing, err))

	// save tx into repo
	tx.Status = jobs.Success
	err = repo.Save(tx)
	assert.Nil(t, err)

	// get tx back
	tx, err = repo.Get(cid, tx.ID)
	assert.Nil(t, err)
	assert.NotNil(t, tx)
	assert.Equal(t, cid, tx.DID)
	assert.Equal(t, jobs.Success, tx.Status)
}

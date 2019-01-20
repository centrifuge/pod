// +build unit

package transactions

import (
	"os"
	"testing"

	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/testingutils/commons"

	"github.com/centrifuge/go-centrifuge/storage/leveldb"

	"github.com/centrifuge/go-centrifuge/identity"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testlogging"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/satori/go.uuid"
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
	ctx[identity.BootstrappedIDService] = &testingcommons.MockIDService{}
	bootstrap.RunTestBootstrappers(ibootstappers, ctx)
	result := m.Run()
	bootstrap.RunTestTeardown(ibootstappers)
	os.Exit(result)
}

func Test_getKey(t *testing.T) {
	cid := identity.RandomCentID()
	id := uuid.UUID([16]byte{})

	// empty id
	key, err := getKey(cid, id)
	assert.Nil(t, key)
	assert.Equal(t, "transaction ID is not valid", err.Error())

	id = uuid.Must(uuid.NewV4())
	key, err = getKey(cid, id)
	assert.Nil(t, err)
	assert.Equal(t, append(cid[:], id.Bytes()...), key)
}

func TestRepository(t *testing.T) {
	cid := identity.RandomCentID()
	bytes := utils.RandomSlice(identity.CentIDLength)
	assert.Equal(t, identity.CentIDLength, copy(cid[:], bytes))

	repo := ctx[BootstrappedRepo].(Repository)
	tx := NewTransaction(cid, "Some transaction")
	assert.NotNil(t, tx.ID)
	assert.NotNil(t, tx.CID)
	assert.Equal(t, Pending, tx.Status)

	// get tx from repo
	_, err := repo.Get(cid, tx.ID)
	assert.True(t, errors.IsOfType(ErrTransactionMissing, err))

	// save tx into repo
	tx.Status = Success
	err = repo.Save(tx)
	assert.Nil(t, err)

	// get tx back
	tx, err = repo.Get(cid, tx.ID)
	assert.Nil(t, err)
	assert.NotNil(t, tx)
	assert.Equal(t, cid, tx.CID)
	assert.Equal(t, Success, tx.Status)
}

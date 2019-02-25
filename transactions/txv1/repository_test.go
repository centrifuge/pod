// +build unit

package txv1

import (
	"os"
	"testing"

	"github.com/centrifuge/go-centrifuge/testingutils/identity"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testlogging"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	"github.com/centrifuge/go-centrifuge/testingutils/commons"
	"github.com/centrifuge/go-centrifuge/transactions"
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
	id := transactions.NilTxID()

	// empty id
	key, err := getKey(cid, id)
	assert.Nil(t, key)
	assert.Equal(t, "transaction ID is not valid", err.Error())

	id = transactions.NewTxID()
	key, err = getKey(cid, id)
	assert.Nil(t, err)
	assert.Equal(t, append(cid[:], id.Bytes()...), key)
}

func TestRepository(t *testing.T) {
	cid := testingidentity.GenerateRandomDID()
	bytes := utils.RandomSlice(identity.DIDLength)
	assert.Equal(t, identity.DIDLength, copy(cid[:], bytes))

	repo := ctx[transactions.BootstrappedRepo].(transactions.Repository)
	tx := transactions.NewTransaction(cid, "Some transaction")
	assert.NotNil(t, tx.ID)
	assert.NotNil(t, tx.DID)
	assert.Equal(t, transactions.Pending, tx.Status)

	// get tx from repo
	_, err := repo.Get(cid, tx.ID)
	assert.True(t, errors.IsOfType(transactions.ErrTransactionMissing, err))

	// save tx into repo
	tx.Status = transactions.Success
	err = repo.Save(tx)
	assert.Nil(t, err)

	// get tx back
	tx, err = repo.Get(cid, tx.ID)
	assert.Nil(t, err)
	assert.NotNil(t, tx)
	assert.Equal(t, cid, tx.DID)
	assert.Equal(t, transactions.Success, tx.Status)
}

package transactions

import (
	"os"
	"testing"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/context/testlogging"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

var ctx = map[string]interface{}{}

func TestMain(m *testing.M) {
	ibootstappers := []bootstrap.TestBootstrapper{
		&testlogging.TestLoggingBootstrapper{},
		&config.Bootstrapper{},
		&storage.Bootstrapper{},
		Bootstrapper{},
	}
	bootstrap.RunTestBootstrappers(ibootstappers, ctx)
	result := m.Run()
	bootstrap.RunTestTeardown(ibootstappers)
	os.Exit(result)
}

func Test_getKey(t *testing.T) {
	identity := common.Address([20]byte{})
	id := uuid.UUID([16]byte{})

	// zero address
	key, err := getKey(identity, id)
	assert.Nil(t, key)
	assert.Equal(t, "identity cannot be empty", err.Error())

	bytes := utils.RandomSlice(common.AddressLength)
	assert.Equal(t, common.AddressLength, copy(identity[:], bytes))

	// empty id
	key, err = getKey(identity, id)
	assert.Nil(t, key)
	assert.Equal(t, "transaction ID is not valid", err.Error())

	id = uuid.Must(uuid.NewV4())
	key, err = getKey(identity, id)
	assert.Nil(t, err)
	assert.Equal(t, append(identity[:], id.Bytes()...), key)
}

func TestRepository(t *testing.T) {
	identity := common.Address([20]byte{})
	bytes := utils.RandomSlice(common.AddressLength)
	assert.Equal(t, common.AddressLength, copy(identity[:], bytes))

	repo := ctx[BootstrappedRepo].(Repository)
	tx := NewTransaction(identity, "Some transaction")
	assert.NotNil(t, tx.ID)
	assert.NotNil(t, tx.Identity)
	assert.Equal(t, Pending, tx.Status)

	// get tx from repo
	_, err := repo.Get(identity, tx.ID)
	assert.True(t, errors.IsOfType(ErrTransactionMissing, err))

	// save tx into repo
	tx.Status = Success
	err = repo.Save(tx)
	assert.Nil(t, err)

	// get tx back
	tx, err = repo.Get(identity, tx.ID)
	assert.Nil(t, err)
	assert.NotNil(t, tx)
	assert.Equal(t, identity, tx.Identity)
	assert.Equal(t, Success, tx.Status)
}

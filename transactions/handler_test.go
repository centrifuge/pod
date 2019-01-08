// +build unit

package transactions

import (
	"context"
	"testing"

	"github.com/centrifuge/go-centrifuge/identity"

	"github.com/centrifuge/go-centrifuge/config/configstore"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/transactions"
	"github.com/stretchr/testify/assert"
)

func TestGRPCHandler_GetTransactionStatus(t *testing.T) {
	cService := ctx[configstore.BootstrappedConfigStorage].(configstore.Service)
	h := GRPCHandler(ctx[BootstrappedService].(Service), cService)
	req := new(transactionspb.TransactionStatusRequest)
	ctxl := context.Background()

	// empty ID
	res, err := h.GetTransactionStatus(ctxl, req)
	assert.Nil(t, res)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(ErrInvalidTransactionID, err))

	// invalid ID
	req.TransactionId = "invalid id"
	res, err = h.GetTransactionStatus(ctxl, req)
	assert.Nil(t, res)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(ErrInvalidTransactionID, err))

	// missing err
	tcs, _ := cService.GetAllTenants()
	cid, err := identity.ToCentID(tcs[0].IdentityID)
	tx := NewTransaction(cid, "")
	req.TransactionId = tx.ID.String()
	res, err = h.GetTransactionStatus(ctxl, req)
	assert.Nil(t, res)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(ErrTransactionMissing, err))

	repo := ctx[BootstrappedRepo].(Repository)
	assert.Nil(t, repo.Save(tx))

	// success
	res, err = h.GetTransactionStatus(ctxl, req)
	assert.Nil(t, err)
	assert.Equal(t, tx.ID.String(), res.TransactionId)
	assert.Equal(t, string(tx.Status), res.Status)
}

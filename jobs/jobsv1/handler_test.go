// +build unit

package jobsv1

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/jobs"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/stretchr/testify/assert"
)

func TestGRPCHandler_GetTransactionStatus(t *testing.T) {
	cService := ctx[config.BootstrappedConfigStorage].(config.Service)
	h := GRPCHandler(ctx[jobs.BootstrappedService].(jobs.Manager), cService)
	req := new(jobspb.JobStatusRequest)
	ctxl := testingconfig.HandlerContext(cService)

	// empty ID
	res, err := h.GetJobStatus(ctxl, req)
	assert.Nil(t, res)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(ErrInvalidJobID, err))

	// invalid ID
	req.JobId = "invalid id"
	res, err = h.GetJobStatus(ctxl, req)
	assert.Nil(t, res)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(ErrInvalidJobID, err))

	// missing err
	tcs, _ := cService.GetAllAccounts()
	accID, _ := tcs[0].GetIdentityID()
	cid, err := identity.NewDIDFromBytes(accID)
	assert.NoError(t, err)
	tx := jobs.NewJob(cid, "")
	req.JobId = tx.ID.String()
	res, err = h.GetJobStatus(ctxl, req)
	assert.Nil(t, res)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(jobs.ErrJobsMissing, err))

	repo := ctx[jobs.BootstrappedRepo].(jobs.Repository)
	assert.Nil(t, repo.Save(tx))

	// success
	res, err = h.GetJobStatus(ctxl, req)
	assert.Nil(t, err)
	assert.Equal(t, tx.ID.String(), res.JobId)
	assert.Equal(t, string(tx.Status), res.Status)
}

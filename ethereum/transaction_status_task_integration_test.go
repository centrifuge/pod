// +build integration

package ethereum_test

import (
	"context"
	"testing"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func enqueueJob(t *testing.T, txHash string) (jobs.Manager, identity.DID, jobs.JobID, chan error) {
	queueSrv := ctx[bootstrap.BootstrappedQueueServer].(*queue.Server)
	jobManager := ctx[jobs.BootstrappedService].(jobs.Manager)

	cid := common.BytesToAddress(utils.RandomSlice(20))
	tx, done, err := jobManager.ExecuteWithinJob(context.Background(), identity.NewDID(cid), jobs.NilJobID(),
		"Check TX status",
		func(accountID identity.DID, jobID jobs.JobID, txMan jobs.Manager, errChan chan<- error) {
			result, err := queueSrv.EnqueueJob(ethereum.EthTXStatusTaskName, map[string]interface{}{
				jobs.JobIDParam:                  jobID.String(),
				ethereum.TransactionAccountParam: cid.String(),
				ethereum.TransactionTxHashParam:  txHash,
			})
			if err != nil {
				errChan <- err
			}
			_, err = result.Get(jobManager.GetDefaultTaskTimeout())
			if err != nil {
				errChan <- err
			}
			errChan <- nil
		})
	assert.NoError(t, err)

	return jobManager, identity.NewDID(cid), tx, done
}

func TestTransactionStatusTask_successful(t *testing.T) {
	t.Parallel()
	txManager, cid, tx, result := enqueueJob(t, "0x1")

	r := <-result
	assert.NoError(t, r)
	trans, err := txManager.GetJob(cid, tx)
	assert.Nil(t, err, "a transaction should be returned")
	assert.Equal(t, string(jobs.Success), string(trans.Status), "transaction should be successful")
}

func TestTransactionStatusTask_failed(t *testing.T) {
	t.Parallel()
	txManager, cid, tx, result := enqueueJob(t, "0x2")

	r := <-result
	assert.Error(t, r)
	trans, err := txManager.GetJob(cid, tx)
	assert.Nil(t, err, "a  centrifuge transaction should be  returned")
	assert.Equal(t, string(jobs.Failed), string(trans.Status), "transaction should fail")
}

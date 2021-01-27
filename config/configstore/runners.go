package configstore

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs/jobsv2"
	"github.com/centrifuge/gocelery/v2"
	"github.com/ethereum/go-ethereum/common"
)

const (
	generateIdentityRunnerName = "GenerateIdentity"
	taskSendTxn                = "Create identity on Chain"
	taskWaitForTxn             = "Wait for Identity transaction to be included"
	txnHash                    = "transaction hash"
)

// generateIdentityRunner does the following
// Send txn to
type generateIdentityRunner struct {
	idFactory identity.FactoryInterface
	ethClient ethereum.Client
	repo      Repository
}

func (g generateIdentityRunner) New() gocelery.Runner {
	return generateIdentityRunner{
		idFactory: g.idFactory,
		repo:      g.repo,
		ethClient: g.ethClient,
	}
}

func (g generateIdentityRunner) RunnerFunc(task string) gocelery.RunnerFunc {
	switch task {
	case taskSendTxn:
		return g.sendTxn
	default:
		return g.checkForTxn
	}
}

func (g generateIdentityRunner) Next(task string) (next string, ok bool) {
	switch task {
	case taskSendTxn:
		return taskWaitForTxn, true
	default:
		return "", false
	}
}

func (g generateIdentityRunner) sendTxn(args []interface{}, overrides map[string]interface{}) (result interface{}, err error) {
	did := args[0].(identity.DID)
	acc, err := g.repo.GetAccount(did[:])
	if err != nil {
		return nil, fmt.Errorf("failed to fetch account from repo: %w", err)
	}

	keys, err := acc.GetKeys()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch keys from the account: %w", err)
	}

	idKeys, err := identity.ConvertAccountKeysToKeyDID(keys)
	if err != nil {
		return nil, fmt.Errorf("failed to convert keys: %w", err)
	}

	txn, err := g.idFactory.CreateIdentity(
		acc.GetEthereumDefaultAccountName(), common.HexToAddress(acc.GetEthereumAccount().Address), idKeys)
	if err != nil {
		return nil, fmt.Errorf("failed to send txn to create identity: %w", err)
	}

	overrides[txnHash] = txn.Hash()
	return nil, nil
}

func (g generateIdentityRunner) checkForTxn(args []interface{}, overrides map[string]interface{}) (result interface{}, err error) {
	did := args[0].(identity.DID)
	txHash, ok := overrides[txnHash].(common.Hash)
	if !ok {
		return nil, errors.New("failed to find the txn hash")
	}

	ctx := context.Background()
	createdDID, err := ethereum.IsTxnSuccessful(ctx, g.ethClient, txHash)
	if err != nil {
		return nil, err
	}

	if !bytes.Equal(createdDID[:], did[:]) {
		return nil, fmt.Errorf("identity mismatch. probably concurrent txn")
	}

	return nil, nil
}

// StartGenerateIdentityJob starts a new job that creates the provided identity on chain
// account must be already stored in the repo
func StartGenerateIdentityJob(
	did identity.DID, dispatcher jobsv2.Dispatcher, validUntil time.Time) (jobID []byte, err error) {
	job := gocelery.NewRunnerJob(
		"Create identity on chain",
		generateIdentityRunnerName,
		taskSendTxn, []interface{}{did}, make(map[string]interface{}), validUntil)
	_, err = dispatcher.Dispatch(did, job)
	return job.ID, err
}

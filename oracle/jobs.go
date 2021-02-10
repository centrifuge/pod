package oracle

import (
	"context"
	"encoding/gob"
	"fmt"
	"time"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/nft"
	"github.com/centrifuge/gocelery/v2"
	"github.com/ethereum/go-ethereum/common"
)

func init() {
	gob.Register([32]byte{})
}

const oraclePushJob = "Push to Oracle"

// PushToOracleJob pushes nft value to oracle
// args are as follows
// did, oracleAddr, tokenID, fingerprint, value
type PushToOracleJob struct {
	jobs.Base
	accountsSrv     config.Service
	identityService identity.Service
	ethClient       ethereum.Client
}

// New returns a new PushToOracleJob instance
func (p *PushToOracleJob) New() gocelery.Runner {
	np := &PushToOracleJob{
		accountsSrv:     p.accountsSrv,
		identityService: p.identityService,
		ethClient:       p.ethClient,
	}

	np.Base = jobs.NewBase(np.getTasks())
	return np
}

func (p *PushToOracleJob) getTasks() map[string]jobs.Task {
	return map[string]jobs.Task{
		"push_to_oracle": {
			RunnerFunc: func(args []interface{}, overrides map[string]interface{}) (result interface{}, err error) {
				did := args[0].(identity.DID)
				acc, err := p.accountsSrv.GetAccount(did[:])
				if err != nil {
					return nil, fmt.Errorf("failed to get account: %w", err)
				}

				ctx := contextutil.WithAccount(context.Background(), acc)
				oracleAddr := args[1].(common.Address)

				// to tokenId *big.Int, bytes32, bytes32
				txn, err := p.identityService.ExecuteAsync(ctx, oracleAddr, updateABI, "update", args[2:]...)
				if err != nil {
					return nil, fmt.Errorf("failed to send oracle txn: %w", err)
				}

				log.Infof("Sent nft details to oracle[%s] with txn[%s]", oracleAddr, txn.Hash())
				overrides["eth_txn"] = txn.Hash()
				return nil, nil
			},
			Next: "wait_for_txn",
		},

		"wait_for_txn": {
			RunnerFunc: func(args []interface{}, overrides map[string]interface{}) (result interface{}, err error) {
				txn := overrides["eth_txn"].(common.Hash)
				_, err = ethereum.IsTxnSuccessful(context.Background(), p.ethClient, txn)
				if err != nil {
					return nil, err
				}

				log.Infof("Document value successfully pushed to Oracle with TX hash: %v\n", txn.Hex())
				return nil, nil
			},
		},
	}
}

func initOraclePushJob(
	dispatcher jobs.Dispatcher,
	did identity.DID, oracleAddr common.Address,
	tokenID nft.TokenID, fp, value [32]byte) (gocelery.JobID, error) {
	job := gocelery.NewRunnerJob(
		oraclePushJob, oraclePushJob, "push_to_oracle",
		[]interface{}{did, oracleAddr, tokenID.BigInt(), fp, value}, make(map[string]interface{}), time.Time{})

	_, err := dispatcher.Dispatch(did, job)
	return job.ID, err
}

// +build integration unit testworld

package testingidentity

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs/jobsv2"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/gocelery/v2"
	"github.com/ethereum/go-ethereum/common"
)

func CreateAccountIDWithKeys(
	contextTimeout time.Duration,
	acc *configstore.Account,
	idService identity.Service,
	idFactory identity.Factory,
	client ethereum.Client,
	dispatcher jobsv2.Dispatcher) (identity.DID, error) {
	ctxh, _ := contextutil.New(context.Background(), acc)
	idKeys, err := acc.GetKeys()
	if err != nil {
		return identity.DID{}, err
	}

	// only create identity if it doesn't exist
	did := identity.NewDID(common.BytesToAddress(acc.IdentityID))
	err = idService.Exists(ctxh, did)
	if err != nil {
		txn, err := idFactory.CreateIdentity(acc.GetEthereumDefaultAccountName(), nil)
		if err != nil {
			return did, err
		}

		ok := dispatcher.RegisterRunnerFunc("ethWaitTxn", func([]interface{}, map[string]interface{}) (interface{}, error) {
			return ethereum.IsTxnSuccessful(context.Background(), client.GetEthClient(), txn.Hash())
		})

		if !ok {
			return did, errors.New("failed to register worker")
		}

		job := gocelery.NewRunnerFuncJob("Wait for Identity creation", "ethWaitTxn", nil, nil, time.Time{})
		res, err := dispatcher.Dispatch(did, job)
		if err != nil {
			return did, fmt.Errorf("failed to dispatch identity create job: %w", err)
		}

		_, err = res.Await(context.Background())
		if err != nil {
			return did, fmt.Errorf("identity creation failed: %w", err)
		}

		// LookupIdentityForId
		err = idService.Exists(ctxh, did)
		if err != nil {
			return did, fmt.Errorf("identity creation failed: %w", err)
		}
	}

	// Add Action key if it doesn't exist
	keys, err := idService.GetKeysByPurpose(did, &(identity.KeyPurposeAction.Value))
	if err != nil {
		return identity.DID{}, err
	}
	ctx, cancel1 := defaultWaitForTransactionMiningContext(contextTimeout)
	ctxh, err = contextutil.New(ctx, acc)
	if err != nil {
		return identity.DID{}, err
	}
	defer cancel1()
	if len(keys) == 0 {
		pk, _ := utils.SliceToByte32(idKeys[identity.KeyPurposeAction.Name].PublicKey)
		keyDID := identity.NewKey(pk, &(identity.KeyPurposeAction.Value), big.NewInt(identity.KeyTypeECDSA), 0)
		err = idService.AddKey(ctxh, keyDID)
		if err != nil {
			return identity.DID{}, err
		}
	}

	// Add Signing key if it doesn't exist
	keys, err = idService.GetKeysByPurpose(did, &(identity.KeyPurposeSigning.Value))
	if err != nil {
		return identity.DID{}, err
	}
	ctx, cancel2 := defaultWaitForTransactionMiningContext(contextTimeout)
	ctxh, _ = contextutil.New(ctx, acc)
	defer cancel2()
	if len(keys) == 0 {
		pk, _ := utils.SliceToByte32(idKeys[identity.KeyPurposeSigning.Name].PublicKey)
		keyDID := identity.NewKey(pk, &(identity.KeyPurposeSigning.Value), big.NewInt(identity.KeyTypeECDSA), 0)
		err = idService.AddKey(ctxh, keyDID)
		if err != nil {
			return identity.DID{}, err
		}
	}

	return did, nil
}

func GenerateRandomDID() identity.DID {
	r := utils.RandomSlice(identity.DIDLength)
	did, _ := identity.NewDIDFromBytes(r)
	return did
}

// defaultWaitForTransactionMiningContext returns context with timeout for write operations
func defaultWaitForTransactionMiningContext(d time.Duration) (ctx context.Context, cancelFunc context.CancelFunc) {
	toBeDone := time.Now().Add(d)
	return context.WithDeadline(context.Background(), toBeDone)
}

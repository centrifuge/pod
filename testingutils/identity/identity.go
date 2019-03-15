// +build integration unit testworld

package testingidentity

import (
	"context"
	"math/big"
	"time"

	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/utils"
)

func CreateAccountIDWithKeys(contextTimeout time.Duration, acc *configstore.Account, idService identity.ServiceDID, idFactory identity.Factory) (identity.DID, error) {
	ctxh, _ := contextutil.New(context.Background(), acc)
	idKeys, err := acc.GetKeys()
	if err != nil {
		return identity.DID{}, err
	}
	var did *identity.DID
	// only create identity if it doesn't exist
	err = idService.Exists(ctxh, identity.NewDIDFromBytes(acc.IdentityID))
	if err != nil {
		did, err = idFactory.CreateIdentity(ctxh)
		if err != nil {
			return identity.DID{}, err
		}
		// LookupIdentityForId
		err = idService.Exists(ctxh, *did)
	}

	// only add key if it doesn't exist
	ctx, cancel := defaultWaitForTransactionMiningContext(contextTimeout)
	ctxh, err = contextutil.New(ctx, acc)
	if err != nil {
		return identity.DID{}, nil
	}
	keys, err := idService.GetKeysByPurpose(*did, &(identity.KeyPurposeSigning.Value))
	ctx, cancel = defaultWaitForTransactionMiningContext(contextTimeout)
	ctxh, _ = contextutil.New(ctx, acc)
	defer cancel()
	if err != nil || len(keys) == 0 {
		pk, _ := utils.SliceToByte32(idKeys[identity.KeyPurposeSigning.Name].PublicKey)
		keyDID := identity.NewKey(pk, &(identity.KeyPurposeSigning.Value), big.NewInt(identity.KeyTypeECDSA))
		err = idService.AddKey(ctxh, keyDID)
		if err != nil {
			return identity.DID{}, nil
		}
	}

	return *did, nil
}

func GenerateRandomDID() identity.DID {
	r := utils.RandomSlice(identity.DIDLength)
	return identity.NewDIDFromBytes(r)
}

// defaultWaitForTransactionMiningContext returns context with timeout for write operations
func defaultWaitForTransactionMiningContext(d time.Duration) (ctx context.Context, cancelFunc context.CancelFunc) {
	toBeDone := time.Now().Add(d)
	return context.WithDeadline(context.Background(), toBeDone)
}

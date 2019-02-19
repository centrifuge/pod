// +build integration unit

package testingidentity

import (
	"context"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"time"

	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/contextutil"

	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"
)

func CreateAccountIDWithKeys(contextTimeout time.Duration, acc *configstore.Account, idService identity.ServiceDID, idFactory identity.Factory) (identity.DID, error) {
	ctxh, _ := contextutil.New(context.Background(), acc)
	idConfig, _ := identity.GetIdentityConfig(acc)
	var did *identity.DID
	// only create identity if it doesn't exist
	err := idService.Exists(ctxh, identity.NewDIDFromBytes(acc.IdentityID))
	if err != nil {
		did, err = idFactory.CreateIdentity(ctxh)
		if err != nil {
			return identity.DID{}, err
		}
		// LookupIdentityForId
		err = idService.Exists(ctxh, *did)
	}

	// only add key if it doesn't exist
	keys, err := idService.GetKeysByPurpose(*did, big.NewInt(identity.KeyPurposeEthMsgAuth))
	ctx, cancel := ethereum.DefaultWaitForTransactionMiningContext(contextTimeout)
	ctxh, _ = contextutil.New(ctx, acc)
	defer cancel()
	if err != nil || len(keys) == 0 {
		pk, _ := utils.SliceToByte32(idConfig.Keys[identity.KeyPurposeEthMsgAuth].PublicKey)
		keyDID := identity.NewKey(pk, big.NewInt(identity.KeyPurposeEthMsgAuth), big.NewInt(identity.KeyTypeECDSA))
		_ = idService.AddKey(ctxh, keyDID)
	}
	keys, err = idService.GetKeysByPurpose(*did, big.NewInt(identity.KeyPurposeSigning))
	ctx, cancel = ethereum.DefaultWaitForTransactionMiningContext(contextTimeout)
	ctxh, _ = contextutil.New(ctx, acc)
	defer cancel()
	if err != nil {
		pk, _ := utils.SliceToByte32(idConfig.Keys[identity.KeyPurposeSigning].PublicKey)
		keyDID := identity.NewKey(pk, big.NewInt(identity.KeyPurposeSigning), big.NewInt(identity.KeyTypeECDSA))
		_ = idService.AddKey(ctxh, keyDID)
	}

	return *did, nil
}

func GenerateRandomDID() identity.DID {
	r := utils.RandomSlice(common.AddressLength)
	return identity.NewDIDFromBytes(r)
}

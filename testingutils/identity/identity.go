// +build integration unit

package testingidentity

import (
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/keytools/ed25519"
	"github.com/centrifuge/go-centrifuge/keytools/secp256k1"
)

func CreateIdentityWithKeys() identity.CentID {
	idConfigEth, _ := secp256k1.GetIDConfig()
	idConfig, _ := ed25519.GetIDConfig()
	centIdTyped, _ := identity.ToCentID(idConfigEth.ID)
	// only create identity if it doesn't exist
	id, err := identity.IDService.LookupIdentityForID(centIdTyped)
	if err != nil {
		_, confirmations, _ := identity.IDService.CreateIdentity(centIdTyped)
		<-confirmations
		// LookupIdentityForId
		id, _ = identity.IDService.LookupIdentityForID(centIdTyped)
	}

	// only add key if it doesn't exist
	_, err = id.GetLastKeyForPurpose(identity.KeyPurposeEthMsgAuth)
	ctx, cancel := ethereum.DefaultWaitForTransactionMiningContext()
	defer cancel()
	if err != nil {
		confirmations, _ := id.AddKeyToIdentity(ctx, identity.KeyPurposeEthMsgAuth, idConfigEth.PublicKey)
		<-confirmations
	}
	_, err = id.GetLastKeyForPurpose(identity.KeyPurposeSigning)
	ctx, cancel = ethereum.DefaultWaitForTransactionMiningContext()
	defer cancel()
	if err != nil {
		confirmations, _ := id.AddKeyToIdentity(ctx, identity.KeyPurposeSigning, idConfig.PublicKey)
		<-confirmations
	}

	return centIdTyped
}

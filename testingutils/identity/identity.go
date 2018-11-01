// +build integration unit

package testingidentity

import (
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"
)

func CreateIdentityWithKeys() identity.CentID {
	idConfig, _ := identity.GetIdentityConfig()
	// only create identity if it doesn't exist
	id, err := identity.IDService.LookupIdentityForID(idConfig.ID)
	if err != nil {
		_, confirmations, _ := identity.IDService.CreateIdentity(idConfig.ID)
		<-confirmations
		// LookupIdentityForId
		id, _ = identity.IDService.LookupIdentityForID(idConfig.ID)
	}

	// only add key if it doesn't exist
	_, err = id.GetLastKeyForPurpose(identity.KeyPurposeEthMsgAuth)
	ctx, cancel := ethereum.DefaultWaitForTransactionMiningContext()
	defer cancel()
	if err != nil {
		confirmations, _ := id.AddKeyToIdentity(ctx, identity.KeyPurposeEthMsgAuth, idConfig.Keys[identity.KeyPurposeEthMsgAuth].PublicKey)
		<-confirmations
	}
	_, err = id.GetLastKeyForPurpose(identity.KeyPurposeSigning)
	ctx, cancel = ethereum.DefaultWaitForTransactionMiningContext()
	defer cancel()
	if err != nil {
		confirmations, _ := id.AddKeyToIdentity(ctx, identity.KeyPurposeSigning, idConfig.Keys[identity.KeyPurposeSigning].PublicKey)
		<-confirmations
	}

	return idConfig.ID
}

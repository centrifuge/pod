// +build integration unit

package testingidentity

import (
	"context"
	"testing"
	"time"

	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/contextutil"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
)

func CreateIdentityWithKeys(cfg config.Configuration, idService identity.Service) identity.CentID {
	idConfig, _ := identity.GetIdentityConfig(cfg)
	// only create identity if it doesn't exist
	id, err := idService.LookupIdentityForID(idConfig.ID)
	if err != nil {
		_, confirmations, _ := idService.CreateIdentity(testingconfig.CreateAccountContext(&testing.T{}, cfg), idConfig.ID)
		<-confirmations
		// LookupIdentityForId
		id, _ = idService.LookupIdentityForID(idConfig.ID)
	}

	// only add key if it doesn't exist
	_, err = id.LastKeyForPurpose(identity.KeyPurposeEthMsgAuth)
	ctx, cancel := ethereum.DefaultWaitForTransactionMiningContext(cfg.GetEthereumContextWaitTimeout())
	defer cancel()
	if err != nil {
		confirmations, _ := id.AddKeyToIdentity(ctx, identity.KeyPurposeEthMsgAuth, idConfig.Keys[identity.KeyPurposeEthMsgAuth].PublicKey)
		<-confirmations
	}
	_, err = id.LastKeyForPurpose(identity.KeyPurposeSigning)
	ctx, cancel = ethereum.DefaultWaitForTransactionMiningContext(cfg.GetEthereumContextWaitTimeout())
	defer cancel()
	if err != nil {
		confirmations, _ := id.AddKeyToIdentity(ctx, identity.KeyPurposeSigning, idConfig.Keys[identity.KeyPurposeSigning].PublicKey)
		<-confirmations
	}

	return idConfig.ID
}

func CreateAccountIDWithKeys(contextTimeout time.Duration, cfg *configstore.Account, idService identity.Service) identity.Identity {
	ctxh, _ := contextutil.New(context.Background(), cfg)
	idConfig, _ := identity.GetIdentityConfig(cfg)
	// only create identity if it doesn't exist
	id, err := idService.LookupIdentityForID(idConfig.ID)
	if err != nil {
		_, confirmations, _ := idService.CreateIdentity(ctxh, idConfig.ID)
		<-confirmations
		// LookupIdentityForId
		id, _ = idService.LookupIdentityForID(idConfig.ID)
	}

	// only add key if it doesn't exist
	_, err = id.LastKeyForPurpose(identity.KeyPurposeEthMsgAuth)
	ctx, cancel := ethereum.DefaultWaitForTransactionMiningContext(contextTimeout)
	defer cancel()
	if err != nil {
		confirmations, _ := id.AddKeyToIdentity(ctx, identity.KeyPurposeEthMsgAuth, idConfig.Keys[identity.KeyPurposeEthMsgAuth].PublicKey)
		<-confirmations
	}
	_, err = id.LastKeyForPurpose(identity.KeyPurposeSigning)
	ctx, cancel = ethereum.DefaultWaitForTransactionMiningContext(contextTimeout)
	defer cancel()
	if err != nil {
		confirmations, _ := id.AddKeyToIdentity(ctx, identity.KeyPurposeSigning, idConfig.Keys[identity.KeyPurposeSigning].PublicKey)
		<-confirmations
	}

	return id
}

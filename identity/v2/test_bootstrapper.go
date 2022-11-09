//go:build integration || testworld

package v2

import (
	"context"
	"errors"
	"fmt"

	keystoreTypes "github.com/centrifuge/chain-custom-types/pkg/keystore"
	proxyType "github.com/centrifuge/chain-custom-types/pkg/proxy"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/crypto"
	"github.com/centrifuge/go-centrifuge/pallets/keystore"
	"github.com/centrifuge/go-centrifuge/pallets/proxy"
	genericUtils "github.com/centrifuge/go-centrifuge/testingutils/generic"
	"github.com/centrifuge/go-centrifuge/testingutils/keyrings"
	proxyUtils "github.com/centrifuge/go-centrifuge/testingutils/proxy"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	logging "github.com/ipfs/go-log"
)

var (
	log = logging.Logger("identity_test_bootstrap")
)

func (b *Bootstrapper) TestBootstrap(serviceCtx map[string]any) error {
	return b.Bootstrap(serviceCtx)
}

func (b *Bootstrapper) TestTearDown() error {
	return nil
}

type AccountTestBootstrapper struct {
	Bootstrapper
}

func (b *AccountTestBootstrapper) TestBootstrap(serviceCtx map[string]any) error {
	if err := b.Bootstrap(serviceCtx); err != nil {
		return err
	}

	log.Info("Generating test account for Alice")

	_, err := BootstrapTestAccount(serviceCtx, keyrings.AliceKeyRingPair)

	if err != nil {
		return fmt.Errorf("couldn't bootstrap test account for Alice: %w", err)
	}

	return nil
}

func (b *AccountTestBootstrapper) TestTearDown() error {
	return nil
}

func BootstrapTestAccount(
	serviceCtx map[string]any,
	accountKeyringPair signature.KeyringPair,
) (config.Account, error) {
	accountID, err := types.NewAccountID(accountKeyringPair.PublicKey)

	if err != nil {
		return nil, fmt.Errorf("couldn't get account ID: %w", err)
	}

	acc, err := CreateTestIdentity(serviceCtx, accountID, "")

	if err != nil {
		return nil, fmt.Errorf("couldn't create test account: %w", err)
	}

	cfgService := genericUtils.GetService[config.Service](serviceCtx)

	podOperator, err := cfgService.GetPodOperator()

	if err != nil {
		return nil, fmt.Errorf("couldn't get pod operator: %w", err)
	}

	proxyPairs := ProxyPairs{
		{
			Delegate:  podOperator.GetAccountID(),
			ProxyType: proxyType.PodOperation,
		},
		{
			Delegate:  podOperator.GetAccountID(),
			ProxyType: proxyType.KeystoreManagement,
		},
	}

	if err := AddAndWaitForTestProxies(serviceCtx, accountKeyringPair, proxyPairs); err != nil {
		return nil, fmt.Errorf("couldn't create test proxies: %w", err)
	}

	if err := AddAccountKeysToStore(serviceCtx, acc); err != nil {
		return nil, fmt.Errorf("couldn't add keys to keystore: %w", err)
	}

	return acc, nil
}

func CreateTestIdentity(
	serviceCtx map[string]any,
	accountID *types.AccountID,
	webhookURL string,
) (config.Account, error) {
	cfgService := genericUtils.GetService[config.Service](serviceCtx)
	identityService := genericUtils.GetService[Service](serviceCtx)

	if acc, err := cfgService.GetAccount(accountID.ToBytes()); err == nil {
		log.Info("Account already created for - ", accountID.ToHexString())

		return acc, nil
	}

	acc, err := identityService.CreateIdentity(context.Background(), &CreateIdentityRequest{
		Identity:         accountID,
		WebhookURL:       webhookURL,
		PrecommitEnabled: true,
	})

	if err != nil {
		return nil, fmt.Errorf("couldn't create identity: %w", err)
	}

	return acc, nil
}

type ProxyPair struct {
	Delegate  *types.AccountID
	ProxyType proxyType.CentrifugeProxyType
}

type ProxyPairs []ProxyPair

func (p ProxyPairs) GetDelegateAccountIDs() []*types.AccountID {
	accountIDMap := make(map[string]struct{})

	var accountIDs []*types.AccountID

	for _, proxyPair := range p {
		if _, ok := accountIDMap[proxyPair.Delegate.ToHexString()]; ok {
			continue
		}

		accountIDMap[proxyPair.Delegate.ToHexString()] = struct{}{}

		accountIDs = append(accountIDs, proxyPair.Delegate)
	}

	return accountIDs
}

func AddAndWaitForTestProxies(
	serviceCtx map[string]any,
	delegatorKrp signature.KeyringPair,
	proxyPairs ProxyPairs,
) error {
	proxyAPI := genericUtils.GetService[proxy.API](serviceCtx)

	delegator, err := types.NewAccountID(delegatorKrp.PublicKey)

	if err != nil {
		return fmt.Errorf("couldn't create delegator account ID: %w", err)
	}

	ctx := context.Background()

	for _, proxyPair := range proxyPairs {
		if err := proxyAPI.AddProxy(ctx, proxyPair.Delegate, proxyPair.ProxyType, 0, delegatorKrp); err != nil {
			return fmt.Errorf("couldn't add proxy to %s: %w", delegator.ToHexString(), err)
		}
	}

	err = proxyUtils.WaitForProxiesToBeAdded(
		ctx,
		serviceCtx,
		delegator,
		proxyPairs.GetDelegateAccountIDs()...,
	)

	if err != nil {
		return fmt.Errorf("proxies were not added: %w", err)
	}

	return nil
}

func AddAccountKeysToStore(
	serviceCtx map[string]any,
	acc config.Account,
) error {
	unstoredAccountKeys, err := getUnstoredAccountKeys(serviceCtx, acc)
	if err != nil {
		return fmt.Errorf("couldn't get account keys: %w", err)
	}

	keystoreAPI := genericUtils.GetService[keystore.API](serviceCtx)

	var keys []*keystoreTypes.AddKey

	for _, unstoredAccountKey := range unstoredAccountKeys {
		keys = append(keys, &keystoreTypes.AddKey{
			Key:     unstoredAccountKey.Hash,
			Purpose: unstoredAccountKey.KeyPurpose,
			KeyType: keystoreTypes.KeyTypeECDSA,
		})
	}

	_, err = keystoreAPI.AddKeys(contextutil.WithAccount(context.Background(), acc), keys)
	if err != nil {
		return fmt.Errorf("couldn't store keys: %w", err)
	}

	return nil
}

func getUnstoredAccountKeys(
	serviceCtx map[string]any,
	acc config.Account,
) ([]*keystoreTypes.KeyID, error) {
	cfgService := genericUtils.GetService[config.Service](serviceCtx)
	cfg, err := cfgService.GetConfig()

	if err != nil {
		return nil, fmt.Errorf("couldn't get config: %w", err)
	}

	_, p2pPublicKey, err := crypto.ObtainP2PKeypair(cfg.GetP2PKeyPair())

	if err != nil {
		return nil, fmt.Errorf("couldn't obtain P2P key pair: %w", err)
	}

	p2pPublicKeyRaw, err := p2pPublicKey.Raw()

	if err != nil {
		return nil, fmt.Errorf("couldn't get raw P2P public key: %w", err)
	}

	keys := []*keystoreTypes.KeyID{
		{
			Hash:       types.NewHash(p2pPublicKeyRaw),
			KeyPurpose: keystoreTypes.KeyPurposeP2PDiscovery,
		},
		{
			Hash:       types.NewHash(acc.GetSigningPublicKey()),
			KeyPurpose: keystoreTypes.KeyPurposeP2PDocumentSigning,
		},
	}

	return filterUnstoredAccountKeys(serviceCtx, acc.GetIdentity(), keys)
}

func filterUnstoredAccountKeys(serviceCtx map[string]any, accountID *types.AccountID, keys []*keystoreTypes.KeyID) ([]*keystoreTypes.KeyID, error) {
	keystoreAPI := genericUtils.GetService[keystore.API](serviceCtx)

	return genericUtils.FilterSlice(keys, func(key *keystoreTypes.KeyID) (bool, error) {
		_, err := keystoreAPI.GetKey(accountID, key)

		if err != nil {
			if errors.Is(err, keystore.ErrKeyNotFound) {
				return true, nil
			}

			return false, err
		}

		return false, nil
	})
}

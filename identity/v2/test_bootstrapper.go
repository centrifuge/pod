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

	acc, err := CreateTestAccount(serviceCtx, accountID, "")

	if err != nil {
		return nil, fmt.Errorf("couldn't create test account: %w", err)
	}

	cfgService := genericUtils.GetService[config.Service](serviceCtx)

	podOperator, err := cfgService.GetPodOperator()

	if err != nil {
		return nil, fmt.Errorf("couldn't get pod operator: %w", err)
	}

	proxyPairs := []ProxyPair{
		{
			Delegate:  podOperator.GetAccountID(),
			ProxyType: proxyType.PodOperation,
		},
		{
			Delegate:  podOperator.GetAccountID(),
			ProxyType: proxyType.KeystoreManagement,
		},
	}

	if err := AddTestProxies(serviceCtx, accountKeyringPair, proxyPairs...); err != nil {
		return nil, fmt.Errorf("couldn't create test proxies: %w", err)
	}

	if err := AddAccountKeysToStore(serviceCtx, acc); err != nil {
		return nil, fmt.Errorf("couldn't add keys to keystore: %w", err)
	}

	return acc, nil
}

func CreateTestAccount(
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

func AddTestProxies(
	serviceCtx map[string]any,
	delegatorKrp signature.KeyringPair,
	proxyPairs ...ProxyPair,
) error {
	proxyAPI := genericUtils.GetService[proxy.API](serviceCtx)

	delegator, err := types.NewAccountID(delegatorKrp.PublicKey)

	if err != nil {
		return fmt.Errorf("couldn't create delegator account ID: %w", err)
	}

	_, err = proxyAPI.GetProxies(delegator)

	if err != nil && !errors.Is(err, proxy.ErrProxiesNotFound) {
		return fmt.Errorf("couldn't retrieve delegator proxies: %w", err)
	}

	ctx := context.Background()

	for _, proxyPair := range proxyPairs {
		if err := proxyAPI.AddProxy(ctx, proxyPair.Delegate, proxyPair.ProxyType, 0, delegatorKrp); err != nil {
			return fmt.Errorf("couldn't add proxy to %s: %w", delegator.ToHexString(), err)
		}
	}

	return nil
}

func AddAccountKeysToStore(
	serviceCtx map[string]any,
	acc config.Account,
) error {
	cfgService := genericUtils.GetService[config.Service](serviceCtx)
	keystoreAPI := genericUtils.GetService[keystore.API](serviceCtx)

	err := addKeyIfNotPresent(keystoreAPI, acc, acc.GetSigningPublicKey(), keystoreTypes.KeyPurposeP2PDocumentSigning)

	if err != nil {
		return fmt.Errorf("couldn't add document signing key to keystore: %w", err)
	}

	cfg, err := cfgService.GetConfig()

	if err != nil {
		return fmt.Errorf("couldn't get config: %w", err)
	}

	_, P2PPublicKey, err := crypto.ObtainP2PKeypair(cfg.GetP2PKeyPair())

	if err != nil {
		return fmt.Errorf("couldn't obtain P2P key pair: %w", err)
	}

	P2PPublicKeyRaw, err := P2PPublicKey.Raw()

	if err != nil {
		return fmt.Errorf("couldn't get raw public key: %w", err)
	}

	err = addKeyIfNotPresent(keystoreAPI, acc, P2PPublicKeyRaw, keystoreTypes.KeyPurposeP2PDiscovery)

	if err != nil {
		return fmt.Errorf("couldn't add P2P discovery key to keystore: %w", err)
	}

	return nil
}

func addKeyIfNotPresent(keystoreAPI keystore.API, acc config.Account, key []byte, keyPurpose keystoreTypes.KeyPurpose) error {
	ctx := context.Background()

	keyHash := types.NewHash(key)

	_, err := keystoreAPI.GetKey(
		acc.GetIdentity(),
		&keystoreTypes.KeyID{
			Hash:       keyHash,
			KeyPurpose: keyPurpose,
		},
	)

	if err == nil {
		return nil
	}

	_, err = keystoreAPI.AddKeys(
		contextutil.WithAccount(ctx, acc),
		[]*keystoreTypes.AddKey{
			{
				Key:     keyHash,
				Purpose: keyPurpose,
				KeyType: keystoreTypes.KeyTypeECDSA,
			},
		},
	)

	if err != nil {
		return fmt.Errorf("couldn't add key to keystore: %w", err)
	}

	return nil
}

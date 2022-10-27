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
	"github.com/centrifuge/go-centrifuge/pallets"
	"github.com/centrifuge/go-centrifuge/pallets/keystore"
	"github.com/centrifuge/go-centrifuge/pallets/proxy"
	"github.com/centrifuge/go-centrifuge/testingutils/keyrings"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	logging "github.com/ipfs/go-log"
)

var (
	log = logging.Logger("identity_test_bootstrap")
)

func (b *Bootstrapper) TestBootstrap(serviceCtx map[string]any) error {
	if err := b.bootstrap(serviceCtx); err != nil {
		return err
	}

	log.Info("Generating test account for Alice")

	_, err := BootstrapTestAccount(serviceCtx, keyrings.AliceKeyRingPair)

	if err != nil {
		return fmt.Errorf("couldn't bootstrap test account for Alice: %w", err)
	}

	return nil
}

func (b *Bootstrapper) TestTearDown() error {
	return nil
}

func BootstrapTestAccount(
	serviceCtx map[string]any,
	accountKeyringPair signature.KeyringPair,
) (config.Account, error) {
	cfgService, ok := serviceCtx[config.BootstrappedConfigStorage].(config.Service)

	if !ok {
		return nil, errors.New("config service not initialised")
	}

	proxyAPI, ok := serviceCtx[pallets.BootstrappedProxyAPI].(proxy.API)

	if !ok {
		return nil, errors.New("proxy API not initialised")
	}

	keystoreAPI, ok := serviceCtx[pallets.BootstrappedKeystoreAPI].(keystore.API)

	if !ok {
		return nil, errors.New("keystore API not initialised")
	}

	identityService, ok := serviceCtx[BootstrappedIdentityServiceV2].(Service)

	if !ok {
		return nil, errors.New("identity API not initialised")
	}

	accountID, err := types.NewAccountID(accountKeyringPair.PublicKey)

	if err != nil {
		return nil, fmt.Errorf("couldn't get account ID: %w", err)
	}

	acc, err := createTestAccount(cfgService, identityService, accountID)

	if err != nil {
		return nil, fmt.Errorf("couldn't create test account: %w", err)
	}

	if err := addTestProxies(cfgService, proxyAPI, accountKeyringPair); err != nil {
		return nil, fmt.Errorf("couldn't create test proxies: %w", err)
	}

	if err := addKeysToStore(cfgService, keystoreAPI, acc); err != nil {
		return nil, fmt.Errorf("couldn't add keys to keystore: %w", err)
	}

	return acc, nil
}

func createTestAccount(cfgService config.Service, identityService Service, accountID *types.AccountID) (config.Account, error) {
	if acc, err := cfgService.GetAccount(accountID.ToBytes()); err == nil {
		log.Info("Account already created for -", accountID.ToHexString())

		return acc, nil
	}

	acc, err := identityService.CreateIdentity(context.Background(), &CreateIdentityRequest{
		Identity:         accountID,
		WebhookURL:       "",
		PrecommitEnabled: true,
	})

	if err != nil {
		return nil, fmt.Errorf("couldn't create identity: %w", err)
	}

	return acc, nil
}

func addTestProxies(cfgService config.Service, proxyAPI proxy.API, krp signature.KeyringPair) error {
	delegator, err := types.NewAccountID(krp.PublicKey)

	if err != nil {
		return fmt.Errorf("couldn't create delegator account ID: %w", err)
	}

	res, err := proxyAPI.GetProxies(delegator)

	switch {
	case err != nil && !errors.Is(err, proxy.ErrProxiesNotFound):
		return fmt.Errorf("couldn't retrieve delegator proxies: %w", err)
	case res != nil:
		log.Infof("Account %s has %d test proxies", delegator.ToHexString(), len(res.ProxyDefinitions))

		return nil
	case errors.Is(err, proxy.ErrProxiesNotFound):
		// Continue
	}

	podOperator, err := cfgService.GetPodOperator()

	if err != nil {
		return fmt.Errorf("couldn't retrieve pod operator: %w", err)
	}

	ctx := context.Background()

	if err := proxyAPI.AddProxy(ctx, podOperator.GetAccountID(), proxyType.PodOperation, 0, krp); err != nil {
		return fmt.Errorf("couldn't add pod operator as pod operation proxy to %s: %w", delegator.ToHexString(), err)
	}

	if err := proxyAPI.AddProxy(ctx, podOperator.GetAccountID(), proxyType.KeystoreManagement, 0, krp); err != nil {
		return fmt.Errorf("couldn't add pod operator as keystore management proxy to %s: %w", delegator.ToHexString(), err)
	}

	return nil
}

func addKeysToStore(cfgService config.Service, keystoreAPI keystore.API, acc config.Account) error {
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

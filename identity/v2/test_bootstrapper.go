//go:build integration || testworld

package v2

import (
	"context"
	"errors"
	"fmt"

	keystoreTypes "github.com/centrifuge/chain-custom-types/pkg/keystore"
	proxyType "github.com/centrifuge/chain-custom-types/pkg/proxy"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/crypto"
	"github.com/centrifuge/go-centrifuge/pallets"
	"github.com/centrifuge/go-centrifuge/pallets/keystore"
	"github.com/centrifuge/go-centrifuge/pallets/proxy"
	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/common"
	"github.com/centrifuge/go-centrifuge/testingutils/keyrings"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	logging "github.com/ipfs/go-log"
)

var (
	log = logging.Logger("identity_test_bootstrap")
)

func (b *Bootstrapper) TestBootstrap(context map[string]interface{}) error {
	if err := b.bootstrap(context); err != nil {
		return err
	}

	log.Info("Generating test account data")

	cfgService, ok := context[config.BootstrappedConfigStorage].(config.Service)

	if !ok {
		return errors.New("config service not initialised")
	}

	proxyAPI, ok := context[pallets.BootstrappedProxyAPI].(proxy.API)

	if !ok {
		return errors.New("proxy API not initialised")
	}

	keystoreAPI, ok := context[pallets.BootstrappedKeystoreAPI].(keystore.API)

	if !ok {
		return errors.New("keystore API not initialised")
	}

	_, err := BootstrapTestAccount(cfgService, proxyAPI, keystoreAPI, keyrings.AliceKeyRingPair.PublicKey)

	if err != nil {
		return fmt.Errorf("couldn't bootstrap test account for Alice: %w", err)
	}

	return nil
}

func (b *Bootstrapper) TestTearDown() error {
	return nil
}

func BootstrapTestAccount(
	cfgService config.Service,
	proxyAPI proxy.API,
	keystoreAPI keystore.API,
	accountPublicKey []byte,
) (config.Account, error) {
	accountID, err := types.NewAccountID(accountPublicKey)

	if err != nil {
		return nil, fmt.Errorf("couldn't get account ID: %w", err)
	}

	acc, err := createTestAccount(cfgService, accountID)

	if err != nil {
		return nil, fmt.Errorf("couldn't create test account: %w", err)
	}

	if err := createTestProxies(cfgService, proxyAPI, keyrings.AliceKeyRingPair); err != nil {
		return nil, fmt.Errorf("couldn't create test proxies: %w", err)
	}

	if err := addKeysToStore(cfgService, keystoreAPI, acc); err != nil {
		return nil, fmt.Errorf("couldn't add keys to keystore: %w", err)
	}

	return acc, nil
}

func createTestAccount(cfgService config.Service, accountID *types.AccountID) (config.Account, error) {
	if acc, err := cfgService.GetAccount(accountID.ToBytes()); err == nil {
		log.Info("Account already created for -", accountID.ToHexString())

		return acc, nil
	}

	signingPublicKey, signingPrivateKey, err := testingcommons.GetTestSigningKeys()

	if err != nil {
		return nil, fmt.Errorf("couldn't generate document signing keys: %w", err)
	}

	acc, err := configstore.NewAccount(
		accountID,
		signingPublicKey,
		signingPrivateKey,
		"",
		false,
	)

	if err != nil {
		return nil, fmt.Errorf("couldn't create new account: %w", err)
	}

	if err = cfgService.CreateAccount(acc); err != nil {
		return nil, fmt.Errorf("couldn't store account: %w", err)
	}

	return acc, nil
}

func createTestProxies(cfgService config.Service, proxyAPI proxy.API, krp signature.KeyringPair) error {
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

	_, publicKey, err := crypto.ObtainP2PKeypair(cfg.GetP2PKeyPair())

	if err != nil {
		return fmt.Errorf("couldn't obtain P2P key pair: %w", err)
	}

	publicKeyRaw, err := publicKey.Raw()

	if err != nil {
		return fmt.Errorf("couldn't get raw public key: %w", err)
	}

	err = addKeyIfNotPresent(keystoreAPI, acc, publicKeyRaw, keystoreTypes.KeyPurposeP2PDiscovery)

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

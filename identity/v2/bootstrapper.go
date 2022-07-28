package v2

import (
	"context"

	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/crypto"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity/v2/keystore"
	"github.com/centrifuge/go-centrifuge/identity/v2/proxy"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/testingutils/keyrings"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

const (
	BootstrappedKeystoreAPI       = "BootstrappedKeystoreAPI"
	BootstrappedProxyAPI          = "BootstrappedProxyAPI"
	BootstrappedIdentityServiceV2 = "BootstrappedIdentityServiceV2"
)

type Bootstrapper struct{}

func (b *Bootstrapper) Bootstrap(context map[string]interface{}) error {
	return b.bootstrap(context)
}

func (b *Bootstrapper) bootstrap(context map[string]interface{}) error {
	centAPI, ok := context[centchain.BootstrappedCentChainClient].(centchain.API)

	if !ok {
		return errors.New("centchain API not initialised")
	}
	keystoreAPI := keystore.NewAPI(centAPI)

	context[BootstrappedKeystoreAPI] = keystoreAPI

	proxyAPI := proxy.NewAPI(centAPI)

	context[BootstrappedProxyAPI] = proxyAPI

	dispatcher, ok := context[jobs.BootstrappedDispatcher].(jobs.Dispatcher)

	if !ok {
		return errors.New("dispatcher not initialised")
	}

	go dispatcher.RegisterRunner(addKeysJob, &AddKeysJob{
		keystoreAPI: keystoreAPI,
	})

	configService, ok := context[config.BootstrappedConfigStorage].(config.Service)

	if !ok {
		return errors.New("config storage not initialised")
	}

	identityServiceV2 := NewService(configService, centAPI, dispatcher, keystoreAPI)

	context[BootstrappedIdentityServiceV2] = identityServiceV2

	return nil
}

func (b *Bootstrapper) TestBootstrap(context map[string]interface{}) error {
	if err := b.bootstrap(context); err != nil {
		return err
	}

	return generateTestAccountData(context)
}

func (b *Bootstrapper) TestTearDown() error {
	return nil
}

func generateTestAccountData(ctx map[string]interface{}) error {
	configSrv := ctx[config.BootstrappedConfigStorage].(config.Service)
	proxyAPI := ctx[BootstrappedProxyAPI].(proxy.API)

	cfg, err := configSrv.GetConfig()

	if err != nil {
		return err
	}

	p2pPrivateKey, p2pPublicKey, err := crypto.ObtainP2PKeypair(cfg.GetP2PKeyPair())

	if err != nil {
		return err
	}

	signingPrivateKey, signingPublicKey, err := crypto.ObtainP2PKeypair(cfg.GetSigningKeyPair())

	if err != nil {
		return err
	}

	aliceAccountID, err := types.NewAccountID(keyrings.AliceKeyRingPair.PublicKey)

	if err != nil {
		return err
	}

	accountProxies, err := getTestAccountProxies()

	if err != nil {
		return err
	}

	acc, err := configstore.NewAccount(
		aliceAccountID,
		p2pPublicKey,
		p2pPrivateKey,
		signingPublicKey,
		signingPrivateKey,
		"someURL",
		false,
		accountProxies,
	)

	if err != nil {
		return err
	}

	if err := addTestAccountProxies(proxyAPI, acc); err != nil {
		return err
	}

	if _, err := configSrv.CreateAccount(acc); err != nil {
		return err
	}

	return nil
}

func addTestAccountProxies(proxyAPI proxy.API, acc config.Account) error {
	ctx := contextutil.WithAccount(context.Background(), acc)
	ctx = context.WithValue(ctx, config.AccountHeaderKey, acc.GetIdentity())

	for _, accountProxy := range acc.GetAccountProxies() {
		err := proxyAPI.AddProxy(
			ctx,
			accountProxy.AccountID,
			accountProxy.ProxyType,
			types.U32(0),
			keyrings.AliceKeyRingPair,
		)

		if err != nil {
			return err
		}
	}

	return nil
}

func getTestAccountProxies() (config.AccountProxies, error) {
	var accountProxies config.AccountProxies

	for _, proxyType := range types.ProxyTypeValue {
		accountID, err := types.NewAccountID(keyrings.BobKeyRingPair.PublicKey)

		if err != nil {
			return nil, err
		}

		accountProxy := &config.AccountProxy{
			Default:     true,
			AccountID:   accountID,
			Secret:      keyrings.BobKeyRingPair.URI,
			SS58Address: keyrings.BobKeyRingPair.Address,
			ProxyType:   proxyType,
		}

		accountProxies = append(accountProxies, accountProxy)
	}

	return accountProxies, nil
}

//go:build unit || integration || testworld

package v2

import (
	"context"
	"fmt"
	"sync"

	"github.com/centrifuge/go-centrifuge/errors"

	logging "github.com/ipfs/go-log"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/crypto"
	"github.com/centrifuge/go-centrifuge/identity/v2/proxy"
	"github.com/centrifuge/go-centrifuge/testingutils/keyrings"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

var (
	log = logging.Logger("identity_test_bootstrap")
)

func (b *Bootstrapper) TestBootstrap(context map[string]interface{}) error {
	if err := b.bootstrap(context); err != nil {
		return err
	}

	return generateTestAccountData(context)
}

func (b *Bootstrapper) TestTearDown() error {
	return nil
}

var (
	once sync.Once
)

// generateTestAccountData creates a node account for Alice and adds Bob as a proxy with each available type.
func generateTestAccountData(ctx map[string]interface{}) error {
	var err error

	once.Do(func() {
		log.Info("Generating test account data")

		configSrv, ok := ctx[config.BootstrappedConfigStorage].(config.Service)

		if !ok {
			err = errors.New("config service not initialised")
			return
		}

		proxyAPI, ok := ctx[BootstrappedProxyAPI].(proxy.API)

		if !ok {
			err = errors.New("proxy API not initialised")
			return
		}

		aliceAccountID, err := types.NewAccountID(keyrings.AliceKeyRingPair.PublicKey)

		if err != nil {
			err = fmt.Errorf("couldn't get account ID for Alice: %w", err)
			return
		}

		hasTestProxies, err := hasTestProxies(proxyAPI, aliceAccountID)

		if err != nil {
			err = fmt.Errorf("couldn't check if account has test proxies: %w", err)
			return
		}

		if hasTestProxies {
			return
		}

		cfg, err := configSrv.GetConfig()

		if err != nil {
			err = fmt.Errorf("couldn't retrieve config: %w", err)
			return
		}

		p2pPrivateKey, p2pPublicKey, err := crypto.ObtainP2PKeypair(cfg.GetP2PKeyPair())

		if err != nil {
			err = fmt.Errorf("couldn't retrieve p2p key pair: %w", err)
			return
		}

		signingPrivateKey, signingPublicKey, err := crypto.ObtainP2PKeypair(cfg.GetSigningKeyPair())

		if err != nil {
			err = fmt.Errorf("couldn't retrieve signing key pair: %w", err)
			return
		}

		accountProxies, err := getTestAccountProxies()

		if err != nil {
			err = fmt.Errorf("couldn't get test account proxies: %w", err)
			return
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
			err = fmt.Errorf("couldn't create new account: %w", err)
			return
		}

		if err := addTestAccountProxies(proxyAPI, acc); err != nil {
			err = fmt.Errorf("couldn't add test account proxies: %w", err)
			return
		}

		if err := configSrv.CreateAccount(acc); err != nil {
			err = fmt.Errorf("couldn't store account: %w", err)
			return
		}
	})

	return err
}

func hasTestProxies(proxyAPI proxy.API, accountID *types.AccountID) (bool, error) {
	proxies, err := proxyAPI.GetProxies(context.Background(), accountID)

	if err != nil {
		return false, err
	}

	if len(proxies.ProxyDefinitions) == len(types.ProxyTypeValue) {
		return true, nil
	}

	return false, nil
}

func addTestAccountProxies(proxyAPI proxy.API, acc config.Account) error {
	ctx := contextutil.WithAccount(context.Background(), acc)

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

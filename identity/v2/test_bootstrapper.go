//go:build unit || integration || testworld

package v2

import (
	"context"
	"fmt"
	"sync"

	proxyType "github.com/centrifuge/chain-custom-types/pkg/proxy"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity/v2/proxy"
	"github.com/centrifuge/go-centrifuge/testingutils/keyrings"
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

		signingPublicKey, signingPrivateKey, err := generateDocumentSigningKeys()

		if err != nil {
			err = fmt.Errorf("couldn't generate document signing keys: %w", err)
		}

		acc, err := configstore.NewAccount(
			aliceAccountID,
			signingPublicKey,
			signingPrivateKey,
			"https://someURL.com",
			false,
		)

		if err != nil {
			err = fmt.Errorf("couldn't create new account: %w", err)
			return
		}

		if err := configSrv.CreateAccount(acc); err != nil {
			err = fmt.Errorf("couldn't store account: %w", err)
			return
		}

		podOperator, err := configSrv.GetPodOperator()

		if err != nil {
			err = fmt.Errorf("couldn't retrieve pod operator: %w", err)
			return
		}

		if err := proxyAPI.AddProxy(context.Background(), podOperator.GetAccountID(), proxyType.PodOperation, 0, keyrings.AliceKeyRingPair); err != nil {
			err = fmt.Errorf("couldn't add pod operator as proxy to test account Alice: %w", err)
			return
		}
	})

	return err
}

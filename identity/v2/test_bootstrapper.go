//go:build unit || integration || testworld

package v2

import (
	"context"
	"errors"
	"fmt"
	"github.com/centrifuge/go-centrifuge/pallets"
	"sync"

	"github.com/centrifuge/go-centrifuge/pallets/keystore"
	"github.com/centrifuge/go-centrifuge/pallets/proxy"

	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/common"

	"github.com/centrifuge/go-centrifuge/contextutil"

	keystoreTypes "github.com/centrifuge/chain-custom-types/pkg/keystore"
	proxyType "github.com/centrifuge/chain-custom-types/pkg/proxy"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
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
	var generateError error

	once.Do(func() {
		log.Info("Generating test account data")

		configSrv, ok := ctx[config.BootstrappedConfigStorage].(config.Service)

		if !ok {
			generateError = errors.New("config service not initialised")
			return
		}

		proxyAPI, ok := ctx[pallets.BootstrappedProxyAPI].(proxy.API)

		if !ok {
			generateError = errors.New("proxy API not initialised")
			return
		}

		keystoreAPI, ok := ctx[pallets.BootstrappedKeystoreAPI].(keystore.API)

		if !ok {
			generateError = errors.New("keystore API not initialised")
			return
		}

		aliceAccountID, err := types.NewAccountID(keyrings.AliceKeyRingPair.PublicKey)

		if err != nil {
			generateError = fmt.Errorf("couldn't get account ID for Alice: %w", err)
			return
		}

		bobAccountID, err := types.NewAccountID(keyrings.BobKeyRingPair.PublicKey)

		if err != nil {
			generateError = fmt.Errorf("couldn't get account ID for Bob: %w", err)
			return
		}

		signingPublicKey, signingPrivateKey, err := testingcommons.GetTestSigningKeys()

		if err != nil {
			generateError = fmt.Errorf("couldn't generate document signing keys: %w", err)
			return
		}

		acc, err := configstore.NewAccount(
			aliceAccountID,
			signingPublicKey,
			signingPrivateKey,
			"https://someURL.com",
			false,
		)

		if err != nil {
			generateError = fmt.Errorf("couldn't create new account: %w", err)
			return
		}

		if err = configSrv.CreateAccount(acc); err != nil {
			generateError = fmt.Errorf("couldn't store account: %w", err)
			return
		}

		podOperator, err := configSrv.GetPodOperator()

		if err != nil {
			generateError = fmt.Errorf("couldn't retrieve pod operator: %w", err)
			return
		}

		ctx := context.Background()

		if err := proxyAPI.AddProxy(ctx, bobAccountID, proxyType.PodAuth, 0, keyrings.AliceKeyRingPair); err != nil {
			generateError = fmt.Errorf("couldn't add Bob as pod auth proxy to test account Alice: %w", err)
			return
		}

		if err := proxyAPI.AddProxy(ctx, podOperator.GetAccountID(), proxyType.PodOperation, 0, keyrings.AliceKeyRingPair); err != nil {
			generateError = fmt.Errorf("couldn't add pod operator as pod operation proxy to test account Alice: %w", err)
			return
		}

		if err := proxyAPI.AddProxy(ctx, podOperator.GetAccountID(), proxyType.KeystoreManagement, 0, keyrings.AliceKeyRingPair); err != nil {
			generateError = fmt.Errorf("couldn't add pod operator as keystore management proxy to test account Alice: %w", err)
			return
		}

		rawSigningKey, err := signingPublicKey.Raw()

		if err != nil {
			generateError = fmt.Errorf("couldn't get raw signing key: %w", err)
			return
		}

		keyHash := types.NewHash(rawSigningKey)

		_, err = keystoreAPI.GetKey(
			contextutil.WithAccount(ctx, acc),
			&keystoreTypes.KeyID{
				Hash:       keyHash,
				KeyPurpose: keystoreTypes.KeyPurposeP2PDocumentSigning,
			},
		)

		if err == nil {
			log.Info("Key already stored in keystore, skipping.")
			return
		}

		if !errors.Is(err, keystore.ErrKeyNotFound) {
			generateError = err
			return
		}

		_, err = keystoreAPI.AddKeys(
			contextutil.WithAccount(ctx, acc),
			[]*keystoreTypes.AddKey{
				{
					Key:     keyHash,
					Purpose: keystoreTypes.KeyPurposeP2PDocumentSigning,
					KeyType: keystoreTypes.KeyTypeECDSA,
				},
			},
		)

		if err != nil {
			generateError = fmt.Errorf("couldn't add document signing key to keystore: %w", err)
			return
		}
	})

	return generateError
}

//go:build testworld

package factory

import (
	"context"
	"fmt"
	"math/big"

	"github.com/centrifuge/go-centrifuge/contextutil"

	proxyType "github.com/centrifuge/chain-custom-types/pkg/proxy"
	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/crypto"
	"github.com/centrifuge/go-centrifuge/errors"
	identityv2 "github.com/centrifuge/go-centrifuge/identity/v2"
	genericUtils "github.com/centrifuge/go-centrifuge/testingutils/generic"
	"github.com/centrifuge/go-centrifuge/testworld/park/host"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

func CreateTestHostAccount(
	serviceCtx map[string]any,
	krp signature.KeyringPair,
	webhookURL string,
) (*host.Account, error) {
	accountID, err := types.NewAccountID(krp.PublicKey)

	if err != nil {
		return nil, fmt.Errorf("couldn't get account ID: %w", err)
	}

	acc, err := identityv2.CreateTestIdentity(serviceCtx, accountID, webhookURL)

	if err != nil {
		return nil, fmt.Errorf("couldn't create test account: %w", err)
	}

	podAuthProxy, err := host.GenerateSignerAccount()

	if err != nil {
		return nil, fmt.Errorf("couldn't generate pod auth proxy: %w", err)
	}

	if err := addTestHostAccountProxies(serviceCtx, krp, podAuthProxy); err != nil {
		return nil, fmt.Errorf("couldn't create test account proxies: %w", err)
	}

	if err := identityv2.AddAccountKeysToStore(serviceCtx, acc); err != nil {
		return nil, fmt.Errorf("couldn't add test account keys to store: %w", err)
	}

	hostCfg := genericUtils.GetService[config.Configuration](serviceCtx)

	podOperator, err := host.GetSignerAccount(hostCfg.GetPodOperatorSecretSeed())
	if err != nil {
		return nil, fmt.Errorf("couldn't get pod operator signer account: %w", err)
	}

	podAdmin, err := host.GetSignerAccount(hostCfg.GetPodAdminSecretSeed())
	if err != nil {
		return nil, fmt.Errorf("couldn't get pod admin signer account: %w", err)
	}

	p2pPublicKey, err := getP2PPublicKey(hostCfg)
	if err != nil {
		return nil, fmt.Errorf("couldn't get P2P public key: %w", err)
	}

	return host.NewAccount(
		acc,
		krp,
		podAuthProxy,
		podAdmin,
		podOperator,
		p2pPublicKey,
	), nil
}

func addTestHostAccountProxies(serviceCtx map[string]any, krp signature.KeyringPair, podAuthProxy *host.SignerAccount) error {
	cfgService := genericUtils.GetService[config.Service](serviceCtx)

	podOperator, err := cfgService.GetPodOperator()

	if err != nil {
		return fmt.Errorf("couldn't get pod operator: %w", err)
	}

	proxyPairs := identityv2.ProxyPairs{
		{
			Delegate:  podOperator.GetAccountID(),
			ProxyType: proxyType.PodOperation,
		},
		{
			Delegate:  podOperator.GetAccountID(),
			ProxyType: proxyType.KeystoreManagement,
		},
		{
			Delegate:  podAuthProxy.AccountID,
			ProxyType: proxyType.PodAuth,
		},
	}

	if err := identityv2.AddAndWaitForTestProxies(serviceCtx, krp, proxyPairs); err != nil {
		return fmt.Errorf("couldn't add test proxies: %w", err)
	}

	return nil
}

func getP2PPublicKey(cfg config.Configuration) ([]byte, error) {
	_, pubKey, err := crypto.ObtainP2PKeypair(cfg.GetP2PKeyPair())

	if err != nil {
		return nil, fmt.Errorf("couldn't get P2P public key: %w", err)
	}

	return pubKey.Raw()
}

func CreateRandomHostAccount(
	serviceCtx map[string]any,
	webhookURL string,
	fundsProvider *host.Account,
) (*host.Account, error) {
	randomHostAccount, err := createRandomAccountOnChain(serviceCtx, fundsProvider)

	if err != nil {
		return nil, err
	}

	return CreateTestHostAccount(serviceCtx, randomHostAccount.ToKeyringPair(), webhookURL)
}

func createRandomAccountOnChain(
	serviceCtx map[string]any,
	fundsProvider *host.Account,
) (*host.SignerAccount, error) {
	randomHostAccount, err := host.GenerateSignerAccount()

	if err != nil {
		return nil, fmt.Errorf("couldn't generate signer account: %w", err)
	}

	if err := addFundsToAccount(serviceCtx, fundsProvider, randomHostAccount.ToKeyringPair()); err != nil {
		return nil, fmt.Errorf("couldn't add funds to account: %w", err)
	}

	return randomHostAccount, nil
}

func addFundsToAccount(serviceCtx map[string]any, fundsProvider *host.Account, to signature.KeyringPair) error {
	centAPI := genericUtils.GetService[centchain.API](serviceCtx)

	addr, err := types.NewMultiAddressFromAccountID(to.PublicKey)

	if err != nil {
		return fmt.Errorf("couldn't get multi address: %w", err)
	}

	meta, err := centAPI.GetMetadataLatest()

	if err != nil {
		return fmt.Errorf("couldn't get latest metadata: %w", err)
	}

	amount, ok := big.NewInt(0).SetString("1000000000000000000000", 10)

	if !ok {
		return errors.New("couldn't create balance amount")
	}

	call, err := types.NewCall(meta, "Balances.transfer", addr, types.NewUCompact(amount))

	if err != nil {
		return fmt.Errorf("couldn't create call: %w", err)
	}

	ctx := contextutil.WithAccount(context.Background(), fundsProvider.GetAccount())

	if _, err = centAPI.SubmitAndWatch(ctx, meta, call, fundsProvider.GetKeyringPair()); err != nil {
		return fmt.Errorf("couldn't submit and watch balance transfer extrinsic: %w", err)
	}

	return nil
}

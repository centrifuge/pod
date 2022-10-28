//go:build testworld

package bootstrap

import (
	"context"
	"fmt"

	proxyType "github.com/centrifuge/chain-custom-types/pkg/proxy"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/crypto"
	identityv2 "github.com/centrifuge/go-centrifuge/identity/v2"
	genericUtils "github.com/centrifuge/go-centrifuge/testingutils/generic"
	proxyUtils "github.com/centrifuge/go-centrifuge/testingutils/proxy"
	"github.com/centrifuge/go-centrifuge/testworld/park/host"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

func bootstrapHostAccount(
	serviceCtx map[string]any,
	krp signature.KeyringPair,
	webhookURL string,
) (config.Account, *host.SignerAccount, error) {
	accountID, err := types.NewAccountID(krp.PublicKey)

	if err != nil {
		return nil, nil, fmt.Errorf("couldn't get account ID: %w", err)
	}

	acc, err := identityv2.CreateTestAccount(serviceCtx, accountID, webhookURL)

	if err != nil {
		return nil, nil, fmt.Errorf("couldn't create test account: %w", err)
	}

	podAuthProxy, err := createTestHostAccountProxies(serviceCtx, krp)

	if err != nil {
		return nil, nil, fmt.Errorf("couldn't create test account proxies: %w", err)
	}

	if err := identityv2.AddAccountKeysToStore(serviceCtx, acc); err != nil {
		return nil, nil, fmt.Errorf("couldn't add test account keys to store: %w", err)
	}

	return acc, podAuthProxy, nil
}

func createTestHostAccountProxies(serviceCtx map[string]any, krp signature.KeyringPair) (*host.SignerAccount, error) {
	cfgService := genericUtils.GetService[config.Service](serviceCtx)

	podOperator, err := cfgService.GetPodOperator()

	if err != nil {
		return nil, fmt.Errorf("couldn't get pod operator: %w", err)
	}

	podAuthProxySeed, err := crypto.GenerateSR25519SecretSeed()

	if err != nil {
		return nil, fmt.Errorf("couldn't generate pod auth proxy seed: %w", err)
	}

	podAuthProxy, err := host.GetSignerAccount(podAuthProxySeed)

	if err != nil {
		return nil, fmt.Errorf("couldn't generate proxy account: %w", err)
	}

	proxyPairs := []identityv2.ProxyPair{
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

	if err := identityv2.AddTestProxies(serviceCtx, krp, proxyPairs...); err != nil {
		return nil, fmt.Errorf("couldn't add test proxies: %w", err)
	}

	delegatorAccountID, err := types.NewAccountID(krp.PublicKey)

	if err != nil {
		return nil, fmt.Errorf("couldn't get account ID: %w", err)
	}

	err = proxyUtils.WaitForProxiesToBeAdded(
		context.Background(),
		serviceCtx,
		delegatorAccountID,
		podOperator.GetAccountID(),
		podAuthProxy.AccountID,
	)

	if err != nil {
		return nil, fmt.Errorf("proxies were not added: %w", err)
	}

	return podAuthProxy, nil
}

func createHost(
	hostControlUnit *host.ControlUnit,
	hostAccount config.Account,
	podAuthProxy *host.SignerAccount,
) (*host.Host, error) {
	hostCfg := genericUtils.GetService[config.Configuration](hostControlUnit.GetServiceCtx())

	podOperator, err := host.GetSignerAccount(hostCfg.GetPodOperatorSecretSeed())
	if err != nil {
		return nil, fmt.Errorf("couldn't get pod operator signer account: %w", err)
	}

	podAdmin, err := host.GetSignerAccount(hostCfg.GetPodAdminSecretSeed())
	if err != nil {
		return nil, fmt.Errorf("couldn't get pod admin signer account: %w", err)
	}

	return host.NewHost(
		hostAccount,
		podAuthProxy,
		podOperator,
		podAdmin,
		hostControlUnit,
	), nil
}

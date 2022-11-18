//go:build testworld

package factory

import (
	"fmt"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/crypto"
	identityv2 "github.com/centrifuge/go-centrifuge/identity/v2"
	genericUtils "github.com/centrifuge/go-centrifuge/testingutils/generic"
	"github.com/centrifuge/go-centrifuge/testworld/park/host"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
)

func CreateTestHostAccount(
	serviceCtx map[string]any,
	originKrp signature.KeyringPair,
	webhookURL string,
) (*host.Account, error) {
	identity, err := identityv2.CreateAnonymousProxy(serviceCtx, originKrp)

	if err != nil {
		return nil, fmt.Errorf("couldn't get account ID: %w", err)
	}

	acc, err := identityv2.CreateTestIdentity(serviceCtx, identity, webhookURL)

	if err != nil {
		return nil, fmt.Errorf("couldn't create test account: %w", err)
	}

	hostCfg := genericUtils.GetService[config.Configuration](serviceCtx)

	podAuthProxy, err := host.GenerateSignerAccount()

	if err != nil {
		return nil, fmt.Errorf("couldn't generate pod auth proxy: %w", err)
	}

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
		originKrp,
		podAuthProxy,
		podAdmin,
		podOperator,
		p2pPublicKey,
	), nil
}

func getP2PPublicKey(cfg config.Configuration) ([]byte, error) {
	_, pubKey, err := crypto.ObtainP2PKeypair(cfg.GetP2PKeyPair())

	if err != nil {
		return nil, fmt.Errorf("couldn't get P2P public key: %w", err)
	}

	return pubKey.Raw()
}

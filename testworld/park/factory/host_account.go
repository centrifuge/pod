//go:build testworld

package factory

import (
	"fmt"

	v2 "github.com/centrifuge/pod/identity/v2"

	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	"github.com/centrifuge/pod/config"
	"github.com/centrifuge/pod/crypto"
	"github.com/centrifuge/pod/pallets"
	genericUtils "github.com/centrifuge/pod/testingutils/generic"
	"github.com/centrifuge/pod/testworld/park/host"
)

func CreateTestHostAccount(
	serviceCtx map[string]any,
	originKrp signature.KeyringPair,
	webhookURL string,
) (*host.Account, error) {
	identity, err := pallets.CreateAnonymousProxy(serviceCtx, originKrp)

	if err != nil {
		return nil, fmt.Errorf("couldn't get account ID: %w", err)
	}

	acc, err := v2.CreateTestIdentity(serviceCtx, identity, webhookURL)

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

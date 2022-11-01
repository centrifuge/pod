//go:build testworld

package host

import (
	"fmt"

	proxyType "github.com/centrifuge/chain-custom-types/pkg/proxy"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/http/auth"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

type Name string

const (
	Alice   Name = "Alice"
	Bob     Name = "Bob"
	Charlie Name = "Charlie"
	Dave    Name = "Dave"
)

type Host struct {
	acc config.Account

	podAuthProxy *SignerAccount
	podAdmin     *SignerAccount
	podOperator  *SignerAccount

	controlUnit *ControlUnit
}

func NewHost(
	acc config.Account,
	podAuthProxy *SignerAccount,
	podAdmin *SignerAccount,
	podOperator *SignerAccount,
	controlUnit *ControlUnit,
) *Host {
	return &Host{
		acc,
		podAuthProxy,
		podAdmin,
		podOperator,
		controlUnit,
	}
}

func (t *Host) GetAPIURL() string {
	return fmt.Sprintf("http://localhost:%d", t.controlUnit.GetPodCfg().GetServerPort())
}

func (t *Host) AccountID() *types.AccountID {
	return t.acc.GetIdentity()
}

func (t *Host) GetJW3Token(pt string) (string, error) {
	tokenArgs, err := t.getTokenArgsForProxyType(pt)

	if err != nil {
		return "", fmt.Errorf("couldn't get token args: %w", err)
	}

	return auth.CreateJW3Token(
		tokenArgs.delegateAccountID,
		tokenArgs.delegatorAccountID,
		tokenArgs.secretSeed,
		tokenArgs.proxyType,
	)
}

type tokenArgs struct {
	delegateAccountID  *types.AccountID
	delegatorAccountID *types.AccountID
	secretSeed         string
	proxyType          string
}

func (t *Host) getTokenArgsForProxyType(pt string) (*tokenArgs, error) {
	var args *tokenArgs

	switch pt {
	case proxyType.ProxyTypeName[proxyType.PodOperation]:
		args = &tokenArgs{
			delegateAccountID:  t.podOperator.AccountID,
			delegatorAccountID: t.acc.GetIdentity(),
			secretSeed:         t.podOperator.SecretSeed,
			proxyType:          pt,
		}
	case proxyType.ProxyTypeName[proxyType.PodAuth]:
		args = &tokenArgs{
			delegateAccountID:  t.podAuthProxy.AccountID,
			delegatorAccountID: t.acc.GetIdentity(),
			secretSeed:         t.podAuthProxy.SecretSeed,
			proxyType:          pt,
		}
	case auth.PodAdminProxyType:
		args = &tokenArgs{
			delegateAccountID:  t.podAdmin.AccountID,
			delegatorAccountID: t.podAdmin.AccountID,
			secretSeed:         t.podAdmin.SecretSeed,
			proxyType:          pt,
		}
	default:
		return nil, fmt.Errorf("unsupported proxy type - %s", pt)
	}

	return args, nil
}

func (t *Host) Stop() error {
	return t.controlUnit.Stop()
}

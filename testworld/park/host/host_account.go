//go:build testworld

package host

import (
	"fmt"

	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"

	proxyType "github.com/centrifuge/chain-custom-types/pkg/proxy"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/http/auth"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

type Account struct {
	acc config.Account
	krp signature.KeyringPair

	podAuthProxy *SignerAccount
	podAdmin     *SignerAccount
	podOperator  *SignerAccount

	p2pPublicKey []byte
}

func NewAccount(
	acc config.Account,
	krp signature.KeyringPair,
	podAuthProxy *SignerAccount,
	podAdmin *SignerAccount,
	podOperator *SignerAccount,
	p2pPublicKey []byte,
) *Account {
	return &Account{
		acc,
		krp,
		podAuthProxy,
		podAdmin,
		podOperator,
		p2pPublicKey,
	}
}

func (a *Account) GetAccount() config.Account {
	return a.acc
}

func (a *Account) GetKeyringPair() signature.KeyringPair {
	return a.krp
}

func (a *Account) GetAccountID() *types.AccountID {
	return a.acc.GetIdentity()
}

func (a *Account) GetPodOperatorAccountID() *types.AccountID {
	return a.podOperator.AccountID
}

func (a *Account) GetP2PPublicKey() []byte {
	return a.p2pPublicKey
}

func (a *Account) GetJW3Token(pt string) (string, error) {
	tokenArgs, err := a.getTokenArgsForProxyType(pt)

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

func (a *Account) getTokenArgsForProxyType(pt string) (*tokenArgs, error) {
	var args *tokenArgs

	switch pt {
	case proxyType.ProxyTypeName[proxyType.PodOperation]:
		args = &tokenArgs{
			delegateAccountID:  a.podOperator.AccountID,
			delegatorAccountID: a.acc.GetIdentity(),
			secretSeed:         a.podOperator.SecretSeed,
			proxyType:          pt,
		}
	case proxyType.ProxyTypeName[proxyType.PodAuth]:
		args = &tokenArgs{
			delegateAccountID:  a.podAuthProxy.AccountID,
			delegatorAccountID: a.acc.GetIdentity(),
			secretSeed:         a.podAuthProxy.SecretSeed,
			proxyType:          pt,
		}
	case auth.PodAdminProxyType:
		args = &tokenArgs{
			delegateAccountID:  a.podAdmin.AccountID,
			delegatorAccountID: a.podAdmin.AccountID,
			secretSeed:         a.podAdmin.SecretSeed,
			proxyType:          pt,
		}
	default:
		return nil, fmt.Errorf("unsupported proxy type - %s", pt)
	}

	return args, nil
}
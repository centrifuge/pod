//go:build testworld

package testworld

import (
	proxyType "github.com/centrifuge/chain-custom-types/pkg/proxy"
	"github.com/centrifuge/go-centrifuge/testingutils/keyrings"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

type testAccount struct {
	name    testAccountName
	keyRing signature.KeyringPair
	proxy   *signerAccount
}

func (t *testAccount) GetJW3Token() (string, error) {
	testAccountID, err := t.AccountID()

	if err != nil {
		return "", err
	}

	return CreateJW3Token(
		t.proxy.AccountID,
		testAccountID,
		t.proxy.SecretSeed,
		proxyType.ProxyTypeName[proxyType.PodAuth],
	)
}

func (t *testAccount) AccountID() (*types.AccountID, error) {
	return types.NewAccountID(t.keyRing.PublicKey)
}

type testAccountName string

const (
	testAccountAlice   testAccountName = "alice"
	testAccountBob     testAccountName = "bob"
	testAccountCharlie testAccountName = "charlie"
	testAccountDave    testAccountName = "dave"

	// Eve is the node admin.
	//testAccountEve     testAccountName = "eve"

	// Ferdie is the pod operator.
	//testAccountFerdie  testAccountName = "ferdie"
)

var (
	testAccountMap = map[testAccountName]*testAccount{
		testAccountAlice: {
			name:    testAccountAlice,
			keyRing: keyrings.AliceKeyRingPair,
		},
		testAccountBob: {
			name:    testAccountBob,
			keyRing: keyrings.BobKeyRingPair,
		},
		testAccountCharlie: {
			name:    testAccountCharlie,
			keyRing: keyrings.CharlieKeyRingPair,
		},
		testAccountDave: {
			name:    testAccountDave,
			keyRing: keyrings.DaveKeyRingPair,
		},
	}
)

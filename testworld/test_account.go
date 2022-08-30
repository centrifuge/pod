//go:build testworld

package testworld

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/centrifuge/go-centrifuge/http/auth"
	"github.com/centrifuge/go-centrifuge/testingutils/keyrings"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

type testAccount struct {
	name    testAccountName
	keyRing signature.KeyringPair
}

func (t *testAccount) toMockJW3T() (string, error) {
	header := &auth.JW3THeader{
		Algorithm:   "sr25519",
		AddressType: "ss58",
		TokenType:   "JW3T",
	}

	now := time.Now()
	exipreTime := now.Add(24 * time.Hour)

	accountID, err := t.AccountID()

	if err != nil {
		return "", err
	}

	payload := auth.JW3TPayload{
		IssuedAt:   fmt.Sprintf("%d", now.Unix()),
		NotBefore:  fmt.Sprintf("%d", now.Unix()),
		ExpiresAt:  fmt.Sprintf("%d", exipreTime.Unix()),
		Address:    accountID.ToHexString(),
		OnBehalfOf: accountID.ToHexString(),
		ProxyType:  auth.NodeAdminProxyType,
	}

	signature := "signature"

	jsonHeader, err := json.Marshal(header)

	if err != nil {
		return "", err
	}

	encodedJSONHeader := base64.RawURLEncoding.EncodeToString(jsonHeader)

	jsonPayload, err := json.Marshal(payload)

	if err != nil {
		return "", err
	}

	encodedJSONPayload := base64.RawURLEncoding.EncodeToString(jsonPayload)

	encodedSignature := base64.RawURLEncoding.EncodeToString([]byte(signature))

	elems := []string{
		encodedJSONHeader,
		encodedJSONPayload,
		encodedSignature,
	}

	return strings.Join(elems, "."), nil
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
	testAccountEve     testAccountName = "eve"
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
		testAccountEve: {
			name:    testAccountEve,
			keyRing: keyrings.EveKeyRingPair,
		},
	}
)

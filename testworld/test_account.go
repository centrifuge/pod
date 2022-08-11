//go:build testworld

package testworld

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/centrifuge/go-centrifuge/http/auth"
	"github.com/centrifuge/go-centrifuge/http/coreapi"
	"github.com/centrifuge/go-centrifuge/testingutils/keyrings"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

type testAccount struct {
	name      testAccountName
	keyRing   signature.KeyringPair
	proxiesFn proxiesFn
}

func (t *testAccount) toMockJW3T() (string, error) {
	header := &auth.JW3THeader{
		Algorithm:   "sr25519",
		AddressType: "ss58",
		TokenType:   "JW3T",
	}

	now := time.Now()
	exipreTime := now.Add(24 * time.Hour)

	proxies, err := t.proxiesFn()

	if err != nil {
		return "", err
	}

	accountID, err := t.AccountID()

	if err != nil {
		return "", err
	}

	payload := auth.JW3TPayload{
		IssuedAt:   fmt.Sprintf("%d", now.Unix()),
		NotBefore:  fmt.Sprintf("%d", now.Unix()),
		ExpiresAt:  fmt.Sprintf("%d", exipreTime.Unix()),
		Address:    proxies[0].AccountID.ToHexString(),
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
	testAccountFerdie  testAccountName = "ferdie"
)

type proxiesFn func() (coreapi.AccountProxies, error)

var (
	testAccountMap = map[testAccountName]*testAccount{
		testAccountAlice: {
			name:    testAccountAlice,
			keyRing: keyrings.AliceKeyRingPair,
			proxiesFn: func() (coreapi.AccountProxies, error) {
				var accountProxies coreapi.AccountProxies

				for proxyTypeString, _ := range types.ProxyTypeValue {
					proxyAccountID, err := types.NewAccountID(keyrings.BobKeyRingPair.PublicKey)

					if err != nil {
						return nil, err
					}

					accountProxies = append(accountProxies, &coreapi.AccountProxy{
						Default:     false,
						AccountID:   proxyAccountID,
						Secret:      keyrings.BobKeyRingPair.URI,
						SS58Address: keyrings.BobKeyRingPair.Address,
						ProxyType:   proxyTypeString,
					})
				}

				accountProxies[0].Default = true

				return accountProxies, nil
			},
		},
		testAccountBob: {
			name:    testAccountBob,
			keyRing: keyrings.BobKeyRingPair,
			proxiesFn: func() (coreapi.AccountProxies, error) {
				proxyAccountID, err := types.NewAccountID(keyrings.AliceKeyRingPair.PublicKey)

				if err != nil {
					return nil, err
				}

				return coreapi.AccountProxies{&coreapi.AccountProxy{
					Default:     true,
					AccountID:   proxyAccountID,
					Secret:      keyrings.AliceKeyRingPair.URI,
					SS58Address: keyrings.AliceKeyRingPair.Address,
					ProxyType:   types.ProxyTypeName[types.Any],
				}}, nil
			},
		},
		//testAccountCharlie: {
		//	name:    testAccountCharlie,
		//	keyRing: keyrings.CharlieKeyRingPair,
		//	proxiesFn: func() (coreapi.AccountProxies, error) {
		//		proxyAccountID, err := types.NewAccountID(keyrings.BobKeyRingPair.PublicKey)
		//
		//		if err != nil {
		//			return nil, err
		//		}
		//
		//		return coreapi.AccountProxies{&coreapi.AccountProxy{
		//			Default:     true,
		//			AccountID:   proxyAccountID,
		//			Secret:      keyrings.BobKeyRingPair.URI,
		//			SS58Address: keyrings.BobKeyRingPair.Address,
		//			ProxyType:   types.ProxyTypeName[types.Any],
		//		}}, nil
		//	},
		//},
		//testAccountDave: {
		//	name:    testAccountDave,
		//	keyRing: keyrings.DaveKeyRingPair,
		//	proxiesFn: func() (coreapi.AccountProxies, error) {
		//		proxyAccountID, err := types.NewAccountID(keyrings.CharlieKeyRingPair.PublicKey)
		//
		//		if err != nil {
		//			return nil, err
		//		}
		//
		//		return coreapi.AccountProxies{&coreapi.AccountProxy{
		//			Default:     true,
		//			AccountID:   proxyAccountID,
		//			Secret:      keyrings.CharlieKeyRingPair.URI,
		//			SS58Address: keyrings.CharlieKeyRingPair.Address,
		//			ProxyType:   types.ProxyTypeName[types.Any],
		//		}}, nil
		//	},
		//},
		//testAccountEve: {
		//	name:    testAccountEve,
		//	keyRing: keyrings.EveKeyRingPair,
		//	proxiesFn: func() (coreapi.AccountProxies, error) {
		//		proxyAccountID, err := types.NewAccountID(keyrings.DaveKeyRingPair.PublicKey)
		//
		//		if err != nil {
		//			return nil, err
		//		}
		//
		//		return coreapi.AccountProxies{&coreapi.AccountProxy{
		//			Default:     true,
		//			AccountID:   proxyAccountID,
		//			Secret:      keyrings.DaveKeyRingPair.URI,
		//			SS58Address: keyrings.DaveKeyRingPair.Address,
		//			ProxyType:   types.ProxyTypeName[types.Any],
		//		}}, nil
		//	},
		//},
		//testAccountFerdie: {
		//	name:    testAccountFerdie,
		//	keyRing: keyrings.FerdieKeyRingPair,
		//	proxiesFn: func() (coreapi.AccountProxies, error) {
		//		proxyAccountID, err := types.NewAccountID(keyrings.EveKeyRingPair.PublicKey)
		//
		//		if err != nil {
		//			return nil, err
		//		}
		//
		//		return coreapi.AccountProxies{&coreapi.AccountProxy{
		//			Default:     true,
		//			AccountID:   proxyAccountID,
		//			Secret:      keyrings.EveKeyRingPair.URI,
		//			SS58Address: keyrings.EveKeyRingPair.Address,
		//			ProxyType:   types.ProxyTypeName[types.Any],
		//		}}, nil
		//	},
		//},
	}
)

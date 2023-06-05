//go:build unit

package token

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/vedhavyas/go-subkey"

	proxyType "github.com/centrifuge/chain-custom-types/pkg/proxy"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/pod/testingutils/keyrings"
	"github.com/stretchr/testify/assert"
)

func TestDecodeJW3Token(t *testing.T) {
	delegate, err := types.NewAccountID(keyrings.AliceKeyRingPair.PublicKey)
	assert.NoError(t, err)

	delegator, err := types.NewAccountID(keyrings.BobKeyRingPair.PublicKey)
	assert.NoError(t, err)

	token, err := CreateJW3Token(
		delegate,
		delegator,
		keyrings.AliceKeyRingPair.URI,
		proxyType.ProxyTypeName[proxyType.Any],
	)
	assert.NoError(t, err)

	res, err := DecodeJW3Token(token)
	assert.NoError(t, err)
	assert.IsType(t, &JW3Token{}, res)
}

func TestDecodeJW3Token_InvalidToken(t *testing.T) {
	tokenHeader := JW3THeader{
		Algorithm:   "sr25519",
		AddressType: "ss58",
		TokenType:   "jw3t",
	}

	tokenHeaderJson, err := json.Marshal(tokenHeader)
	assert.NoError(t, err)

	tokenHeaderBase64 := base64.RawURLEncoding.EncodeToString(tokenHeaderJson)

	now := time.Now()
	expireTime := now.Add(1 * time.Hour)

	tokenPayload := &JW3TPayload{
		IssuedAt:   fmt.Sprintf("%d", now.Unix()),
		NotBefore:  fmt.Sprintf("%d", now.Unix()),
		ExpiresAt:  fmt.Sprintf("%d", expireTime.Unix()),
		Address:    "delegateAddress",
		OnBehalfOf: "delegatorAddress",
		ProxyType:  "proxyType",
	}

	tokenPayloadJson, err := json.Marshal(tokenPayload)
	assert.NoError(t, err)

	tokenPayloadBase64 := base64.RawURLEncoding.EncodeToString(tokenPayloadJson)

	tests := []struct {
		Name           string
		TokenHeader    string
		TokenPayload   string
		TokenSignature string
		ExpectedError  error
	}{
		{
			Name:           "Invalid base64 header",
			TokenHeader:    "!!",
			TokenPayload:   "",
			TokenSignature: "",
			ExpectedError:  ErrBase64HeaderDecoding,
		},
		{
			Name:           "Invalid json header",
			TokenHeader:    base64.RawURLEncoding.EncodeToString([]byte("invalid_json")),
			TokenPayload:   "",
			TokenSignature: "",
			ExpectedError:  ErrJSONHeaderDecoding,
		},
		{
			Name:           "Invalid base64 payload",
			TokenHeader:    tokenHeaderBase64,
			TokenPayload:   "!!",
			TokenSignature: "",
			ExpectedError:  ErrBase64PayloadDecoding,
		},
		{
			Name:           "Invalid json payload",
			TokenHeader:    tokenHeaderBase64,
			TokenPayload:   base64.RawURLEncoding.EncodeToString([]byte("invalid_json")),
			TokenSignature: "",
			ExpectedError:  ErrJSONPayloadDecoding,
		},
		{
			Name:           "Invalid base64 signature",
			TokenHeader:    tokenHeaderBase64,
			TokenPayload:   tokenPayloadBase64,
			TokenSignature: "!!",
			ExpectedError:  ErrBase64SignatureDecoding,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			tokenStr := strings.Join([]string{test.TokenHeader, test.TokenPayload, test.TokenSignature}, tokenSeparator)

			res, err := DecodeJW3Token(tokenStr)
			assert.ErrorIs(t, err, test.ExpectedError)
			assert.Nil(t, res)
		})
	}
}

func TestDecodeSS58Address(t *testing.T) {
	accountID, err := types.NewAccountID(keyrings.AliceKeyRingPair.PublicKey)
	assert.NoError(t, err)

	ss58Address, err := subkey.SS58Address(accountID.ToBytes(), CentrifugeNetworkID)
	assert.NoError(t, err)

	res, err := DecodeSS58Address(ss58Address)
	assert.NoError(t, err)
	assert.True(t, accountID.Equal(res))

	res, err = DecodeSS58Address(ss58Address[:len(ss58Address)-2])
	assert.Error(t, err)
	assert.Nil(t, res)
}

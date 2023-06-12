//go:build unit

package token

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	proxyType "github.com/centrifuge/chain-custom-types/pkg/proxy"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/pod/testingutils/keyrings"
	"github.com/centrifuge/pod/utils"
	"github.com/stretchr/testify/assert"
	"github.com/vedhavyas/go-subkey/v2"
	"github.com/vedhavyas/go-subkey/v2/sr25519"
)

func Test_SR25519HeaderValidationFn(t *testing.T) {
	header := &JW3THeader{
		Algorithm:   sr25519Algorithm,
		AddressType: ss58AddressType,
	}

	err := SR25519HeaderValidationFn(header)
	assert.NoError(t, err)

	header = &JW3THeader{
		Algorithm:   "ed25519",
		AddressType: ss58AddressType,
	}

	err = SR25519HeaderValidationFn(header)
	assert.ErrorIs(t, err, ErrInvalidJW3TAlgorithm)

	header = &JW3THeader{
		Algorithm:   sr25519Algorithm,
		AddressType: "ss59",
	}

	err = SR25519HeaderValidationFn(header)
	assert.ErrorIs(t, err, ErrInvalidJW3TAddressType)
}

func Test_SR25519ProxyTypeValidationFn(t *testing.T) {
	delegateAccountID, err := types.NewAccountID(utils.RandomSlice(32))
	assert.NoError(t, err)

	delegatorAccountID, err := types.NewAccountID(utils.RandomSlice(32))
	assert.NoError(t, err)

	delegateAddress := subkey.SS58Encode(delegateAccountID.ToBytes(), CentrifugeNetworkID)
	delegatorAddress := subkey.SS58Encode(delegatorAccountID.ToBytes(), CentrifugeNetworkID)

	tests := []struct {
		Name             string
		ProxyType        string
		DelegateAddress  string
		DelegatorAddress string
		ExpectedError    error
	}{
		{
			Name:             "Valid proxy type 1",
			ProxyType:        proxyType.ProxyTypeName[proxyType.Any],
			DelegateAddress:  "",
			DelegatorAddress: "",
			ExpectedError:    nil,
		},
		{
			Name:             "Valid proxy type 2",
			ProxyType:        proxyType.ProxyTypeName[proxyType.PodOperation],
			DelegateAddress:  "",
			DelegatorAddress: "",
			ExpectedError:    nil,
		},
		{
			Name:             "Valid proxy type 3",
			ProxyType:        proxyType.ProxyTypeName[proxyType.PodAuth],
			DelegateAddress:  "",
			DelegatorAddress: "",
			ExpectedError:    nil,
		},
		{
			Name:             "Valid proxy type 4",
			ProxyType:        PodAdminProxyType,
			DelegateAddress:  delegateAddress,
			DelegatorAddress: delegateAddress,
			ExpectedError:    nil,
		},
		{
			Name:             "Invalid admin proxy type",
			ProxyType:        PodAdminProxyType,
			DelegateAddress:  delegateAddress,
			DelegatorAddress: delegatorAddress,
			ExpectedError:    ErrAdminAddressesMismatch,
		},
		{
			Name:             "Invalid proxy type 1",
			ProxyType:        "invalid_proxy_type",
			DelegateAddress:  "",
			DelegatorAddress: "",
			ExpectedError:    ErrInvalidProxyType,
		},
		{
			Name:             "Invalid proxy type 2",
			ProxyType:        proxyType.ProxyTypeName[proxyType.Borrow],
			DelegateAddress:  "",
			DelegatorAddress: "",
			ExpectedError:    ErrInvalidProxyType,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			payload := &JW3TPayload{
				Address:    test.DelegateAddress,
				OnBehalfOf: test.DelegatorAddress,
				ProxyType:  test.ProxyType,
			}

			err := SR25519ProxyTypeValidationFn(payload)

			if test.ExpectedError == nil {
				assert.NoError(t, err)
				return
			}

			assert.ErrorIs(t, err, test.ExpectedError)
		})
	}
}

func Test_SR25519SignatureValidationFn(t *testing.T) {
	tokenHeader := &JW3THeader{
		Algorithm:   sr25519Algorithm,
		AddressType: ss58AddressType,
		TokenType:   jw3TokenType,
	}

	jsonHeader, err := json.Marshal(tokenHeader)
	assert.NoError(t, err)

	delegateAccountID, err := types.NewAccountID(keyrings.AliceKeyRingPair.PublicKey)
	assert.NoError(t, err)

	delegatorAccountID, err := types.NewAccountID(keyrings.BobKeyRingPair.PublicKey)
	assert.NoError(t, err)

	now := time.Now()
	expireTime := now.Add(24 * time.Hour)

	delegateAddress := subkey.SS58Encode(delegateAccountID.ToBytes(), CentrifugeNetworkID)
	delegatorAddress := subkey.SS58Encode(delegatorAccountID.ToBytes(), CentrifugeNetworkID)

	tokenPayload := &JW3TPayload{
		IssuedAt:   fmt.Sprintf("%d", now.Unix()),
		NotBefore:  fmt.Sprintf("%d", now.Unix()),
		ExpiresAt:  fmt.Sprintf("%d", expireTime.Unix()),
		Address:    delegateAddress,
		OnBehalfOf: delegatorAddress,
		ProxyType:  proxyType.ProxyTypeName[proxyType.Any],
	}

	jsonPayload, err := json.Marshal(tokenPayload)
	assert.NoError(t, err)

	message := strings.Join([]string{
		string(jsonHeader),
		string(jsonPayload),
	},
		tokenSeparator,
	)

	kp, err := subkey.DeriveKeyPair(sr25519.Scheme{}, keyrings.AliceKeyRingPair.URI)
	assert.NoError(t, err)

	sig, err := kp.Sign([]byte(message))
	assert.NoError(t, err)

	token := &JW3Token{
		Payload: &JW3TPayload{
			Address: delegateAddress,
		},
		JSONHeader:  jsonHeader,
		JSONPayload: jsonPayload,
		Signature:   sig,
	}

	err = sr25519SignatureValidationFn(token)
	assert.NoError(t, err)
}

func Test_SR25519SignatureValidationFn_WithWrappedSignedMessage(t *testing.T) {
	tokenHeader := &JW3THeader{
		Algorithm:   sr25519Algorithm,
		AddressType: ss58AddressType,
		TokenType:   jw3TokenType,
	}

	jsonHeader, err := json.Marshal(tokenHeader)
	assert.NoError(t, err)

	delegateAccountID, err := types.NewAccountID(keyrings.AliceKeyRingPair.PublicKey)
	assert.NoError(t, err)

	delegatorAccountID, err := types.NewAccountID(keyrings.BobKeyRingPair.PublicKey)
	assert.NoError(t, err)

	now := time.Now()
	expireTime := now.Add(24 * time.Hour)

	delegateAddress := subkey.SS58Encode(delegateAccountID.ToBytes(), CentrifugeNetworkID)
	delegatorAddress := subkey.SS58Encode(delegatorAccountID.ToBytes(), CentrifugeNetworkID)

	tokenPayload := &JW3TPayload{
		IssuedAt:   fmt.Sprintf("%d", now.Unix()),
		NotBefore:  fmt.Sprintf("%d", now.Unix()),
		ExpiresAt:  fmt.Sprintf("%d", expireTime.Unix()),
		Address:    delegateAddress,
		OnBehalfOf: delegatorAddress,
		ProxyType:  proxyType.ProxyTypeName[proxyType.Any],
	}

	jsonPayload, err := json.Marshal(tokenPayload)
	assert.NoError(t, err)

	message := strings.Join([]string{
		string(jsonHeader),
		string(jsonPayload),
	},
		tokenSeparator,
	)

	kp, err := subkey.DeriveKeyPair(sr25519.Scheme{}, keyrings.AliceKeyRingPair.URI)
	assert.NoError(t, err)

	sig, err := kp.Sign(wrapSignedMessage(message))
	assert.NoError(t, err)

	token := &JW3Token{
		Payload: &JW3TPayload{
			Address: delegateAddress,
		},
		JSONHeader:  jsonHeader,
		JSONPayload: jsonPayload,
		Signature:   sig,
	}

	err = sr25519SignatureValidationFn(token)
	assert.NoError(t, err)
}

func Test_SR25519SignatureValidationFn_InvalidDelegateAddress(t *testing.T) {
	token := &JW3Token{
		Payload: &JW3TPayload{
			Address: "invalid_address",
		},
	}

	err := sr25519SignatureValidationFn(token)
	assert.ErrorIs(t, err, ErrSS58AddressDecode)
}

func Test_SR25519SignatureValidationFn_InvalidSignature(t *testing.T) {
	tokenHeader := &JW3THeader{
		Algorithm:   sr25519Algorithm,
		AddressType: ss58AddressType,
		TokenType:   jw3TokenType,
	}

	jsonHeader, err := json.Marshal(tokenHeader)
	assert.NoError(t, err)

	delegateAccountID, err := types.NewAccountID(keyrings.AliceKeyRingPair.PublicKey)
	assert.NoError(t, err)

	delegatorAccountID, err := types.NewAccountID(keyrings.BobKeyRingPair.PublicKey)
	assert.NoError(t, err)

	now := time.Now()
	expireTime := now.Add(24 * time.Hour)

	delegateAddress := subkey.SS58Encode(delegateAccountID.ToBytes(), CentrifugeNetworkID)
	delegatorAddress := subkey.SS58Encode(delegatorAccountID.ToBytes(), CentrifugeNetworkID)

	tokenPayload := &JW3TPayload{
		IssuedAt:   fmt.Sprintf("%d", now.Unix()),
		NotBefore:  fmt.Sprintf("%d", now.Unix()),
		ExpiresAt:  fmt.Sprintf("%d", expireTime.Unix()),
		Address:    delegateAddress,
		OnBehalfOf: delegatorAddress,
		ProxyType:  proxyType.ProxyTypeName[proxyType.Any],
	}

	jsonPayload, err := json.Marshal(tokenPayload)
	assert.NoError(t, err)

	token := &JW3Token{
		Payload: &JW3TPayload{
			Address: delegateAddress,
		},
		JSONHeader:  jsonHeader,
		JSONPayload: jsonPayload,
		Signature:   []byte("invalid_signature"),
	}

	err = sr25519SignatureValidationFn(token)
	assert.ErrorIs(t, err, ErrInvalidSignature)
}

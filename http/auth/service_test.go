//go:build unit

package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	configMocks "github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/pallets/proxy"
	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/common"
	"github.com/centrifuge/go-centrifuge/testingutils/keyrings"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
	"github.com/vedhavyas/go-subkey/v2"
	"github.com/vedhavyas/go-subkey/v2/sr25519"
)

func TestService_Validate(t *testing.T) {
	proxyAPIMock := proxy.NewAPIMock(t)
	configServiceMock := configMocks.NewServiceMock(t)

	srv := NewService(true, proxyAPIMock, configServiceMock)

	// Bob is a proxy of Alice.

	delegateAccountID, err := types.NewAccountID(keyrings.BobKeyRingPair.PublicKey)
	assert.NoError(t, err)

	delegatorAccountID, err := types.NewAccountID(keyrings.AliceKeyRingPair.PublicKey)
	assert.NoError(t, err)

	token, err := CreateJW3Token(
		delegateAccountID,
		delegatorAccountID,
		keyrings.BobKeyRingPair.URI,
		"any",
	)
	assert.NoError(t, err)

	ctx := context.Background()

	configServiceMock.On("GetAccount", delegatorAccountID.ToBytes()).
		Return(nil, nil).
		Once()

	proxyRes := &types.ProxyStorageEntry{
		ProxyDefinitions: []types.ProxyDefinition{
			{
				Delegate:  *delegateAccountID,
				ProxyType: 0,
			},
		},
	}

	proxyAPIMock.On("GetProxies", delegatorAccountID).
		Return(proxyRes, nil).
		Once()

	res, err := srv.Validate(ctx, token)
	assert.NoError(t, err)
	assert.Equal(t, delegatorAccountID, res.Identity)
	assert.False(t, res.IsAdmin)
}

func TestService_Validate_DecodeError(t *testing.T) {
	proxyAPIMock := proxy.NewAPIMock(t)
	configServiceMock := configMocks.NewServiceMock(t)

	srv := NewService(true, proxyAPIMock, configServiceMock)

	ctx := context.Background()

	res, err := srv.Validate(ctx, "invalid_token")
	assert.ErrorIs(t, err, ErrInvalidJW3Token)
	assert.Nil(t, res)
}

func TestService_ParseError(t *testing.T) {
	proxyAPIMock := proxy.NewAPIMock(t)
	configServiceMock := configMocks.NewServiceMock(t)

	srv := NewService(true, proxyAPIMock, configServiceMock)

	ctx := context.Background()

	// Token that can be parsed and base64 decoded, however, with invalid data.
	token := "aaa.bbb.ccc"

	res, err := srv.Validate(ctx, token)
	assert.ErrorIs(t, err, ErrJSONHeaderDecoding)
	assert.Nil(t, res)
}

func TestService_Validate_InvalidHeader(t *testing.T) {
	proxyAPIMock := proxy.NewAPIMock(t)
	configServiceMock := configMocks.NewServiceMock(t)

	srv := NewService(true, proxyAPIMock, configServiceMock)

	// Bob is a proxy of Alice.

	delegateAccountID, err := types.NewAccountID(keyrings.BobKeyRingPair.PublicKey)
	assert.NoError(t, err)

	delegatorAccountID, err := types.NewAccountID(keyrings.AliceKeyRingPair.PublicKey)
	assert.NoError(t, err)

	token, err := CreateJW3Token(
		delegateAccountID,
		delegatorAccountID,
		keyrings.BobKeyRingPair.URI,
		"any",
		func(header *JW3THeader, payload *JW3TPayload) {
			header.Algorithm = "invalid-algorithm"
		},
	)
	assert.NoError(t, err)

	ctx := context.Background()

	res, err := srv.Validate(ctx, token)
	assert.ErrorIs(t, err, ErrInvalidJW3TAlgorithm)
	assert.Nil(t, res)
}

func TestService_Validate_InvalidPayload(t *testing.T) {
	proxyAPIMock := proxy.NewAPIMock(t)
	configServiceMock := configMocks.NewServiceMock(t)

	srv := NewService(true, proxyAPIMock, configServiceMock)

	// Bob is a proxy of Alice.

	delegateAccountID, err := types.NewAccountID(keyrings.BobKeyRingPair.PublicKey)
	assert.NoError(t, err)

	delegatorAccountID, err := types.NewAccountID(keyrings.AliceKeyRingPair.PublicKey)
	assert.NoError(t, err)

	token, err := CreateJW3Token(
		delegateAccountID,
		delegatorAccountID,
		keyrings.BobKeyRingPair.URI,
		"any",
		func(header *JW3THeader, payload *JW3TPayload) {
			payload.ProxyType = "invalid-proxy"
		},
	)
	assert.NoError(t, err)

	ctx := context.Background()

	res, err := srv.Validate(ctx, token)
	assert.ErrorIs(t, err, ErrInvalidProxyType)
	assert.Nil(t, res)
}

func TestService_Validate_InvalidDelegateAddress(t *testing.T) {
	proxyAPIMock := proxy.NewAPIMock(t)
	configServiceMock := configMocks.NewServiceMock(t)

	srv := NewService(true, proxyAPIMock, configServiceMock)

	// Bob is a proxy of Alice.

	delegateAccountID, err := types.NewAccountID(keyrings.BobKeyRingPair.PublicKey)
	assert.NoError(t, err)

	delegatorAccountID, err := types.NewAccountID(keyrings.AliceKeyRingPair.PublicKey)
	assert.NoError(t, err)

	token, err := CreateJW3Token(
		delegateAccountID,
		delegatorAccountID,
		keyrings.BobKeyRingPair.URI,
		"any",
		func(header *JW3THeader, payload *JW3TPayload) {
			payload.Address = "invalid_address"
		},
	)
	assert.NoError(t, err)

	ctx := context.Background()

	res, err := srv.Validate(ctx, token)
	assert.ErrorIs(t, err, ErrSS58AddressDecode)
	assert.Nil(t, res)
}

func TestService_Validate_InvalidSignature(t *testing.T) {
	proxyAPIMock := proxy.NewAPIMock(t)
	configServiceMock := configMocks.NewServiceMock(t)

	srv := NewService(true, proxyAPIMock, configServiceMock)

	// Bob is a proxy of Alice.

	delegateAccountID, err := types.NewAccountID(keyrings.BobKeyRingPair.PublicKey)
	assert.NoError(t, err)

	delegatorAccountID, err := types.NewAccountID(keyrings.AliceKeyRingPair.PublicKey)
	assert.NoError(t, err)

	token, err := CreateJW3Token(
		delegateAccountID,
		delegatorAccountID,
		keyrings.BobKeyRingPair.URI,
		"any",
		func(header *JW3THeader, payload *JW3TPayload) {
			// Replace the delegate address with the delegator address so that signature validation fails.
			payload.Address = subkey.SS58Encode(delegatorAccountID.ToBytes(), CentrifugeNetworkID)
		},
	)
	assert.NoError(t, err)

	ctx := context.Background()

	res, err := srv.Validate(ctx, token)
	assert.ErrorIs(t, err, ErrInvalidSignature)
	assert.Nil(t, res)
}

func TestService_Validate_InvalidDelegatorAddress(t *testing.T) {
	proxyAPIMock := proxy.NewAPIMock(t)
	configServiceMock := configMocks.NewServiceMock(t)

	srv := NewService(true, proxyAPIMock, configServiceMock)

	// Bob is a proxy of Alice.

	delegateAccountID, err := types.NewAccountID(keyrings.BobKeyRingPair.PublicKey)
	assert.NoError(t, err)

	delegatorAccountID, err := types.NewAccountID(keyrings.AliceKeyRingPair.PublicKey)
	assert.NoError(t, err)

	token, err := CreateJW3Token(
		delegateAccountID,
		delegatorAccountID,
		keyrings.BobKeyRingPair.URI,
		"any",
		func(header *JW3THeader, payload *JW3TPayload) {
			payload.OnBehalfOf = "invalid-address"
		},
	)
	assert.NoError(t, err)

	ctx := context.Background()

	res, err := srv.Validate(ctx, token)
	assert.ErrorIs(t, err, ErrSS58AddressDecode)
	assert.Nil(t, res)
}

func TestService_Validate_ConfigServiceError(t *testing.T) {
	proxyAPIMock := proxy.NewAPIMock(t)
	configServiceMock := configMocks.NewServiceMock(t)

	srv := NewService(true, proxyAPIMock, configServiceMock)

	// Bob is a proxy of Alice.

	delegateAccountID, err := types.NewAccountID(keyrings.BobKeyRingPair.PublicKey)
	assert.NoError(t, err)

	delegatorAccountID, err := types.NewAccountID(keyrings.AliceKeyRingPair.PublicKey)
	assert.NoError(t, err)

	token, err := CreateJW3Token(
		delegateAccountID,
		delegatorAccountID,
		keyrings.BobKeyRingPair.URI,
		"any",
	)
	assert.NoError(t, err)

	ctx := context.Background()

	configServiceMock.On("GetAccount", delegatorAccountID.ToBytes()).
		Return(nil, errors.New("error")).
		Once()

	res, err := srv.Validate(ctx, token)
	assert.ErrorIs(t, err, ErrInvalidIdentity)
	assert.Nil(t, res)
}

func TestService_Validate_ProxyServiceError(t *testing.T) {
	proxyAPIMock := proxy.NewAPIMock(t)
	configServiceMock := configMocks.NewServiceMock(t)

	srv := NewService(true, proxyAPIMock, configServiceMock)

	// Bob is a proxy of Alice.

	delegateAccountID, err := types.NewAccountID(keyrings.BobKeyRingPair.PublicKey)
	assert.NoError(t, err)

	delegatorAccountID, err := types.NewAccountID(keyrings.AliceKeyRingPair.PublicKey)
	assert.NoError(t, err)

	token, err := CreateJW3Token(
		delegateAccountID,
		delegatorAccountID,
		keyrings.BobKeyRingPair.URI,
		"any",
	)
	assert.NoError(t, err)

	ctx := context.Background()

	configServiceMock.On("GetAccount", delegatorAccountID.ToBytes()).
		Return(nil, nil).
		Once()

	proxyAPIMock.On("GetProxies", delegatorAccountID).
		Return(nil, errors.New("error")).
		Once()

	res, err := srv.Validate(ctx, token)
	assert.ErrorIs(t, err, ErrAccountProxiesRetrieval)
	assert.Nil(t, res)
}

func TestService_Validate_NotAProxy(t *testing.T) {
	proxyAPIMock := proxy.NewAPIMock(t)
	configServiceMock := configMocks.NewServiceMock(t)

	srv := NewService(true, proxyAPIMock, configServiceMock)

	// Bob is a proxy of Alice.

	delegateAccountID, err := types.NewAccountID(keyrings.BobKeyRingPair.PublicKey)
	assert.NoError(t, err)

	delegatorAccountID, err := types.NewAccountID(keyrings.AliceKeyRingPair.PublicKey)
	assert.NoError(t, err)

	token, err := CreateJW3Token(
		delegateAccountID,
		delegatorAccountID,
		keyrings.BobKeyRingPair.URI,
		"any",
	)
	assert.NoError(t, err)

	ctx := context.Background()

	configServiceMock.On("GetAccount", delegatorAccountID.ToBytes()).
		Return(nil, nil).
		Once()

	// The delegate is not part of the proxy response.
	proxyRes := &types.ProxyStorageEntry{
		ProxyDefinitions: []types.ProxyDefinition{
			{
				Delegate:  *delegatorAccountID,
				ProxyType: 0,
			},
		},
	}

	proxyAPIMock.On("GetProxies", delegatorAccountID).
		Return(proxyRes, nil).
		Once()

	res, err := srv.Validate(ctx, token)
	assert.ErrorIs(t, err, ErrInvalidDelegate)
	assert.Nil(t, res)
}

func TestService_Validate_ProxyTypeMismatch(t *testing.T) {
	proxyAPIMock := proxy.NewAPIMock(t)
	configServiceMock := configMocks.NewServiceMock(t)

	srv := NewService(true, proxyAPIMock, configServiceMock)

	// Bob is a proxy of Alice.

	delegateAccountID, err := types.NewAccountID(keyrings.BobKeyRingPair.PublicKey)
	assert.NoError(t, err)

	delegatorAccountID, err := types.NewAccountID(keyrings.AliceKeyRingPair.PublicKey)
	assert.NoError(t, err)

	token, err := CreateJW3Token(
		delegateAccountID,
		delegatorAccountID,
		keyrings.BobKeyRingPair.URI,
		"any",
	)
	assert.NoError(t, err)

	ctx := context.Background()

	configServiceMock.On("GetAccount", delegatorAccountID.ToBytes()).
		Return(nil, nil).
		Once()

	proxyRes := &types.ProxyStorageEntry{
		ProxyDefinitions: []types.ProxyDefinition{
			{
				Delegate: *delegateAccountID,
				// Proxy type any is 0.
				ProxyType: 11,
			},
		},
	}

	proxyAPIMock.On("GetProxies", delegatorAccountID).
		Return(proxyRes, nil).
		Once()

	res, err := srv.Validate(ctx, token)
	assert.ErrorIs(t, err, ErrInvalidDelegate)
	assert.Nil(t, res)
}

func TestService_Validate_PodAdmin(t *testing.T) {
	proxyAPIMock := proxy.NewAPIMock(t)
	configServiceMock := configMocks.NewServiceMock(t)

	srv := NewService(true, proxyAPIMock, configServiceMock)

	podAdminKeyPair, err := sr25519.Scheme{}.Generate()
	assert.NoError(t, err)

	podAdminAccountID, err := types.NewAccountID(podAdminKeyPair.AccountID())
	assert.NoError(t, err)

	token, err := CreateJW3Token(
		podAdminAccountID,
		podAdminAccountID,
		hexutil.Encode(podAdminKeyPair.Seed()),
		PodAdminProxyType,
	)
	assert.NoError(t, err)

	ctx := context.Background()

	podAdmin := configstore.NewPodAdmin(podAdminAccountID)

	configServiceMock.On("GetPodAdmin").
		Return(podAdmin, nil).
		Once()

	res, err := srv.Validate(ctx, token)
	assert.NoError(t, err)
	assert.Equal(t, podAdmin.GetAccountID(), res.Identity)
	assert.True(t, res.IsAdmin)
}

func TestService_Validate_PodAdmin_ConfigServiceError(t *testing.T) {
	proxyAPIMock := proxy.NewAPIMock(t)
	configServiceMock := configMocks.NewServiceMock(t)

	srv := NewService(true, proxyAPIMock, configServiceMock)

	podAdminKeyPair, err := sr25519.Scheme{}.Generate()
	assert.NoError(t, err)

	podAdminAccountID, err := types.NewAccountID(podAdminKeyPair.AccountID())
	assert.NoError(t, err)

	token, err := CreateJW3Token(
		podAdminAccountID,
		podAdminAccountID,
		hexutil.Encode(podAdminKeyPair.Seed()),
		PodAdminProxyType,
	)
	assert.NoError(t, err)

	ctx := context.Background()

	configServiceMock.On("GetPodAdmin").
		Return(nil, errors.New("error")).
		Once()

	res, err := srv.Validate(ctx, token)
	assert.ErrorIs(t, err, ErrPodAdminRetrieval)
	assert.Nil(t, res)
}

func TestService_Validate_PodAdmin_AdminAccountMismatch(t *testing.T) {
	proxyAPIMock := proxy.NewAPIMock(t)
	configServiceMock := configMocks.NewServiceMock(t)

	srv := NewService(true, proxyAPIMock, configServiceMock)

	podAdminKeyPair, err := sr25519.Scheme{}.Generate()
	assert.NoError(t, err)

	podAdminAccountID, err := types.NewAccountID(podAdminKeyPair.AccountID())
	assert.NoError(t, err)

	token, err := CreateJW3Token(
		podAdminAccountID,
		podAdminAccountID,
		hexutil.Encode(podAdminKeyPair.Seed()),
		PodAdminProxyType,
	)
	assert.NoError(t, err)

	ctx := context.Background()

	randomAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	podAdmin := configstore.NewPodAdmin(randomAccountID)

	configServiceMock.On("GetPodAdmin").
		Return(podAdmin, nil).
		Once()

	res, err := srv.Validate(ctx, token)
	assert.ErrorIs(t, err, ErrNotAdminAccount)
	assert.Nil(t, res)
}

func TestService_validateSignature(t *testing.T) {
	proxyAPIMock := proxy.NewAPIMock(t)
	configServiceMock := configMocks.NewServiceMock(t)

	srv := NewService(true, proxyAPIMock, configServiceMock)

	service := srv.(*service)

	delegateAccountID, err := types.NewAccountID(keyrings.BobKeyRingPair.PublicKey)
	assert.NoError(t, err)

	delegatorAccountID, err := types.NewAccountID(keyrings.AliceKeyRingPair.PublicKey)
	assert.NoError(t, err)

	token, err := CreateJW3Token(
		delegateAccountID,
		delegatorAccountID,
		keyrings.BobKeyRingPair.URI,
		"any",
	)
	assert.NoError(t, err)

	header, payload, signature, err := decodeJW3Token(token)
	assert.NoError(t, err)

	tests := []struct {
		Name              string
		Header            []byte
		Payload           []byte
		DelegateAccountID []byte
		Signature         []byte
		ExpectedError     bool
	}{
		{
			Name:              "valid signature",
			Header:            header,
			Payload:           payload,
			DelegateAccountID: delegateAccountID.ToBytes(),
			Signature:         signature,
			ExpectedError:     false,
		},
		{
			Name:              "invalid header",
			Header:            utils.RandomSlice(32),
			Payload:           payload,
			DelegateAccountID: delegateAccountID.ToBytes(),
			Signature:         signature,
			ExpectedError:     true,
		},
		{
			Name:              "invalid payload",
			Header:            header,
			Payload:           utils.RandomSlice(32),
			DelegateAccountID: delegateAccountID.ToBytes(),
			Signature:         signature,
			ExpectedError:     true,
		},
		{
			Name:              "invalid delegate",
			Header:            header,
			Payload:           payload,
			DelegateAccountID: delegatorAccountID.ToBytes(),
			Signature:         signature,
			ExpectedError:     true,
		},
		{
			Name:              "invalid signature",
			Header:            header,
			Payload:           payload,
			DelegateAccountID: delegateAccountID.ToBytes(),
			Signature:         utils.RandomSlice(32),
			ExpectedError:     true,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			err := service.validateSignature(test.Header, test.Payload, test.DelegateAccountID, test.Signature)

			if test.ExpectedError {
				assert.NotNil(t, err)
				return
			}

			assert.Nil(t, err)
		})
	}
}

func Test_NewAccountHeader(t *testing.T) {
	randomAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	randomAccountAddress := subkey.SS58Encode(randomAccountID.ToBytes(), CentrifugeNetworkID)

	tests := []struct {
		Name          string
		Payload       *JW3TPayload
		ExpectedError bool
		ExpectedAdmin bool
	}{
		{
			Name: "valid payload",
			Payload: &JW3TPayload{
				OnBehalfOf: randomAccountAddress,
				ProxyType:  "any",
			},
			ExpectedError: false,
			ExpectedAdmin: false,
		},
		{
			Name: "valid admin payload",
			Payload: &JW3TPayload{
				OnBehalfOf: randomAccountAddress,
				ProxyType:  PodAdminProxyType,
			},
			ExpectedError: false,
			ExpectedAdmin: true,
		},
		{
			Name: "invalid delegator address",
			Payload: &JW3TPayload{
				OnBehalfOf: "invalid-address",
				ProxyType:  "any",
			},
			ExpectedError: true,
			ExpectedAdmin: false,
		},
		{
			Name: "invalid proxy type",
			Payload: &JW3TPayload{
				OnBehalfOf: randomAccountAddress,
				ProxyType:  "invalid-proxy",
			},
			ExpectedError: true,
			ExpectedAdmin: false,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			res, err := NewAccountHeader(test.Payload)

			if test.ExpectedError {
				assert.NotNil(t, err)
				assert.Nil(t, res)
				return
			}

			assert.Nil(t, err)
			assert.Equal(t, test.ExpectedAdmin, res.IsAdmin)
		})
	}
}

func Test_decodeSS58Address(t *testing.T) {
	randomAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	randomAccountAddress := subkey.SS58Encode(randomAccountID.ToBytes(), CentrifugeNetworkID)

	res, err := decodeSS58Address(randomAccountAddress)
	assert.NoError(t, err)
	assert.Equal(t, randomAccountID, res)

	res, err = decodeSS58Address("invalid-address")
	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func Test_decodeJW3Token(t *testing.T) {
	delegateAccountID, err := types.NewAccountID(keyrings.BobKeyRingPair.PublicKey)
	assert.NoError(t, err)

	delegatorAccountID, err := types.NewAccountID(keyrings.AliceKeyRingPair.PublicKey)
	assert.NoError(t, err)

	token, err := CreateJW3Token(
		delegateAccountID,
		delegatorAccountID,
		keyrings.BobKeyRingPair.URI,
		"any",
	)
	assert.NoError(t, err)

	tests := []struct {
		Name          string
		Token         string
		ExpectedError error
	}{
		{
			Name:          "valid token",
			Token:         token,
			ExpectedError: nil,
		},
		{
			Name:          "invalid token",
			Token:         "aaa",
			ExpectedError: ErrInvalidJW3Token,
		},
		{
			Name:          "invalid token parts",
			Token:         "aaa.bbb",
			ExpectedError: ErrInvalidJW3Token,
		},
		// Use + as invalid base64 since we are using raw url encoding.
		{
			Name:          "invalid header part",
			Token:         "+++.bbb.ccc",
			ExpectedError: ErrBase64HeaderDecoding,
		},
		{
			Name:          "invalid payload part",
			Token:         "aaa.+++.ccc",
			ExpectedError: ErrBase64PayloadDecoding,
		},
		{
			Name:          "invalid signature part",
			Token:         "aaa.bbb.+++",
			ExpectedError: ErrBase64SignatureDecoding,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			header, payload, signature, err := decodeJW3Token(test.Token)

			if test.ExpectedError != nil {
				assert.ErrorIs(t, err, test.ExpectedError)
				assert.Nil(t, header)
				assert.Nil(t, payload)
				assert.Nil(t, signature)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, header)
			assert.NotNil(t, payload)
			assert.NotNil(t, signature)
		})
	}
}

func Test_parseHeaderAndPayload(t *testing.T) {
	jw3tHeader := JW3THeader{
		Algorithm:   "sr25519",
		AddressType: "ss58",
		TokenType:   "jw3t",
	}

	headerBytes, err := json.Marshal(jw3tHeader)
	assert.NoError(t, err)

	now := time.Now()

	jw3tPayload := JW3TPayload{
		IssuedAt:   fmt.Sprintf("%d", now.Unix()),
		NotBefore:  fmt.Sprintf("%d", now.Unix()),
		ExpiresAt:  fmt.Sprintf("%d", now.Unix()),
		Address:    "address",
		OnBehalfOf: "on-behalf-of",
		ProxyType:  "proxy-type",
	}

	payloadBytes, err := json.Marshal(jw3tPayload)
	assert.NoError(t, err)

	header, payload, err := parseHeaderAndPayload(headerBytes, payloadBytes)
	assert.NoError(t, err)
	assert.NotNil(t, header)
	assert.NotNil(t, payload)

	header, payload, err = parseHeaderAndPayload(utils.RandomSlice(32), payloadBytes)
	assert.ErrorIs(t, err, ErrJSONHeaderDecoding)
	assert.Nil(t, header)
	assert.Nil(t, payload)

	header, payload, err = parseHeaderAndPayload(headerBytes, utils.RandomSlice(32))
	assert.ErrorIs(t, err, ErrJSONPayloadDecoding)
	assert.Nil(t, header)
	assert.Nil(t, payload)
}

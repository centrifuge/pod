//go:build unit

package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	proxyType "github.com/centrifuge/chain-custom-types/pkg/proxy"

	"github.com/vedhavyas/go-subkey/v2"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"

	configMocks "github.com/centrifuge/go-centrifuge/config"
	v2proxy "github.com/centrifuge/go-centrifuge/identity/v2/proxy"

	"github.com/stretchr/testify/assert"
	"github.com/vedhavyas/go-subkey/v2/sr25519"
)

func formSignaturePayload(t *testing.T, header JW3THeader, payload JW3TPayload) string {
	headerJson, err := json.Marshal(header)
	assert.NoError(t, err)
	payloadJson, err := json.Marshal(payload)
	assert.NoError(t, err)

	return fmt.Sprintf("%s.%s", headerJson, payloadJson)
}

//func TestValidateSkipValidation(t *testing.T) {
//	h := JW3THeader{
//		Algorithm:   "sr25519",
//		AddressType: "ss58",
//		TokenType:   "JW3T",
//	}
//
//	kp, err := sr25519.Scheme{}.Generate()
//	assert.NoError(t, err)
//
//	issued := time.Now().UTC()
//	p := JW3TPayload{
//		IssuedAt:   fmt.Sprintf("%d", issued.Unix()),
//		NotBefore:  fmt.Sprintf("%d", issued.Unix()),
//		ExpiresAt:  fmt.Sprintf("%d", issued.Add(time.Hour).Unix()),
//		Address:    kp.SS58Address(36),
//		OnBehalfOf: "kANXoeY7KYbrzhyoDFypEWpijPRgKRx5G34ZX7TKbDBJwrVjp",
//		ProxyType:  "Any",
//	}
//
//	sigPayload := formSignaturePayload(t, h, p)
//	s, err := kp.Sign([]byte(sigPayload))
//	assert.NoError(t, err)
//
//	jw3tString := fmt.Sprintf("%s.%s", sigPayload, base64.RawURLEncoding.EncodeToString(s))
//	service := &service{}
//	accHeader, err := service.Validate(context.Background(), jw3tString)
//	assert.NoError(t, err)
//	assert.NotNil(t, accHeader)
//	assert.Equal(t, p.OnBehalfOf, accHeader.Identity)
//}

func TestValidate(t *testing.T) {
	ctx := context.Background()
	cfgSvc := configMocks.NewServiceMock(t)
	proxySvc := v2proxy.NewProxyAPIMock(t)

	authSrv := NewAuth(true, proxySvc, cfgSvc)

	header := JW3THeader{
		Algorithm:   "sr25519",
		AddressType: "ss58",
		TokenType:   "JW3T",
	}

	jsonHeader, err := json.Marshal(header)
	assert.NoError(t, err)

	base64Header := base64.RawURLEncoding.EncodeToString(jsonHeader)

	delegateKeyPair, err := sr25519.Scheme{}.Generate()
	assert.NoError(t, err)

	issued := time.Now().UTC()
	payload := JW3TPayload{
		IssuedAt:   fmt.Sprintf("%d", issued.Unix()),
		NotBefore:  fmt.Sprintf("%d", issued.Unix()),
		ExpiresAt:  fmt.Sprintf("%d", issued.Add(time.Hour).Unix()),
		Address:    delegateKeyPair.SS58Address(36),
		OnBehalfOf: "kANXoeY7KYbrzhyoDFypEWpijPRgKRx5G34ZX7TKbDBJwrVjp",
		ProxyType:  "any",
	}

	jsonPayload, err := json.Marshal(payload)
	assert.NoError(t, err)

	base64Payload := base64.RawURLEncoding.EncodeToString(jsonPayload)

	s, err := delegateKeyPair.Sign([]byte(formSignaturePayload(t, header, payload)))
	assert.NoError(t, err)

	base64Signature := base64.RawURLEncoding.EncodeToString(s)

	jw3tString := fmt.Sprintf("%s.%s.%s", base64Header, base64Payload, base64Signature)

	_, delegatorPubKey, err := subkey.SS58Decode(payload.OnBehalfOf)
	assert.NoError(t, err)

	delegatorAccountID, err := types.NewAccountID(delegatorPubKey)
	assert.NoError(t, err)

	cfgSvc.On("GetAccount", delegatorAccountID.ToBytes()).
		Once().
		Return(nil, errors.New("account not found"))

	// Account not found in storage
	accHeader, err := authSrv.Validate(ctx, jw3tString)
	assert.Error(t, err)
	assert.Nil(t, accHeader)

	// Account found in storage but invalid proxy
	cfgSvc.On("GetAccount", delegatorAccountID.ToBytes()).
		Once().
		Return(nil, nil)

	proxySvc.On("GetProxies", ctx, delegatorAccountID).
		Once().
		Return(nil, errors.New("invalid proxy"))

	accHeader, err = authSrv.Validate(ctx, jw3tString)
	assert.Error(t, err)
	assert.Nil(t, accHeader)

	// Proxy Account not member of the proxied Identity

	delegateAccountID, err := types.NewAccountID(delegateKeyPair.AccountID())
	assert.NoError(t, err)

	proxyRes := &types.ProxyStorageEntry{
		ProxyDefinitions: []types.ProxyDefinition{
			{
				Delegate:  *delegateAccountID,
				ProxyType: 0,
			},
		},
	}

	cfgSvc.On("GetAccount", delegatorAccountID.ToBytes()).
		Once().
		Return(nil, nil)

	proxySvc.On("GetProxies", ctx, delegatorAccountID).
		Once().
		Return(proxyRes, nil)

	accHeader, err = authSrv.Validate(ctx, jw3tString)
	assert.ErrorIs(t, err, ErrInvalidProxyType)
	assert.Nil(t, accHeader)

	// Proxy Account is member but not proxy type
	//proxyDef.Delegates[0].Delegate = types.NewAccountID(delegateKeyPair.Public())
	//proxyDef.Delegates[0].ProxyType = 1
	//proxySvc = new(proxy.MockService)
	//proxySvc.On("GetProxy", p.OnBehalfOf).Return(proxyDef, nil)
	//accHeader, err = authSrv.Validate(jw3tString, false)
	//assert.EqualError(t, err, JW3TInvalidProxyError)
	//assert.Nil(t, accHeader)

	// Success flow
	proxyRes.ProxyDefinitions[0].ProxyType = types.U8(proxyType.PodAuth)
	proxySvc.On("GetProxies", ctx, delegatorAccountID).Return(proxyRes, nil)

	accHeader, err = authSrv.Validate(ctx, jw3tString)
	assert.NoError(t, err)
	assert.NotNil(t, accHeader)
	assert.Equal(t, payload.OnBehalfOf, accHeader.Identity)
}

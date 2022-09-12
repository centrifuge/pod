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
	return fmt.Sprintf("%s.%s", base64.RawURLEncoding.EncodeToString(headerJson), base64.RawURLEncoding.EncodeToString(payloadJson))
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

	h := JW3THeader{
		Algorithm:   "sr25519",
		AddressType: "ss58",
		TokenType:   "JW3T",
	}

	kp, err := sr25519.Scheme{}.Generate()
	assert.NoError(t, err)

	issued := time.Now().UTC()
	p := JW3TPayload{
		IssuedAt:   fmt.Sprintf("%d", issued.Unix()),
		NotBefore:  fmt.Sprintf("%d", issued.Unix()),
		ExpiresAt:  fmt.Sprintf("%d", issued.Add(time.Hour).Unix()),
		Address:    kp.SS58Address(36),
		OnBehalfOf: "kANXoeY7KYbrzhyoDFypEWpijPRgKRx5G34ZX7TKbDBJwrVjp",
		ProxyType:  "Any",
	}

	_, proxiedPk, err := subkey.SS58Decode(p.OnBehalfOf)
	assert.NoError(t, err)

	sigPayload := formSignaturePayload(t, h, p)
	s, err := kp.Sign([]byte(sigPayload))
	assert.NoError(t, err)

	jw3tString := fmt.Sprintf("%s.%s", sigPayload, base64.RawURLEncoding.EncodeToString(s))

	cfgSvc := configMocks.NewServiceMock(t)
	cfgSvc.On("GetAccount", []byte(p.OnBehalfOf)).Return(nil, errors.New("account not found"))

	ctx := context.Background()
	// Account not found in storage
	authSrv := NewAuth(true, nil, cfgSvc)
	accHeader, err := authSrv.Validate(ctx, jw3tString)
	assert.Error(t, err)
	assert.Nil(t, accHeader)

	// Account found in storage but invalid proxy
	cfgSvc.On("GetAccount", []byte(p.OnBehalfOf)).Return(nil, nil)
	proxySvc := v2proxy.NewProxyAPIMock(t)
	proxySvc.On("GetProxy", p.OnBehalfOf).Return(nil, errors.New("invalid proxy"))
	authSrv = NewAuth(true, proxySvc, cfgSvc)

	accHeader, err = authSrv.Validate(ctx, jw3tString)
	assert.Error(t, err)
	assert.Nil(t, accHeader)

	// Proxy Account not member of the proxied Identity
	_, delegatorPubKey, err := subkey.SS58Decode(p.OnBehalfOf)
	assert.NoError(t, err)

	delegatorAccountID, err := types.NewAccountID(delegatorPubKey)
	assert.NoError(t, err)

	proxiedAccountID, err := types.NewAccountID(proxiedPk)
	assert.NoError(t, err)

	res := &types.ProxyStorageEntry{
		ProxyDefinitions: []types.ProxyDefinition{
			{
				Delegate:  *proxiedAccountID,
				ProxyType: 0,
			},
		},
	}

	proxySvc.On("GetProxies", ctx, delegatorAccountID).Return(res, nil)
	authSrv = NewAuth(true, proxySvc, cfgSvc)
	accHeader, err = authSrv.Validate(ctx, jw3tString)
	assert.ErrorIs(t, err, ErrInvalidProxyType)
	assert.Nil(t, accHeader)

	// Proxy Account is member but not proxy type
	//proxyDef.Delegates[0].Delegate = types.NewAccountID(kp.Public())
	//proxyDef.Delegates[0].ProxyType = 1
	//proxySvc = new(proxy.MockService)
	//proxySvc.On("GetProxy", p.OnBehalfOf).Return(proxyDef, nil)
	//authSrv = NewAuth(cfgSvc, proxySvc)
	//accHeader, err = authSrv.Validate(jw3tString, false)
	//assert.EqualError(t, err, JW3TInvalidProxyError)
	//assert.Nil(t, accHeader)

	// Success flow
	proxyDef.Delegates[0].ProxyType = 0
	proxySvc = new(proxy.MockService)
	proxySvc.On("GetProxy", p.OnBehalfOf).Return(proxyDef, nil)
	authSrv = NewAuth(cfgSvc, proxySvc)
	accHeader, err = authSrv.Validate(jw3tString, false)
	assert.NoError(t, err)
	assert.NotNil(t, accHeader)
	assert.Equal(t, p.OnBehalfOf, accHeader.Identity)
	assert.Equal(t, p.Address, accHeader.Signer)
	assert.Equal(t, p.ProxyType, accHeader.ProxyType)
}

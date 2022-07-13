//go:build unit

package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/proxy"
	"testing"
	"time"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/stretchr/testify/assert"
	"github.com/vedhavyas/go-subkey/v2"
	"github.com/vedhavyas/go-subkey/v2/sr25519"
)

func formSignaturePayload(t *testing.T, header JW3THeader, payload JW3TPayload) string {
	headerJson, err := json.Marshal(header)
	assert.NoError(t, err)
	payloadJson, err := json.Marshal(payload)
	assert.NoError(t, err)
	return fmt.Sprintf("%s.%s", base64.RawURLEncoding.EncodeToString(headerJson), base64.RawURLEncoding.EncodeToString(payloadJson))
}

func TestValidateSkipValidation(t *testing.T) {
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

	sigPayload := formSignaturePayload(t, h, p)
	s, err := kp.Sign([]byte(sigPayload))
	assert.NoError(t, err)

	jw3tString := fmt.Sprintf("%s.%s", sigPayload, base64.RawURLEncoding.EncodeToString(s))
	service := &service{}
	accHeader, err := service.Validate(context.Background(), jw3tString)
	assert.NoError(t, err)
	assert.NotNil(t, accHeader)
	assert.Equal(t, p.OnBehalfOf, accHeader.Identity)
	assert.Equal(t, p.Address, accHeader.Signer)
	assert.Equal(t, p.ProxyType, accHeader.ProxyType)
}

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

	cfgSvc := new(configstore.MockService)
	cfgSvc.On("GetAccount", []byte(p.OnBehalfOf)).Return(nil, errors.New("account not found"))

	// Account not found in storage
	authSrv := NewAuth(cfgSvc, nil)
	accHeader, err := authSrv.Validate(jw3tString, false)
	assert.Error(t, err)
	assert.Nil(t, accHeader)

	// Account found in storage but invalid proxy
	cfgSvc = new(configstore.MockService)
	cfgSvc.On("GetAccount", []byte(p.OnBehalfOf)).Return(nil, nil)
	proxySvc := new(proxy.MockService)
	proxySvc.On("GetProxy", p.OnBehalfOf).Return(nil, errors.New("invalid proxy"))
	authSrv = NewAuth(cfgSvc, proxySvc)
	accHeader, err = authSrv.Validate(jw3tString, false)
	assert.Error(t, err)
	assert.Nil(t, accHeader)

	// Proxy Account not member of the proxied Identity
	proxyDef := &proxy.Definition{
		Delegates: []proxy.Delegate{
			{
				Delegate:  types.NewAccountID(proxiedPk),
				ProxyType: 0,
				Delay:     0,
			},
		},
		Amount: types.U128{},
	}
	proxySvc = new(proxy.MockService)
	proxySvc.On("GetProxy", p.OnBehalfOf).Return(proxyDef, nil)
	authSrv = NewAuth(cfgSvc, proxySvc)
	accHeader, err = authSrv.Validate(jw3tString, false)
	assert.EqualError(t, err, JW3TInvalidProxyError)
	assert.Nil(t, accHeader)

	// Proxy Account is member but not proxy type
	proxyDef.Delegates[0].Delegate = types.NewAccountID(kp.Public())
	proxyDef.Delegates[0].ProxyType = 1
	proxySvc = new(proxy.MockService)
	proxySvc.On("GetProxy", p.OnBehalfOf).Return(proxyDef, nil)
	authSrv = NewAuth(cfgSvc, proxySvc)
	accHeader, err = authSrv.Validate(jw3tString, false)
	assert.EqualError(t, err, JW3TInvalidProxyError)
	assert.Nil(t, accHeader)

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

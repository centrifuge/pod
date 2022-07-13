package auth

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/crypto"
	"github.com/centrifuge/go-centrifuge/proxy"
	"github.com/vedhavyas/go-subkey/v2"
)

const (
	BadJW3TFormatError        = "bad JW3T format"
	InvalidJW3TAlgError       = "invalid JW3T algorithm"
	JW3TNotActiveError        = "JW3T not active yet"
	JW3TExpiredError          = "JW3T expired"
	JW3TSignatureInvalidError = "JW3T signature not valid"
	JW3TInvalidProxyError     = "not valid delegate for proxy"
)

type JW3THeader struct {
	Algorithm   string `json:"algorithm"`
	AddressType string `json:"address-type"`
	TokenType   string `json:"token-type"`
}

type JW3TPayload struct {
	IssuedAt   string `json:"issued-at"`
	NotBefore  string `json:"not-before"`
	ExpiresAt  string `json:"expires-at"`
	Address    string `json:"address"`
	OnBehalfOf string `json:"on-behalf-of"`
	ProxyType  string `json:"proxy-type"`
}

type AccountHeader struct {
	Identity  string
	Signer    string
	ProxyType string
}

type Auth interface {
	Validate(jw3t string, skipAuthentication bool) (*AccountHeader, error)
}

type auth struct {
	proxySrv  proxy.Service
	configSrv config.Service
}

func NewAuth(configSrv config.Service, proxySrv proxy.Service) Auth {
	return &auth{
		proxySrv:  proxySrv,
		configSrv: configSrv,
	}
}

func decodeJW3T(jw3t string) (*JW3THeader, *JW3TPayload, []byte, error) {
	fragments := strings.Split(jw3t, ".")
	if len(fragments) != 3 {
		return nil, nil, nil, errors.New(BadJW3TFormatError)
	}

	headerJSONText, err := base64.RawURLEncoding.DecodeString(fragments[0])
	if err != nil {
		return nil, nil, nil, err
	}

	var jw3tHeader JW3THeader
	err = json.Unmarshal(headerJSONText, &jw3tHeader)
	if err != nil {
		return nil, nil, nil, err
	}

	payloadJSONText, err := base64.RawURLEncoding.DecodeString(fragments[1])
	if err != nil {
		return nil, nil, nil, err
	}

	var jw3tPayload JW3TPayload
	err = json.Unmarshal(payloadJSONText, &jw3tPayload)
	if err != nil {
		return nil, nil, nil, err
	}

	signature, err := base64.RawURLEncoding.DecodeString(fragments[2])
	if err != nil {
		return nil, nil, nil, err
	}

	return &jw3tHeader, &jw3tPayload, signature, nil
}

func (a auth) Validate(jw3t string, skipAuthentication bool) (*AccountHeader, error) {
	jw3tHeader, jw3tPayload, signature, err := decodeJW3T(jw3t)
	if err != nil {
		return nil, err
	}

	if skipAuthentication {
		return &AccountHeader{
			Identity:  jw3tPayload.OnBehalfOf,
			Signer:    jw3tPayload.Address,
			ProxyType: jw3tPayload.ProxyType,
		}, nil
	}

	// Check on supported algorithms
	if jw3tHeader.Algorithm != "sr25519" {
		return nil, fmt.Errorf("%s: %s", InvalidJW3TAlgError, jw3tHeader.Algorithm)
	}

	// Validating Timestamps
	i, err := strconv.ParseInt(jw3tPayload.NotBefore, 10, 64)
	if err != nil {
		return nil, err
	}

	tm := time.Unix(i, 0).UTC()
	if tm.After(time.Now().UTC()) {
		return nil, errors.New(JW3TNotActiveError)
	}

	i, err = strconv.ParseInt(jw3tPayload.ExpiresAt, 10, 64)
	if err != nil {
		return nil, err
	}
	tm = time.Unix(i, 0).UTC()
	if tm.Before(time.Now().UTC()) {
		return nil, errors.New(JW3TExpiredError)
	}

	// Validate Signature
	_, publicKey, err := subkey.SS58Decode(jw3tPayload.Address)
	if err != nil {
		return nil, err
	}
	toSign := strings.Join(strings.Split(jw3t, ".")[:2], ".")
	valid := crypto.VerifyMessage(publicKey, []byte(toSign), signature, crypto.CurveSr25519)
	if !valid {
		return nil, errors.New(JW3TSignatureInvalidError)
	}

	if jw3tPayload.ProxyType == proxy.NodeAdminRole {
		// TODO: check that a.configSrv.GetAdminAccount() matches with jw3tPayload.Address return error otherwise
		return &AccountHeader{
			Identity:  jw3tPayload.Address,
			Signer:    jw3tPayload.Address,
			ProxyType: proxy.NodeAdminRole,
		}, nil
	}

	// Verify OnBehalfOf is a valid Identity on the node
	_, err = a.configSrv.GetAccount([]byte(jw3tPayload.OnBehalfOf))
	if err != nil {
		return nil, err
	}

	// Verify that Address is a valid proxy of OnBehalfOf against the Proxy Pallet with the desired level ProxyType
	proxyDef, err := a.proxySrv.GetProxy(jw3tPayload.OnBehalfOf)
	if err != nil {
		return nil, err
	}

	if !a.proxySrv.ProxyHasProxyType(proxyDef, publicKey, jw3tPayload.ProxyType) {
		return nil, errors.New(JW3TInvalidProxyError)
	}

	return &AccountHeader{
		Identity:  jw3tPayload.OnBehalfOf,
		Signer:    jw3tPayload.Address,
		ProxyType: jw3tPayload.ProxyType,
	}, nil
}

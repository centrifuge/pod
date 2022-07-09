package auth

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/centrifuge/go-centrifuge/centchain"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/crypto"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/vedhavyas/go-subkey/v2"
	"strconv"
	"strings"
	"time"
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

type Auth struct {
	centchainSrv centchain.API
	configSrv    config.Service
}

func NewAuth(centchainSrv centchain.API, configSrv config.Service) Auth {
	return Auth{
		centchainSrv: centchainSrv,
		configSrv:    configSrv,
	}
}

func decodeJW3T(jw3t string) (*JW3THeader, *JW3TPayload, []byte, error) {
	fragments := strings.Split(jw3t, ".")
	if len(fragments) != 3 {
		return nil, nil, nil, errors.New("bad JW3T format")
	}

	headerJsonText, err := base64.RawURLEncoding.DecodeString(fragments[0])
	if err != nil {
		return nil, nil, nil, err
	}

	var jw3tHeader JW3THeader
	err = json.Unmarshal(headerJsonText, &jw3tHeader)
	if err != nil {
		return nil, nil, nil, err
	}

	payloadJsonText, err := base64.RawURLEncoding.DecodeString(fragments[1])
	if err != nil {
		return nil, nil, nil, err
	}

	var jw3tPayload JW3TPayload
	err = json.Unmarshal(payloadJsonText, &jw3tPayload)
	if err != nil {
		return nil, nil, nil, err
	}

	signature, err := base64.RawURLEncoding.DecodeString(fragments[2])
	if err != nil {
		return nil, nil, nil, err
	}

	return &jw3tHeader, &jw3tPayload, signature, nil
}

func (a Auth) Validate(jw3t string) (*AccountHeader, error) {
	jw3tHeader, jw3tPayload, signature, err := decodeJW3T(jw3t)
	if err != nil {
		return nil, err
	}

	// Check on supported algorithms
	if jw3tHeader.Algorithm != "sr25519" {
		return nil, errors.New(fmt.Sprintf("Invalid JW3T Algorithm: %s", jw3tHeader.Algorithm))
	}

	// Validating Timestamps
	i, err := strconv.ParseInt(jw3tPayload.NotBefore, 10, 64)
	if err != nil {
		return nil, err
	}

	tm := time.Unix(i, 0).UTC()
	if tm.After(time.Now().UTC()) {
		return nil, errors.New("JW3T Not active yet")
	}

	i, err = strconv.ParseInt(jw3tPayload.ExpiresAt, 10, 64)
	if err != nil {
		return nil, err
	}
	tm = time.Unix(i, 0).UTC()
	if tm.Before(time.Now().UTC()) {
		return nil, errors.New("JW3T Expired")
	}

	// Validate Signature
	_, publicKey, err := subkey.SS58Decode(jw3tPayload.Address)
	if err != nil {
		return nil, err
	}
	toSign := strings.Join(strings.Split(jw3t, ".")[:2], ".")
	valid := crypto.VerifyMessage(publicKey, []byte(toSign), signature, crypto.CurveSr25519)
	if !valid {
		return nil, errors.New("JW3T signature not valid")
	}

	if jw3tPayload.ProxyType == identity.NodeAdminRole {
		// check that a.configSrv.GetAdminAccount() matches with jw3tPayload.Address return error otherwise
		return &AccountHeader{
			Identity:  jw3tPayload.Address,
			Signer:    jw3tPayload.Address,
			ProxyType: identity.NodeAdminRole,
		}, nil
	}

	// Verify OnBehalfOf is a valid Identity on the node
	_, err = a.configSrv.GetAccount([]byte(jw3tPayload.OnBehalfOf))
	if err != nil {
		return nil, err
	}

	// Verify that Address is a valid proxy of OnBehalfOf against the Proxy Pallet with the desired level ProxyType
	_, proxyPublicKey, err := subkey.SS58Decode(jw3tPayload.OnBehalfOf)
	if err != nil {
		return nil, err
	}
	meta, err := a.centchainSrv.GetMetadataLatest()
	if err != nil {
		return nil, err
	}

	key, err := types.CreateStorageKey(meta, identity.ProxyPallet, identity.ProxiesMethod, proxyPublicKey)
	if err != nil {
		return nil, err
	}

	var proxyDef identity.ProxyDefinition
	err = a.centchainSrv.GetStorageLatest(key, &proxyDef)
	if err != nil {
		return nil, fmt.Errorf("failed to get the proxy definition: %w", err)
	}

	valid = false
	for _, d := range proxyDef.Delegates {
		if utils.IsSameByteSlice(utils.Byte32ToSlice(d.Delegate), publicKey) {
			pxInt, err := strconv.Atoi(jw3tPayload.ProxyType)
			if err != nil {
				return nil, err
			}

			if uint8(d.ProxyType) == uint8(pxInt) {
				valid = true
				break
			}
		}

	}
	if !valid {
		return nil, errors.New("not valid delegate for proxy")
	}

	return &AccountHeader{
		Identity:  jw3tPayload.OnBehalfOf,
		Signer:    jw3tPayload.Address,
		ProxyType: jw3tPayload.ProxyType,
	}, nil
}

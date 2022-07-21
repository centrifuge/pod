package auth

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/crypto"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/identity/v2/proxy"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	logging "github.com/ipfs/go-log"
	"github.com/vedhavyas/go-subkey/v2"
)

type Service interface {
	Validate(ctx context.Context, token string) (*AccountHeader, error)
}

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
	Identity identity.DID
	IsAdmin  bool
}

func NewAccountHeader(payload *JW3TPayload) (*AccountHeader, error) {
	delegator, err := identity.NewDIDFromString(payload.OnBehalfOf)

	if err != nil {
		return nil, fmt.Errorf("couldn't create delegator identity: %w", err)
	}

	accountHeader := &AccountHeader{
		Identity: delegator,
	}

	payloadProxyType := strings.ToLower(payload.ProxyType)

	switch payloadProxyType {
	case NodeAdminProxyType:
		accountHeader.IsAdmin = true
	default:
		if _, ok := types.ProxyTypeValue[payloadProxyType]; !ok {
			return nil, fmt.Errorf("invalid proxy type - %s", payload.ProxyType)
		}
	}

	return accountHeader, nil
}

type service struct {
	authenticationEnabled bool
	log                   *logging.ZapEventLogger
	proxyAPI              proxy.API
	configSrv             config.Service
}

func NewAuth(
	authenticationEnabled bool,
	proxyAPI proxy.API,
	configSrv config.Service,
) Service {
	log := logging.Logger("http-auth")

	return &service{
		authenticationEnabled: authenticationEnabled,
		log:                   log,
		proxyAPI:              proxyAPI,
		configSrv:             configSrv,
	}
}

const (
	NodeAdminProxyType = "node_admin"

	tokenSeparator = "."
)

func (s *service) Validate(ctx context.Context, token string) (*AccountHeader, error) {
	jw3tHeader, jw3tPayload, signature, err := decodeJW3T(token)

	if err != nil {
		s.log.Errorf("Couldn't decode JW3T: %s", err)
		return nil, err
	}

	if !s.authenticationEnabled {
		return NewAccountHeader(jw3tPayload)
	}

	// Check on supported algorithms
	if jw3tHeader.Algorithm != "sr25519" {
		s.log.Errorf("Invalid JW3T algorithm")

		return nil, ErrInvalidJW3TAlgorithm
	}

	// Validating Timestamps
	i, err := strconv.ParseInt(jw3tPayload.NotBefore, 10, 64)
	if err != nil {
		s.log.Errorf("Invalid NotBefore timestamp: %s", err)

		return nil, ErrInvalidNotBeforeTimestamp
	}

	tm := time.Unix(i, 0).UTC()

	if tm.After(time.Now().UTC()) {
		s.log.Errorf("Inactive token")

		return nil, ErrInactiveToken
	}

	i, err = strconv.ParseInt(jw3tPayload.ExpiresAt, 10, 64)
	if err != nil {
		s.log.Errorf("Invalid ExpiresAt timestamp: %s", err)

		return nil, ErrInvalidExpiresAtTimestamp
	}

	tm = time.Unix(i, 0).UTC()
	if tm.Before(time.Now().UTC()) {
		s.log.Errorf("Token expired")

		return nil, ErrExpiredToken
	}

	// Validate Signature
	_, delegatePublicKey, err := subkey.SS58Decode(jw3tPayload.Address)
	if err != nil {
		s.log.Errorf("Invalid delegate address: %s", err)

		return nil, ErrInvalidDelegateAddress
	}

	toSign := strings.Join(strings.Split(token, tokenSeparator)[:2], tokenSeparator)
	valid := crypto.VerifyMessage(delegatePublicKey, []byte(toSign), signature, crypto.CurveSr25519)
	if !valid {
		s.log.Errorf("Invalid signature")

		return nil, ErrInvalidSignature
	}

	if jw3tPayload.ProxyType == NodeAdminProxyType {
		// TODO(cdamian): Check that a.configSrv.GetAdminAccount() matches with jw3tPayload.Address return error otherwise
		// TODO(cdamian): Check if we really want to combine known proxy types with node specific proxy types
		return NewAccountHeader(jw3tPayload)
	}

	// Verify OnBehalfOf is a valid Identity on the node
	_, err = s.configSrv.GetAccount([]byte(jw3tPayload.OnBehalfOf))
	if err != nil {
		s.log.Errorf("Invalid identity: %s", err)

		return nil, ErrInvalidIdentity
	}

	// Verify that Address is a valid proxy of OnBehalfOf against the Proxy Pallet with the desired level ProxyType
	_, delegatorPublicKey, err := subkey.SS58Decode(jw3tPayload.OnBehalfOf)
	if err != nil {
		s.log.Errorf("Invalid identity address: %s", err)

		return nil, ErrInvalidIdentityAddress
	}

	accID := types.NewAccountID(delegatorPublicKey)

	proxyStorageEntry, err := s.proxyAPI.GetProxies(ctx, &accID)

	valid = false
	for _, proxyDefinition := range proxyStorageEntry.ProxyDefinitions {
		if bytes.Equal(proxyDefinition.Delegate[:], delegatePublicKey) {
			pt, ok := types.ProxyTypeValue[strings.ToLower(jw3tPayload.ProxyType)]

			if !ok {
				s.log.Errorf("Invalid proxy type: %s", jw3tPayload.ProxyType)

				return nil, ErrInvalidProxyType
			}

			if uint8(proxyDefinition.ProxyType) == uint8(pt) {
				valid = true
				break
			}
		}

	}
	if !valid {
		s.log.Errorf("Invalid delegate")

		return nil, ErrInvalidDelegate
	}

	return NewAccountHeader(jw3tPayload)
}

func decodeJW3T(jw3t string) (*JW3THeader, *JW3TPayload, []byte, error) {
	fragments := strings.Split(jw3t, tokenSeparator)
	if len(fragments) != 3 {
		return nil, nil, nil, ErrInvalidJW3Token
	}

	headerJsonText, err := base64.RawURLEncoding.DecodeString(fragments[0])
	if err != nil {
		return nil, nil, nil, ErrBase64HeaderDecoding
	}

	var jw3tHeader JW3THeader
	err = json.Unmarshal(headerJsonText, &jw3tHeader)
	if err != nil {
		return nil, nil, nil, ErrJSONHeaderDecoding
	}

	payloadJsonText, err := base64.RawURLEncoding.DecodeString(fragments[1])
	if err != nil {
		return nil, nil, nil, ErrBase64PayloadDecoding
	}

	var jw3tPayload JW3TPayload
	err = json.Unmarshal(payloadJsonText, &jw3tPayload)
	if err != nil {
		return nil, nil, nil, ErrJSONPayloadDecoding
	}

	signature, err := base64.RawURLEncoding.DecodeString(fragments[2])
	if err != nil {
		return nil, nil, nil, ErrBase64SignatureDecoding
	}

	return &jw3tHeader, &jw3tPayload, signature, nil
}

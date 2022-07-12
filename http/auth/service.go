package auth

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/crypto"
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
	Identity  string
	Signer    string
	ProxyType string
}

type service struct {
	log       *logging.ZapEventLogger
	proxyAPI  proxy.API
	configSrv config.Service
}

func NewAuth(proxyAPI proxy.API, configSrv config.Service) Service {
	log := logging.Logger("http-auth")

	return &service{
		log:       log,
		proxyAPI:  proxyAPI,
		configSrv: configSrv,
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
	_, publicKey, err := subkey.SS58Decode(jw3tPayload.Address)
	if err != nil {
		s.log.Errorf("Invalid delegate address: %s", err)

		return nil, ErrInvalidDelegateAddress
	}

	toSign := strings.Join(strings.Split(token, tokenSeparator)[:2], tokenSeparator)
	valid := crypto.VerifyMessage(publicKey, []byte(toSign), signature, crypto.CurveSr25519)
	if !valid {
		s.log.Errorf("Invalid signature")

		return nil, ErrInvalidSignature
	}

	if jw3tPayload.ProxyType == NodeAdminProxyType {
		// TODO(cdamian): Check that a.configSrv.GetAdminAccount() matches with jw3tPayload.Address return error otherwise
		return &AccountHeader{
			Identity:  jw3tPayload.Address,
			Signer:    jw3tPayload.Address,
			ProxyType: NodeAdminProxyType,
		}, nil
	}

	// Verify OnBehalfOf is a valid Identity on the node
	_, err = s.configSrv.GetAccount([]byte(jw3tPayload.OnBehalfOf))
	if err != nil {
		s.log.Errorf("Invalid identity: %s", err)

		return nil, ErrInvalidIdentity
	}

	// Verify that Address is a valid proxy of OnBehalfOf against the Proxy Pallet with the desired level ProxyType
	_, proxyPublicKey, err := subkey.SS58Decode(jw3tPayload.OnBehalfOf)
	if err != nil {
		s.log.Errorf("Invalid identity address: %s", err)

		return nil, ErrInvalidIdentityAddress
	}

	accID := types.NewAccountID(proxyPublicKey)

	proxyStorageEntry, err := s.proxyAPI.GetProxies(ctx, &accID)

	valid = false
	for _, proxyDefinition := range proxyStorageEntry.ProxyDefinitions {
		if bytes.Equal(proxyDefinition.Delegate[:], publicKey) {
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

	return &AccountHeader{
		Identity:  jw3tPayload.OnBehalfOf,
		Signer:    jw3tPayload.Address,
		ProxyType: jw3tPayload.ProxyType,
	}, nil
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

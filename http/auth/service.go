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

	"github.com/centrifuge/go-centrifuge/errors"

	proxyType "github.com/centrifuge/chain-custom-types/pkg/proxy"

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
	AddressType string `json:"address_type"`
	TokenType   string `json:"token_type"`
}

type JW3TPayload struct {
	IssuedAt   string `json:"issued_at"`
	NotBefore  string `json:"not_before"`
	ExpiresAt  string `json:"expires_at"`
	Address    string `json:"address"`
	OnBehalfOf string `json:"on_behalf_of"`
	ProxyType  string `json:"proxy_type"`
}

type AccountHeader struct {
	Identity *types.AccountID
	IsAdmin  bool
}

func NewAccountHeader(payload *JW3TPayload) (*AccountHeader, error) {
	_, delegatorPublicKey, err := subkey.SS58Decode(payload.OnBehalfOf)

	if err != nil {
		return nil, fmt.Errorf("couldn't decode delegator public key: %w", err)
	}

	delegator, err := types.NewAccountID(delegatorPublicKey)

	if err != nil {
		return nil, fmt.Errorf("couldn't create delegator account ID: %w", err)
	}

	accountHeader := &AccountHeader{
		Identity: delegator,
	}

	payloadProxyType := strings.ToLower(payload.ProxyType)

	switch payloadProxyType {
	case NodeAdminProxyType:
		accountHeader.IsAdmin = true
	default:
		if _, ok := proxyType.ProxyTypeValue[payloadProxyType]; !ok {
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

var (
	allowedProxyTypes = map[string]struct{}{
		NodeAdminProxyType:                              {},
		proxyType.ProxyTypeName[proxyType.Any]:          {},
		proxyType.ProxyTypeName[proxyType.PodOperation]: {},
		proxyType.ProxyTypeName[proxyType.PodAuth]:      {},
	}
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

	if _, ok := allowedProxyTypes[jw3tPayload.ProxyType]; !ok {
		s.log.Errorf("Unsupported proxy type: %s", jw3tPayload.ProxyType)

		return nil, ErrInvalidProxyType
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

	if err := s.validateSignature(token, delegatePublicKey, signature); err != nil {
		s.log.Errorf("Invalid signature: %s", err)

		return nil, ErrInvalidSignature

	}

	if jw3tPayload.ProxyType == NodeAdminProxyType {
		if err := s.validateAdminAccount(delegatePublicKey); err != nil {
			s.log.Errorf("Invalid admin account: %s", err)

			return nil, err
		}

		return NewAccountHeader(jw3tPayload)
	}

	_, delegatorPublicKey, err := subkey.SS58Decode(jw3tPayload.OnBehalfOf)
	if err != nil {
		s.log.Errorf("Invalid identity address: %s", err)

		return nil, ErrInvalidIdentityAddress
	}

	delegatorAccountID, err := types.NewAccountID(delegatorPublicKey)

	if err != nil {
		s.log.Errorf("Couldn't create delegator account ID: %s", err)

		return nil, ErrDelegatorAccountIDCreation
	}

	// Verify OnBehalfOf is a valid Identity on the node
	_, err = s.configSrv.GetAccount(delegatorAccountID.ToBytes())
	if err != nil {
		s.log.Errorf("Invalid identity: %s", err)

		return nil, ErrInvalidIdentity
	}

	// Verify that Address is a valid proxy of OnBehalfOf against the Proxy Pallet with the desired level ProxyType
	proxyStorageEntry, err := s.proxyAPI.GetProxies(ctx, delegatorAccountID)

	if err != nil {
		s.log.Errorf("Couldn't retrieve account proxies: %s", err)

		return nil, ErrAccountProxiesRetrieval
	}

	valid := false
	for _, proxyDefinition := range proxyStorageEntry.ProxyDefinitions {
		if bytes.Equal(proxyDefinition.Delegate[:], delegatePublicKey) {
			pt, ok := proxyType.ProxyTypeValue[strings.ToLower(jw3tPayload.ProxyType)]

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

func (s *service) validateSignature(
	token string,
	delegatePublicKey []byte,
	signature []byte,
) error {
	tokenParts := strings.Split(token, tokenSeparator)

	if len(tokenParts) != 3 {
		return errors.New("invalid token")
	}

	jsonHeader, err := base64.RawURLEncoding.DecodeString(tokenParts[0])

	if err != nil {
		return fmt.Errorf("couldn't decode header: %w", err)
	}

	jsonPayload, err := base64.RawURLEncoding.DecodeString(tokenParts[1])

	if err != nil {
		return fmt.Errorf("couldn't decode payload: %w", err)
	}

	// The message that is signed is in the form:
	//
	// <Bytes>json_header.json_payload</Bytes>
	//
	// Example:
	//
	// <Bytes>{
	//  "algorithm": "sr25519",
	//  "token_type": "JW3T",
	//  "address_type": "ss58"
	// }.{
	//  "address": "delegate_address",
	//  "on_behalf_of": "delegator_address",
	//  "proxy_type": "proxy_type",
	//  "expires_at": "1663070957",
	//  "issued_at": "1662984557",
	//  "not_before": "1662984557"
	// }</Bytes>
	wrappedMessage := wrapSignedMessage(
		strings.Join(
			[]string{
				string(jsonHeader),
				string(jsonPayload),
			},
			tokenSeparator,
		),
	)

	if !crypto.VerifyMessage(delegatePublicKey, wrappedMessage, signature, crypto.CurveSr25519) {
		return errors.New("invalid signature")
	}

	return nil
}

const (
	BytesPrefix = "<Bytes>"
	BytesSuffix = "</Bytes>"
)

// The polkadot JS extension that is signing the token header and payload
// is wrapping the initial message with BytesPrefix and BytesSuffix.
//
// As per:
// https://github.com/polkadot-js/extension/blob/607f4b3e3b045020659587771fd3eba7b3214862/packages/extension-base/src/background/RequestBytesSign.ts#L20
// https://github.com/polkadot-js/common/blob/11ab3a4f6ba652e8fcfe54b2a6b74e91bd30c693/packages/util/src/u8a/wrap.ts#L13-L14
func wrapSignedMessage(msg string) []byte {
	return []byte(BytesPrefix + msg + BytesSuffix)
}

func (s *service) validateAdminAccount(pubKey []byte) error {
	accountID, err := types.NewAccountID(pubKey)

	if err != nil {
		return ErrInvalidIdentityAddress
	}

	nodeAdmin, err := s.configSrv.GetNodeAdmin()

	if err != nil {
		return ErrNodeAdminRetrieval
	}

	if !nodeAdmin.GetAccountID().Equal(accountID) {
		return ErrNotAdminAccount
	}

	return nil
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

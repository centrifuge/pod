package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	proxyTypes "github.com/centrifuge/chain-custom-types/pkg/proxy"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/crypto"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/pallets/proxy"
	"github.com/centrifuge/go-centrifuge/validation"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	logging "github.com/ipfs/go-log"
	"github.com/vedhavyas/go-subkey/v2"
)

//go:generate mockery --name Service --structname ServiceMock --filename service_mock.go --inpackage

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
	delegatorAccountID, err := decodeSS58Address(payload.OnBehalfOf)

	if err != nil {
		return nil, fmt.Errorf("couldn't decode delegator address: %w", err)
	}

	accountHeader := &AccountHeader{
		Identity: delegatorAccountID,
	}

	switch payload.ProxyType {
	case PodAdminProxyType:
		accountHeader.IsAdmin = true
	default:
		if _, ok := proxyTypes.ProxyTypeValue[payload.ProxyType]; !ok {
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

func NewService(
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
	// PodAdminProxyType is a special type only used in the pod.
	PodAdminProxyType = "PodAdmin"

	tokenSeparator = "."
)

var (
	allowedProxyTypes = map[string]struct{}{
		PodAdminProxyType:                                 {},
		proxyTypes.ProxyTypeName[proxyTypes.Any]:          {},
		proxyTypes.ProxyTypeName[proxyTypes.PodOperation]: {},
		proxyTypes.ProxyTypeName[proxyTypes.PodAuth]:      {},
	}
)

func (s *service) Validate(_ context.Context, token string) (*AccountHeader, error) {
	header, payload, signature, err := decodeJW3Token(token)

	if err != nil {
		s.log.Errorf("Couldn't decode token: %s", err)

		return nil, err
	}

	jw3tHeader, jw3tPayload, err := parseHeaderAndPayload(header, payload)

	if err != nil {
		s.log.Errorf("Couldn't parse header and payload: %s", err)

		return nil, err
	}

	if !s.authenticationEnabled {
		return NewAccountHeader(jw3tPayload)
	}

	err = validation.Validate(
		validation.NewValidator(jw3tHeader, headerValidationFn),
		validation.NewValidator(jw3tPayload, payloadValidationFn),
	)

	if err != nil {
		s.log.Errorf("Invalid token: %s", err)

		return nil, err
	}

	// Validate Signature
	delegateAccountID, err := decodeSS58Address(jw3tPayload.Address)

	if err != nil {
		s.log.Errorf("Couldn't decode delegate address: %s", err)

		return nil, ErrSS58AddressDecode
	}

	if err := s.validateSignature(header, payload, delegateAccountID.ToBytes(), signature); err != nil {
		s.log.Errorf("Invalid signature: %s", err)

		return nil, ErrInvalidSignature

	}

	if jw3tPayload.ProxyType == PodAdminProxyType {
		if err := s.validateAdminAccount(delegateAccountID); err != nil {
			s.log.Errorf("Invalid admin account: %s", err)

			return nil, err
		}

		return NewAccountHeader(jw3tPayload)
	}

	delegatorAccountID, err := decodeSS58Address(jw3tPayload.OnBehalfOf)

	if err != nil {
		s.log.Errorf("Couldn't decode delegator address: %s", err)

		return nil, ErrSS58AddressDecode
	}

	// Verify OnBehalfOf is a valid Identity on the pod
	_, err = s.configSrv.GetAccount(delegatorAccountID.ToBytes())
	if err != nil {
		s.log.Errorf("Invalid identity: %s", err)

		return nil, ErrInvalidIdentity
	}

	// Verify that Address is a valid proxy of OnBehalfOf against the Proxy Pallet with the desired level ProxyType
	proxyStorageEntry, err := s.proxyAPI.GetProxies(delegatorAccountID)

	if err != nil {
		s.log.Errorf("Couldn't retrieve account proxies: %s", err)

		return nil, ErrAccountProxiesRetrieval
	}

	pt := proxyTypes.ProxyTypeValue[jw3tPayload.ProxyType]

	valid := false
	for _, proxyDefinition := range proxyStorageEntry.ProxyDefinitions {
		if proxyDefinition.Delegate.Equal(delegateAccountID) {
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

func decodeSS58Address(address string) (*types.AccountID, error) {
	_, publicKey, err := subkey.SS58Decode(address)
	if err != nil {
		return nil, err
	}

	return types.NewAccountID(publicKey)
}

func (s *service) validateSignature(
	header []byte,
	payload []byte,
	delegatePublicKey []byte,
	signature []byte,
) error {
	// A normal signed message would be in the form:
	//
	// json_header.json_payload
	signedMessage := strings.Join(
		[]string{
			string(header),
			string(payload),
		},
		tokenSeparator,
	)

	// The message that is signed by polkadot JS is in the form:
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

	// To avoid unnecessary client logic, check both possible variants.
	signedMessages := [][]byte{[]byte(signedMessage), wrapSignedMessage(signedMessage)}

	for _, signedMessage := range signedMessages {
		if crypto.VerifyMessage(delegatePublicKey, signedMessage, signature, crypto.CurveSr25519) {
			return nil
		}
	}

	return errors.New("invalid signature")
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

func (s *service) validateAdminAccount(accountID *types.AccountID) error {
	podAdmin, err := s.configSrv.GetPodAdmin()

	if err != nil {
		return ErrPodAdminRetrieval
	}

	if !podAdmin.GetAccountID().Equal(accountID) {
		return ErrNotAdminAccount
	}

	return nil
}

func decodeJW3Token(token string) ([]byte, []byte, []byte, error) {
	fragments := strings.Split(token, tokenSeparator)
	if len(fragments) != 3 {
		return nil, nil, nil, ErrInvalidJW3Token
	}

	headerBytes, err := base64.RawURLEncoding.DecodeString(fragments[0])
	if err != nil {
		return nil, nil, nil, ErrBase64HeaderDecoding
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(fragments[1])
	if err != nil {
		return nil, nil, nil, ErrBase64PayloadDecoding
	}

	signature, err := base64.RawURLEncoding.DecodeString(fragments[2])

	if err != nil {
		return nil, nil, nil, ErrBase64SignatureDecoding
	}

	return headerBytes, payloadBytes, signature, nil
}

func parseHeaderAndPayload(header, payload []byte) (*JW3THeader, *JW3TPayload, error) {
	var jw3tHeader JW3THeader

	if err := json.Unmarshal(header, &jw3tHeader); err != nil {
		return nil, nil, ErrJSONHeaderDecoding
	}

	var jw3tPayload JW3TPayload

	if err := json.Unmarshal(payload, &jw3tPayload); err != nil {
		return nil, nil, ErrJSONPayloadDecoding
	}

	return &jw3tHeader, &jw3tPayload, nil
}

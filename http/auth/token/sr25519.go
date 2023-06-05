package token

import (
	"strings"

	proxyTypes "github.com/centrifuge/chain-custom-types/pkg/proxy"
	"github.com/centrifuge/pod/crypto"
	"github.com/centrifuge/pod/validation"
)

type sr25519TokenValidatorFn func(token *JW3Token) error

func (v sr25519TokenValidatorFn) Validate(token *JW3Token) error {
	return v(token)
}

func NewSR25519TokenValidator(
	headerValidationFns []func(header *JW3THeader) error,
	payloadValidationFn []func(payload *JW3TPayload) error,
) Validator {
	return sr25519TokenValidatorFn(func(token *JW3Token) error {
		return validation.Validate(
			validation.NewValidator(token.Header, headerValidationFns...),
			validation.NewValidator(token.Payload, payloadValidationFn...),
			validation.NewValidator(token, sr25519SignatureValidationFn),
		)
	})
}

func DefaultSR25519TokenValidator() Validator {
	return NewSR25519TokenValidator(
		[]func(header *JW3THeader) error{
			BasicHeaderValidationFn,
			SR25519HeaderValidationFn,
		},
		[]func(payload *JW3TPayload) error{
			BasicPayloadValidationFn,
			SR25519ProxyTypeValidationFn,
		},
	)
}

const (
	sr25519Algorithm = "sr25519"
	ss58AddressType  = "ss58"
)

var (
	SR25519HeaderValidationFn = func(header *JW3THeader) error {
		if strings.ToLower(header.Algorithm) != sr25519Algorithm {
			return ErrInvalidJW3TAlgorithm
		}

		if strings.ToLower(header.AddressType) != ss58AddressType {
			return ErrInvalidJW3TAddressType
		}

		return nil
	}
)

const (
	// PodAdminProxyType is a special type only used in the POD.
	PodAdminProxyType = "PodAdmin"
)

var (
	allowedProxyTypes = map[string]struct{}{
		PodAdminProxyType:                                 {},
		proxyTypes.ProxyTypeName[proxyTypes.Any]:          {},
		proxyTypes.ProxyTypeName[proxyTypes.PodOperation]: {},
		proxyTypes.ProxyTypeName[proxyTypes.PodAuth]:      {},
	}
)

var (
	SR25519ProxyTypeValidationFn = func(payload *JW3TPayload) error {
		if _, ok := allowedProxyTypes[payload.ProxyType]; !ok {
			return ErrInvalidProxyType
		}

		if payload.ProxyType == PodAdminProxyType {
			return sr25519AdminPayloadValidation(payload)
		}

		return nil
	}

	sr25519AdminPayloadValidation = func(payload *JW3TPayload) error {
		if payload.Address != payload.OnBehalfOf {
			return ErrAdminAddressesMismatch
		}

		return nil
	}
)

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

var (
	sr25519SignatureValidationFn = func(token *JW3Token) error {
		delegateAccountID, err := DecodeSS58Address(token.Payload.Address)

		if err != nil {
			return ErrSS58AddressDecode
		}

		// A normal signed message would be in the form:
		//
		// json_header.json_payload
		signedMessage := strings.Join(
			[]string{
				string(token.JSONHeader),
				string(token.JSONPayload),
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
			if crypto.VerifyMessage(delegateAccountID.ToBytes(), signedMessage, token.Signature, crypto.CurveSr25519) {
				return nil
			}
		}

		return ErrInvalidSignature
	}
)

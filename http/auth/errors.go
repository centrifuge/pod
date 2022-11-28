package auth

import "github.com/centrifuge/go-centrifuge/errors"

const (
	ErrInvalidJW3Token           = errors.Error("invalid JW3T token")
	ErrBase64HeaderDecoding      = errors.Error("couldn't decode header JSON from base 64")
	ErrJSONHeaderDecoding        = errors.Error("couldn't unmarshal JW3T header from JSON")
	ErrBase64PayloadDecoding     = errors.Error("couldn't decode payload JSON from base 64")
	ErrJSONPayloadDecoding       = errors.Error("couldn't unmarshal payload from JSON")
	ErrBase64SignatureDecoding   = errors.Error("couldn't decode signature from base 64")
	ErrInvalidJW3TAlgorithm      = errors.Error("invalid JW3T algorithm")
	ErrInvalidJW3TAddressType    = errors.Error("invalid JW3T address type")
	ErrInvalidJW3TTokenType      = errors.Error("invalid JW3T token type")
	ErrInvalidNotBeforeTimestamp = errors.Error("invalid NotBefore timestamp")
	ErrInactiveToken             = errors.Error("token is not active yet")
	ErrInvalidExpiresAtTimestamp = errors.Error("invalid ExpiresAt timestamp")
	ErrExpiredToken              = errors.Error("token expired")
	ErrSS58AddressDecode         = errors.Error("couldn't decode SS58 address")
	ErrInvalidSignature          = errors.Error("invalid signature")
	ErrInvalidIdentity           = errors.Error("invalid identity")
	ErrAccountProxiesRetrieval   = errors.Error("couldn't retrieve account proxies")
	ErrInvalidProxyType          = errors.Error("invalid proxy type")
	ErrInvalidDelegate           = errors.Error("invalid delegate")
	ErrPodAdminRetrieval         = errors.Error("couldn't retrieve pod admin")
	ErrNotAdminAccount           = errors.Error("provided account is not an admin")
)

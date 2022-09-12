package auth

import "github.com/centrifuge/go-centrifuge/errors"

const (
	ErrInvalidJW3Token            = errors.Error("invalid JW3T token")
	ErrBase64HeaderDecoding       = errors.Error("couldn't decode header JSON from base 64")
	ErrJSONHeaderDecoding         = errors.Error("couldn't unmarshal JW3T header from JSON")
	ErrBase64PayloadDecoding      = errors.Error("couldn't decode payload JSON from base 64")
	ErrJSONPayloadDecoding        = errors.Error("couldn't unmarshal payload from JSON")
	ErrBase64SignatureDecoding    = errors.Error("couldn't decode signature from base 64")
	ErrInvalidJW3TAlgorithm       = errors.Error("invalid JW3T algorithm")
	ErrInvalidNotBeforeTimestamp  = errors.Error("invalid NotBefore timestamp")
	ErrInactiveToken              = errors.Error("token is not active yet")
	ErrInvalidExpiresAtTimestamp  = errors.Error("invalid ExpiresAt timestamp")
	ErrExpiredToken               = errors.Error("token expired")
	ErrInvalidDelegateAddress     = errors.Error("invalid delegate address")
	ErrInvalidSignature           = errors.Error("invalid signature")
	ErrInvalidIdentity            = errors.Error("invalid identity")
	ErrAccountProxiesRetrieval    = errors.Error("couldn't retrieve account proxies")
	ErrInvalidIdentityAddress     = errors.Error("invalid identity address")
	ErrInvalidProxyType           = errors.Error("invalid proxy type")
	ErrInvalidDelegate            = errors.Error("invalid delegate")
	ErrDelegatorAccountIDCreation = errors.Error("couldn't create delegator account ID")
	ErrNodeAdminRetrieval         = errors.Error("couldn't retrieve node admin")
	ErrNotAdminAccount            = errors.Error("provided account is not an admin")
)

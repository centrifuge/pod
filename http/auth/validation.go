package auth

import (
	"strconv"
	"strings"
	"time"

	proxyTypes "github.com/centrifuge/chain-custom-types/pkg/proxy"
)

const (
	expectedAlgorithm   = "sr25519"
	expectedAddressType = "ss58"
	expectedTokenType   = "jw3t"
)

var (
	headerValidationFn = func(header *JW3THeader) error {
		if strings.ToLower(header.Algorithm) != expectedAlgorithm {
			return ErrInvalidJW3TAlgorithm
		}

		if strings.ToLower(header.AddressType) != expectedAddressType {
			return ErrInvalidJW3TAddressType
		}

		if strings.ToLower(header.TokenType) != expectedTokenType {
			return ErrInvalidJW3TTokenType
		}

		return nil
	}
)

var (
	payloadValidationFn = func(payload *JW3TPayload) error {
		i, err := strconv.ParseInt(payload.NotBefore, 10, 64)
		if err != nil {
			return ErrInvalidNotBeforeTimestamp
		}

		tm := time.Unix(i, 0).UTC()

		if tm.After(time.Now().UTC()) {
			return ErrInactiveToken
		}

		i, err = strconv.ParseInt(payload.ExpiresAt, 10, 64)
		if err != nil {
			return ErrInvalidExpiresAtTimestamp
		}

		tm = time.Unix(i, 0).UTC()
		if tm.Before(time.Now().UTC()) {
			return ErrExpiredToken
		}

		if _, ok := allowedProxyTypes[payload.ProxyType]; !ok {
			return ErrInvalidProxyType
		}

		if payload.ProxyType == PodAdminProxyType {
			return nil
		}

		if _, ok := proxyTypes.ProxyTypeValue[payload.ProxyType]; !ok {
			return ErrInvalidProxyType
		}

		return nil
	}
)

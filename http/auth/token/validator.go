package token

import (
	"strconv"
	"strings"
	"time"
)

type Validator interface {
	Validate(token *JW3Token) error
}

const (
	expectedTokenType = "jw3t"
)

var (
	BasicHeaderValidationFn = func(header *JW3THeader) error {
		if strings.ToLower(header.TokenType) != expectedTokenType {
			return ErrInvalidJW3TTokenType
		}

		return nil
	}

	BasicPayloadValidationFn = func(payload *JW3TPayload) error {
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

		return nil
	}
)

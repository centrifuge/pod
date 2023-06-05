//go:build unit

package token

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_BasicHeaderValidationFn(t *testing.T) {
	header := &JW3THeader{TokenType: jw3TokenType}

	err := BasicHeaderValidationFn(header)
	assert.NoError(t, err)

	header = &JW3THeader{TokenType: "test"}

	err = BasicHeaderValidationFn(header)
	assert.ErrorIs(t, err, ErrInvalidJW3TTokenType)
}

func Test_BasicPayloadValidationFn(t *testing.T) {
	now := time.Now()
	expireTime := now.Add(24 * time.Hour)

	payload := &JW3TPayload{
		NotBefore: fmt.Sprintf("%d", now.Unix()),
		ExpiresAt: fmt.Sprintf("%d", expireTime.Unix()),
	}

	err := BasicPayloadValidationFn(payload)
	assert.NoError(t, err)

	payload.NotBefore = "invalid_not_before"

	err = BasicPayloadValidationFn(payload)
	assert.ErrorIs(t, err, ErrInvalidNotBeforeTimestamp)

	payload.NotBefore = fmt.Sprintf("%d", now.Add(1*time.Hour).Unix())

	err = BasicPayloadValidationFn(payload)
	assert.ErrorIs(t, err, ErrInactiveToken)

	payload.NotBefore = fmt.Sprintf("%d", now.Unix())
	payload.ExpiresAt = "invalid_expires_at"

	err = BasicPayloadValidationFn(payload)
	assert.ErrorIs(t, err, ErrInvalidExpiresAtTimestamp)

	payload.ExpiresAt = fmt.Sprintf("%d", now.Add(-1*time.Hour).Unix())

	err = BasicPayloadValidationFn(payload)
	assert.ErrorIs(t, err, ErrExpiredToken)
}

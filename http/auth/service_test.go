//go:build unit

package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vedhavyas/go-subkey/v2/sr25519"
)

func formSignaturePayload(t *testing.T, header JW3THeader, payload JW3TPayload) string {
	headerJson, err := json.Marshal(header)
	assert.NoError(t, err)
	payloadJson, err := json.Marshal(payload)
	assert.NoError(t, err)
	return fmt.Sprintf("%s.%s", base64.RawURLEncoding.EncodeToString(headerJson), base64.RawURLEncoding.EncodeToString(payloadJson))
}

func TestValidate(t *testing.T) {
	h := JW3THeader{
		Algorithm:   "sr25519",
		AddressType: "ss58",
		TokenType:   "JW3T",
	}

	kp, err := sr25519.Scheme{}.Generate()
	assert.NoError(t, err)

	issued := time.Now().UTC()
	p := JW3TPayload{
		IssuedAt:   fmt.Sprintf("%d", issued.Unix()),
		NotBefore:  fmt.Sprintf("%d", issued.Unix()),
		ExpiresAt:  fmt.Sprintf("%d", issued.Add(time.Hour).Unix()),
		Address:    kp.SS58Address(36),
		OnBehalfOf: "kANXoeY7KYbrzhyoDFypEWpijPRgKRx5G34ZX7TKbDBJwrVjp",
		ProxyType:  "Any",
	}

	sigPayload := formSignaturePayload(t, h, p)
	s, err := kp.Sign([]byte(sigPayload))
	assert.NoError(t, err)

	jw3tString := fmt.Sprintf("%s.%s", sigPayload, base64.RawURLEncoding.EncodeToString(s))
	service := &service{}
	accHeader, err := service.Validate(context.Background(), jw3tString)
	assert.NoError(t, err)
	assert.NotNil(t, accHeader)
}

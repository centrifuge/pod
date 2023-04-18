package token

import (
	"encoding/base64"
	"encoding/json"
	"strings"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/vedhavyas/go-subkey/v2"
)

const (
	tokenSeparator = "."
)

func DecodeJW3Token(token string) (*JW3Token, error) {
	fragments := strings.Split(token, tokenSeparator)
	if len(fragments) != 3 {
		return nil, ErrInvalidJW3Token
	}

	base64Header := fragments[0]

	jsonHeader, err := base64.RawURLEncoding.DecodeString(base64Header)
	if err != nil {
		return nil, ErrBase64HeaderDecoding
	}

	var jw3tHeader JW3THeader

	if err := json.Unmarshal(jsonHeader, &jw3tHeader); err != nil {
		return nil, ErrJSONHeaderDecoding
	}

	base64Payload := fragments[1]

	jsonPayload, err := base64.RawURLEncoding.DecodeString(base64Payload)
	if err != nil {
		return nil, ErrBase64PayloadDecoding
	}

	var jw3tPayload JW3TPayload

	if err := json.Unmarshal(jsonPayload, &jw3tPayload); err != nil {
		return nil, ErrJSONPayloadDecoding
	}

	signature, err := base64.RawURLEncoding.DecodeString(fragments[2])

	if err != nil {
		return nil, ErrBase64SignatureDecoding
	}

	return &JW3Token{
		Header:        &jw3tHeader,
		Base64Header:  base64Header,
		JSONHeader:    jsonHeader,
		Payload:       &jw3tPayload,
		Base64Payload: base64Payload,
		JSONPayload:   jsonPayload,
		Signature:     signature,
	}, nil
}

func DecodeSS58Address(address string) (*types.AccountID, error) {
	_, publicKey, err := subkey.SS58Decode(address)
	if err != nil {
		return nil, err
	}

	return types.NewAccountID(publicKey)
}

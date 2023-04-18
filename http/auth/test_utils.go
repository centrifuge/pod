//go:build unit || integration || testworld

package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/pod/http/auth/token"
	"github.com/vedhavyas/go-subkey/v2"
	"github.com/vedhavyas/go-subkey/v2/sr25519"
)

const (
	CentrifugeNetworkID = 36
)

type MutateOpt func(header *token.JW3THeader, payload *token.JW3TPayload)

func CreateJW3Token(
	delegateAccountID *types.AccountID,
	delegatorAccountID *types.AccountID,
	delegateURI string,
	proxyType string,
	mutateOpts ...MutateOpt,
) (string, error) {
	header := &token.JW3THeader{
		Algorithm:   "sr25519",
		AddressType: "ss58",
		TokenType:   "JW3T",
	}

	now := time.Now()
	exipreTime := now.Add(24 * time.Hour)

	delegateAddress := subkey.SS58Encode(delegateAccountID.ToBytes(), CentrifugeNetworkID)
	delegatorAddress := subkey.SS58Encode(delegatorAccountID.ToBytes(), CentrifugeNetworkID)

	payload := &token.JW3TPayload{
		IssuedAt:   fmt.Sprintf("%d", now.Unix()),
		NotBefore:  fmt.Sprintf("%d", now.Unix()),
		ExpiresAt:  fmt.Sprintf("%d", exipreTime.Unix()),
		Address:    delegateAddress,
		OnBehalfOf: delegatorAddress,
		ProxyType:  proxyType,
	}

	for _, opt := range mutateOpts {
		opt(header, payload)
	}

	jsonHeader, err := json.Marshal(header)

	if err != nil {
		return "", err
	}

	encodedJSONHeader := base64.RawURLEncoding.EncodeToString(jsonHeader)

	jsonPayload, err := json.Marshal(payload)

	if err != nil {
		return "", err
	}

	encodedJSONPayload := base64.RawURLEncoding.EncodeToString(jsonPayload)

	signatureMessage := strings.Join(
		[]string{
			string(jsonHeader),
			string(jsonPayload),
		},
		".",
	)

	kp, err := subkey.DeriveKeyPair(sr25519.Scheme{}, delegateURI)

	if err != nil {
		return "", err
	}

	// Note that we are not wrapping the message because both formats should work.
	sig, err := kp.Sign([]byte(signatureMessage))

	if err != nil {
		return "", err
	}

	encodedSignature := base64.RawURLEncoding.EncodeToString(sig)

	elems := []string{
		encodedJSONHeader,
		encodedJSONPayload,
		encodedSignature,
	}

	return strings.Join(elems, "."), nil
}

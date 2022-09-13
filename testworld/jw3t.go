package testworld

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/vedhavyas/go-subkey/v2"
	"github.com/vedhavyas/go-subkey/v2/sr25519"

	"github.com/centrifuge/go-centrifuge/http/auth"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

const (
	centrifugeNetworkID = 36
)

func CreateJW3Token(
	delegateAccountID *types.AccountID,
	delegatorAccountID *types.AccountID,
	delegateURI string,
	proxyType string,
) (string, error) {
	header := &auth.JW3THeader{
		Algorithm:   "sr25519",
		AddressType: "ss58",
		TokenType:   "JW3T",
	}

	now := time.Now()
	exipreTime := now.Add(24 * time.Hour)

	delegateAddress := subkey.SS58Encode(delegateAccountID.ToBytes(), centrifugeNetworkID)
	delegatorAddress := subkey.SS58Encode(delegatorAccountID.ToBytes(), centrifugeNetworkID)

	payload := auth.JW3TPayload{
		IssuedAt:   fmt.Sprintf("%d", now.Unix()),
		NotBefore:  fmt.Sprintf("%d", now.Unix()),
		ExpiresAt:  fmt.Sprintf("%d", exipreTime.Unix()),
		Address:    delegateAddress,
		OnBehalfOf: delegatorAddress,
		ProxyType:  proxyType,
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

	sig, err := kp.Sign(wrapSignatureMessage(signatureMessage))

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

func wrapSignatureMessage(msg string) []byte {
	return []byte(auth.BytesPrefix + msg + auth.BytesSuffix)
}

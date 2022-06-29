package crypto

import (
	"strings"

	"github.com/centrifuge/go-centrifuge/errors"
	"golang.org/x/crypto/ed25519"
)

// SignMessage signs the message using the private key as the curveType provided.
func SignMessage(privateKey, message []byte, curveType string) ([]byte, error) {
	curveType = strings.ToLower(curveType)
	switch curveType {
	case CurveEd25519:
		return ed25519.Sign(privateKey, message), nil
	default:
		return nil, errors.New("curve %s not supported", curveType)
	}
}

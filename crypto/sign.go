package crypto

import (
	"github.com/centrifuge/pod/errors"
	"golang.org/x/crypto/ed25519"
)

// SignMessage signs the message using the private key as the curveType provided.
func SignMessage(privateKey, message []byte, curveType CurveType) ([]byte, error) {
	switch curveType {
	case CurveEd25519:
		return ed25519.Sign(privateKey, message), nil
	default:
		return nil, errors.New("curve %s not supported", curveType)
	}
}

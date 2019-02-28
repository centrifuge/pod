package crypto

import (
	"strings"

	"github.com/centrifuge/go-centrifuge/crypto/secp256k1"
	"github.com/centrifuge/go-centrifuge/errors"
	"golang.org/x/crypto/ed25519"
)

// SignMessage signs the message using the private key as the curveType provided.
// if Secp256K1 curve provided, then the signature format is specific to ethereum.
func SignMessage(privateKey, message []byte, curveType string) ([]byte, error) {
	curveType = strings.ToLower(curveType)
	switch curveType {
	case CurveSecp256K1:
		return secp256k1.SignEthereum(message, privateKey)
	case CurveEd25519:
		return ed25519.Sign(privateKey, message), nil
	default:
		return nil, errors.New("curve %s not supported", curveType)
	}
}

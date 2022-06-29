package crypto

import (
	"github.com/centrifuge/go-centrifuge/crypto/ed25519"
)

// VerifyMessage verifies message using the public key as per the curve type.
func VerifyMessage(publicKey, message []byte, signature []byte, curveType string) bool {
	switch curveType {
	case CurveEd25519:
		return ed25519.VerifySignature(publicKey, message, signature)
	default:
		return false
	}
}

package crypto

import (
	"fmt"

	"github.com/centrifuge/pod/crypto/ed25519"
	"github.com/vedhavyas/go-subkey/v2/sr25519"
)

// VerifyMessage verifies message using the public key as per the curve type.
func VerifyMessage(publicKey, message []byte, signature []byte, curveType CurveType) bool {
	switch curveType {
	case CurveEd25519:
		return ed25519.VerifySignature(publicKey, message, signature)
	case CurveSr25519:
		pub, err := sr25519.Scheme{}.FromPublicKey(publicKey)
		if err != nil {
			fmt.Println(err.Error())
			return false
		}
		return pub.Verify(message, signature)
	default:
		return false
	}
}

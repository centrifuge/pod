package keytools

import (
	"github.com/centrifuge/go-centrifuge/keytools/secp256k1"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// VerifyMessage verifies message using the public key as per the curve type.
// if ethereumVerify is true, ethereum specific verification is done
func VerifyMessage(publicKey, message []byte, signature []byte, curveType string, ethereumVerify bool) bool {
	signatureBytes := make([]byte, len(signature))
	copy(signatureBytes, signature)

	switch curveType {
	case CurveSecp256K1:
		msg := make([]byte, MaxMsgLen)
		copy(msg, message)
		if ethereumVerify {
			address := secp256k1.GetAddress(publicKey)
			return secp256k1.VerifySignatureWithAddress(address, hexutil.Encode(signatureBytes), msg)
		}

		return secp256k1.VerifySignature(publicKey, msg, signatureBytes)
	case CurveEd25519:
		return false
	default:
		return false
	}
}

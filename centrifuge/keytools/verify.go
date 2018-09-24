package keytools

import (
	"github.com/centrifuge/go-centrifuge/centrifuge/keytools/secp256k1"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

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

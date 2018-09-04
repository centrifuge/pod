package keytools

import (
	"fmt"

	"github.com/CentrifugeInc/go-centrifuge/centrifuge/utils"

	"github.com/CentrifugeInc/go-centrifuge/centrifuge/keytools/secp256k1"
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
			return secp256k1.VerifySignatureWithAddress(address, utils.ByteArrayToHex(signatureBytes), msg)
		}
		return secp256k1.VerifySignature(publicKey, msg, signatureBytes)

	case CurveEd25519:
		fmt.Errorf("curve ed25519 not supported yet")
		return false

	default:
		fmt.Errorf("curve %s not supported yet", curveType)
		return false
	}

}

package keytools

import (
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/keytools/secp256k1"
	"fmt"
)


func VerifyMessage(publicKeyPath string, message string, signature []byte, curveType string) bool {

	publicKey, err := readKeyFromPemFile(publicKeyPath, PUBLIC_KEY)

	if err != nil {
		log.Fatal(err)
	}

	msg := make([]byte, MAX_MSG_LEN)
	copy(msg, message)

	signatureBytes := make([]byte, len(signature))
	copy(signatureBytes, signature)

	switch curveType {

	case CURVE_SECP256K1:
		return secp256k1.VerifySignature(publicKey, msg, signatureBytes)

	case CURVE_ED25519:
		fmt.Println("curve ed25519 not support yet")
		return false

	default:
		return secp256k1.VerifySignature(publicKey, msg, signatureBytes)
	}

}

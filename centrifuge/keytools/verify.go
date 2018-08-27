package keytools

import (
	"fmt"

	"github.com/CentrifugeInc/go-centrifuge/centrifuge/keytools/io"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/keytools/secp256k1"
)

func VerifyMessage(publicKeyPath string, message string, signature []byte, curveType string) bool {

	publicKey, err := io.ReadKeyFromPemFile(publicKeyPath, PublicKey)

	if err != nil {
		log.Fatal(err)
	}

	msg := make([]byte, MaxMsgLen)
	copy(msg, message)

	signatureBytes := make([]byte, len(signature))
	copy(signatureBytes, signature)

	switch curveType {

	case CurveSecp256K1:
		return secp256k1.VerifySignature(publicKey, msg, signatureBytes)

	case CurveEd25519:
		fmt.Println("curve ed25519 not support yet")
		return false

	default:
		return secp256k1.VerifySignature(publicKey, msg, signatureBytes)
	}

}

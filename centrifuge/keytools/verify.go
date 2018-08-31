package keytools

import (
	"fmt"

	"github.com/CentrifugeInc/go-centrifuge/centrifuge/utils"

	"github.com/CentrifugeInc/go-centrifuge/centrifuge/keytools/io"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/keytools/secp256k1"
)

func VerifyMessage(publicKeyPath string, message string, signature []byte, curveType string, ethereumVerify bool) bool {

	publicKey, err := io.ReadKeyFromPemFile(publicKeyPath, PublicKey)

	if err != nil {
		log.Fatal(err)
	}

	var msg []byte
	if ethereumVerify == true {
		msg = []byte(message)
	} else {
		if len(message) > MaxMsgLen {
			log.Fatal("max message len is 32 bytes current len:", len(message))
		}

		msg = make([]byte, MaxMsgLen)
		copy(msg, message)
	}

	signatureBytes := make([]byte, len(signature))
	copy(signatureBytes, signature)

	switch curveType {

	case CurveSecp256K1:
		if ethereumVerify {
			address := secp256k1.GetAddress(publicKey)
			return secp256k1.VerifySignatureWithAddress(address, utils.ByteArrayToHex(signatureBytes), msg)
		}
		return secp256k1.VerifySignature(publicKey, msg, signatureBytes)

	case CurveEd25519:
		fmt.Println("curve ed25519 not support yet")
		return false

	default:
		if ethereumVerify {
			address := secp256k1.GetAddress(publicKey)
			return secp256k1.VerifySignatureWithAddress(address, utils.ByteArrayToHex(signatureBytes), msg)
		}
		return secp256k1.VerifySignature(publicKey, msg, signatureBytes)
	}

}

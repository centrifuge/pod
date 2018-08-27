package keytools

import ("strings"
		"github.com/CentrifugeInc/go-centrifuge/centrifuge/keytools/secp256k1"
	"fmt"
)

func SignMessage(privateKeyPath, message, curveType string) []byte {

	privateKey, err := readKeyFromPemFile(privateKeyPath, PRIVATE_KEY)

	if err != nil {
		log.Fatal(err)
	}

	curveType = strings.ToLower(curveType)

	if len(message) > MAX_MSG_LEN {
		log.Fatal("max message len is 32 bytes current len:", len(message))
	}

	msg := make([]byte, MAX_MSG_LEN)
	copy(msg, message)

	switch curveType {

	case CURVE_SECP256K1:
		return secp256k1.Sign(msg, privateKey)

	case CURVE_ED25519:
		fmt.Println("curve ed25519 not support yet")
		return []byte("")

	default:
		return secp256k1.Sign(msg, privateKey)

	}

}

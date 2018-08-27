package keytools

import (
	"strings"

	"github.com/CentrifugeInc/go-centrifuge/centrifuge/keytools/secp256k1"

	"github.com/CentrifugeInc/go-centrifuge/centrifuge/keytools/ed25519"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/keytools/io"
)

func GenerateSigningKeyPair(publicFileName, privateFileName, curveType string) {

	curveType = strings.ToLower(curveType)

	var publicKey, privateKey []byte

	switch curveType {

	case CURVE_SECP256K1:
		publicKey, privateKey = secp256k1.GenerateSigningKeyPair()

	case CURVE_ED25519:
		publicKey, privateKey = ed25519.GenerateSigningKeyPair()

	default:
		publicKey, privateKey = ed25519.GenerateSigningKeyPair()

	}

	io.WriteKeyToPemFile(privateFileName,PRIVATE_KEY, privateKey)
	io.WriteKeyToPemFile(publicFileName, PUBLIC_KEY, publicKey)

}

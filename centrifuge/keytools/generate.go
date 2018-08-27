package keytools

import (
	"strings"

	"github.com/CentrifugeInc/go-centrifuge/centrifuge/keytools/secp256k1"

)

func GenerateSigningKeyPair(publicFileName, privateFileName, curveType string) {

	curveType = strings.ToLower(curveType)

	var publicKey, privateKey []byte

	switch curveType {

	case CURVE_SECP256K1:
		publicKey, privateKey = secp256k1.GenerateSigningKeyPair()

	case CURVE_ED25519:
		publicKey, privateKey = GenerateSigningKeyPairED25519()

	default:
		publicKey, privateKey = GenerateSigningKeyPairED25519()

	}

	writeKeyToPemFile(privateFileName, "PRIVATE KEY", privateKey)
	writeKeyToPemFile(publicFileName, "PUBLIC KEY", publicKey)

}

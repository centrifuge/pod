package keytools

import (
	"strings"

	"github.com/CentrifugeInc/go-centrifuge/centrifuge/keytools/ed25519"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/keytools/secp256k1"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/utils"
)

func GenerateSigningKeyPair(publicFileName, privateFileName, curveType string) {
	var publicKey, privateKey []byte
	switch strings.ToLower(curveType) {
	case CurveSecp256K1:
		publicKey, privateKey = secp256k1.GenerateSigningKeyPair()
	case CurveEd25519:
		publicKey, privateKey = ed25519.GenerateSigningKeyPair()
	default:
		publicKey, privateKey = ed25519.GenerateSigningKeyPair()
	}

	utils.WriteKeyToPemFile(privateFileName, PrivateKey, privateKey)
	utils.WriteKeyToPemFile(publicFileName, PublicKey, publicKey)
}

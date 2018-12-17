package keytools

import (
	"strings"

	"github.com/centrifuge/go-centrifuge/keytools/ed25519"
	"github.com/centrifuge/go-centrifuge/keytools/secp256k1"
	"github.com/centrifuge/go-centrifuge/utils"
)

// GenerateSigningKeyPair generates based on the curveType and writes keys to file paths given.
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

	utils.WriteKeyToPemFile(privateFileName, utils.PrivateKey, privateKey)
	utils.WriteKeyToPemFile(publicFileName, utils.PublicKey, publicKey)
}

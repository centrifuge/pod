package crypto

import (
	"strings"

	"github.com/centrifuge/go-centrifuge/crypto/ed25519"
	"github.com/centrifuge/go-centrifuge/crypto/secp256k1"
	"github.com/centrifuge/go-centrifuge/utils"
)

// GenerateCryptoKeyPair generates based on the curveType and writes keys to file paths given.
func GenerateCryptoKeyPair(publicFileName, privateFileName, curveType string) (err error) {
	var publicKey, privateKey []byte
	switch strings.ToLower(curveType) {
	case CurveSecp256K1:
		publicKey, privateKey, err = secp256k1.GenerateSigningKeyPair()
	case CurveEd25519:
		publicKey, privateKey, err = ed25519.GenerateSigningKeyPair()
	default:
		publicKey, privateKey, err = ed25519.GenerateSigningKeyPair()
	}
	if err != nil {
		return err
	}

	err = utils.WriteKeyToPemFile(privateFileName, utils.PrivateKey, privateKey)
	if err != nil {
		return err
	}

	err = utils.WriteKeyToPemFile(publicFileName, utils.PublicKey, publicKey)
	if err != nil {
		return err
	}
	return nil
}

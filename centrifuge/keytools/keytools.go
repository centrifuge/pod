package keytools

import (
	"fmt"
	"strings"

	"github.com/CentrifugeInc/go-centrifuge/centrifuge/errors"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/keytools/ed25519"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/keytools/secp256k1"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/utils"
	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("keytools")

const (
	PublicKey  = "PUBLIC KEY"
	PrivateKey = "PRIVATE KEY"
)

const (
	CurveEd25519   string = "ed25519"
	CurveSecp256K1 string = "secp256k1"
)

const MaxMsgLen = 32

// GenerateSigningKeyPair generates key pair using the given curve and writes them to the disk
func GenerateSigningKeyPair(curveType string) (publicKey, privateKey []byte) {
	switch strings.ToLower(curveType) {
	case CurveSecp256K1:
		publicKey, privateKey = secp256k1.GenerateSigningKeyPair()
	case CurveEd25519:
		publicKey, privateKey = ed25519.GenerateSigningKeyPair()
	default:
		publicKey, privateKey = ed25519.GenerateSigningKeyPair()
	}

	return publicKey, privateKey
}

// SaveKeyPair saves the public and private keys to the underlying storage
func SaveKeyPair(publicFileName, privateFileName string, publicKey, privateKey []byte) error {
	err := utils.WriteKeyToPemFile(publicFileName, PublicKey, publicKey)
	if err != nil {
		return errors.Wrap(err, "failed to save public key")
	}

	err = utils.WriteKeyToPemFile(privateFileName, PrivateKey, privateKey)
	if err != nil {
		return errors.Wrap(err, "failed to save private key")
	}

	return nil

}

// SignMessage signs the message using private key as per the curve provided
// ethereumSign to signal if the signature should be generated with specific message format
func SignMessage(privateKey, message []byte, curveType string, ethereumSign bool) ([]byte, error) {
	curveType = strings.ToLower(curveType)

	switch curveType {
	case CurveSecp256K1:
		msg := make([]byte, MaxMsgLen)
		copy(msg, message)
		if ethereumSign {
			return secp256k1.SignEthereum(msg, privateKey)
		}
		return secp256k1.Sign(msg, privateKey)
	default:
		return nil, fmt.Errorf("curve %s not supported", curveType)
	}

}

// VerifyMessage verifies the signature using the public key and message based on the curve type
// ethereumVerify to signal to extract the key from the specified format
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
	default:
		log.Warning("curve %s not supported yet", curveType)
		return false
	}

}

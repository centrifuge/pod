package crypto

import (
	"strings"

	"github.com/centrifuge/go-centrifuge/crypto/secp256k1"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"golang.org/x/crypto/ed25519"
)

// SignMessage signs the message using the private key as the curveType provided.
// if ethereumSign is true, then the signature format is specific to ethereum.
func SignMessage(privateKey, message []byte, curveType string, ethereumSign bool) ([]byte, error) {
	curveType = strings.ToLower(curveType)
	switch curveType {
	case CurveSecp256K1:
		if ethereumSign {
			return secp256k1.SignEthereum(message, privateKey)
		}

		return secp256k1.Sign(message, privateKey)
	case CurveEd25519:
		return ed25519.Sign(privateKey, message), nil
	default:
		return nil, errors.New("curve %s not supported", curveType)
	}

}

// VerifySignature verifies the signature using secp256k1
func VerifySignature(pubKey, message, signature []byte) error {
	valid := secp256k1.VerifySignatureWithAddress(common.BytesToAddress(pubKey).String(), hexutil.Encode(signature), message)
	if !valid {
		return errors.New("invalid signature")
	}

	return nil
}

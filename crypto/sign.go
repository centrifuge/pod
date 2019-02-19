package crypto

import (
	"strings"
	"time"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/crypto/secp256k1"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/utils"
	"golang.org/x/crypto/ed25519"
)

// SignMessage signs the message using the private key as the curveType provided.
// if ethereumSign is true, then the signature format is specific to ethereum.
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
	case CurveEd25519:
		return ed25519.Sign(privateKey, message), nil
	default:
		return nil, errors.New("curve %s not supported", curveType)
	}

}

// VerifySignature verifies the signature using ed25519
func VerifySignature(pubKey, message, signature []byte) error {
	valid := ed25519.Verify(pubKey, message, signature)
	if !valid {
		return errors.New("invalid signature")
	}

	return nil
}

// Sign the document with the private key and return the signature along with the public key for the verification
// assumes that signing root for the document is generated
// Deprecated
func Sign(didBytes []byte, privateKey []byte, pubKey []byte, payload []byte) *coredocumentpb.Signature {
	return &coredocumentpb.Signature{
		EntityId:  didBytes,
		PublicKey: pubKey,
		Signature: ed25519.Sign(privateKey, payload),
		Timestamp: utils.ToTimestamp(time.Now().UTC()),
	}
}

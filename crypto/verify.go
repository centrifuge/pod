package crypto

import (
	"github.com/centrifuge/go-centrifuge/crypto/ed25519"
	"github.com/centrifuge/go-centrifuge/crypto/secp256k1"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// VerifyMessage verifies message using the public key as per the curve type.
// for Secp256K1 curve the verification is done with ethereum prefix
// for Secp256K1, public key should be the address of the original public key, following ethereum standards
func VerifyMessage(publicKey, message []byte, signature []byte, curveType string) bool {
	switch curveType {
	case CurveSecp256K1:
		return secp256k1.VerifySignatureWithAddress(common.BytesToAddress(publicKey).String(), hexutil.Encode(signature), message)
	case CurveEd25519:
		return ed25519.VerifySignature(publicKey, message, signature)
	default:
		return false
	}
}

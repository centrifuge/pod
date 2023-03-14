package crypto

import (
	"github.com/centrifuge/pod/crypto/ed25519"
	"github.com/centrifuge/pod/errors"
	"github.com/libp2p/go-libp2p-core/crypto"
)

type CurveType string

// CurveEd25519 Constants shared within subfolders
const CurveEd25519 CurveType = "ed25519"

// CurveSr25519 Constants shared within subfolders
const CurveSr25519 CurveType = "sr25519"

// ObtainP2PKeypair obtains a key pair from given file paths
func ObtainP2PKeypair(pubKeyFile, privKeyFile string) (priv crypto.PrivKey, pub crypto.PubKey, err error) {
	// Create the signing key for the host
	publicKey, privateKey, err := ed25519.GetSigningKeyPair(pubKeyFile, privKeyFile)
	if err != nil {
		return nil, nil, errors.New("failed to get keys: %v", err)
	}

	var key []byte
	key = append(key, privateKey...)
	key = append(key, publicKey...)

	priv, err = crypto.UnmarshalEd25519PrivateKey(key)
	if err != nil {
		return nil, nil, err
	}

	pub = priv.GetPublic()
	return priv, pub, nil
}

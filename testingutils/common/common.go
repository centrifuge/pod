//go:build unit || integration || testworld

package testingcommons

import (
	"fmt"
	"math/rand"
	"os"
	"path"

	"github.com/centrifuge/go-centrifuge/crypto/ed25519"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	libp2pcrypto "github.com/libp2p/go-libp2p-core/crypto"
)

func GetRandomAccountID() (*types.AccountID, error) {
	b := make([]byte, 32)

	if _, err := rand.Read(b); err != nil {
		return nil, err
	}

	return types.NewAccountID(b)
}

const (
	storageDir = "go-centrifuge-test"
)

// GetRandomTestStoragePath generates a random path for DB storage
func GetRandomTestStoragePath(pattern string) (string, error) {
	tempDirPath := path.Join(os.TempDir(), storageDir)

	if err := os.MkdirAll(tempDirPath, os.ModePerm); err != nil {
		return "", fmt.Errorf("couldn't create temp dir: %w", err)
	}

	return os.MkdirTemp(tempDirPath, pattern)
}

const (
	TestPublicSigningKeyPath  = "testingutils/common/keys/testSigningKey.pub.pem"
	TestPrivateSigningKeyPath = "testingutils/common/keys/testSigningKey.key.pem"
)

func GetTestSigningKeys() (libp2pcrypto.PubKey, libp2pcrypto.PrivKey, error) {
	signingPublicKey, signingPrivateKey, err := ed25519.GetSigningKeyPair(TestPublicSigningKeyPath, TestPrivateSigningKeyPath)

	if err != nil {
		return nil, nil, fmt.Errorf("couldn't get test signing keys: %w", err)
	}

	privateKey, err := libp2pcrypto.UnmarshalEd25519PrivateKey(signingPrivateKey)

	if err != nil {
		return nil, nil, err
	}

	publicKey, err := libp2pcrypto.UnmarshalEd25519PublicKey(signingPublicKey)

	if err != nil {
		return nil, nil, err
	}

	return publicKey, privateKey, nil
}

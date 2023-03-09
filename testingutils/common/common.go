//go:build unit || integration || testworld

package testingcommons

import (
	"fmt"
	"math/rand"
	"os"
	"path"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/pod/crypto/ed25519"
	pathUtils "github.com/centrifuge/pod/testingutils/path"
	"github.com/centrifuge/pod/utils"
	libp2pcrypto "github.com/libp2p/go-libp2p-core/crypto"
)

func CopyServiceContext(in map[string]any) map[string]any {
	out := make(map[string]any)

	for serviceName, service := range in {
		out[serviceName] = service
	}

	return out
}

func GetRandomAccountID() (*types.AccountID, error) {
	b := make([]byte, 32)

	if _, err := rand.Read(b); err != nil {
		return nil, err
	}

	return types.NewAccountID(b)
}

const (
	storageDir = "pod-test"
)

// GetRandomTestStoragePath generates a random path for DB storage
func GetRandomTestStoragePath(pattern string) (string, error) {
	tempDirPath := path.Join(os.TempDir(), storageDir)

	if err := os.MkdirAll(tempDirPath, os.ModePerm); err != nil {
		return "", fmt.Errorf("couldn't create temp dir: %w", err)
	}

	return os.MkdirTemp(tempDirPath, pattern)
}

var (
	TestPublicSigningKeyPath  = pathUtils.AppendPathToProjectRoot("testingutils/common/keys/testSigningKey.pub.pem")
	TestPrivateSigningKeyPath = pathUtils.AppendPathToProjectRoot("testingutils/common/keys/testSigningKey.key.pem")
	TestPublicP2PKeyPath      = pathUtils.AppendPathToProjectRoot("testingutils/common/keys/testP2PKey.pub.pem")
	TestPrivateP2PKeyPath     = pathUtils.AppendPathToProjectRoot("testingutils/common/keys/testP2PKey.key.pem")
)

func GetTestSigningKeys() (libp2pcrypto.PubKey, libp2pcrypto.PrivKey, error) {
	return GetTestKeys(TestPublicSigningKeyPath, TestPrivateSigningKeyPath)
}

func GetTestP2PKeys() (libp2pcrypto.PubKey, libp2pcrypto.PrivKey, error) {
	return GetTestKeys(TestPublicP2PKeyPath, TestPrivateP2PKeyPath)
}

func GetTestKeys(publicKeyPath, privateKeyPath string) (libp2pcrypto.PubKey, libp2pcrypto.PrivKey, error) {
	ed25519PublicKey, ed25519PrivateKey, err := ed25519.GetSigningKeyPair(publicKeyPath, privateKeyPath)

	if err != nil {
		return nil, nil, fmt.Errorf("couldn't get test keys: %w", err)
	}

	privateKey, err := libp2pcrypto.UnmarshalEd25519PrivateKey(ed25519PrivateKey)

	if err != nil {
		return nil, nil, fmt.Errorf("couldn't unmarshal private key: %w", err)
	}

	publicKey, err := libp2pcrypto.UnmarshalEd25519PublicKey(ed25519PublicKey)

	if err != nil {
		return nil, nil, fmt.Errorf("couldn't unmarshal public key: %w", err)
	}

	return publicKey, privateKey, nil
}

func MustGetFreePort() int {
	_, port, err := utils.GetFreeAddrPort()

	if err != nil {
		panic("couldn't get free port")
	}

	return port
}

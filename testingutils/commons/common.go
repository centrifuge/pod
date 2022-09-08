//go:build unit || integration || testworld

package testingcommons

import (
	"fmt"
	"math/rand"
	"os"
	"path"

	"github.com/centrifuge/chain-custom-types/pkg/proxy"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

func GetRandomAccountID() (*types.AccountID, error) {
	b := make([]byte, 32)

	if _, err := rand.Read(b); err != nil {
		return nil, err
	}

	return types.NewAccountID(b)
}

func GetRandomProxyType() proxy.CentrifugeProxyType {
	c := rand.Intn(len(proxy.ProxyTypeName))

	return proxy.CentrifugeProxyType(c)
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

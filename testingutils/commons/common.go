//go:build unit || integration || testworld

package testingcommons

import (
	"math/rand"
	"os"

	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

func GetRandomAccountID() (*types.AccountID, error) {
	b := make([]byte, 32)

	if _, err := rand.Read(b); err != nil {
		return nil, err
	}

	return types.NewAccountID(b)
}

func GetRandomProxyType() types.ProxyType {
	c := rand.Intn(len(types.ProxyTypeName))

	return types.ProxyType(c)
}

const (
	storageDirPattern = "go-centrifuge-test-*"
)

// GetRandomTestStoragePath generates a random path for DB storage
func GetRandomTestStoragePath() (string, error) {
	return os.MkdirTemp(os.TempDir(), storageDirPattern)
}

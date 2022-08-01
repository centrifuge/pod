//go:build unit || integration || testworld

package testingcommons

import (
	"math/rand"

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

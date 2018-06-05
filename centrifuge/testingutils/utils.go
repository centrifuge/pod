package testingutils

import (
	"crypto/rand"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/config"
)

func MockConfigOption(key string, value interface{}) func() {
	mockedValue := config.Config.V.Get(key)
	config.Config.V.Set(key, value)
	return func() {
		config.Config.V.Set(key, mockedValue)
	}
}

func Rand32Bytes() []byte {
	randbytes := make([]byte, 32)
	rand.Read(randbytes)
	return randbytes
}

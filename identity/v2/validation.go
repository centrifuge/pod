package v2

import (
	"github.com/centrifuge/go-centrifuge/errors"
)

const (
	ErrInvalidPublicKey = errors.Error("invalid public key")

	publicKeyExpectedLen = 32
)

var (
	publicKeyValidatorFn = func(pubKey []byte) error {
		if len(pubKey) != publicKeyExpectedLen {
			return ErrInvalidPublicKey
		}

		return nil
	}
)

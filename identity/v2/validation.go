package v2

import (
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

const (
	ErrInvalidAccountID = errors.Error("invalid account ID")
	ErrInvalidPublicKey = errors.Error("invalid public key")

	publicKeyExpectedLen = 32
)

var (
	accountIDValidatorFn = func(accountID *types.AccountID) error {
		if accountID == nil {
			return ErrInvalidAccountID
		}

		return nil
	}

	publicKeyValidatorFn = func(pubKey []byte) error {
		if len(pubKey) != publicKeyExpectedLen {
			return ErrInvalidPublicKey
		}

		return nil
	}
)

package keystore

import (
	keystoreType "github.com/centrifuge/chain-custom-types/pkg/keystore"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

const (
	ErrNoKeysProvided      = errors.Error("no keys provided")
	ErrNoKeyHashesProvided = errors.Error("no key hashes provided")
	ErrInvalidKey          = errors.Error("invalid key")
	ErrInvalidKeyID        = errors.Error("invalid key ID")
	ErrInvalidKeyIDHash    = errors.Error("invalid key ID hash")
)

var (
	emptyKeyHash types.Hash

	addKeysValidationFn = func(keys []*keystoreType.AddKey) error {
		if keys == nil {
			return ErrNoKeysProvided
		}

		for _, key := range keys {
			if key.Key == emptyKeyHash {
				return ErrInvalidKey
			}
		}

		return nil
	}

	keyHashesValidationFn = func(keyHashes []*types.Hash) error {
		if keyHashes == nil {
			return ErrNoKeyHashesProvided
		}

		for _, keyHash := range keyHashes {
			if keyHash == nil || *keyHash == emptyKeyHash {
				return ErrInvalidKey
			}
		}

		return nil
	}

	keyIDValidationFn = func(keyID *keystoreType.KeyID) error {
		if keyID == nil {
			return ErrInvalidKeyID
		}

		if keyID.Hash == emptyKeyHash {
			return ErrInvalidKeyIDHash
		}

		return nil
	}
)

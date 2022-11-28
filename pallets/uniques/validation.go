package uniques

import (
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

const (
	ErrInvalidCollectionID = errors.Error("invalid collection ID")
	ErrInvalidItemID       = errors.Error("invalid item ID")
	ErrMissingMetadata     = errors.Error("missing metadata")
	ErrMetadataTooBig      = errors.Error("metadata too big")
	ErrMissingKey          = errors.Error("missing key")
	ErrKeyTooBig           = errors.Error("key too big")
	ErrMissingValue        = errors.Error("missing value")
	ErrValueTooBig         = errors.Error("value too big")
)

var (
	ItemIDValidatorFn = func(itemID types.U128) error {
		if itemID.BitLen() == 0 {
			return ErrInvalidItemID
		}

		return nil
	}

	CollectionIDValidatorFn = func(collectionID types.U64) error {
		if collectionID == 0 {
			return ErrInvalidCollectionID
		}

		return nil
	}

	metadataValidatorFn = func(data []byte) error {
		if len(data) > MetadataLimit {
			return ErrMetadataTooBig
		}

		if len(data) == 0 {
			return ErrMissingMetadata
		}

		return nil
	}

	KeyValidatorFn = func(key []byte) error {
		if len(key) > KeyLimit {
			return ErrKeyTooBig
		}

		if len(key) == 0 {
			return ErrMissingKey
		}

		return nil
	}

	valueValidatorFn = func(value []byte) error {
		if len(value) > ValueLimit {
			return ErrValueTooBig
		}

		if len(value) == 0 {
			return ErrMissingValue
		}

		return nil
	}
)

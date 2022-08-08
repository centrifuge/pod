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
)

var (
	ItemIDValidatorFn = func(instanceID types.U128) error {
		if instanceID.BitLen() == 0 {
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
		if len(data) > StringLimit {
			return ErrMetadataTooBig
		}

		if len(data) == 0 {
			return ErrMissingMetadata
		}

		return nil
	}
)

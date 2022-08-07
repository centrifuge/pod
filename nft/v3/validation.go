package v3

import (
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"go.uber.org/multierr"
)

const (
	ErrRequestNil          = errors.Error("request nil")
	ErrMissingDocumentID   = errors.Error("missing document ID")
	ErrInvalidCollectionID = errors.Error("invalid collection ID")
	ErrInvalidItemID       = errors.Error("invalid item ID")
	ErrMissingMetadata     = errors.Error("missing metadata")
	ErrMetadataTooBig      = errors.Error("metadata too big")
)

type validator struct {
	errs []error
}

func newValidator() *validator {
	return &validator{}
}

func (v *validator) error() error {
	return multierr.Combine(v.errs...)
}

func (v *validator) addError(err error) *validator {
	v.errs = append(v.errs, err)

	return v
}

func (v *validator) validateBool(expr bool, err error) *validator {
	if expr {
		v.addError(err)
	}

	return v
}

var (
	mintNFTRequestValidatorFn = func(req *MintNFTRequest) error {
		if req == nil {
			return ErrRequestNil
		}

		if len(req.DocumentID) == 0 {
			return ErrMissingDocumentID
		}

		return collectionIDValidatorFn(req.CollectionID)
	}

	ownerOfValidatorFn = func(req *OwnerOfRequest) error {
		if req == nil {
			return ErrRequestNil
		}

		if err := collectionIDValidatorFn(req.CollectionID); err != nil {
			return err
		}

		return itemIDValidatorFn(req.ItemID)
	}

	createNFTClassRequestValidatorFn = func(req *CreateNFTClassRequest) error {
		if req == nil {
			return ErrRequestNil
		}

		return collectionIDValidatorFn(req.CollectionID)
	}

	instanceMetadataOfRequestValidatorFn = func(req *ItemMetadataOfRequest) error {
		if req == nil {
			return ErrRequestNil
		}

		if err := collectionIDValidatorFn(req.CollectionID); err != nil {
			return err
		}

		return itemIDValidatorFn(req.ItemID)
	}

	itemIDValidatorFn = func(instanceID types.U128) error {
		if instanceID.BitLen() == 0 {
			return ErrInvalidItemID
		}

		return nil
	}

	collectionIDValidatorFn = func(collectionID types.U64) error {
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

package v3

import (
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/nft/v3/uniques"
)

const (
	ErrMissingDocumentID = errors.Error("missing document ID")
)

var (
	mintNFTRequestValidatorFn = func(req *MintNFTRequest) error {
		if req == nil {
			return errors.ErrRequestNil
		}

		if len(req.DocumentID) == 0 {
			return ErrMissingDocumentID
		}

		return uniques.CollectionIDValidatorFn(req.CollectionID)
	}

	ownerOfValidatorFn = func(req *OwnerOfRequest) error {
		if req == nil {
			return errors.ErrRequestNil
		}

		if err := uniques.CollectionIDValidatorFn(req.CollectionID); err != nil {
			return err
		}

		return uniques.ItemIDValidatorFn(req.ItemID)
	}

	createNFTCollectionRequestValidatorFn = func(req *CreateNFTCollectionRequest) error {
		if req == nil {
			return errors.ErrRequestNil
		}

		return uniques.CollectionIDValidatorFn(req.CollectionID)
	}

	itemMetadataRequestValidatorFn = func(req *GetItemMetadataRequest) error {
		if req == nil {
			return errors.ErrRequestNil
		}

		if err := uniques.CollectionIDValidatorFn(req.CollectionID); err != nil {
			return err
		}

		return uniques.ItemIDValidatorFn(req.ItemID)
	}

	itemAttributeRequestValidatorFn = func(req *GetItemAttributeRequest) error {
		if req == nil {
			return errors.ErrRequestNil
		}

		if err := uniques.CollectionIDValidatorFn(req.CollectionID); err != nil {
			return err
		}

		if err := uniques.ItemIDValidatorFn(req.ItemID); err != nil {
			return err
		}

		return uniques.KeyValidatorFn([]byte(req.Key))
	}
)

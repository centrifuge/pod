package v3

import (
	"github.com/centrifuge/pod/errors"
	"github.com/centrifuge/pod/pallets/uniques"
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
)

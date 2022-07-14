package entity

import (
	"context"

	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	v2 "github.com/centrifuge/go-centrifuge/identity/v2"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
)

// fieldValidateFunc validates the fields of the entity model
func fieldValidator(identityService v2.Service) documents.Validator {
	return documents.ValidatorFunc(func(_, new documents.Document) error {
		if new == nil {
			return documents.ErrDocumentNil
		}

		entity, ok := new.(*Entity)
		if !ok {
			return documents.ErrDocumentInvalidType
		}

		if entity.Data.Identity == nil {
			return errors.New("entity identity is empty")
		}

		// TODO(cdamian): Get a proper context here
		ctx := context.Background()
		accID := types.NewAccountID((*entity.Data.Identity)[:])

		err := identityService.ValidateAccount(ctx, &accID)
		if err != nil {
			return documents.ErrIdentityInvalid
		}

		return nil
	})
}

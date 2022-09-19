package entity

import (
	"context"

	"github.com/centrifuge/go-centrifuge/documents"
	v2 "github.com/centrifuge/go-centrifuge/identity/v2"
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
			return ErrEntityDataNoIdentity
		}

		ctx := context.Background()

		err := identityService.ValidateAccount(ctx, entity.Data.Identity)
		if err != nil {
			return documents.ErrIdentityInvalid
		}

		return nil
	})
}

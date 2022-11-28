package entity

import (
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

		if err := identityService.ValidateAccount(entity.Data.Identity); err != nil {
			return documents.ErrIdentityInvalid
		}

		return nil
	})
}

package entity

import (
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
)

// fieldValidateFunc validates the fields of the entity model
func fieldValidator(factory identity.Factory) documents.Validator {
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

		valid, err := factory.IdentityExists(*entity.Data.Identity)
		if err != nil || !valid {
			return errors.New("identity not created from identity factory")
		}

		return nil
	})
}

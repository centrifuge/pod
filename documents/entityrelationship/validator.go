package entityrelationship

import "github.com/centrifuge/go-centrifuge/documents"

// fieldValidateFunc validates the fields of the entity model
func fieldValidator() documents.Validator {
	return documents.ValidatorFunc(func(_, new documents.Model) error {
		if new == nil {
			return documents.ErrDocumentNil
		}

		_, ok := new.(*EntityRelationship)
		if !ok {
			return documents.ErrDocumentInvalidType
		}

		return nil
	})
}

// CreateValidator returns a validator group that should be run before creating the entity and persisting it to DB
func CreateValidator() documents.ValidatorGroup {
	return documents.ValidatorGroup{
		fieldValidator(),
	}
}

// UpdateValidator returns a validator group that should be run before updating the entity
func UpdateValidator() documents.ValidatorGroup {
	return documents.ValidatorGroup{
		fieldValidator(),
		documents.UpdateVersionValidator(),
	}
}
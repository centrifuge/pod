package transferdetails

import (
	"github.com/centrifuge/go-centrifuge/documents"
)

// fieldValidateFunc validates the fields of the funding extension
func fieldValidator() documents.Validator {
	return documents.ValidatorFunc(func(_, new documents.Model) error {
		//todo implement validator
		return nil
	})
}

// CreateValidator returns a validator group that should be run before adding the funding extension
func CreateValidator() documents.ValidatorGroup {
	return documents.ValidatorGroup{
		fieldValidator(),
	}
}

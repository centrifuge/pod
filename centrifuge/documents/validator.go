package documents

import (
	"github.com/centrifuge/go-centrifuge/centrifuge/utils"
)

// Validator is an interface every Validator (atomic or group) should implement
type Validator interface {
	// Validate validates the updates to the model in newState.
	Validate(oldState Model, newState Model) []error
}

// ValidatorGroup implements Validator for validating a set of validators.
type ValidatorGroup []Validator

//Validate will execute all group specific atomic validations
func (group ValidatorGroup) Validate(oldState Model, newState Model) (validationErrors []error) {

	for _, v := range group {
		centErrors := v.Validate(oldState, newState)
		validationErrors = append(validationErrors, centErrors...)
	}
	return validationErrors
}

// IsCurrencyValid checks if the currency is of length 3
func IsCurrencyValid(cur string) bool {
	return utils.StringLengthEqual(cur, 3)
}

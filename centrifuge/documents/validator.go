package documents

// Validator is an interface every Validator (atomic or group) should implement
type Validator interface {
	// Validate validates the updates to the model in newState.
	Validate(oldState Model, newState Model) []error
}

// ValidatorGroup implements Validator for validating a set of atomic validators.
type ValidatorGroup struct {
	validators []Validator
}

//Validate will execute all group specific atomic validations
func (group ValidatorGroup) Validate(oldState Model, newState Model) (validationErrors []error) {

	for _, v := range group.validators {
		centErrors := v.Validate(oldState, newState)
		validationErrors = append(validationErrors, centErrors...)
	}
	return validationErrors
}

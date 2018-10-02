// +build unit

package documents

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/centrifuge/centerrors"
	"github.com/centrifuge/go-centrifuge/centrifuge/code"
	"github.com/stretchr/testify/assert"
)

type MockValidator struct {
}

func (m MockValidator) Validate(oldState Model, newState Model) []error {
	return nil
}

type MockValidatorWithError struct {
}

func (m MockValidatorWithError) Validate(oldState Model, newState Model) []error {

	var errors []error

	errors = append(errors, centerrors.New(code.DocumentInvalid, "first sample error"))
	errors = append(errors, centerrors.New(code.DocumentInvalid, "second sample error "))

	return errors
}

func TestValidatorInterface(t *testing.T) {

	var validator Validator

	validator = MockValidator{}

	errors := validator.Validate(nil, nil)

	assert.Nil(t, errors, "")

	validator = MockValidatorWithError{}

	errors = validator.Validate(nil, nil)

	assert.Equal(t, 2, len(errors), "Validate should return 2 errors")

}

func TestValidatorGroup_Validate(t *testing.T) {

	var testValidatorGroup = ValidatorGroup{
		validators: []Validator{
			MockValidator{},
			MockValidatorWithError{},
		},
	}

	errors := testValidatorGroup.Validate(nil, nil)

	assert.Equal(t, 2, len(errors), "Validate should return 2 errors")

}

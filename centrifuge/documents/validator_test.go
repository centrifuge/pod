// +build unit

package documents

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockValidator struct {
}

func (m MockValidator) Validate(oldState Model, newState Model) error {
	return nil
}

type MockValidatorWithErrors struct {
}

func (m MockValidatorWithErrors) Validate(oldState Model, newState Model) error {

	err := NewError("error_test", "error msg 1")
	err = AppendError(err, NewError("error_test2", "error msg 2"))

	return err
}

type MockValidatorWithOneError struct {
}

func (m MockValidatorWithOneError) Validate(oldState Model, newState Model) error {

	return fmt.Errorf("one error")
}

func TestValidatorInterface(t *testing.T) {

	var validator Validator

	// no error
	validator = MockValidator{}
	errors := validator.Validate(nil, nil)
	assert.Nil(t, errors, "")

	//one error
	validator = MockValidatorWithOneError{}
	errors = validator.Validate(nil, nil)
	assert.Error(t, errors, "error should be returned")
	assert.Equal(t, 1, LenError(errors), "errors should include one error")

	// more than one error
	validator = MockValidatorWithErrors{}
	errors = validator.Validate(nil, nil)
	assert.Error(t, errors, "error should be returned")
	assert.Equal(t, 2, LenError(errors), "errors should include two error")

	errorArray := Errors(errors)
	assert.Equal(t, 2, len(errorArray), "error array should include two error")

}

func TestValidatorGroup_Validate(t *testing.T) {

	var testValidatorGroup = ValidatorGroup{
		MockValidator{},
		MockValidatorWithOneError{},
		MockValidatorWithErrors{},
	}
	errors := testValidatorGroup.Validate(nil, nil)
	assert.Equal(t, 3, len(Errors(errors)), "Validate should return 2 errors")

	testValidatorGroup = ValidatorGroup{
		MockValidator{},
		MockValidatorWithErrors{},
		MockValidatorWithErrors{},
	}
	errors = testValidatorGroup.Validate(nil, nil)
	assert.Equal(t, 4, len(Errors(errors)), "Validate should return 4 errors")

	// empty group
	testValidatorGroup = ValidatorGroup{}
	errors = testValidatorGroup.Validate(nil, nil)
	assert.Equal(t, 0, len(Errors(errors)), "Validate should return no error")

	// group with no errors at all
	testValidatorGroup = ValidatorGroup{
		MockValidator{},
		MockValidator{},
		MockValidator{},
	}
	errors = testValidatorGroup.Validate(nil, nil)
	assert.Equal(t, 0, len(Errors(errors)), "Validate should return no error")

}

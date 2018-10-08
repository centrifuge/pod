// +build unit

package documents

import (
	"fmt"
	"testing"

	"github.com/centrifuge/go-centrifuge/centrifuge/documenterror"
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

	err := documenterror.New("error 1")
	err = documenterror.Append(err, documenterror.New("error 2"))

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
	assert.Equal(t, 1, documenterror.Len(errors), "errors should include one error")

	// more than one error
	validator = MockValidatorWithErrors{}
	errors = validator.Validate(nil, nil)
	assert.Error(t, errors, "error should be returned")
	assert.Equal(t, 2, documenterror.Len(errors), "errors should include two error")

	errorArray := documenterror.Errors(errors)
	assert.Equal(t, 2, len(errorArray), "error array should include two error")

}

func TestValidatorGroup_Validate(t *testing.T) {

	var testValidatorGroup = ValidatorGroup{
		MockValidator{},
		MockValidatorWithOneError{},
		MockValidatorWithErrors{},
	}
	errors := testValidatorGroup.Validate(nil, nil)
	assert.Equal(t, 3, len(documenterror.Errors(errors)), "Validate should return 2 errors")

	testValidatorGroup = ValidatorGroup{
		MockValidator{},
		MockValidatorWithErrors{},
		MockValidatorWithErrors{},
	}
	errors = testValidatorGroup.Validate(nil, nil)
	assert.Equal(t, 4, len(documenterror.Errors(errors)), "Validate should return 4 errors")

	// empty group
	testValidatorGroup = ValidatorGroup{}
	errors = testValidatorGroup.Validate(nil, nil)
	assert.Equal(t, 0, len(documenterror.Errors(errors)), "Validate should return no error")

	// group with no errors at all
	testValidatorGroup = ValidatorGroup{
		MockValidator{},
		MockValidator{},
		MockValidator{},
	}
	errors = testValidatorGroup.Validate(nil, nil)
	assert.Equal(t, 0, len(documenterror.Errors(errors)), "Validate should return no error")

}

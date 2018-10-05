// +build unit

package documents

import (
	"fmt"
	"testing"

	"github.com/centrifuge/go-centrifuge/centrifuge/centerrors"
	"github.com/centrifuge/go-centrifuge/centrifuge/code"
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

	errors := make(map[string]string)
	errors["error1"] = "first error"
	errors["error2"] = "second error"

	return centerrors.NewWithErrors(code.DocumentInvalid, "document is invalid", errors)
}

type MockValidatorWithOneError struct {
}

func (m MockValidatorWithOneError) Validate(oldState Model, newState Model) error {

	err := fmt.Errorf("one error")
	centerrors.Wrap(err, "second error")
	return err
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
	centErrors, ok := centerrors.FromError(errors)
	assert.False(t, ok, "should contain an unkown error")
	assert.Equal(t, code.Unknown, centErrors.Code(), "code should be unkown")
	assert.Equal(t, 0, len(centErrors.Errors()), "map should be empty")

	// more than one error
	validator = MockValidatorWithErrors{}
	errors = validator.Validate(nil, nil)
	assert.NotNil(t, errors, "should return some errors")

	centErrors, ok = centerrors.FromError(errors)
	assert.True(t, ok, "errors should contain centerrors")

	assert.Equal(t, 2, len(centErrors.Errors()), "Validate should return two entries in error map")
	assert.Equal(t, code.DocumentInvalid, centErrors.Code(), "error code should be DocumentInvalid")

}

func TestValidatorGroup_Validate(t *testing.T) {

	var testValidatorGroup = ValidatorGroup{
		MockValidator{},
		MockValidatorWithOneError{},
		MockValidatorWithErrors{},
	}
	errors := testValidatorGroup.Validate(nil, nil)
	centErrors, _ := centerrors.FromError(errors)
	assert.Equal(t, 2, len(centErrors.Errors()), "Validate should return 2 errors")

	testValidatorGroup = ValidatorGroup{
		MockValidator{},
		MockValidatorWithErrors{},
		MockValidatorWithErrors{},
	}
	errors = testValidatorGroup.Validate(nil, nil)
	centErrors, _ = centerrors.FromError(errors)
	assert.Equal(t, 4, len(centErrors.Errors()), "Validate should return 4 errors")

	// empty group
	testValidatorGroup = ValidatorGroup{}
	errors = testValidatorGroup.Validate(nil, nil)

	centErrors, _ = centerrors.FromError(errors)
	assert.Equal(t, 0, len(centErrors.Errors()), "Validate should return no errors")

	// group with no errors at all
	testValidatorGroup = ValidatorGroup{
		MockValidator{},
		MockValidator{},
		MockValidator{},
	}

	errors = testValidatorGroup.Validate(nil, nil)
	centErrors, _ = centerrors.FromError(errors)
	assert.Equal(t, 0, len(centErrors.Errors()), "Validate should return no errors")

}

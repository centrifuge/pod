// +build unit

package documents

import (
	"fmt"
	"testing"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/stretchr/testify/assert"
)

type MockValidator struct{}

func (m MockValidator) Validate(oldState Model, newState Model) error {
	return nil
}

type MockValidatorWithErrors struct{}

func (m MockValidatorWithErrors) Validate(oldState Model, newState Model) error {

	err := NewError("error_test", "error msg 1")
	err = errors.AppendError(err, NewError("error_test2", "error msg 2"))

	return err
}

type MockValidatorWithOneError struct{}

func (m MockValidatorWithOneError) Validate(oldState Model, newState Model) error {
	return fmt.Errorf("one error")
}

func TestValidatorInterface(t *testing.T) {
	var validator Validator

	// no error
	validator = MockValidator{}
	errs := validator.Validate(nil, nil)
	assert.Nil(t, errs, "")

	//one error
	validator = MockValidatorWithOneError{}
	errs = validator.Validate(nil, nil)
	assert.Error(t, errs, "error should be returned")
	assert.Equal(t, 1, errors.Len(errs), "errors should include one error")

	// more than one error
	validator = MockValidatorWithErrors{}
	errs = validator.Validate(nil, nil)
	assert.Error(t, errs, "error should be returned")
	assert.Equal(t, 2, errors.Len(errs), "errors should include two errors")
}

func TestValidatorGroup_Validate(t *testing.T) {

	var testValidatorGroup = ValidatorGroup{
		MockValidator{},
		MockValidatorWithOneError{},
		MockValidatorWithErrors{},
	}
	errs := testValidatorGroup.Validate(nil, nil)
	assert.Equal(t, 3, errors.Len(errs), "Validate should return 2 errors")

	testValidatorGroup = ValidatorGroup{
		MockValidator{},
		MockValidatorWithErrors{},
		MockValidatorWithErrors{},
	}
	errs = testValidatorGroup.Validate(nil, nil)
	assert.Equal(t, 4, errors.Len(errs), "Validate should return 4 errors")

	// empty group
	testValidatorGroup = ValidatorGroup{}
	errs = testValidatorGroup.Validate(nil, nil)
	assert.Equal(t, 0, errors.Len(errs), "Validate should return no error")

	// group with no errors at all
	testValidatorGroup = ValidatorGroup{
		MockValidator{},
		MockValidator{},
		MockValidator{},
	}
	errs = testValidatorGroup.Validate(nil, nil)
	assert.Equal(t, 0, errors.Len(errs), "Validate should return no error")
}

func TestIsCurrencyValid(t *testing.T) {
	tests := []struct {
		cur   string
		valid bool
	}{
		{
			cur:   "EUR",
			valid: true,
		},

		{
			cur:   "INR",
			valid: true,
		},

		{
			cur:   "some currency",
			valid: false,
		},
	}

	for _, c := range tests {
		got := IsCurrencyValid(c.cur)
		assert.Equal(t, c.valid, got, "result must match")
	}
}

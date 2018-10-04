// +build unit

package documents

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/centrifuge/centerrors"
	"github.com/centrifuge/go-centrifuge/centrifuge/code"
	"github.com/stretchr/testify/assert"
)

type MockValidator struct{}

func (m MockValidator) Validate(oldState Model, newState Model) []error {
	return nil
}

type MockValidatorWithError struct{}

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
		MockValidator{},
		MockValidatorWithError{},
	}
	errors := testValidatorGroup.Validate(nil, nil)
	assert.Equal(t, 2, len(errors), "Validate should return 2 errors")

	testValidatorGroup = ValidatorGroup{
		MockValidator{},
		MockValidatorWithError{},
		MockValidatorWithError{},
	}
	errors = testValidatorGroup.Validate(nil, nil)
	assert.Equal(t, 4, len(errors), "Validate should return 4 errors")

	testValidatorGroup = ValidatorGroup{}
	errors = testValidatorGroup.Validate(nil, nil)
	assert.Equal(t, 0, len(errors), "Validate should return no errors")
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

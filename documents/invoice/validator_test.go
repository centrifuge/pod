// +build unit

package invoice

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/stretchr/testify/assert"
)

func TestFieldValidator_Validate(t *testing.T) {
	fv := fieldValidator()

	//  nil error
	err := fv.Validate(nil, nil)
	assert.Error(t, err)
	errs := errors.GetErrs(err)
	assert.Len(t, errs, 1, "errors length must be one")
	assert.Contains(t, errs[0].Error(), "nil document")

	// unknown type
	err = fv.Validate(nil, &mockModel{})
	assert.Error(t, err)
	errs = errors.GetErrs(err)
	assert.Len(t, errs, 1, "errors length must be one")
	assert.Contains(t, errs[0].Error(), "unknown document type")

	// fail
	err = fv.Validate(nil, new(Invoice))
	assert.Error(t, err)
	errs = errors.GetErrs(err)
	assert.Len(t, errs, 1, "errors length must be 2")
	assert.Contains(t, errs[0].Error(), "currency is invalid")

	// success
	err = fv.Validate(nil, &Invoice{
		Currency: "EUR",
	})
	assert.Nil(t, err)
}

func TestCreateValidator(t *testing.T) {
	cv := CreateValidator()
	assert.Len(t, cv, 1)
}

func TestUpdateValidator(t *testing.T) {
	uv := UpdateValidator()
	assert.Len(t, uv, 2)
}

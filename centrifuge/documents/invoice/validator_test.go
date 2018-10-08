// +build unit

package invoice

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/centrifuge/documents"
	"github.com/stretchr/testify/assert"
)

func TestFieldValidator_Validate(t *testing.T) {
	fv := fieldValidator()

	//  nil error
	err := fv.Validate(nil, nil)
	assert.Error(t, err)
	errs := documents.Errors(err)
	assert.Len(t, errs, 1, "errors length must be one")
	assert.Contains(t, errs[0].Error(), "nil document")

	// unknown type
	err = fv.Validate(nil, &mockModel{})
	assert.Error(t, err)
	errs = documents.Errors(err)
	assert.Len(t, errs, 1, "errors length must be one")
	assert.Contains(t, errs[0].Error(), "unknown document type")

	// fail
	err = fv.Validate(nil, new(InvoiceModel))
	assert.Error(t, err)
	errs = documents.Errors(err)
	assert.Len(t, errs, 1, "errors length must be 2")
	assert.Contains(t, errs[0].Error(), "currency is invalid")

	// success
	err = fv.Validate(nil, &InvoiceModel{
		Currency: "EUR",
	})
	assert.Nil(t, err)
}

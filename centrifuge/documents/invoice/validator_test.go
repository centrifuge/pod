// +build unit

package invoice

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/centrifuge/identity"
	"github.com/stretchr/testify/assert"
)

func TestFieldValidator_Validate(t *testing.T) {
	fv := fieldValidator()

	//  nil error
	errs := fv.Validate(nil, nil)
	assert.Len(t, errs, 1, "errors length must be one")
	assert.Contains(t, errs[0].Error(), "nil document")

	// unknown type
	errs = fv.Validate(nil, &mockModel{})
	assert.Len(t, errs, 1, "errors length must be one")
	assert.Contains(t, errs[0].Error(), "unknown document type")

	// fail
	errs = fv.Validate(nil, new(InvoiceModel))
	assert.Len(t, errs, 2, "errors length must be 2")
	assert.Contains(t, errs[0].Error(), "currency is invalid")
	assert.Contains(t, errs[1].Error(), "payee is invalid")

	// success
	errs = fv.Validate(nil, &InvoiceModel{
		Currency: "EUR",
		Payee:    identity.CentID([identity.CentIDByteLength]byte{1, 2, 3, 4, 5, 6}),
	})
	assert.Len(t, errs, 0, "errors must be nil")
}

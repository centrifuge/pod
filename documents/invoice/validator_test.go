// +build unit

package invoice

import (
	"fmt"
	"testing"

	"context"

	"github.com/centrifuge/go-centrifuge/coredocument"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/header"
	"github.com/centrifuge/go-centrifuge/testingutils/documents"
	"github.com/centrifuge/go-centrifuge/utils"
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
	err = fv.Validate(nil, new(Invoice))
	assert.Error(t, err)
	errs = documents.Errors(err)
	assert.Len(t, errs, 1, "errors length must be 2")
	assert.Contains(t, errs[0].Error(), "currency is invalid")

	// success
	err = fv.Validate(nil, &Invoice{
		Currency: "EUR",
	})
	assert.Nil(t, err)
}

func TestDataRootValidation_Validate(t *testing.T) {
	drv := dataRootValidator()

	// nil error
	err := drv.Validate(nil, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nil document")

	// pack coredoc failed
	model := &mockModel{}
	model.On("PackCoreDocument").Return(nil, fmt.Errorf("error")).Once()
	err = drv.Validate(nil, model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to pack coredocument")

	// missing data root
	model = &mockModel{}
	model.On("PackCoreDocument").Return(coredocument.New(), nil).Once()
	err = drv.Validate(nil, model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "data root missing")

	// unknown doc type
	cd := coredocument.New()
	cd.DataRoot = utils.RandomSlice(32)
	model = &mockModel{}
	model.On("PackCoreDocument").Return(cd, nil).Once()
	err = drv.Validate(nil, model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown document type")

	// mismatch
	contextHeader, err := header.NewContextHeader(context.Background(), cfg)
	assert.Nil(t, err)
	inv := new(Invoice)
	err = inv.InitInvoiceInput(testingdocuments.CreateInvoicePayload(), contextHeader)
	assert.Nil(t, err)
	inv.CoreDocument = cd
	err = drv.Validate(nil, inv)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mismatched data root")

	// success
	inv = new(Invoice)
	err = inv.InitInvoiceInput(testingdocuments.CreateInvoicePayload(), contextHeader)
	assert.Nil(t, err)
	err = inv.calculateDataRoot()
	assert.Nil(t, err)
	err = drv.Validate(nil, inv)
	assert.Nil(t, err)
}

func TestCreateValidator(t *testing.T) {
	cv := CreateValidator()
	assert.Len(t, cv, 2)
}

func TestUpdateValidator(t *testing.T) {
	uv := UpdateValidator()
	assert.Len(t, uv, 3)
}

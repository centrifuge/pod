// +build unit

package coredocument

import (
	"fmt"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/centrifuge/tools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockModel struct {
	mock.Mock
	documents.Model
}

func (m mockModel) PackCoreDocument() (*coredocumentpb.CoreDocument, error) {
	args := m.Called()
	cd, _ := args.Get(0).(*coredocumentpb.CoreDocument)
	return cd, args.Error(1)
}

func Test_getCoreDocument(t *testing.T) {
	// nil document
	cd, err := getCoreDocument(nil)
	assert.Error(t, err)
	assert.Nil(t, cd)

	// pack core document fail
	model := mockModel{}
	model.On("PackCoreDocument").Return(nil, fmt.Errorf("err")).Once()
	cd, err = getCoreDocument(model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Nil(t, cd)

	// success
	model = mockModel{}
	cd = New()
	model.On("PackCoreDocument").Return(cd, nil).Once()
	got, err := getCoreDocument(model)
	model.AssertExpectations(t)
	assert.Nil(t, err)
	assert.Equal(t, cd, got)
}

func TestValidator_fieldValidator(t *testing.T) {
	fv := fieldValidator()

	// fail getCoreDocument
	model := mockModel{}
	model.On("PackCoreDocument").Return(nil, fmt.Errorf("err")).Once()
	err := fv.Validate(nil, model)
	model.AssertExpectations(t)
	assert.Error(t, err)

	// failed validator
	model = mockModel{}
	cd := New()
	model.On("PackCoreDocument").Return(cd, nil).Once()
	err = fv.Validate(nil, model)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cd_salts : Required field")

	// success
	model = mockModel{}
	cd.DataRoot = tools.RandomSlice(32)
	FillSalts(cd)
	model.On("PackCoreDocument").Return(cd, nil).Once()
	err = fv.Validate(nil, model)
	assert.Nil(t, err)
}

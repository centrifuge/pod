package coredocument

import (
	"fmt"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents"
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

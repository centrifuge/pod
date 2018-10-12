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
	documents.Model
	mock.Mock
}

func (m *mockModel) PackCoreDocument() (*coredocumentpb.CoreDocument, error) {
	args := m.Called()
	cd, _ := args.Get(0).(*coredocumentpb.CoreDocument)
	return cd, args.Error(1)
}

func TestUpdateVersionValidator(t *testing.T) {
	uvv := UpdateVersionValidator()

	// nil models
	err := uvv.Validate(nil, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "need both the old and new model")

	// old model pack core doc fail
	old := &mockModel{}
	old.On("PackCoreDocument").Return(nil, fmt.Errorf("error")).Once()
	err = uvv.Validate(old, nil)
	old.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch old core document")

	// new model pack core doc fail
	new := &mockModel{}
	oldCD := New()
	oldCD.DocumentRoot = tools.RandomSlice(32)
	old.On("PackCoreDocument").Return(oldCD, nil).Once()
	new.On("PackCoreDocument").Return(nil, fmt.Errorf("error")).Once()
	err = uvv.Validate(old, new)
	old.AssertExpectations(t)
	new.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch new core document")

	// mismatched identifiers
	newCD := New()
	newCD.NextVersion = nil
	old.On("PackCoreDocument").Return(oldCD, nil).Once()
	new.On("PackCoreDocument").Return(newCD, nil).Once()
	err = uvv.Validate(old, new)
	old.AssertExpectations(t)
	new.AssertExpectations(t)
	assert.Error(t, err)
	assert.Len(t, documents.ConvertToMap(err), 4)

	// success
	newCD, err = PrepareNewVersion(*oldCD, nil)
	assert.Nil(t, err)
	old.On("PackCoreDocument").Return(oldCD, nil).Once()
	new.On("PackCoreDocument").Return(newCD, nil).Once()
	err = uvv.Validate(old, new)
	old.AssertExpectations(t)
	new.AssertExpectations(t)
	assert.Nil(t, err)
}

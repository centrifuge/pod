// +build unit

package documents_test

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/testingutils/documents"
	"github.com/stretchr/testify/assert"
)

func TestRegistry_Register_LocateService_successful(t *testing.T) {
	registry := documents.NewServiceRegistry()
	a := &testingdocuments.MockService{}
	dm, err := testingdocuments.GenerateCoreDocumentModel()
	assert.NoError(t, err)
	documentType := dm.Document.EmbeddedData.TypeUrl
	err = registry.Register(documentType, a)
	assert.Nil(t, err, "register didn't work with unused id")

	b, err := registry.LocateService(documentType)
	assert.Nil(t, err, "service hasn't been registered properly")
	assert.Equal(t, a, b, "locateService should return the same service ")
}

func TestRegistry_Register_invalidId(t *testing.T) {
	registry := documents.NewServiceRegistry()
	a := &testingdocuments.MockService{}
	dm, err := testingdocuments.GenerateCoreDocumentModel()
	assert.NoError(t, err)
	cd := dm.Document
	cd.EmbeddedData.TypeUrl = "testID_1"

	err = registry.Register(cd.EmbeddedData.TypeUrl, a)
	assert.Nil(t, err, "register didn't work with unused id")

	err = registry.Register(cd.EmbeddedData.TypeUrl, a)
	assert.Error(t, err, "register shouldn't work with same id")

	err = registry.Register("testId", a)
	assert.Nil(t, err, "register didn't work with unused id")
}

func TestRegistry_LocateService_invalid(t *testing.T) {
	registry := documents.NewServiceRegistry()
	dm, err := testingdocuments.GenerateCoreDocumentModel()
	assert.NoError(t, err)
	cd := dm.Document
	cd.EmbeddedData.TypeUrl = "testID_2"
	documentType := cd.EmbeddedData.TypeUrl
	_, err = registry.LocateService(documentType)
	assert.Error(t, err, "should throw an error because no services is registered")
}

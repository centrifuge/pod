// +build unit

package documents_test

import (
	"testing"

	cd "github.com/centrifuge/go-centrifuge/coredocument"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/testingutils/coredocument"
	"github.com/centrifuge/go-centrifuge/testingutils/documents"
	"github.com/stretchr/testify/assert"
)

func TestServiceRegistry_FindService(t *testing.T) {
	registry := documents.NewServiceRegistry()
	a := &testingdocuments.MockService{}
	b := &testingdocuments.MockService{}
	a.On("Exists").Return(true)
	b.On("Exists").Return(false)
	err := registry.Register("a service", a)
	err = registry.Register("b service", b)

	service, err := registry.FindService([]byte{})
	assert.Nil(t, err, "findService should be successful")
	assert.Equal(t, a, service, "service a should be returned")
}

func TestRegistry_Register_LocateService_successful(t *testing.T) {
	registry := documents.NewServiceRegistry()
	a := &testingdocuments.MockService{}
	coreDocument := testingcoredocument.GenerateCoreDocument()
	documentType, err := cd.GetTypeURL(coreDocument)
	assert.Nil(t, err, "should not throw an error because core document contains a type")

	err = registry.Register(documentType, a)
	assert.Nil(t, err, "register didn't work with unused id")

	b, err := registry.LocateService(documentType)
	assert.Nil(t, err, "service hasn't been registered properly")
	assert.Equal(t, a, b, "locateService should return the same service ")
}

func TestRegistry_Register_invalidId(t *testing.T) {
	registry := documents.NewServiceRegistry()
	a := &testingdocuments.MockService{}
	coreDocument := testingcoredocument.GenerateCoreDocument()
	coreDocument.EmbeddedData.TypeUrl = "testID_1"

	err := registry.Register(coreDocument.EmbeddedData.TypeUrl, a)
	assert.Nil(t, err, "register didn't work with unused id")

	err = registry.Register(coreDocument.EmbeddedData.TypeUrl, a)
	assert.Error(t, err, "register shouldn't work with same id")

	err = registry.Register("testId", a)
	assert.Nil(t, err, "register didn't work with unused id")
}

func TestRegistry_LocateService_invalid(t *testing.T) {
	registry := documents.NewServiceRegistry()
	coreDocument := testingcoredocument.GenerateCoreDocument()
	coreDocument.EmbeddedData.TypeUrl = "testID_2"
	documentType, err := cd.GetTypeURL(coreDocument)
	assert.Nil(t, err, "should not throw an error because core document contains a type")

	_, err = registry.LocateService(documentType)
	assert.Error(t, err, "should throw an error because no services is registered")
}

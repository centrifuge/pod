// +build unit

package documents_test

import (
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/centrifuge/testingutils"
	"github.com/stretchr/testify/assert"
)

type MockService struct{}

func (m *MockService) DeriveWithCoreDocument(cd *coredocumentpb.CoreDocument) (documents.Model, error) {
	return nil, nil
}

func TestRegistry_GetRegistryInstance(t *testing.T) {

	registryFirst := documents.GetRegistryInstance()
	registrySecond := documents.GetRegistryInstance()

	assert.Equal(t, &registryFirst, &registrySecond, "only one instance of registry should exist")
}

func TestRegistry_Register_LocateService_successful(t *testing.T) {

	registry := documents.GetRegistryInstance()

	a := &MockService{}

	coreDocument := testingutils.GenerateCoreDocument()
	err := registry.Register(coreDocument.EmbeddedData.TypeUrl, a)

	assert.Nil(t, err, "register didn't work with unused id")

	b, err := registry.LocateService(coreDocument)
	assert.Nil(t, err, "service hasn't been registered properly")

	assert.Equal(t, a, b, "locateService should return the same service ")

	documents.KillRegistry()

}

func TestRegistry_Register_invalidId(t *testing.T) {

	registry := documents.GetRegistryInstance()

	a := &MockService{}

	coreDocument := testingutils.GenerateCoreDocument()

	err := registry.Register(coreDocument.EmbeddedData.TypeUrl, a)
	assert.Nil(t, err, "register didn't work with unused id")

	err = registry.Register(coreDocument.EmbeddedData.TypeUrl, a)
	assert.Error(t, err, "register shouldn't work with same id")

	err = registry.Register("testId", a)
	assert.Nil(t, err, "register didn't work with unused id")

	documents.KillRegistry()

}

func TestRegistry_LocateService_invalid(t *testing.T) {

	registry := documents.GetRegistryInstance()
	coreDocument := testingutils.GenerateCoreDocument()

	_, err := registry.LocateService(coreDocument)
	assert.Error(t, err, "should throw an error because no services is registered")

	documents.KillRegistry()

}

func TestRegistry_Unregister(t *testing.T) {

	registry := documents.GetRegistryInstance()
	err := registry.Unregister("testId")
	assert.Error(t, err, "unregister should fail because no services is registered")

	a := &MockService{}

	coreDocument := testingutils.GenerateCoreDocument()

	err = registry.Register(coreDocument.EmbeddedData.TypeUrl, a)
	assert.Nil(t, err, "register didn't work with unused id")

	err = registry.Unregister(coreDocument.EmbeddedData.TypeUrl)
	assert.Nil(t, err, "unregister should work if a service is registered")

	documents.KillRegistry()

}

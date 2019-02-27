// +build unit

package documents_test

import (
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/testingutils/documents"
	"github.com/stretchr/testify/assert"
)

func TestRegistry_Register_LocateService_successful(t *testing.T) {
	registry := documents.NewServiceRegistry()
	a := &testingdocuments.MockService{}
	docType := documenttypes.InvoiceDataTypeUrl
	err := registry.Register(docType, a)
	assert.Nil(t, err)

	b, err := registry.LocateService(docType)
	assert.Nil(t, err, "service hasn't been registered properly")
	assert.Equal(t, a, b, "locateService should return the same service ")
}

func TestRegistry_Register_invalidId(t *testing.T) {
	registry := documents.NewServiceRegistry()
	a := &testingdocuments.MockService{}
	docType := documenttypes.InvoiceDataTypeUrl
	err := registry.Register(docType, a)
	assert.Nil(t, err, "register didn't work with unused id")

	err = registry.Register(docType, a)
	assert.Error(t, err, "register shouldn't work with same id")

	err = registry.Register("testId", a)
	assert.Nil(t, err, "register didn't work with unused id")
}

func TestRegistry_LocateService_invalid(t *testing.T) {
	registry := documents.NewServiceRegistry()
	_, err := registry.LocateService(documenttypes.InvoiceDataTypeUrl)
	assert.Error(t, err, "should throw an error because no services is registered")
}

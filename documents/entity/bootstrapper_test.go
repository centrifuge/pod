// +build unit

package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBootstrapper_Bootstrap(t *testing.T) {
	err := (&Bootstrapper{}).Bootstrap(map[string]interface{}{})
	assert.Error(t, err, "Should throw an error because of empty context")
}

func TestBootstrapper_registerEntityService(t *testing.T) {
	//context := map[string]interface{}{}
	//context[bootstrap.BootstrappedDb] = true
	//err := (&Bootstrapper{}).Bootstrap(context)
	//assert.Nil(t, err, "Should throw because context is passed")
	//
	////coreDocument embeds a entity
	//coreDocument := testingutils.GenerateCoreDocument()
	//registry := document.GetRegistryInstance()
	//
	//documentType, err := cd.GetTypeUrl(coreDocument)
	//assert.Nil(t, err, "should not throw an error because document contains a type")
	//
	//service, err := registry.LocateService(documentType)
	//assert.Nil(t, err, "service should be successful registered and able to locate")
	//assert.NotNil(t, service, "service should be returned")
}

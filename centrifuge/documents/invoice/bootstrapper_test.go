// +build unit

package invoice

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/centrifuge/bootstrapper"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/centrifuge/testingutils"

	"github.com/stretchr/testify/assert"
)

func TestBootstrapper_Bootstrap(t *testing.T) {
	err := (&Bootstrapper{}).Bootstrap(map[string]interface{}{})
	assert.Error(t, err, "Should throw an error because of empty context")
}

func TestBootstrapper_registerInvoiceService(t *testing.T) {

	context := map[string]interface{}{}
	context[bootstrapper.BootstrappedLevelDb] = true
	err := (&Bootstrapper{}).Bootstrap(context)
	assert.Nil(t, err, "Should throw because context is passed")

	registry := documents.GetRegistryInstance()

	//coreDocument embeds a invoice
	coreDocument := testingutils.GenerateCoreDocument()

	service, err := registry.LocateService(coreDocument)
	assert.Nil(t, err, "service should be successful registered and able to locate")

	assert.NotNil(t, service, "service should be returned")

}

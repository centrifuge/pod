// +build unit

package purchaseorder

import (
	"os"
	"testing"

	cc "github.com/centrifuge/go-centrifuge/centrifuge/context/testingbootstrap"
		"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	cc.TestIntegrationBootstrap()
	result := m.Run()
	cc.TestIntegrationTearDown()
	os.Exit(result)
}

func TestBootstrapper_Bootstrap(t *testing.T) {
	err := (&Bootstrapper{}).Bootstrap(map[string]interface{}{})
	assert.Error(t, err, "Should throw an error because of empty context")
}

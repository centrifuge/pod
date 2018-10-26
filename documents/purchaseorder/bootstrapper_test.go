// +build unit

package purchaseorder

import (
	"flag"
	"os"
	"testing"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/context/testlogging"
	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	ibootstappers := []bootstrap.TestBootstrapper{
		&testlogging.TestLoggingBootstrapper{},
		&config.Bootstrapper{},
		&storage.Bootstrapper{},
		&Bootstrapper{},
	}

	bootstrap.RunTestBootstrappers(ibootstappers, nil)
	flag.Parse()
	result := m.Run()
	bootstrap.RunTestTeardown(ibootstappers)
	os.Exit(result)
}

func TestBootstrapper_Bootstrap(t *testing.T) {
	err := (&Bootstrapper{}).Bootstrap(map[string]interface{}{})
	assert.Error(t, err, "Should throw an error because of empty context")
}

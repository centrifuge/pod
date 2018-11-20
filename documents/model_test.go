package documents

import (
	"os"
	"testing"

	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/context/testlogging"
	"github.com/centrifuge/go-centrifuge/storage"
)

var ctx = map[string]interface{}{}
var cfg *config.Configuration

func TestMain(m *testing.M) {
	ibootstappers := []bootstrap.TestBootstrapper{
		&testlogging.TestLoggingBootstrapper{},
		&config.Bootstrapper{},
		&storage.Bootstrapper{},
		&Bootstrapper{},
	}
	bootstrap.RunTestBootstrappers(ibootstappers, ctx)
	cfg = ctx[config.BootstrappedConfig].(*config.Configuration)
	result := m.Run()
	bootstrap.RunTestTeardown(ibootstappers)
	os.Exit(result)
}

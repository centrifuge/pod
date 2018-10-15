package testlogging

import (
	"os"

	logging "github.com/ipfs/go-log"
	gologging "github.com/whyrusleeping/go-logging"
)

type TestLoggingBootstrapper struct{}

func (TestLoggingBootstrapper) TestBootstrap(context map[string]interface{}) error {
	logging.SetAllLoggers(gologging.DEBUG)
	backend := gologging.NewLogBackend(os.Stdout, "", 0)
	gologging.SetBackend(backend)
	return nil
}

func (TestLoggingBootstrapper) TestTearDown() error {
	return nil
}

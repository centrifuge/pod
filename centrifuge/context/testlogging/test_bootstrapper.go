// +build integration unit

package testlogging

import (
	"github.com/centrifuge/go-centrifuge/centrifuge/utils"
	"os"

	logging "github.com/ipfs/go-log"
	gologging "github.com/whyrusleeping/go-logging"
)

type TestLoggingBootstrapper struct{}

func (TestLoggingBootstrapper) TestBootstrap(context map[string]interface{}) error {

	var format = gologging.MustStringFormatter(utils.GetCentLogFormat())

	logging.SetAllLoggers(gologging.DEBUG)
	gologging.SetFormatter(format)

	backend := gologging.NewLogBackend(os.Stdout, "", 0)
	gologging.SetBackend(backend)


	return nil
}

func (TestLoggingBootstrapper) TestTearDown() error {
	return nil
}

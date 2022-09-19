//go:build integration || testworld

package anchors

import (
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/errors"
)

func (b Bootstrapper) TestBootstrap(context map[string]interface{}) error {
	if _, ok := context[bootstrap.BootstrappedConfig]; !ok {
		return errors.New("config hasn't been initialized")
	}

	return b.Bootstrap(context)
}

func (b Bootstrapper) TestTearDown() error {
	return nil
}

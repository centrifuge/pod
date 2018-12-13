// +build integration unit

package anchors

import (
	"errors"

	"github.com/centrifuge/go-centrifuge/bootstrap"
)

const BootstrappedAnchorRepository string = "BootstrappedAnchorRepository"

func (b Bootstrapper) TestBootstrap(context map[string]interface{}) error {
	if _, ok := context[bootstrap.BootstrappedConfig]; !ok {
		return errors.New("config hasn't been initialized")
	}

	return b.Bootstrap(context)
}

func (b Bootstrapper) TestTearDown() error {
	return nil
}

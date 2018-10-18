// +build integration unit

package anchors

import (
	"github.com/centrifuge/go-centrifuge/centrifuge/bootstrap"
	"errors"
)

func (b *Bootstrapper) TestBootstrap(context map[string]interface{}) error {
	if _, ok := context[bootstrap.BootstrappedConfig]; !ok {
		return errors.New("config hasn't been initialized")
	}

	if repo, ok := context[bootstrap.BootstrappedAnchorRepository]; ok {
		setAnchorRepository(repo.(AnchorRepository))
	}

	return nil
}

func (b *Bootstrapper) TestTearDown() error {
	return nil
}

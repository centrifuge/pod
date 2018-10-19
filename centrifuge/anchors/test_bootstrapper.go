// +build integration unit

package anchors

import (
	"errors"

	"github.com/centrifuge/go-centrifuge/centrifuge/bootstrap"
)

func (b *Bootstrapper) TestBootstrap(context map[string]interface{}) error {
	if _, ok := context[bootstrap.BootstrappedConfig]; !ok {
		return errors.New("config hasn't been initialized")
	}

	if repo, ok := context[bootstrap.BootstrappedAnchorRepository]; ok {
		setAnchorRepository(repo.(AnchorRepository))
	} else {
		b.Bootstrap(context)
	}

	return nil
}

func (b *Bootstrapper) TestTearDown() error {
	return nil
}

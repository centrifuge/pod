// +build integration unit

package anchors

import (
	"errors"

	"github.com/centrifuge/go-centrifuge/config"
)

const BootstrappedAnchorRepository string = "BootstrappedAnchorRepository"

func (b *Bootstrapper) TestBootstrap(context map[string]interface{}) error {
	if _, ok := context[config.BootstrappedConfig]; !ok {
		return errors.New("config hasn't been initialized")
	}

	if repo, ok := context[BootstrappedAnchorRepository]; ok {
		setAnchorRepository(repo.(AnchorRepository))
	} else {
		b.Bootstrap(context)
	}

	return nil
}

func (b *Bootstrapper) TestTearDown() error {
	return nil
}

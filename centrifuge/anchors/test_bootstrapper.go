// +build integration unit

package anchors

import (
	"errors"
	"fmt"

	"github.com/centrifuge/go-centrifuge/centrifuge/bootstrap"
)

func (b *Bootstrapper) TestBootstrap(context map[string]interface{}) error {
	if _, ok := context[bootstrap.BootstrappedConfig]; !ok {
		return errors.New("config hasn't been initialized")
	}

	if repo, ok := context[bootstrap.BootstrappedAnchorRepository]; ok {
		fmt.Printf("Bootstrapped! %v\n", repo.(AnchorRepository))
		setAnchorRepository(repo.(AnchorRepository))
	} else {
		b.Bootstrap(context)
	}

	return nil
}

func (b *Bootstrapper) TestTearDown() error {
	return nil
}

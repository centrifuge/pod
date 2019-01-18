// +build integration unit

package documents

import (
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/storage"
)

func (b Bootstrapper) TestBootstrap(context map[string]interface{}) error {
	if _, ok := context[storage.BootstrappedDB]; !ok {
		return errors.New("initializing LevelDB repository failed")
	}
	return b.Bootstrap(context)
}

func (Bootstrapper) TestTearDown() error {
	return nil
}

func (b PostBootstrapper) TestBootstrap(ctx map[string]interface{}) error {
	return b.Bootstrap(ctx)
}

func (PostBootstrapper) TestTearDown() error {
	return nil
}

func (b DocumentServiceBootstrapper) TestBootstrap(ctx map[string]interface{}) error {
	return b.Bootstrap(ctx)
}

func (DocumentServiceBootstrapper) TestTearDown() error {
	return nil
}

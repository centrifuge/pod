// +build integration unit

package queue

import "github.com/centrifuge/go-centrifuge/bootstrap"

func (b *Bootstrapper) TestBootstrap(context map[string]interface{}) error {
	return b.Bootstrap(context)
}

func (b *Bootstrapper) TestTearDown() error {
	srv := b.context[bootstrap.BootstrappedQueueServer].(*Server)
	return srv.Stop()
}

//go:build unit || integration || testworld

package p2p

func (b Bootstrapper) TestBootstrap(ctx map[string]interface{}) error {
	return b.Bootstrap(ctx)
}

func (b Bootstrapper) TestTearDown() error {
	return nil
}

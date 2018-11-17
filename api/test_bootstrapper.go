// +build unit integration

package api

func (b Bootstrapper) TestBootstrap(ctx map[string]interface{}) error {
	return b.Bootstrap(ctx)
}

func (b Bootstrapper) TestTearDown() error {
	return nil
}

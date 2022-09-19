//go:build integration || testworld

package documents

func (b Bootstrapper) TestBootstrap(ctx map[string]interface{}) error {
	return b.Bootstrap(ctx)
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

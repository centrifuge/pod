package documents

func (Bootstrapper) TestTearDown() error {
	return nil
}

func (b PostBootstrapper) TestBootstrap(ctx map[string]interface{}) error {
	return b.Bootstrap(ctx)
}

func (PostBootstrapper) TestTearDown() error {
	return nil
}

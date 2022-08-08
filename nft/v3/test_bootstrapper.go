//go:build unit || integration || testworld

package v3

func (b *Bootstrapper) TestBootstrap(context map[string]interface{}) error {
	return b.Bootstrap(context)
}

func (b *Bootstrapper) TestTearDown() error {
	return nil
}

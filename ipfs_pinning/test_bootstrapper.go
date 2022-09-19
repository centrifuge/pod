//go:build unit || integration || testworld

package ipfs_pinning

func (b *Bootstrapper) TestBootstrap(context map[string]interface{}) error {
	return b.Bootstrap(context)
}

func (*Bootstrapper) TestTearDown() error {
	return nil
}

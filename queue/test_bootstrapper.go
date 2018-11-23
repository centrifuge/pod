// +build integration unit

package queue

func (b Bootstrapper) TestBootstrap(context map[string]interface{}) error {
	return b.Bootstrap(context)
}

func (b Bootstrapper) TestTearDown() error {
	StopQueue()
	return nil
}

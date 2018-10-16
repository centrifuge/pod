package documents

type Bootstrapper struct{}

// Bootstrap sets the required storage and registers
func (*Bootstrapper) Bootstrap(context map[string]interface{}) error {
	return nil
}

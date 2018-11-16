package documents

// BootstrappedRegistry is the key to ServiceRegistry in Bootstrap context
const BootstrappedRegistry = "BootstrappedRegistry"

type Bootstrapper struct{}

// Bootstrap sets the required storage and registers
func (Bootstrapper) Bootstrap(ctx map[string]interface{}) error {
	ctx[BootstrappedRegistry] = &ServiceRegistry{
		services: make(map[string]Service),
	}
	return nil
}

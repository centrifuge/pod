package bootstrap

// DO NOT PUT any app logic in this package to avoid any dependency cycles

// Bootstrap constants are keys to mapped value in bootstrapped context
const (
	BootstrappedP2PServer   string = "BootstrappedP2PServer"
	BootstrappedAPIServer   string = "BootstrappedAPIServer"
	BootstrappedQueueServer string = "BootstrappedQueueServer"
	NodeObjRegistry         string = "NodeObjRegistry"
)

// Bootstrapper must be implemented by all packages that needs bootstrapping at application start
type Bootstrapper interface {
	Bootstrap(context map[string]interface{}) error
}

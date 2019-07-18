package bootstrap

// DO NOT PUT any app logic in this package to avoid any dependency cycles

// Bootstrap constants are keys to mapped value in bootstrapped context
const (
	BootstrappedConfig      string = "BootstrappedConfig"
	BootstrappedPeer        string = "BootstrappedPeer"
	BootstrappedAPIServer   string = "BootstrappedAPIServer"
	BootstrappedQueueServer string = "BootstrappedQueueServer"
	NodeObjRegistry         string = "NodeObjRegistry"
	// BootstrappedInvoiceUnpaid is the key to InvoiceUnpaid NFT in bootstrap context.
	BootstrappedInvoiceUnpaid = "BootstrappedInvoiceUnpaid"
)

// Bootstrapper must be implemented by all packages that needs bootstrapping at application start
type Bootstrapper interface {
	Bootstrap(context map[string]interface{}) error
}

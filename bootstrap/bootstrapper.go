package bootstrap

// DO NOT PUT any app logic in this package to avoid any dependency cycles

// Bootstrap constants are keys to mapped value in bootstrapped context
const (
	BootstrappedConfig    string = "BootstrappedConfig"
	BootstrappedPeer      string = "BootstrappedPeer"
	BootstrappedAPIServer string = "BootstrappedAPIServer"
	// BootstrappedNFTService is the key to NFT Service in bootstrap context.
	BootstrappedNFTService = "BootstrappedNFTService"
	// BootstrappedNFTV3Service is the key to the v3 NFT Service in bootstrap context.
	BootstrappedNFTV3Service = "BootstrappedNFTV3Service"
)

// NodeObjRegistry key for context
var NodeObjRegistry struct{}

// Bootstrapper must be implemented by all packages that needs bootstrapping at application start
type Bootstrapper interface {
	Bootstrap(context map[string]interface{}) error
}

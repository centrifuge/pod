package bootstrap

// DO NOT PUT any app logic in this package to avoid any dependency cycles

// Bootstrapper must be implemented by all packages that needs bootstrapping at application start
type Bootstrapper interface {
	Bootstrap(context map[string]interface{}) error
}

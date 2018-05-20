package networks

var networkConfigurationLoader NetworkConfigurationLoader

func GetNetworkConfigurationLoader() NetworkConfigurationLoader {
	return networkConfigurationLoader
}

// NetworkConfiguration holds all information required for a Centrifuge node to
// connnect to a given network.
type NetworkConfiguration interface {
	GetNetworkString() string
	GetContractAddress(string) ([]byte, error)
	GetBootstrapPeers() []string
}

// NetworkConfigurationLoader is used to get the NetworkConfiguration based on
// the network identifier defined in the application's configuration.
type NetworkConfigurationLoader interface {
	GetConfigurationFromKey(key string) (NetworkConfiguration, error)
}

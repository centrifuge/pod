package registry

import (
	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("anchor")

type AnchorRegistry interface {
	RegisterAsAnchor(anchorID [32]byte, rootHash [32]byte) (<-chan *WatchAnchor, error)
}

// RegisterAsAnchor registers the given AnchorID and RootHash as an anchor on the configured anchor registry
func RegisterAsAnchor(anchorID [32]byte, rootHash [32]byte) (<-chan *WatchAnchor, error) {
	registry, _ := getConfiguredRegistry()

	confirmations, err := registry.RegisterAsAnchor(anchorID, rootHash)
	if err != nil {
		log.Errorf("Failed to register the anchor [id:%x, hash:%x ]: %v", anchorID, rootHash, err)
	}
	return confirmations, err
}

// getConfiguredRegistry will later pull a configured registry (if not only using Ethereum as the anchor registry)
// For now hard-coded to the Ethereum setup
func getConfiguredRegistry() (AnchorRegistry, error) {
	return &EthereumAnchorRegistry{}, nil
}

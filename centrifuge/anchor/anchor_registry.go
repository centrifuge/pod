package anchor

import (
	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("anchor")

type AnchorRegistry interface {
	RegisterAsAnchor(anchorID string, rootHash string, confirmations chan<- *Anchor) error
}

// RegisterAsAnchor registers the given AnchorID and RootHash as an anchor on the configured anchor registry
func RegisterAsAnchor(anchorID string, rootHash string, confirmations chan<- *Anchor) error {
	registry, _ := getConfiguredRegistry()

	err := registry.RegisterAsAnchor(anchorID, rootHash, confirmations)
	if err != nil {
		log.Fatalf("Failed to register the anchor [id:%x, hash:%x ]: %v", anchorID, rootHash, err)
	}
	return err
}

// getConfiguredRegistry will later pull a configured registry (if not only using Ethereum as the anchor registry)
// For now hard-coded to the Ethereum setup
func getConfiguredRegistry() (AnchorRegistry, error) {
	return &EthereumAnchorRegistry{}, nil
}

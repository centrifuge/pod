package anchor

import (
	"log"
)

type AnchorRegistry interface {
	RegisterAsAnchor(anchorID string, rootHash string, confirmations chan<- *Anchor) ( error)
}

// Register the given anchorID and root has as an anchor
func RegisterAsAnchor(anchorID string, rootHash string, confirmations chan<- *Anchor) ( error) {
	registry, _ := getConfiguredRegistry()

	 err := registry.RegisterAsAnchor(anchorID, rootHash, confirmations )
	if err != nil {
		log.Fatalf("Failed to register the anchor [id:%v, hash:%v ]: %v", anchorID, rootHash, err)
	}
	return  err
}

// This will later pull a configured registry (if not only using Ethereum as the anchor registry)
// For now hard-coded to the Ethereum setup
func getConfiguredRegistry() (AnchorRegistry, error) {
	return &EthereumAnchorRegistry{}, nil
}

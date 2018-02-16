package anchor

import (
	"log"
)

type AnchorRegistry interface {
	RegisterAnchor(*Anchor) (Anchor, error)
	RegisterAsAnchor(anchorID string, rootHash string) (Anchor, error)
}

// Register the given Anchor with the configured public registry.
func RegisterAnchor(anchor *Anchor) (Anchor, error) {
	registry, _ := getConfiguredRegistry()
	ret, err := registry.RegisterAnchor(anchor)
	if err != nil {
		log.Fatalf("Failed to register the anchor [id:%v, hash:%v, schemaVersion:%v]: %v", anchor.anchorID, anchor.rootHash, anchor.schemaVersion, err)
	}
	return ret, err
}

// Register the given anchorID and root has as an anchor
func RegisterAsAnchor(anchorID string, rootHash string) (Anchor, error) {
	registry, _ := getConfiguredRegistry()
	ret, _ := registry.RegisterAsAnchor(anchorID, rootHash)

	return ret, nil
}

func getConfiguredRegistry() (AnchorRegistry, error) {
	return &EthereumAnchorRegistry{}, nil
}

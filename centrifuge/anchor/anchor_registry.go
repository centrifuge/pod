package anchor

import (
	"crypto/rand"
	"log"
)

type AnchorRegistry interface {
	RegisterAnchor(*Anchor) (Anchor, error)
	RegisterAsAnchor(anchorID string, rootHash string) (Anchor, error)
}

func createRandomByte32() (out [32]byte) {
	r := make([]byte, 32)
	_, err := rand.Read(r)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		panic(err)
	}
	copy(out[:], r[:32])
	return
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

package anchor

import (
	"crypto/rand"
	//"github.com/CentrifugeInc/go-centrifuge/centrifuge/ethereum"
)

//Supported anchor schema version as stored on public registry
const ANCHOR_SCHEMA_VERSION uint = 1

type Anchor struct {
	anchorID      string
	rootHash      string
	schemaVersion uint
}

type UsableAnchorRegistry interface {
	registerAnchor(*Anchor) (Anchor, error)
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

type EthereumAnchorRegistry struct {

}

func (ethRegistry EthereumAnchorRegistry) registerAnchor(anchor *Anchor) (Anchor, error) {
	return Anchor{anchor.anchorID, anchor.rootHash, anchor.schemaVersion}, nil
}

// Register the given Anchor with the configured public registry.
func RegisterAnchor(anchor *Anchor) (Anchor, error) {
	returnAnchor := Anchor{
		anchorID:      anchor.anchorID,
		rootHash:      anchor.rootHash,
		schemaVersion: anchor.schemaVersion,
	}
	return returnAnchor, nil
}

func getConfiguredRegistry() (UsableAnchorRegistry, error) {
	return EthereumAnchorRegistry{}, nil
}

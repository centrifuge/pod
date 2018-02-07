package anchor

import "crypto/rand"

//Supported anchor schema version as stored on public registry
const ANCHOR_SCHEMA_VERSION uint = 1

type Anchor struct {
	anchorID      string
	rootHash      string
	schemaVersion uint
}

type RegisteredAnchor struct {
	EthereumTransaction string
	Anchor              Anchor
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

func (anchor *Anchor) registerAnchor() (*Anchor, error) {
	a := &Anchor{
		anchorID:      "string(createRandomByte32()[:])",
		rootHash:      "string(createRandomByte32()[:])",
		schemaVersion: ANCHOR_SCHEMA_VERSION,
	}
	return a, nil
}

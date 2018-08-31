package registry

type Anchor struct {
	AnchorID      [32]byte
	RootHash      [32]byte
	SchemaVersion uint
}

type WatchAnchor struct {
	Anchor *Anchor
	Error  error
}

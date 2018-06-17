package anchor

type Anchor struct {
	AnchorID      string
	RootHash      string
	SchemaVersion uint
}

type WatchAnchor struct {
	Anchor *Anchor
	Error error
}

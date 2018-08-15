package anchor

type Anchor struct {
	AnchorID      [32]byte
	RootHash      [32]byte
	SchemaVersion uint
}

func (*Anchor) MarshalBinary() (data []byte, err error) {
	panic("implement me")
}

type WatchAnchor struct {
	Anchor *Anchor
	Error error
}

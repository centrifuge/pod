package anchor

import (
	"testing"
	
)

func TestRegisterAnchor(t *testing.T) {
	storeThisAnchor := &Anchor{anchorID: "asdde", rootHash: "dafrwfrw", schemaVersion: 1}
	anc, _ := RegisterAnchor(storeThisAnchor)
	if anc.anchorID == "" {
		t.Fatal("No Anchor ID set")
	}
}

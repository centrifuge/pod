package anchor

import (
	"testing"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
)

func TestRegisterAnchor(t *testing.T) {
	storeThisAnchor := &Anchor{anchorID: tools.RandomString32(), rootHash: tools.RandomString32(), schemaVersion: 1}
	anc, _ := RegisterAnchor(storeThisAnchor)
	if anc.anchorID == "" {
		t.Fatal("No Anchor ID set")
	}
}

func TestRegisterAsAnchor(t *testing.T) {

	anc, _ := RegisterAsAnchor(tools.RandomString32(), tools.RandomString32())
	if anc.anchorID == "" {
		t.Fatal("No Anchor ID set")
	}
}

//func TestRegisterAnchorIntegration(t *testing.T){
//	anc, _ := RegisterAsAnchor("asdde", "dafrwfrw")
//	if anc.anchorID == "" {
//		t.Fatal("No Anchor ID set")
//	}
//}

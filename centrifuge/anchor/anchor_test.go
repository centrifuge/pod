package anchor

import (
	"testing"
	
)

func TestRegisterAnchor(t *testing.T) {
	storeThisAnchor := &Anchor{anchorID: "asddeasddeasddeasddeasddeasdde32", rootHash: "dafrwfrwdafrwfrwdafrwfrwdafrwf32", schemaVersion: 1}
	anc, _ := RegisterAnchor(storeThisAnchor)
	if anc.anchorID == "" {
		t.Fatal("No Anchor ID set")
	}
}

func TestRegisterAsAnchor(t *testing.T) {

	anc, _ := RegisterAsAnchor("asddeasddeasddeasddeasddeasdde32", "dafrwfrwdafrwfrwdafrwfrwdafrwf32")
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

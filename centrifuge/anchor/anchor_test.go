package anchor

import (
	"testing"
)

func TestRegisterAnchor(t *testing.T) {
	a := Anchor{}
	anchor, _ := a.registerAnchor()

	if anchor.anchorID == "" {
		t.Fatal("No Anchor ID set")
	}
}

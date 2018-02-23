package anchor

import (
	"testing"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
	"github.com/stretchr/testify/assert"
)

func TestRegisterAsAnchor(t *testing.T) {

	confirmations := make(chan *Anchor)
	id := tools.RandomString32()
	rootHash := tools.RandomString32()
	err := RegisterAsAnchor(id, rootHash, confirmations)
	if err != nil {
		t.Fatalf("Error registering Ancho %v", err)
	}
	registeredAnchor := <-confirmations
	assert.Equal(t, registeredAnchor.anchorID, id, "Resulting anchor should have the same ID as the input")
	assert.Equal(t, registeredAnchor.rootHash, rootHash, "Resulting anchor should have the same root hash as the input")
}
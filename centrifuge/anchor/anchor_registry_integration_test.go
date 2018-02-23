package anchor_test

import (
	"testing"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/anchor"
	"github.com/stretchr/testify/assert"
	"flag"
)

var (
	ethereumTest = flag.Bool("ethereum", false, "run Ethereum integration tests")
)

func TestRegisterAsAnchor_Integration(t *testing.T) {
	if !*ethereumTest{
		return
	}
	confirmations := make(chan *anchor.Anchor)
	id := tools.RandomString32()
	rootHash := tools.RandomString32()
	err := anchor.RegisterAsAnchor(id, rootHash, confirmations)
	if err != nil {
		t.Fatalf("Error registering Anchor %v", err)
	}

	registeredAnchor := <-confirmations
	assert.Equal(t, registeredAnchor.AnchorID, id, "Resulting anchor should have the same ID as the input")
	assert.Equal(t, registeredAnchor.RootHash, rootHash, "Resulting anchor should have the same root hash as the input")
}

// +build ethereum

package anchor_test

import (
	"os"
	"testing"

	"github.com/CentrifugeInc/go-centrifuge/centrifuge/anchor"
	cc "github.com/CentrifugeInc/go-centrifuge/centrifuge/context/testing"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	cc.TestFunctionalEthereumBootstrap()
	result := m.Run()
	cc.TestIntegrationTearDown()
	os.Exit(result)
}

func TestRegisterAsAnchor_Integration(t *testing.T) {
	id := tools.RandomByte32()
	rootHash := tools.RandomByte32()
	confirmations, err := anchor.RegisterAsAnchor(id, rootHash)
	if err != nil {
		t.Fatalf("Error registering Anchor %v", err)
	}

	watchRegisteredAnchor := <-confirmations
	assert.Nil(t, watchRegisteredAnchor.Error, "No error thrown by context")
	assert.Equal(t, watchRegisteredAnchor.Anchor.AnchorID, id, "Resulting anchor should have the same ID as the input")
	assert.Equal(t, watchRegisteredAnchor.Anchor.RootHash, rootHash, "Resulting anchor should have the same root hash as the input")
}

func TestRegisterAsAnchor_Integration_Concurrent(t *testing.T) {
	for ix := 0; ix < 5; ix++ {
		id := tools.RandomByte32()
		rootHash := tools.RandomByte32()

		confirmations, err := anchor.RegisterAsAnchor(id, rootHash)
		watchSingleAnchor := <- confirmations
		assert.Nil(t, err, "should not error out upon anchor registration")
		assert.Nil(t, watchSingleAnchor.Error, "No error thrown by context")
		assert.Equal(t, id, watchSingleAnchor.Anchor.AnchorID, "Should have the ID that was passed into create function [%v]", watchSingleAnchor.Anchor.AnchorID)
		assert.Equal(t, rootHash, watchSingleAnchor.Anchor.RootHash, "Should have the RootHash that was passed into create function [%v]", watchSingleAnchor.Anchor.RootHash)

	}
}

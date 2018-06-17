// +build ethereum

package anchor_test

import (
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/anchor"
	cc "github.com/CentrifugeInc/go-centrifuge/centrifuge/context"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	cc.TestBootstrap()
	result := m.Run()
	cc.TestTearDown()
	os.Exit(result)
}

func TestRegisterAsAnchor_Integration(t *testing.T) {
	confirmations := make(chan *anchor.WatchAnchor, 1)
	id := tools.RandomString32()
	rootHash := tools.RandomString32()
	err := anchor.RegisterAsAnchor(id, rootHash, confirmations)
	if err != nil {
		t.Fatalf("Error registering Anchor %v", err)
	}

	watchRegisteredAnchor := <-confirmations
	assert.Nil(t, watchRegisteredAnchor.Error, "No error thrown by context")
	assert.Equal(t, watchRegisteredAnchor.Anchor.AnchorID, id, "Resulting anchor should have the same ID as the input")
	assert.Equal(t, watchRegisteredAnchor.Anchor.RootHash, rootHash, "Resulting anchor should have the same root hash as the input")
}

func TestRegisterAsAnchor_Integration_Concurrent(t *testing.T) {
	var submittedIds [5]string
	var submittedRhs [5]string

	howMany := cap(submittedIds)
	confirmations := make(chan *anchor.WatchAnchor, howMany)

	for ix := 0; ix < howMany; ix++ {
		id := tools.RandomString32()
		submittedIds[ix] = id

		rootHash := tools.RandomString32()
		submittedRhs[ix] = rootHash

		err := anchor.RegisterAsAnchor(id, rootHash, confirmations)
		assert.Nil(t, err, "should not error out upon anchor registration")
	}

	for ix := 0; ix < howMany; ix++ {
		watchSingleAnchor := <-confirmations
		assert.Nil(t, watchSingleAnchor.Error, "No error thrown by context")
		assert.Contains(t, submittedIds, watchSingleAnchor.Anchor.AnchorID, "Should have the ID that was passed into create function [%v]", watchSingleAnchor.Anchor.AnchorID)
		assert.Contains(t, submittedRhs, watchSingleAnchor.Anchor.RootHash, "Should have the RootHash that was passed into create function [%v]", watchSingleAnchor.Anchor.RootHash)
	}
}

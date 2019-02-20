// +build integration

package anchors_test

import (
	"os"
	"testing"

	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	cc "github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testingbootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
)

var (
	identityService identity.Service
	anchorRepo      anchors.AnchorRepository
	cfg             config.Configuration
)

func TestMain(m *testing.M) {
	ctx := cc.TestFunctionalEthereumBootstrap()
	anchorRepo = ctx[anchors.BootstrappedAnchorRepo].(anchors.AnchorRepository)
	identityService = ctx[identity.BootstrappedIDService].(identity.Service)
	cfg = ctx[bootstrap.BootstrappedConfig].(config.Configuration)
	result := m.Run()
	cc.TestFunctionalEthereumTearDown()
	os.Exit(result)
}


func TestCommitAnchor_Integration(t *testing.T) {
	anchorID, _ := hexutil.Decode("0x154cc26833dec2f4ad7ead9d65f9ec968a1aa5efbf6fe762f8f2a67d18a2d9b1")
	documentRoot, _ := hexutil.Decode("0x65a35574f70281ae4d1f6c9f3adccd5378743f858c67a802a49a08ce185bc975")


	anchorIDTyped, _ := anchors.ToAnchorID(anchorID)
	docRootTyped, _ := anchors.ToDocumentRoot(documentRoot)
	commitAnchor(t, anchorID, documentRoot,[][anchors.DocumentProofLength]byte{utils.RandomByte32()})
	gotDocRoot, err := anchorRepo.GetDocumentRootOf(anchorIDTyped)
	assert.Nil(t, err)
	assert.Equal(t, docRootTyped, gotDocRoot)
}

func commitAnchor(t *testing.T, anchorID, documentRoot []byte, documentProofs [][32]byte) {
	anchorIDTyped, _ := anchors.ToAnchorID(anchorID)
	docRootTyped, _ := anchors.ToDocumentRoot(documentRoot)

	ctx := testingconfig.CreateAccountContext(t, cfg)
	confirmations, err := anchorRepo.CommitAnchor(ctx, anchorIDTyped, docRootTyped, documentProofs)
	if err != nil {
		t.Fatalf("Error commit Anchor %v", err)
	}

	watchCommittedAnchor := <-confirmations
	assert.Nil(t, watchCommittedAnchor.Error, "No error should be thrown by context")
	assert.Equal(t, watchCommittedAnchor.CommitData.AnchorID, anchorIDTyped, "Resulting anchor should have the same ID as the input")
	assert.Equal(t, watchCommittedAnchor.CommitData.DocumentRoot, docRootTyped, "Resulting anchor should have the same document hash as the input")
}

/*
func TestCommitAnchor_Integration_Concurrent(t *testing.T) {
	var commitDataList [5]*anchors.CommitData
	var confirmationList [5]<-chan *anchors.WatchCommit
	testPrivateKey, _ := hexutil.Decode("0x17e063fa17dd8274b09c14b253697d9a20afff74ace3c04fdb1b9c814ce0ada5")
	centrifugeId := utils.RandomSlice(identity.CentIDLength)


	for ix := 0; ix < 5; ix++ {
		currentAnchorId := utils.RandomByte32()
		currentDocumentRoot := utils.RandomByte32()
		centIdFixed, _ := identity.ToCentID(centrifugeId)
		messageToSign := anchors.GenerateCommitHash(currentAnchorId, centIdFixed, currentDocumentRoot)
		documentProofs := [][anchors.DocumentProofLength]byte{utils.RandomByte32()}
		h, err := ethereum.GetClient().GetEthClient().HeaderByNumber(context.Background(), nil)
		assert.Nil(t, err, " error must be nil")
		commitDataList[ix] = anchors.NewCommitData(h.Number.Uint64(), currentAnchorId, currentDocumentRoot, documentProofs)
		cfg.Set("identityId", centIdFixed.String())
		ctx := testingconfig.CreateAccountContext(t, cfg)
		confirmationList[ix], err = anchorRepo.CommitAnchor(ctx, currentAnchorId, currentDocumentRoot, documentProofs)
		if err != nil {
			t.Fatalf("Error commit Anchor %v", err)
		}
		assert.Nil(t, err, "should not error out upon anchor registration")
	}

	for ix := 0; ix < 5; ix++ {
		watchSingleAnchor := <-confirmationList[ix]
		assert.Nil(t, watchSingleAnchor.Error, "No error thrown by context")
		assert.Equal(t, commitDataList[ix].AnchorID, watchSingleAnchor.CommitData.AnchorID, "Should have the ID that was passed into create function [%v]", watchSingleAnchor.CommitData.AnchorID)
		assert.Equal(t, commitDataList[ix].DocumentRoot, watchSingleAnchor.CommitData.DocumentRoot, "Should have the document root that was passed into create function [%v]", watchSingleAnchor.CommitData.DocumentRoot)
		anchorID := commitDataList[ix].AnchorID
		docRoot := commitDataList[ix].DocumentRoot
		gotDocRoot, err := anchorRepo.GetDocumentRootOf(anchorID)
		assert.Nil(t, err)
		assert.Equal(t, docRoot, gotDocRoot)
	}
}
*/

// +build integration

package anchors_test

import (
	"os"
	"testing"

	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	cc "github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testingbootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/stretchr/testify/assert"
)

var (
	anchorRepo      anchors.AnchorRepository
	cfg             config.Configuration
)

func TestMain(m *testing.M) {
	ctx := cc.TestFunctionalEthereumBootstrap()
	anchorRepo = ctx[anchors.BootstrappedAnchorRepo].(anchors.AnchorRepository)
	cfg = ctx[bootstrap.BootstrappedConfig].(config.Configuration)
	result := m.Run()
	cc.TestFunctionalEthereumTearDown()
	os.Exit(result)
}


func TestCommitAnchor_Integration(t *testing.T) {
	anchorID  := utils.RandomSlice(32)
	documentRoot := utils.RandomSlice(32)

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
	err := anchorRepo.CommitAnchor(ctx, anchorIDTyped, docRootTyped, documentProofs)

	assert.Nil(t,err)

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

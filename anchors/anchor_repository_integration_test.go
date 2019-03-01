// +build integration

package anchors_test

import (
	"context"
	"crypto/sha256"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	cc "github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testingbootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/stretchr/testify/assert"
)

var (
	anchorRepo anchors.AnchorRepository
	cfg        config.Configuration
)

func TestMain(m *testing.M) {
	ctx := cc.TestFunctionalEthereumBootstrap()
	anchorRepo = ctx[anchors.BootstrappedAnchorRepo].(anchors.AnchorRepository)
	cfg = ctx[bootstrap.BootstrappedConfig].(config.Configuration)
	result := m.Run()
	cc.TestFunctionalEthereumTearDown()
	os.Exit(result)
}

func TestPreCommitAnchor_Integration(t *testing.T) {
	anchorID := utils.RandomSlice(32)
	signingRoot := utils.RandomSlice(32)

	anchorIDTyped, _ := anchors.ToAnchorID(anchorID)
	preCommitAnchor(t, anchorID, signingRoot)
	valid := anchorRepo.HasValidPreCommit(anchorIDTyped)
	assert.True(t, valid)
}

func TestPreCommit_CommitAnchor_Integration(t *testing.T) {
	anchorIDPreImage := utils.RandomSlice(32)
	h := sha256.New()
	_, err := h.Write(anchorIDPreImage)
	assert.NoError(t, err)
	var anchorID []byte
	anchorID = h.Sum(anchorID)
	proofStr := "0xec4f7c791a848bfae2d45e815090e96cc2a7f9a9b851074162812d687883c2b1"
	signingRootStr := "0x5b5e0bf237698cc39fdc052063aa8f445d28ef295c089754009f5f91dc66a686"
	documentRootStr := "0x91410b6658b692d1cb0b68aceabfe2b0020d4c6bd0bca9dbecae0c43c326ab34"

	signingRoot, err := hexutil.Decode(signingRootStr)
	assert.NoError(t, err)

	documentRoot, err := hexutil.Decode(documentRootStr)
	assert.NoError(t, err)

	proof, err := hexutil.Decode(proofStr)
	assert.NoError(t, err)

	var proofB [32]byte
	copy(proofB[:], proof)

	anchorIDTyped, _ := anchors.ToAnchorID(anchorID)
	preCommitAnchor(t, anchorID, signingRoot)
	valid := anchorRepo.HasValidPreCommit(anchorIDTyped)
	assert.True(t, valid)

	docRootTyped, _ := anchors.ToDocumentRoot(documentRoot)
	commitAnchor(t, anchorIDPreImage, documentRoot, [][anchors.DocumentProofLength]byte{proofB})
	gotDocRoot, err := anchorRepo.GetDocumentRootOf(anchorIDTyped)
	assert.Nil(t, err)
	assert.Equal(t, docRootTyped, gotDocRoot)
}

func TestCommitAnchor_Integration(t *testing.T) {
	anchorIDPreImage := utils.RandomSlice(32)
	h := sha256.New()
	_, err := h.Write(anchorIDPreImage)
	assert.NoError(t, err)
	var anchorID []byte
	anchorID = h.Sum(anchorID)
	documentRoot := utils.RandomSlice(32)

	anchorIDTyped, _ := anchors.ToAnchorID(anchorID)
	docRootTyped, _ := anchors.ToDocumentRoot(documentRoot)
	commitAnchor(t, anchorIDPreImage, documentRoot, [][anchors.DocumentProofLength]byte{utils.RandomByte32()})
	gotDocRoot, err := anchorRepo.GetDocumentRootOf(anchorIDTyped)
	assert.Nil(t, err)
	assert.Equal(t, docRootTyped, gotDocRoot)
}

func commitAnchor(t *testing.T, anchorID, documentRoot []byte, documentProofs [][32]byte) {
	anchorIDTyped, _ := anchors.ToAnchorID(anchorID)
	docRootTyped, _ := anchors.ToDocumentRoot(documentRoot)

	ctx := testingconfig.CreateAccountContext(t, cfg)
	done, err := anchorRepo.CommitAnchor(ctx, anchorIDTyped, docRootTyped, documentProofs)

	isDone := <-done

	assert.True(t, isDone, "isDone should be true")

	assert.Nil(t, err)
}

func preCommitAnchor(t *testing.T, anchorID, documentRoot []byte) {
	anchorIDTyped, _ := anchors.ToAnchorID(anchorID)
	docRootTyped, _ := anchors.ToDocumentRoot(documentRoot)

	ctx := testingconfig.CreateAccountContext(t, cfg)
	done, err := anchorRepo.PreCommitAnchor(ctx, anchorIDTyped, docRootTyped)

	isDone := <-done

	assert.True(t, isDone, "isDone should be true")
	assert.Nil(t, err)
}

func TestCommitAnchor_Integration_Concurrent(t *testing.T) {
	var commitDataList [5]*anchors.CommitData
	var doneList [5]chan bool

	for ix := 0; ix < 5; ix++ {
		currentAnchorId := utils.RandomByte32()
		currentDocumentRoot := utils.RandomByte32()
		documentProofs := [][anchors.DocumentProofLength]byte{utils.RandomByte32()}
		h, err := ethereum.GetClient().GetEthClient().HeaderByNumber(context.Background(), nil)
		assert.Nil(t, err, " error must be nil")
		commitDataList[ix] = anchors.NewCommitData(h.Number.Uint64(), currentAnchorId, currentDocumentRoot, documentProofs)
		ctx := testingconfig.CreateAccountContext(t, cfg)
		doneList[ix], err = anchorRepo.CommitAnchor(ctx, currentAnchorId, currentDocumentRoot, documentProofs)
		if err != nil {
			t.Fatalf("Error commit Anchor %v", err)
		}
		assert.Nil(t, err, "should not error out upon anchor registration")
	}

	for ix := 0; ix < 5; ix++ {
		isDone := <-doneList[ix]
		assert.True(t, isDone)
		anchorID := commitDataList[ix].AnchorID
		docRoot := commitDataList[ix].DocumentRoot
		gotDocRoot, err := anchorRepo.GetDocumentRootOf(anchorID)
		assert.Nil(t, err)
		assert.Equal(t, docRoot, gotDocRoot)
	}
}

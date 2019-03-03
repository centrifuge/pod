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
	proofStr := []string{"0x73bb733279cd232d72732afad693f80510e71262738b1205a061a1c34497e49c", "0x408b2caa80ace6ac3a37be957235011c0053f0a561f5a8dcf66d223bfffccecb"}
	signingRootStr := "0x6ae9e6cc91cded82896d2439942fd42412b5a2ff5fd45bbed0f5a20de0b962c2"
	documentRootStr := "0xf5c8f866f4acf2e2e74a803f86cd2a7ac9285721259b172ef121417e886ca22a"

	signingRoot, err := hexutil.Decode(signingRootStr)
	assert.NoError(t, err)

	documentRoot, err := hexutil.Decode(documentRootStr)
	assert.NoError(t, err)

	proof1, err := hexutil.Decode(proofStr[0])
	assert.NoError(t, err)

	proof2, err := hexutil.Decode(proofStr[1])
	assert.NoError(t, err)

	var proofB1 [32]byte
	copy(proofB1[:], proof1)
	var proofB2 [32]byte
	copy(proofB2[:], proof2)

	anchorIDTyped, _ := anchors.ToAnchorID(anchorID)
	preCommitAnchor(t, anchorID, signingRoot)
	valid := anchorRepo.HasValidPreCommit(anchorIDTyped)
	assert.True(t, valid)

	docRootTyped, _ := anchors.ToDocumentRoot(documentRoot)
	commitAnchor(t, anchorIDPreImage, documentRoot, [][anchors.DocumentProofLength]byte{proofB1, proofB2})
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
		anchorIDPreImage := utils.RandomSlice(32)
		anchorIDPreImageID, err := anchors.ToAnchorID(anchorIDPreImage)
		assert.NoError(t, err)
		h := sha256.New()
		_, err = h.Write(anchorIDPreImage)
		assert.NoError(t, err)
		var cAnchorId []byte
		cAnchorId = h.Sum(cAnchorId)
		currentAnchorId, err := anchors.ToAnchorID(cAnchorId)
		assert.NoError(t, err)
		currentDocumentRoot := utils.RandomByte32()
		documentProofs := [][anchors.DocumentProofLength]byte{utils.RandomByte32()}
		hd, err := ethereum.GetClient().GetEthClient().HeaderByNumber(context.Background(), nil)
		assert.Nil(t, err, " error must be nil")
		commitDataList[ix] = anchors.NewCommitData(hd.Number.Uint64(), currentAnchorId, currentDocumentRoot, documentProofs)
		ctx := testingconfig.CreateAccountContext(t, cfg)
		doneList[ix], err = anchorRepo.CommitAnchor(ctx, anchorIDPreImageID, currentDocumentRoot, documentProofs)
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

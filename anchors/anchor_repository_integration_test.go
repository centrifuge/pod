// +build integration

package anchors_test

import (
	"context"
	"os"
	"testing"

	"github.com/centrifuge/go-centrifuge/anchors"
	cc "github.com/centrifuge/go-centrifuge/context/testingbootstrap"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/keytools/secp256k1"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
)

var (
	identityService identity.Service
	anchorRepo      anchors.AnchorRepository
)

func TestMain(m *testing.M) {
	ctx := cc.TestFunctionalEthereumBootstrap()
	anchorRepo = ctx[anchors.BootstrappedAnchorRepo].(anchors.AnchorRepository)
	identityService = ctx[identity.BootstrappedIDService].(identity.Service)
	result := m.Run()
	cc.TestFunctionalEthereumTearDown()
	os.Exit(result)
}

func createIdentityWithKeys(t *testing.T, centrifugeId []byte) []byte {
	centIdTyped, _ := identity.ToCentID(centrifugeId)
	id, confirmations, err := identityService.CreateIdentity(centIdTyped)
	assert.Nil(t, err, "should not error out when creating identity")
	watchRegisteredIdentity := <-confirmations
	assert.Nil(t, watchRegisteredIdentity.Error, "No error thrown by context")
	// LookupIdentityForId
	id, err = identityService.LookupIdentityForID(centIdTyped)
	assert.Nil(t, err, "should not error out when resolving identity")

	pubKey, _ := hexutil.Decode("0xc8dd3d66e112fae5c88fe6a677be24013e53c33e")

	confirmations, err = id.AddKeyToIdentity(context.Background(), identity.KeyPurposeEthMsgAuth, pubKey)
	assert.Nil(t, err, "should not error out when adding keys")
	assert.NotNil(t, confirmations, "confirmations channel should not be nil")
	watchRegisteredIdentityKey := <-confirmations
	assert.Nil(t, watchRegisteredIdentityKey.Error, "No error thrown by context")

	return centrifugeId
}

func TestCommitAnchor_Integration(t *testing.T) {
	anchorID, _ := hexutil.Decode("0x154cc26833dec2f4ad7ead9d65f9ec968a1aa5efbf6fe762f8f2a67d18a2d9b1")
	documentRoot, _ := hexutil.Decode("0x65a35574f70281ae4d1f6c9f3adccd5378743f858c67a802a49a08ce185bc975")
	centrifugeId := utils.RandomSlice(identity.CentIDLength)
	createIdentityWithKeys(t, centrifugeId)
	testPrivateKey, _ := hexutil.Decode("0x17e063fa17dd8274b09c14b253697d9a20afff74ace3c04fdb1b9c814ce0ada5")
	anchorIDTyped, _ := anchors.ToAnchorID(anchorID)
	centIdTyped, _ := identity.ToCentID(centrifugeId)
	docRootTyped, _ := anchors.ToDocumentRoot(documentRoot)
	messageToSign := anchors.GenerateCommitHash(anchorIDTyped, centIdTyped, docRootTyped)
	signature, _ := secp256k1.SignEthereum(messageToSign, testPrivateKey)
	commitAnchor(t, anchorID, centrifugeId, documentRoot, signature, [][anchors.DocumentProofLength]byte{utils.RandomByte32()})
	gotDocRoot, err := anchorRepo.GetDocumentRootOf(anchorIDTyped)
	assert.Nil(t, err)
	assert.Equal(t, docRootTyped, gotDocRoot)
}

func commitAnchor(t *testing.T, anchorID, centrifugeId, documentRoot, signature []byte, documentProofs [][32]byte) {
	anchorIDTyped, _ := anchors.ToAnchorID(anchorID)
	docRootTyped, _ := anchors.ToDocumentRoot(documentRoot)
	centIdFixed, _ := identity.ToCentID(centrifugeId)

	confirmations, err := anchorRepo.CommitAnchor(anchorIDTyped, docRootTyped, centIdFixed, documentProofs, signature)
	if err != nil {
		t.Fatalf("Error commit Anchor %v", err)
	}

	watchCommittedAnchor := <-confirmations
	assert.Nil(t, watchCommittedAnchor.Error, "No error should be thrown by context")
	assert.Equal(t, watchCommittedAnchor.CommitData.AnchorID, anchorIDTyped, "Resulting anchor should have the same ID as the input")
	assert.Equal(t, watchCommittedAnchor.CommitData.DocumentRoot, docRootTyped, "Resulting anchor should have the same document hash as the input")
}

func TestCommitAnchor_Integration_Concurrent(t *testing.T) {
	var commitDataList [5]*anchors.CommitData
	var confirmationList [5]<-chan *anchors.WatchCommit
	testPrivateKey, _ := hexutil.Decode("0x17e063fa17dd8274b09c14b253697d9a20afff74ace3c04fdb1b9c814ce0ada5")
	centrifugeId := utils.RandomSlice(identity.CentIDLength)
	createIdentityWithKeys(t, centrifugeId)

	for ix := 0; ix < 5; ix++ {
		currentAnchorId := utils.RandomByte32()
		currentDocumentRoot := utils.RandomByte32()
		centIdFixed, _ := identity.ToCentID(centrifugeId)
		messageToSign := anchors.GenerateCommitHash(currentAnchorId, centIdFixed, currentDocumentRoot)
		signature, _ := secp256k1.SignEthereum(messageToSign, testPrivateKey)
		documentProofs := [][anchors.DocumentProofLength]byte{utils.RandomByte32()}
		h, err := ethereum.GetClient().GetEthClient().HeaderByNumber(context.Background(), nil)
		assert.Nil(t, err, " error must be nil")
		commitDataList[ix] = anchors.NewCommitData(h.Number.Uint64(), currentAnchorId, currentDocumentRoot, centIdFixed, documentProofs, signature)
		confirmationList[ix], err = anchorRepo.CommitAnchor(currentAnchorId, currentDocumentRoot, centIdFixed, documentProofs, signature)
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

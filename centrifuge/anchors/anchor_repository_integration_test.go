// +build integration

package anchors_test

import (
	"os"
	"testing"

	"github.com/CentrifugeInc/go-centrifuge/centrifuge/anchors"
	cc "github.com/CentrifugeInc/go-centrifuge/centrifuge/context/testingbootstrap"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/identity"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/keytools/secp256k1"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
)

var identityService identity.Service

// Add Key
var testAddress string
var testPrivateKey string

func TestMain(m *testing.M) {

	identityService = &identity.EthereumIdentityService{}
	cc.TestFunctionalEthereumBootstrap()
	result := m.Run()
	cc.TestFunctionalEthereumTearDown()
	os.Exit(result)
}

func createIdentityWithKeys(t *testing.T, centrifugeId []byte) []byte {

	centIdTyped, _ := identity.NewCentID(centrifugeId)
	id, confirmations, err := identityService.CreateIdentity(centIdTyped)
	assert.Nil(t, err, "should not error out when creating identity")

	watchRegisteredIdentity := <-confirmations
	assert.Nil(t, watchRegisteredIdentity.Error, "No error thrown by context")

	// LookupIdentityForId
	id, err = identityService.LookupIdentityForID(centIdTyped)
	assert.Nil(t, err, "should not error out when resolving identity")

	testAddress = "0xc8dd3d66e112fae5c88fe6a677be24013e53c33e"
	testPrivateKey = "0x17e063fa17dd8274b09c14b253697d9a20afff74ace3c04fdb1b9c814ce0ada5"
	pubKey, _ := hexutil.Decode(testAddress)

	confirmations, err = id.AddKeyToIdentity(identity.KeyPurposeEthMsgAuth, pubKey)
	assert.Nil(t, err, "should not error out when adding keys")
	assert.NotNil(t, confirmations, "confirmations channel should not be nil")
	watchRegisteredIdentityKey := <-confirmations
	assert.Nil(t, watchRegisteredIdentityKey.Error, "No error thrown by context")

	return centrifugeId

}

func TestCommitAnchor_Integration(t *testing.T) {
	anchorID, _ := hexutil.Decode("0x154cc26833dec2f4ad7ead9d65f9ec968a1aa5efbf6fe762f8f2a67d18a2d9b1")
	documentRoot, _ := hexutil.Decode("0x65a35574f70281ae4d1f6c9f3adccd5378743f858c67a802a49a08ce185bc975")
	centrifugeId := tools.RandomSlice(identity.CentIDByteLength)

	createIdentityWithKeys(t, centrifugeId)

	testPrivateKey, _ := hexutil.Decode("0x17e063fa17dd8274b09c14b253697d9a20afff74ace3c04fdb1b9c814ce0ada5")

	anchorIDTyped, _ := anchors.NewAnchorID(anchorID)
	centIdTyped, _ := identity.NewCentID(centrifugeId)
	docRootTyped, _ := anchors.NewDocRoot(documentRoot)

	messageToSign := anchors.GenerateCommitHash(anchorIDTyped, centIdTyped, docRootTyped)

	signature, _ := secp256k1.SignEthereum(messageToSign, testPrivateKey)

	commitAnchor(t, anchorID, centrifugeId, documentRoot, signature, [][anchors.DocumentProofLength]byte{tools.RandomByte32()})

}

func commitAnchor(t *testing.T, anchorID, centrifugeId, documentRoot, signature []byte, documentProofs [][32]byte) {
	anchorIDTyped, _ := anchors.NewAnchorID(anchorID)
	docRootTyped, _ := anchors.NewDocRoot(documentRoot)
	centIdFixed, _ := identity.NewCentID(centrifugeId)

	confirmations, err := anchors.CommitAnchor(anchorIDTyped, docRootTyped, centIdFixed, documentProofs, signature)

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
	var err error
	testPrivateKey, _ := hexutil.Decode("0x17e063fa17dd8274b09c14b253697d9a20afff74ace3c04fdb1b9c814ce0ada5")

	centrifugeId := tools.RandomSlice(identity.CentIDByteLength)

	createIdentityWithKeys(t, centrifugeId)

	for ix := 0; ix < 5; ix++ {
		currentAnchorId := tools.RandomByte32()
		currentDocumentRoot := tools.RandomByte32()
		centIdFixed, _ := identity.NewCentID(centrifugeId)
		messageToSign := anchors.GenerateCommitHash(currentAnchorId, centIdFixed, currentDocumentRoot)
		signature, _ := secp256k1.SignEthereum(messageToSign, testPrivateKey)
		documentProofs := [][anchors.DocumentProofLength]byte{tools.RandomByte32()}

		commitDataList[ix] = anchors.NewCommitData(currentAnchorId, currentDocumentRoot, centIdFixed, documentProofs, signature)

		confirmationList[ix], err = anchors.CommitAnchor(currentAnchorId, currentDocumentRoot, centIdFixed, documentProofs, signature)

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
	}
}

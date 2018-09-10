

package anchoring

import (
	"math/big"
	"os"
	"testing"

	"github.com/CentrifugeInc/go-centrifuge/centrifuge/anchor/repository"
	cc "github.com/CentrifugeInc/go-centrifuge/centrifuge/context/testing"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/identity"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/keytools/secp256k1"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
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

	id, confirmations, err := identityService.CreateIdentity(centrifugeId)
	assert.Nil(t, err, "should not error out when creating identity")

	watchRegisteredIdentity := <-confirmations
	assert.Nil(t, watchRegisteredIdentity.Error, "No error thrown by context")

	// LookupIdentityForId
	id, err = identityService.LookupIdentityForID(centrifugeId)
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

func generateCommitHash(anchorIdByte []byte, centrifugeIdByte []byte, documentRootByte []byte) []byte {

	message := append(anchorIdByte, documentRootByte...)

	message = append(message, centrifugeIdByte...)

	messageToSign := crypto.Keccak256(message)
	return messageToSign
}

func TestCommitAnchor_Integration(t *testing.T) {

	documentProof := tools.RandomByte32()

	anchorId := utils.HexToByteArray("0x154cc26833dec2f4ad7ead9d65f9ec968a1aa5efbf6fe762f8f2a67d18a2d9b1")
	documentRoot := utils.HexToByteArray("0x65a35574f70281ae4d1f6c9f3adccd5378743f858c67a802a49a08ce185bc975")
	centrifugeId := utils.HexToByteArray("0x1851943e76d2")

	//centrifugeId is a constant value
	// the identity will only be created once.
	// re-running the test against the same node will not create a new identity
	createIdentityWithKeys(t, centrifugeId)

	correctCommitSignature := "0xb4051d6d03c3bf39f4ec4ba949a91a358b0cacb4804b82ed2ba978d338f5e747770c00b63c8e50c1a7aa5ba629870b54c2068a56f8b43460aa47891c6635d36d01"

	testPrivateKey := "0x17e063fa17dd8274b09c14b253697d9a20afff74ace3c04fdb1b9c814ce0ada5"

	var documentProofs [][repository.AnchorIdLength]byte

	documentProofs = append(documentProofs, documentProof)

	messageToSign := generateCommitHash(anchorId, centrifugeId, documentRoot)

	signature, _ := secp256k1.SignEthereum(messageToSign, utils.HexToByteArray(testPrivateKey))

	assert.Equal(t, correctCommitSignature, utils.ByteArrayToHex(signature), "signature not correct")

	commitAnchor(t, anchorId, centrifugeId, documentRoot, signature, documentProofs)

}

func commitAnchor(t *testing.T, anchorId, centrifugeId, documentRoot, signature []byte, documentProofs [][32]byte) {
	// Big endian encoding is need for ethereum
	var anchorIdBigInt = new(big.Int)
	anchorIdBigInt.SetBytes(anchorId)

	var centrifugeIdBigInt = new(big.Int)
	centrifugeIdBigInt.SetBytes(centrifugeId)

	var documentRoot32Bytes [repository.AnchorIdLength]byte
	copy(documentRoot32Bytes[:], documentRoot[:repository.AnchorIdLength])

	confirmations, err := repository.CommitAnchor(anchorIdBigInt, documentRoot32Bytes, centrifugeIdBigInt, documentProofs, signature)

	if err != nil {
		t.Fatalf("Error commit Anchor %v", err)
	}

	watchCommittedAnchor := <-confirmations
	assert.Nil(t, watchCommittedAnchor.Error, "No error should be thrown by context")
	assert.Equal(t, watchCommittedAnchor.CommitData.AnchorId, anchorIdBigInt, "Resulting anchor should have the same ID as the input")
	assert.Equal(t, watchCommittedAnchor.CommitData.DocumentRoot, documentRoot32Bytes, "Resulting anchor should have the same document hash as the input")
}

/*
func TestCommitAnchor_Integration_Concurrent(t *testing.T) {
	var commitDataList [5] *repository.CommitData
	var confirmationList [5]<-chan *repository.WatchCommit
	var err error
	testPrivateKey := "0x17e063fa17dd8274b09c14b253697d9a20afff74ace3c04fdb1b9c814ce0ada5"

	centrifugeId := tools.RandomSlice(identity.CentIdByteLength)

	createIdentityWithKeys(t,centrifugeId)

	for ix := 0; ix < 5; ix++ {
		currentAnchorId := tools.RandomByte32()
		currentDocumentRoot := tools.RandomByte32()
		documentProof := tools.RandomByte32()

		var documentProofs [][32]byte

		documentProofs = append(documentProofs, documentProof)
		messageToSign := generateCommitHash(currentAnchorId[:],centrifugeId,currentDocumentRoot[:])
		signature := secp256k1.SignEthereum(messageToSign, utils.HexToByteArray(testPrivateKey))

		var anchorIdBigInt  = new(big.Int)
		anchorIdBigInt.SetBytes(currentAnchorId[:])

		var centrifugeIdBigInt = new(big.Int)
		centrifugeIdBigInt.SetBytes(centrifugeId)

		var documentRoot32Bytes [32]byte
		copy(documentRoot32Bytes[:], currentDocumentRoot[:32])

		commitDataList[ix],err = repository.NewCommitData(anchorIdBigInt,documentRoot32Bytes,centrifugeIdBigInt,documentProofs,signature)

		confirmationList[ix], err = repository.CommitAnchor(anchorIdBigInt, documentRoot32Bytes, centrifugeIdBigInt, documentProofs, signature)

		if err != nil {
			t.Fatalf("Error commit Anchor %v", err)
		}
		assert.Nil(t, err, "should not error out upon anchor registration")
	}
	for ix := 0; ix < 5; ix++ {
		watchSingleAnchor := <-confirmationList[ix]
		assert.Nil(t, watchSingleAnchor.Error, "No error thrown by context")
		assert.Equal(t, commitDataList[ix].AnchorId, watchSingleAnchor.CommitData.AnchorId, "Should have the ID that was passed into create function [%v]", watchSingleAnchor.CommitData.AnchorId)
		assert.Equal(t, commitDataList[ix].DocumentRoot, watchSingleAnchor.CommitData.DocumentRoot, "Should have the document root that was passed into create function [%v]", watchSingleAnchor.CommitData.DocumentRoot)
	}
}*/

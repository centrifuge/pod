// +build unit

package anchors

import (
	"testing"

	"github.com/centrifuge/go-centrifuge/crypto/secp256k1"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
)

func TestCorrectCommitSignatureGen(t *testing.T) {
	anchorID, err := hexutil.Decode("0x154cc26833dec2f4ad7ead9d65f9ec968a1aa5efbf6fe762f8f2a67d18a2d9b1")
	assert.NoError(t, err)
	documentRoot, err := hexutil.Decode("0x65a35574f70281ae4d1f6c9f3adccd5378743f858c67a802a49a08ce185bc975")
	assert.NoError(t, err)
	address, err := hexutil.Decode("0x89b0a86583c4444acfd71b463e0d3c55ae1412a5")
	assert.NoError(t, err)
	correctCommitToSign := "0x004a050342f1edda2462288b9e0123a2e1bcc4f978efdc08c07bbf0c3ccc8ddd"
	correctCommitSignature := "0x4a73286521114f528967674bae4ecdc6cc94789255495429a7f58ca3ef0158ae257dd02a0ccb71d817e480d06f60f640ec021ade2ff90fe601bb7a5f4ddc569700"
	testPrivateKey, err := hexutil.Decode("0x17e063fa17dd8274b09c14b253697d9a20afff74ace3c04fdb1b9c814ce0ada5")
	assert.NoError(t, err)
	anchorIDTyped, err := ToAnchorID(anchorID)
	assert.NoError(t, err)
	didTyped, err := identity.NewDIDFromBytes(address)
	assert.NoError(t, err)
	docRootTyped, err := ToDocumentRoot(documentRoot)
	assert.NoError(t, err)
	messageToSign := GenerateCommitHash(anchorIDTyped, didTyped, docRootTyped)
	assert.Equal(t, correctCommitToSign, hexutil.Encode(messageToSign), "messageToSign not calculated correctly")
	signature, err := secp256k1.SignEthereum(messageToSign, testPrivateKey)
	assert.NoError(t, err)
	assert.Equal(t, correctCommitSignature, hexutil.Encode(signature), "signature not correct")
}

func TestGenerateAnchor(t *testing.T) {
	currentAnchorID := utils.RandomByte32()
	currentDocumentRoot := utils.RandomByte32()
	documentProof := utils.RandomByte32()

	var documentRoot32Bytes [32]byte
	copy(documentRoot32Bytes[:], currentDocumentRoot[:32])

	commitData := NewCommitData(currentAnchorID, documentRoot32Bytes, documentProof)
	anchorID, err := ToAnchorID(currentAnchorID[:])
	assert.NoError(t, err)
	docRoot, err := ToDocumentRoot(documentRoot32Bytes[:])
	assert.NoError(t, err)
	assert.Equal(t, commitData.AnchorID, anchorID, "Anchor should have the passed ID")
	assert.Equal(t, commitData.DocumentRoot, docRoot, "Anchor should have the passed document root")
	assert.Equal(t, commitData.DocumentProof, documentProof, "Anchor should have the document proofs")
}

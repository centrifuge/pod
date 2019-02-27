// +build unit

package anchors

import (
	"math/big"
	"testing"

	"github.com/centrifuge/go-centrifuge/crypto/secp256k1"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/testingutils/commons"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockAnchorRepo struct {
	mock.Mock
	anchorRepositoryContract
}

func (m *mockAnchorRepo) GetAnchorById(opts *bind.CallOpts, anchorID *big.Int) (struct {
	AnchorId     *big.Int
	DocumentRoot [32]byte
}, error) {
	args := m.Called(opts, anchorID)
	type Response struct {
		AnchorId     *big.Int
		DocumentRoot [32]byte
	}
	r := Response{}
	dr := args.Get(0).([32]byte)
	r.DocumentRoot = dr

	return r, args.Error(1)
}

func TestCorrectCommitSignatureGen(t *testing.T) {
	anchorID, _ := hexutil.Decode("0x154cc26833dec2f4ad7ead9d65f9ec968a1aa5efbf6fe762f8f2a67d18a2d9b1")
	documentRoot, _ := hexutil.Decode("0x65a35574f70281ae4d1f6c9f3adccd5378743f858c67a802a49a08ce185bc975")
	address, _ := hexutil.Decode("0x89b0a86583c4444acfd71b463e0d3c55ae1412a5")
	correctCommitToSign := "0x004a050342f1edda2462288b9e0123a2e1bcc4f978efdc08c07bbf0c3ccc8ddd"
	correctCommitSignature := "0x4a73286521114f528967674bae4ecdc6cc94789255495429a7f58ca3ef0158ae257dd02a0ccb71d817e480d06f60f640ec021ade2ff90fe601bb7a5f4ddc569700"
	testPrivateKey, _ := hexutil.Decode("0x17e063fa17dd8274b09c14b253697d9a20afff74ace3c04fdb1b9c814ce0ada5")
	anchorIDTyped, _ := ToAnchorID(anchorID)
	centIdTyped := identity.NewDIDFromByte(address)
	docRootTyped, _ := ToDocumentRoot(documentRoot)
	messageToSign := GenerateCommitHash(anchorIDTyped, centIdTyped, docRootTyped)
	assert.Equal(t, correctCommitToSign, hexutil.Encode(messageToSign), "messageToSign not calculated correctly")
	signature, _ := secp256k1.SignEthereum(messageToSign, testPrivateKey)
	assert.Equal(t, correctCommitSignature, hexutil.Encode(signature), "signature not correct")
}

func TestGenerateAnchor(t *testing.T) {
	currentAnchorID := utils.RandomByte32()
	currentDocumentRoot := utils.RandomByte32()
	documentProof := utils.RandomByte32()

	var documentProofs [][32]byte
	documentProofs = append(documentProofs, documentProof)

	var documentRoot32Bytes [32]byte
	copy(documentRoot32Bytes[:], currentDocumentRoot[:32])

	commitData := NewCommitData(0, currentAnchorID, documentRoot32Bytes, documentProofs)

	anchorID, _ := ToAnchorID(currentAnchorID[:])
	docRoot, _ := ToDocumentRoot(documentRoot32Bytes[:])

	assert.Equal(t, commitData.AnchorID, anchorID, "Anchor should have the passed ID")
	assert.Equal(t, commitData.DocumentRoot, docRoot, "Anchor should have the passed document root")

	assert.Equal(t, commitData.DocumentProofs, documentProofs, "Anchor should have the document proofs")

}

func TestGetDocumentRootOf(t *testing.T) {
	repo := &mockAnchorRepo{}
	anchorID, err := ToAnchorID(utils.RandomSlice(32))
	assert.Nil(t, err)

	ethClient := &testingcommons.MockEthClient{}
	ethClient.On("GetGethCallOpts").Return(nil)
	ethRepo := newService(cfg, repo, nil, ethClient, nil)
	docRoot := utils.RandomByte32()
	repo.On("GetAnchorById", mock.Anything, mock.Anything).Return(docRoot, nil)
	gotRoot, err := ethRepo.GetDocumentRootOf(anchorID)
	repo.AssertExpectations(t)
	assert.Nil(t, err)
	assert.Equal(t, docRoot[:], gotRoot[:])
}

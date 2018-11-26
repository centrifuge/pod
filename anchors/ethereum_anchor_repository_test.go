// +build unit

package anchors

import (
	"math/big"
	"testing"

	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/keytools/secp256k1"
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

func (m *mockAnchorRepo) Commits(opts *bind.CallOpts, anchorID *big.Int) (docRoot [32]byte, err error) {
	args := m.Called(opts, anchorID)
	docRoot, _ = args.Get(0).([32]byte)
	return docRoot, args.Error(1)
}

func TestCorrectCommitSignatureGen(t *testing.T) {
	// hardcoded values are generated with centrifuge-ethereum-contracts
	anchorID, _ := hexutil.Decode("0x154cc26833dec2f4ad7ead9d65f9ec968a1aa5efbf6fe762f8f2a67d18a2d9b1")
	documentRoot, _ := hexutil.Decode("0x65a35574f70281ae4d1f6c9f3adccd5378743f858c67a802a49a08ce185bc975")
	centrifugeId, _ := hexutil.Decode("0x1851943e76d2")
	correctCommitToSign := "0x15f9cb57608a7ef31428fd6b1cb7ea2002ab032211d882b920c1474334004d6b"
	correctCommitSignature := "0xb4051d6d03c3bf39f4ec4ba949a91a358b0cacb4804b82ed2ba978d338f5e747770c00b63c8e50c1a7aa5ba629870b54c2068a56f8b43460aa47891c6635d36d01"
	testPrivateKey, _ := hexutil.Decode("0x17e063fa17dd8274b09c14b253697d9a20afff74ace3c04fdb1b9c814ce0ada5")
	anchorIDTyped, _ := ToAnchorID(anchorID)
	centIdTyped, _ := identity.ToCentID(centrifugeId)
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
	centrifugeId := utils.RandomSlice(identity.CentIDLength)
	testPrivateKey, _ := hexutil.Decode("0x17e063fa17dd8274b09c14b253697d9a20afff74ace3c04fdb1b9c814ce0ada5")

	var documentProofs [][32]byte
	documentProofs = append(documentProofs, documentProof)
	centIdTyped, _ := identity.ToCentID(centrifugeId)
	messageToSign := GenerateCommitHash(currentAnchorID, centIdTyped, currentDocumentRoot)
	signature, _ := secp256k1.SignEthereum(messageToSign, testPrivateKey)

	var documentRoot32Bytes [32]byte
	copy(documentRoot32Bytes[:], currentDocumentRoot[:32])

	commitData := NewCommitData(0, currentAnchorID, documentRoot32Bytes, centIdTyped, documentProofs, signature)

	anchorID, _ := ToAnchorID(currentAnchorID[:])
	docRoot, _ := ToDocumentRoot(documentRoot32Bytes[:])

	assert.Equal(t, commitData.AnchorID, anchorID, "Anchor should have the passed ID")
	assert.Equal(t, commitData.DocumentRoot, docRoot, "Anchor should have the passed document root")
	assert.Equal(t, commitData.CentrifugeID, centIdTyped, "Anchor should have the centrifuge id")
	assert.Equal(t, commitData.DocumentProofs, documentProofs, "Anchor should have the document proofs")
	assert.Equal(t, commitData.Signature, signature, "Anchor should have the signature")
}

func TestGetDocumentRootOf(t *testing.T) {
	repo := &mockAnchorRepo{}
	anchorID, err := ToAnchorID(utils.RandomSlice(32))
	assert.Nil(t, err)

	ethClient := &testingcommons.MockEthClient{}
	ethClient.On("GetGethCallOpts").Return(nil)
	ethRepo := newEthereumAnchorRepository(cfg, repo, nil, func() ethereum.Client {
		return ethClient
	})
	docRoot := utils.RandomByte32()
	repo.On("Commits", mock.Anything, mock.Anything).Return(docRoot, nil)
	gotRoot, err := ethRepo.GetDocumentRootOf(anchorID)
	repo.AssertExpectations(t)
	assert.Nil(t, err)
	assert.Equal(t, docRoot[:], gotRoot[:])
}

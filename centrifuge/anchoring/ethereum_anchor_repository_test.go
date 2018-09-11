// +build unit

package anchoring_test

import (
	"math/big"
	"testing"

	"github.com/CentrifugeInc/go-centrifuge/centrifuge/anchoring"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/identity"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/keytools/secp256k1"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	"github.com/go-errors/errors"
	"github.com/stretchr/testify/assert"
)

// ----- MOCKING -----
type MockRepositoryAnchor struct {
	shouldFail bool
}

func (mra *MockRepositoryAnchor) Commit(opts *bind.TransactOpts, _anchorId *big.Int, _documentRoot [32]byte, _centrifugeId *big.Int, _documentProofs [][32]byte, _signature []byte) (*types.Transaction, error) {
	if mra.shouldFail == true {
		return nil, errors.New("for testing - error if identifier == merkleRoot")
	}
	hashableTransaction := types.NewTransaction(1, common.HexToAddress("0x0000000000000000001"), big.NewInt(1000), 1000, big.NewInt(1000), nil)

	return hashableTransaction, nil
}

type MockWatchAnchorRegistered struct {
	shouldFail bool
	sink       chan<- *anchoring.EthereumAnchorRepositoryContractAnchorCommitted
}

func (mwar *MockWatchAnchorRegistered) WatchAnchorCommitted(opts *bind.WatchOpts, sink chan<- *anchoring.EthereumAnchorRepositoryContractAnchorCommitted,
	from []common.Address, anchorId []*big.Int, centrifugeId []*big.Int) (event.Subscription, error) {
	if mwar.shouldFail == true {
		return nil, errors.New("forced error during test")
	} else {
		if sink != nil {
			mwar.sink = sink
		}
		return nil, nil
	}
}

// END ----- MOCKING -----

func TestCorrectCommitSignatureGen(t *testing.T) {

	// hardcoded values are generated with centrifuge-ethereum-contracts
	anchorId, _ := hexutil.Decode("0x154cc26833dec2f4ad7ead9d65f9ec968a1aa5efbf6fe762f8f2a67d18a2d9b1")
	documentRoot, _ := hexutil.Decode("0x65a35574f70281ae4d1f6c9f3adccd5378743f858c67a802a49a08ce185bc975")
	centrifugeId, _ := hexutil.Decode("0x1851943e76d2")

	correctCommitToSign := "0x15f9cb57608a7ef31428fd6b1cb7ea2002ab032211d882b920c1474334004d6b"
	correctCommitSignature := "0xb4051d6d03c3bf39f4ec4ba949a91a358b0cacb4804b82ed2ba978d338f5e747770c00b63c8e50c1a7aa5ba629870b54c2068a56f8b43460aa47891c6635d36d01"

	testPrivateKey, _ := hexutil.Decode("0x17e063fa17dd8274b09c14b253697d9a20afff74ace3c04fdb1b9c814ce0ada5")

	anchorIdTyped, _ := anchoring.NewAnchorId(anchorId)
	centIdTyped, _ := identity.NewCentId(centrifugeId)
	docRootTyped, _ := anchoring.NewDocRoot(documentRoot)

	messageToSign := anchoring.GenerateCommitHash(anchorIdTyped, centIdTyped, docRootTyped)

	assert.Equal(t, correctCommitToSign, hexutil.Encode(messageToSign), "messageToSign not calculated correctly")

	signature, _ := secp256k1.SignEthereum(messageToSign, testPrivateKey)

	assert.Equal(t, correctCommitSignature, hexutil.Encode(signature), "signature not correct")

}

func TestGenerateAnchor(t *testing.T) {

	currentAnchorId := tools.RandomByte32()
	currentDocumentRoot := tools.RandomByte32()
	documentProof := tools.RandomByte32()
	centrifugeId := tools.RandomSlice(identity.CentIdByteLength)
	testPrivateKey, _ := hexutil.Decode("0x17e063fa17dd8274b09c14b253697d9a20afff74ace3c04fdb1b9c814ce0ada5")

	var documentProofs [][32]byte

	documentProofs = append(documentProofs, documentProof)
	centIdTyped, _ := identity.NewCentId(centrifugeId)
	messageToSign := anchoring.GenerateCommitHash(currentAnchorId, centIdTyped, currentDocumentRoot)
	signature, _ := secp256k1.SignEthereum(messageToSign, testPrivateKey)

	var documentRoot32Bytes [32]byte
	copy(documentRoot32Bytes[:], currentDocumentRoot[:32])

	commitData := anchoring.NewCommitData(currentAnchorId, documentRoot32Bytes, centIdTyped, documentProofs, signature)

	anchorId, _ := anchoring.NewAnchorId(currentAnchorId[:])
	docRoot, _ := anchoring.NewDocRoot(documentRoot32Bytes[:])

	assert.Equal(t, commitData.AnchorId, anchorId, "Anchor should have the passed ID")
	assert.Equal(t, commitData.DocumentRoot, docRoot, "Anchor should have the passed document root")
	assert.Equal(t, commitData.CentrifugeId, centIdTyped, "Anchor should have the centrifuge id")
	assert.Equal(t, commitData.DocumentProofs, documentProofs, "Anchor should have the document proofs")
	assert.Equal(t, commitData.Signature, signature, "Anchor should have the signature")

}

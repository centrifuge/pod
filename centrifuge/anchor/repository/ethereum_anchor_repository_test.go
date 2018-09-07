// +build unit

package repository

import (
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/identity"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/keytools/secp256k1"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/utils"
	"github.com/ethereum/go-ethereum/crypto"
	"math/big"
	"testing"

	"github.com/CentrifugeInc/go-centrifuge/centrifuge/tools"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
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
	sink       chan<- *EthereumAnchorRepositoryContractAnchorCommitted
}

func (mwar *MockWatchAnchorRegistered)WatchAnchorCommitted(opts *bind.WatchOpts, sink chan<- *EthereumAnchorRepositoryContractAnchorCommitted,
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

func generateCommitHash(anchorIdByte []byte,centrifugeIdByte []byte,documentRootByte []byte) ([]byte) {

	message := append(anchorIdByte, documentRootByte...)

	message = append(message, centrifugeIdByte...)

	messageToSign := crypto.Keccak256(message)
	return messageToSign
}

func TestCorrectCommitSignatureGen(t *testing.T){

	// hardcoded values are generated with centrifuge-ethereum-contracts
	anchorId := "0x154cc26833dec2f4ad7ead9d65f9ec968a1aa5efbf6fe762f8f2a67d18a2d9b1"
	documentRoot := "0x65a35574f70281ae4d1f6c9f3adccd5378743f858c67a802a49a08ce185bc975"
	centrifugeId := "0x1851943e76d2"

	correctCommitToSign := "0x15f9cb57608a7ef31428fd6b1cb7ea2002ab032211d882b920c1474334004d6b"
	correctCommitSignature := "0xb4051d6d03c3bf39f4ec4ba949a91a358b0cacb4804b82ed2ba978d338f5e747770c00b63c8e50c1a7aa5ba629870b54c2068a56f8b43460aa47891c6635d36d01"

	testPrivateKey := "0x17e063fa17dd8274b09c14b253697d9a20afff74ace3c04fdb1b9c814ce0ada5"

	anchorIdByte := utils.HexToByteArray(anchorId)
	documentRootByte := utils.HexToByteArray(documentRoot)
	centrifugeIdByte := utils.HexToByteArray(centrifugeId)

	messageToSign := generateCommitHash(anchorIdByte,centrifugeIdByte,documentRootByte)

	assert.Equal(t,correctCommitToSign,utils.ByteArrayToHex(messageToSign),"messageToSign not calculated correctly")

	signature := secp256k1.SignEthereum(messageToSign, utils.HexToByteArray(testPrivateKey))

	assert.Equal(t,correctCommitSignature,utils.ByteArrayToHex(signature),"signature not correct")

}

func TestGenerateAnchor(t *testing.T) {

	currentAnchorId := tools.RandomByte32()
	currentDocumentRoot := tools.RandomByte32()
	documentProof := tools.RandomByte32()
	centrifugeId := tools.RandomSlice(identity.CentIdByteLength)
	testPrivateKey := "0x17e063fa17dd8274b09c14b253697d9a20afff74ace3c04fdb1b9c814ce0ada5"

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

	commitData,err := NewCommitData(anchorIdBigInt,documentRoot32Bytes,centrifugeIdBigInt,documentProofs,signature)
	assert.Nil(t,err,"err should be nil")

	assert.Equal(t, commitData.AnchorId, anchorIdBigInt, "Anchor should have the passed ID")
	assert.Equal(t, commitData.DocumentRoot, documentRoot32Bytes, "Anchor should have the passed document root")
	assert.Equal(t, commitData.CentrifugeId, centrifugeIdBigInt, "Anchor should have the centrifuge id")
	assert.Equal(t, commitData.DocumentProofs, documentProofs, "Anchor should have the document proofs")
	assert.Equal(t, commitData.Signature, signature, "Anchor should have the signature")


}



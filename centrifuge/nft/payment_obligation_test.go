package nft

import (
	"testing"
	"github.com/centrifuge/precise-proofs/proofs/proto"
	"github.com/centrifuge/go-centrifuge/centrifuge/tools"
	"github.com/stretchr/testify/assert"
	"errors"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/centrifuge/go-centrifuge/centrifuge/identity"
	"github.com/stretchr/testify/mock"
	)

func TestCreateProofData(t *testing.T) {
	sortedHashes := [][]byte{tools.RandomSlice(32), tools.RandomSlice(32)}
	salt := tools.RandomSlice(32)
	tests := []struct {
		name string
		proofs []*proofspb.Proof
		result proofData
		err error
	}{
		{
			"happypath",
			[]*proofspb.Proof{
				{
					Property: "prop1",
					Value: "value1",
					Salt: salt,
					SortedHashes: sortedHashes,
				},
				{
					Property: "prop2",
					Value: "value2",
					Salt: salt,
					SortedHashes: sortedHashes,
				},
			},
			proofData{
				Values: [3]string{"value1", "value2"},
				Proofs: [3][][32]byte{{byteSliceToByteArray32(sortedHashes[0]), byteSliceToByteArray32(sortedHashes[1])}, {byteSliceToByteArray32(sortedHashes[0]), byteSliceToByteArray32(sortedHashes[1])}},
				Salts: [3][32]byte{byteSliceToByteArray32(salt), byteSliceToByteArray32(salt)},
			},
			nil,
		},
		{
			"invalid hashes",
			[]*proofspb.Proof{
				{
					Property: "prop1",
					Value: "value1",
					Salt: salt,
					SortedHashes: [][]byte{tools.RandomSlice(33), tools.RandomSlice(31)},
				},
				{
					Property: "prop2",
					Value: "value2",
					Salt: salt,
					SortedHashes: [][]byte{tools.RandomSlice(33), tools.RandomSlice(31)},
				},
			},
			proofData{
				Values: [3]string{"value1", "value2"},
				Proofs: [3][][32]byte{{byteSliceToByteArray32(sortedHashes[0]), byteSliceToByteArray32(sortedHashes[1])}, {byteSliceToByteArray32(sortedHashes[0]), byteSliceToByteArray32(sortedHashes[1])}},
				Salts: [3][32]byte{byteSliceToByteArray32(salt), byteSliceToByteArray32(salt)},
			},
			errors.New("input exceeds length of 32"),
		},
		{
			"invalid salts",
			[]*proofspb.Proof{
				{
					Property: "prop1",
					Value: "value1",
					Salt: tools.RandomSlice(33),
					SortedHashes: sortedHashes,
				},
				{
					Property: "prop2",
					Value: "value2",
					Salt: tools.RandomSlice(32),
					SortedHashes: sortedHashes,
				},
			},
			proofData{
				Values: [3]string{"value1", "value2"},
				Proofs: [3][][32]byte{{byteSliceToByteArray32(sortedHashes[0]), byteSliceToByteArray32(sortedHashes[1])}, {byteSliceToByteArray32(sortedHashes[0]), byteSliceToByteArray32(sortedHashes[1])}},
				Salts: [3][32]byte{byteSliceToByteArray32(salt), byteSliceToByteArray32(salt)},
			},
			errors.New("input exceeds length of 32"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			proofData, err := createProofData(test.proofs)
			if test.err != nil {
				assert.Equal(t, test.err.Error(), err.Error())
			} else if err != nil {
				panic(err)
			} else {
				assert.Equal(t, test.result.Values, proofData.Values)
				assert.Equal(t, test.result.Proofs, proofData.Proofs)
				assert.Equal(t, test.result.Salts, proofData.Salts)
			}
		})
	}
}

type MockPaymentObligation struct{
	mock.Mock
}

func (m *MockPaymentObligation) Mint(opts *bind.TransactOpts, _to common.Address, _tokenId *big.Int, _tokenURI string, _anchorId *big.Int, _merkleRoot [32]byte, _values [3]string, _salts [3][32]byte, _proofs [3][][32]byte) (*types.Transaction, error) {
	args := m.Called(opts, _to, _tokenId, _tokenURI, _anchorId, _merkleRoot, _values, _salts, _proofs)
	return args.Get(0).(*types.Transaction), args.Error(1)
}

func TestPaymentObligationService(t *testing.T) {
	tests := []struct {
		name string
		documentService documents.Service
		paymentObligation PaymentObligation
		idService identity.Service
		config Config
	}{
		// TODO: test cases
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

		})
	}
}

func byteSliceToByteArray32(input []byte) (out [32]byte) {
	out, _ = tools.SliceToByte32(input)
	return out
}

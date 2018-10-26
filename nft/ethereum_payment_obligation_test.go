// +build unit

package nft

import (
	"errors"
	"math/big"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/coredocument"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/invoice"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/nft"
	"github.com/centrifuge/go-centrifuge/testingutils/commons"
	"github.com/centrifuge/go-centrifuge/testingutils/documents"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/precise-proofs/proofs/proto"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateProofData(t *testing.T) {
	sortedHashes := [][]byte{utils.RandomSlice(32), utils.RandomSlice(32)}
	salt := utils.RandomSlice(32)
	tests := []struct {
		name   string
		proofs []*proofspb.Proof
		result proofData
		err    error
	}{
		{
			"happypath",
			[]*proofspb.Proof{
				{
					Property:     "prop1",
					Value:        "value1",
					Salt:         salt,
					SortedHashes: sortedHashes,
				},
				{
					Property:     "prop2",
					Value:        "value2",
					Salt:         salt,
					SortedHashes: sortedHashes,
				},
			},
			proofData{
				Values: [3]string{"value1", "value2"},
				Proofs: [3][][32]byte{{byteSliceToByteArray32(sortedHashes[0]), byteSliceToByteArray32(sortedHashes[1])}, {byteSliceToByteArray32(sortedHashes[0]), byteSliceToByteArray32(sortedHashes[1])}},
				Salts:  [3][32]byte{byteSliceToByteArray32(salt), byteSliceToByteArray32(salt)},
			},
			nil,
		},
		{
			"invalid hashes",
			[]*proofspb.Proof{
				{
					Property:     "prop1",
					Value:        "value1",
					Salt:         salt,
					SortedHashes: [][]byte{utils.RandomSlice(33), utils.RandomSlice(31)},
				},
				{
					Property:     "prop2",
					Value:        "value2",
					Salt:         salt,
					SortedHashes: [][]byte{utils.RandomSlice(33), utils.RandomSlice(31)},
				},
			},
			proofData{
				Values: [3]string{"value1", "value2"},
				Proofs: [3][][32]byte{{byteSliceToByteArray32(sortedHashes[0]), byteSliceToByteArray32(sortedHashes[1])}, {byteSliceToByteArray32(sortedHashes[0]), byteSliceToByteArray32(sortedHashes[1])}},
				Salts:  [3][32]byte{byteSliceToByteArray32(salt), byteSliceToByteArray32(salt)},
			},
			errors.New("input exceeds length of 32"),
		},
		{
			"invalid salts",
			[]*proofspb.Proof{
				{
					Property:     "prop1",
					Value:        "value1",
					Salt:         utils.RandomSlice(33),
					SortedHashes: sortedHashes,
				},
				{
					Property:     "prop2",
					Value:        "value2",
					Salt:         utils.RandomSlice(32),
					SortedHashes: sortedHashes,
				},
			},
			proofData{
				Values: [3]string{"value1", "value2"},
				Proofs: [3][][32]byte{{byteSliceToByteArray32(sortedHashes[0]), byteSliceToByteArray32(sortedHashes[1])}, {byteSliceToByteArray32(sortedHashes[0]), byteSliceToByteArray32(sortedHashes[1])}},
				Salts:  [3][32]byte{byteSliceToByteArray32(salt), byteSliceToByteArray32(salt)},
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

type MockPaymentObligation struct {
	mock.Mock
}

func (m *MockPaymentObligation) Mint(opts *bind.TransactOpts, _to common.Address, _tokenId *big.Int, _tokenURI string, _anchorId *big.Int, _merkleRoot [32]byte, _values [3]string, _salts [3][32]byte, _proofs [3][][32]byte) (*types.Transaction, error) {
	args := m.Called(opts, _to, _tokenId, _tokenURI, _anchorId, _merkleRoot, _values, _salts, _proofs)
	return args.Get(0).(*types.Transaction), args.Error(1)
}

type MockConfig struct {
	mock.Mock
}

func (m *MockConfig) GetIdentityID() ([]byte, error) {
	args := m.Called()
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockConfig) GetEthereumDefaultAccountName() string {
	args := m.Called()
	return args.Get(0).(string)
}

func TestPaymentObligationService(t *testing.T) {
	tests := []struct {
		name    string
		mocker  func() (testingdocuments.MockService, *MockPaymentObligation, testingcommons.MockIDService, testingcommons.MockEthClient, MockConfig)
		request *nftpb.NFTMintRequest
		err     error
		result  string
	}{
		{
			"happypath",
			func() (testingdocuments.MockService, *MockPaymentObligation, testingcommons.MockIDService, testingcommons.MockEthClient, MockConfig) {
				centIDByte := utils.RandomSlice(6)
				centID, _ := identity.ToCentID(centIDByte)
				address := common.BytesToAddress(utils.RandomSlice(32))
				coreDoc := coredocument.New()
				coreDoc.DocumentRoot = utils.RandomSlice(32)
				proof := getDummyProof(coreDoc)
				docServiceMock := testingdocuments.MockService{}
				docServiceMock.On("GetCurrentVersion", decodeHex("0x1212")).Return(&invoice.Invoice{InvoiceNumber: "1232", CoreDocument: coreDoc}, nil)
				docServiceMock.On("CreateProofs", decodeHex("0x1212"), []string{"somefield"}).Return(proof, nil)
				paymentObligationMock := &MockPaymentObligation{}
				idServiceMock := testingcommons.MockIDService{}
				idServiceMock.On("GetIdentityAddress", centID).Return(address, nil)
				ethClientMock := testingcommons.MockEthClient{}
				ethClientMock.On("GetTxOpts", "ethacc").Return(&bind.TransactOpts{}, nil)
				ethClientMock.On("SubmitTransactionWithRetries",
					mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything,
				).Return(&types.Transaction{}, nil)
				configMock := MockConfig{}
				configMock.On("GetEthereumDefaultAccountName").Return("ethacc")
				configMock.On("GetIdentityID").Return(centIDByte, nil)
				return docServiceMock, paymentObligationMock, idServiceMock, ethClientMock, configMock
			},
			&nftpb.NFTMintRequest{Identifier: "0x1212", Type: "happypath", ProofFields: []string{"somefield"}},
			nil,
			"",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// get mocks
			docService, paymentOb, idService, ethClient, config := test.mocker()
			// with below config the documentType has to be test.name to avoid conflicts since registry is a singleton
			documents.GetRegistryInstance().Register(test.name, &docService)
			confirmations := make(chan *WatchTokenMinted)
			service := NewEthereumPaymentObligation(paymentOb, &idService, &ethClient, &config,
				func(tokenID *big.Int) (chan *WatchTokenMinted, error) {
					return confirmations, nil
				})
			_, err := service.MintNFT(decodeHex(test.request.Identifier), test.request.Type, test.request.RegistryAddress, test.request.DepositAddress, test.request.ProofFields)
			if test.err != nil {
				assert.Equal(t, test.err.Error(), err.Error())
			} else if err != nil {
				panic(err)
			}
			docService.AssertExpectations(t)
			paymentOb.AssertExpectations(t)
			idService.AssertExpectations(t)
			ethClient.AssertExpectations(t)
			config.AssertExpectations(t)
		})
	}
}

func getDummyProof(coreDoc *coredocumentpb.CoreDocument) *documents.DocumentProof {
	return &documents.DocumentProof{
		DocumentId: coreDoc.DocumentIdentifier,
		VersionId:  coreDoc.CurrentVersion,
		State:      "state",
		FieldProofs: []*proofspb.Proof{
			{
				Property: "prop1",
				Value:    "val1",
				Salt:     utils.RandomSlice(32),
				Hash:     utils.RandomSlice(32),
				SortedHashes: [][]byte{
					utils.RandomSlice(32),
					utils.RandomSlice(32),
					utils.RandomSlice(32),
				},
			},
			{
				Property: "prop2",
				Value:    "val2",
				Salt:     utils.RandomSlice(32),
				Hash:     utils.RandomSlice(32),
				SortedHashes: [][]byte{
					utils.RandomSlice(32),
					utils.RandomSlice(32),
				},
			},
		},
	}
}

func byteSliceToByteArray32(input []byte) (out [32]byte) {
	out, _ = utils.SliceToByte32(input)
	return out
}

func decodeHex(hex string) []byte {
	h, _ := hexutil.Decode(hex)
	return h
}

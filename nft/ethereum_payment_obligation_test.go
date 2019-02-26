// +build unit

package nft

import (
	"math/big"
	"testing"
	"time"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/invoice"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/nft"
	"github.com/centrifuge/go-centrifuge/testingutils"
	"github.com/centrifuge/go-centrifuge/testingutils/commons"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/testingutils/documents"
	"github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/centrifuge/go-centrifuge/testingutils/testingtx"
	"github.com/centrifuge/go-centrifuge/transactions"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/precise-proofs/proofs"
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
	v1hex := "0x76616c756531"
	v2hex := "0x76616c756532"
	v1, err := hexutil.Decode(v1hex)
	assert.NoError(t, err)
	v2, err := hexutil.Decode(v2hex)
	assert.NoError(t, err)
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
					Property:     proofs.ReadableName("prop1"),
					Value:        v1,
					Salt:         salt,
					SortedHashes: sortedHashes,
				},
				{
					Property:     proofs.ReadableName("prop2"),
					Value:        v2,
					Salt:         salt,
					SortedHashes: sortedHashes,
				},
			},
			proofData{
				Values: [][]byte{v1, v2},
				Proofs: [][][32]byte{{byteSliceToByteArray32(sortedHashes[0]), byteSliceToByteArray32(sortedHashes[1])}, {byteSliceToByteArray32(sortedHashes[0]), byteSliceToByteArray32(sortedHashes[1])}},
				Salts:  [][32]byte{byteSliceToByteArray32(salt), byteSliceToByteArray32(salt)},
			},
			nil,
		},
		{
			"invalid hashes",
			[]*proofspb.Proof{
				{
					Property:     proofs.ReadableName("prop1"),
					Value:        v1,
					Salt:         salt,
					SortedHashes: [][]byte{utils.RandomSlice(33), utils.RandomSlice(31)},
				},
				{
					Property:     proofs.ReadableName("prop2"),
					Value:        v2,
					Salt:         salt,
					SortedHashes: [][]byte{utils.RandomSlice(33), utils.RandomSlice(31)},
				},
			},
			proofData{
				Values: [][]byte{v1, v2},
				Proofs: [][][32]byte{{byteSliceToByteArray32(sortedHashes[0]), byteSliceToByteArray32(sortedHashes[1])}, {byteSliceToByteArray32(sortedHashes[0]), byteSliceToByteArray32(sortedHashes[1])}},
				Salts:  [][32]byte{byteSliceToByteArray32(salt), byteSliceToByteArray32(salt)},
			},
			errors.New("input exceeds length of 32"),
		},
		{
			"invalid salts",
			[]*proofspb.Proof{
				{
					Property:     proofs.ReadableName("prop1"),
					Value:        v1,
					Salt:         utils.RandomSlice(33),
					SortedHashes: sortedHashes,
				},
				{
					Property:     proofs.ReadableName("prop2"),
					Value:        v2,
					Salt:         utils.RandomSlice(32),
					SortedHashes: sortedHashes,
				},
			},
			proofData{
				Values: [][]byte{v1, v2},
				Proofs: [][][32]byte{{byteSliceToByteArray32(sortedHashes[0]), byteSliceToByteArray32(sortedHashes[1])}, {byteSliceToByteArray32(sortedHashes[0]), byteSliceToByteArray32(sortedHashes[1])}},
				Salts:  [][32]byte{byteSliceToByteArray32(salt), byteSliceToByteArray32(salt)},
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

func (m *MockPaymentObligation) Mint(opts *bind.TransactOpts, _to common.Address, _tokenId *big.Int, _tokenURI string, _anchorId *big.Int, _merkleRoot [32]byte, _values []string, _salts [][32]byte, _proofs [][][32]byte) (*types.Transaction, error) {
	args := m.Called(opts, _to, _tokenId, _tokenURI, _anchorId, _merkleRoot, _values, _salts, _proofs)
	return args.Get(0).(*types.Transaction), args.Error(1)
}

func TestPaymentObligationService(t *testing.T) {
	tests := []struct {
		name    string
		mocker  func() (testingdocuments.MockService, *MockPaymentObligation, testingcommons.MockIdentityService, testingcommons.MockEthClient, testingconfig.MockConfig, *testingutils.MockQueue, *testingtx.MockTxManager)
		request *nftpb.NFTMintRequest
		err     error
		result  string
	}{
		{
			"happypath",
			func() (testingdocuments.MockService, *MockPaymentObligation, testingcommons.MockIdentityService, testingcommons.MockEthClient, testingconfig.MockConfig, *testingutils.MockQueue, *testingtx.MockTxManager) {
				cd, err := documents.NewCoreDocumentWithCollaborators(nil)
				assert.NoError(t, err)
				cd.Document.DocumentRoot = utils.RandomSlice(32)
				proof := getDummyProof(&cd.Document)
				docServiceMock := testingdocuments.MockService{}
				docServiceMock.On("GetCurrentVersion", decodeHex("0x1212")).Return(&invoice.Invoice{InvoiceNumber: "1232", CoreDocument: cd}, nil)
				docServiceMock.On("CreateProofs", decodeHex("0x1212"), []string{"collaborators[0]"}).Return(proof, nil)
				paymentObligationMock := &MockPaymentObligation{}
				idServiceMock := testingcommons.MockIdentityService{}
				ethClientMock := testingcommons.MockEthClient{}
				ethClientMock.On("GetTxOpts", "ethacc").Return(&bind.TransactOpts{}, nil)
				ethClientMock.On("SubmitTransactionWithRetries",
					mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything,
				).Return(&types.Transaction{}, nil)
				configMock := testingconfig.MockConfig{}
				configMock.On("GetEthereumDefaultAccountName").Return("ethacc")
				cid := testingidentity.GenerateRandomDID()
				configMock.On("GetIdentityID").Return(cid[:], nil)
				configMock.On("GetEthereumAccount").Return(&config.AccountConfig{}, nil)
				configMock.On("GetEthereumContextWaitTimeout").Return(time.Second)
				configMock.On("GetReceiveEventNotificationEndpoint").Return("")
				configMock.On("GetP2PKeyPair").Return("", "")
				configMock.On("GetSigningKeyPair").Return("", "")
				configMock.On("GetEthAuthKeyPair").Return("", "")
				queueSrv := new(testingutils.MockQueue)
				txMan := &testingtx.MockTxManager{}
				txMan.On("ExecuteWithinTX", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything).Return(transactions.NilTxID(), make(chan bool), nil)
				return docServiceMock, paymentObligationMock, idServiceMock, ethClientMock, configMock, queueSrv, txMan
			},
			&nftpb.NFTMintRequest{Identifier: "0x1212", ProofFields: []string{"collaborators[0]"}, DepositAddress: "0xf72855759a39fb75fc7341139f5d7a3974d4da08"},
			nil,
			"",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// get mocks
			docService, paymentOb, idService, ethClient, mockCfg, queueSrv, txMan := test.mocker()
			// with below config the documentType has to be test.name to avoid conflicts since registry is a singleton
			queueSrv.On("EnqueueJobWithMaxTries", mock.Anything, mock.Anything).Return(nil, nil).Once()
			service := newEthereumPaymentObligation(&mockCfg, &idService, &ethClient, queueSrv, &docService, func(address common.Address, client ethereum.Client) (*EthereumPaymentObligationContract, error) {
				return &EthereumPaymentObligationContract{}, nil
			}, txMan, func() (uint64, error) { return 10, nil })
			ctxh := testingconfig.CreateAccountContext(t, &mockCfg)
			req := MintNFTRequest{
				DocumentID:      decodeHex(test.request.Identifier),
				RegistryAddress: common.HexToAddress(test.request.RegistryAddress),
				DepositAddress:  common.HexToAddress(test.request.DepositAddress),
				ProofFields:     test.request.ProofFields,
			}
			_, _, err := service.MintNFT(ctxh, req)
			if test.err != nil {
				assert.Equal(t, test.err.Error(), err.Error())
			} else if err != nil {
				panic(err)
			}
			paymentOb.AssertExpectations(t)
			idService.AssertExpectations(t)
			mockCfg.AssertExpectations(t)
		})
	}
}

func getDummyProof(coreDoc *coredocumentpb.CoreDocument) *documents.DocumentProof {
	v1, _ := hexutil.Decode("0x76616c756531")
	v2, _ := hexutil.Decode("0x76616c756532")
	return &documents.DocumentProof{
		DocumentID: coreDoc.DocumentIdentifier,
		VersionID:  coreDoc.CurrentVersion,
		State:      "state",
		FieldProofs: []*proofspb.Proof{
			{
				Property: proofs.ReadableName("prop1"),
				Value:    v1,
				Salt:     utils.RandomSlice(32),
				Hash:     utils.RandomSlice(32),
				SortedHashes: [][]byte{
					utils.RandomSlice(32),
					utils.RandomSlice(32),
					utils.RandomSlice(32),
				},
			},
			{
				Property: proofs.ReadableName("prop2"),
				Value:    v2,
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

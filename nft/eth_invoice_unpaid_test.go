// +build unit

package nft

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/invoice"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/nft"
	"github.com/centrifuge/go-centrifuge/testingutils"
	"github.com/centrifuge/go-centrifuge/testingutils/commons"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/testingutils/documents"
	"github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/centrifuge/go-centrifuge/testingutils/testingjobs"
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
			proofData, err := convertToProofData(test.proofs)
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

type MockInvoiceUnpaid struct {
	mock.Mock
}

func (m *MockInvoiceUnpaid) Mint(opts *bind.TransactOpts, _to common.Address, _tokenId *big.Int, _tokenURI string, _anchorId *big.Int, _merkleRoot [32]byte, _values []string, _salts [][32]byte, _proofs [][][32]byte) (*types.Transaction, error) {
	args := m.Called(opts, _to, _tokenId, _tokenURI, _anchorId, _merkleRoot, _values, _salts, _proofs)
	return args.Get(0).(*types.Transaction), args.Error(1)
}

func TestInvoiceUnpaid(t *testing.T) {
	tests := []struct {
		name    string
		mocker  func() (testingdocuments.MockService, *MockInvoiceUnpaid, testingcommons.MockIdentityService, testingcommons.MockEthClient, testingconfig.MockConfig, *testingutils.MockQueue, *testingjobs.MockJobManager)
		request *nftpb.NFTMintRequest
		err     error
		result  string
	}{
		{
			"happypath",
			func() (testingdocuments.MockService, *MockInvoiceUnpaid, testingcommons.MockIdentityService, testingcommons.MockEthClient, testingconfig.MockConfig, *testingutils.MockQueue, *testingjobs.MockJobManager) {
				cd, err := documents.NewCoreDocumentForDoc(nil, documents.CollaboratorsAccess{}, nil)
				assert.NoError(t, err)
				proof := getDummyProof(cd.GetTestCoreDocWithReset())
				docServiceMock := testingdocuments.MockService{}
				docServiceMock.On("GetCurrentVersion", decodeHex("0x1212")).Return(&invoice.Invoice{Number: "1232", CoreDocument: cd}, nil)
				docServiceMock.On("CreateProofs", decodeHex("0x1212"), []string{"collaborators[0]"}).Return(proof, nil)
				invoiceUnpaidMock := &MockInvoiceUnpaid{}
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
				configMock.On("GetPrecommitEnabled").Return(false)
				configMock.On("GetLowEntropyNFTTokenEnabled").Return(false)
				queueSrv := new(testingutils.MockQueue)
				jobMan := new(testingjobs.MockJobManager)
				jobMan.On("ExecuteWithinJob", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything).Return(jobs.NilJobID(), make(chan bool), nil)
				return docServiceMock, invoiceUnpaidMock, idServiceMock, ethClientMock, configMock, queueSrv, jobMan
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
			service := newEthInvoiceUnpaid(&mockCfg, &idService, &ethClient, queueSrv, &docService, func(address common.Address, client ethereum.Client) (*InvoiceUnpaidContract, error) {
				return &InvoiceUnpaidContract{}, nil
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

func TestEthereumInvoiceUnpaid_GetRequiredInvoiceUnpaidProofFields(t *testing.T) {
	service := newEthInvoiceUnpaid(nil, nil, nil, nil, nil, nil, nil, nil)

	//missing account in context
	ctxh := context.Background()
	proofList, err := service.GetRequiredInvoiceUnpaidProofFields(ctxh)
	assert.Error(t, err)
	assert.Nil(t, proofList)

	//error identity keys
	tc, err := configstore.NewAccount("main", cfg)
	assert.Nil(t, err)
	acc := tc.(*configstore.Account)
	acc.EthereumAccount = &config.AccountConfig{
		Key: "blabla",
	}
	ctxh, err = contextutil.New(ctxh, acc)
	assert.Nil(t, err)
	proofList, err = service.GetRequiredInvoiceUnpaidProofFields(ctxh)
	assert.Error(t, err)
	assert.Nil(t, proofList)

	//success assertions
	tc, err = configstore.NewAccount("main", cfg)
	assert.Nil(t, err)
	ctxh, err = contextutil.New(ctxh, tc)
	assert.Nil(t, err)
	proofList, err = service.GetRequiredInvoiceUnpaidProofFields(ctxh)
	assert.NoError(t, err)
	assert.Len(t, proofList, 8)
	accDIDBytes, err := tc.GetIdentityID()
	assert.NoError(t, err)
	keys, err := tc.GetKeys()
	assert.NoError(t, err)
	signerID := hexutil.Encode(append(accDIDBytes, keys[identity.KeyPurposeSigning.Name].PublicKey...))
	signatureSender := fmt.Sprintf("%s.signatures[%s].signature", documents.SignaturesTreePrefix, signerID)
	assert.Equal(t, signatureSender, proofList[6])
}

func TestFilterMintProofs(t *testing.T) {
	service := newEthInvoiceUnpaid(nil, nil, nil, nil, nil, nil, nil, nil)
	indexKey := utils.RandomSlice(52)
	docProof := &documents.DocumentProof{
		FieldProofs: []*proofspb.Proof{
			{
				Property: proofs.CompactName([]byte{10, 100, 5, 20, 69, 1, 0, 1}...),
				Value:    utils.RandomSlice(32),
				Salt:     utils.RandomSlice(32),
				Hash:     utils.RandomSlice(32),
				SortedHashes: [][]byte{
					utils.RandomSlice(32),
					utils.RandomSlice(32),
					utils.RandomSlice(32),
				},
			},
			{
				Property: proofs.CompactName(append(documents.CompactProperties(documents.DRTreePrefix), documents.CompactProperties(documents.SigningRootField)...)...),
				Value:    utils.RandomSlice(32),
				Salt:     utils.RandomSlice(32),
				Hash:     utils.RandomSlice(32),
				SortedHashes: [][]byte{
					utils.RandomSlice(32),
					utils.RandomSlice(32),
					utils.RandomSlice(32),
				},
			},
			{
				Property: proofs.CompactName(append([]byte{3, 0, 0, 0, 0, 0, 0, 1}, append(indexKey, []byte{0, 0, 0, 4}...)...)...),
				Value:    utils.RandomSlice(32),
				Salt:     utils.RandomSlice(32),
				Hash:     utils.RandomSlice(32),
				SortedHashes: [][]byte{
					utils.RandomSlice(32),
					utils.RandomSlice(32),
					utils.RandomSlice(32),
				},
			},
		},
	}

	docProofAux := service.filterMintProofs(docProof)
	assert.Len(t, docProofAux.FieldProofs[0].SortedHashes, 2)
	assert.Len(t, docProofAux.FieldProofs[1].SortedHashes, 3)
	assert.Len(t, docProofAux.FieldProofs[2].SortedHashes, 3)
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

// +build unit

package nft

import (
	"math/big"
	"testing"
	"time"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/generic"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/jobs"
	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/commons"
	testingconfig "github.com/centrifuge/go-centrifuge/testingutils/config"
	testingdocuments "github.com/centrifuge/go-centrifuge/testingutils/documents"
	testingidentity "github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/precise-proofs/proofs"
	proofspb "github.com/centrifuge/precise-proofs/proofs/proto"
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
		name   string
		mocker func() (testingdocuments.MockService, *MockInvoiceUnpaid, testingcommons.MockIdentityService,
			ethereum.MockEthClient, testingconfig.MockConfig, *jobs.MockDispatcher)
		request MintNFTRequest
		err     error
		result  string
	}{
		{
			"happypath",
			func() (testingdocuments.MockService, *MockInvoiceUnpaid, testingcommons.MockIdentityService,
				ethereum.MockEthClient, testingconfig.MockConfig, *jobs.MockDispatcher) {
				cd, err := documents.NewCoreDocument(nil, documents.CollaboratorsAccess{}, nil)
				assert.NoError(t, err)
				proof := getDummyProof(cd.GetTestCoreDocWithReset())
				docServiceMock := testingdocuments.MockService{}
				docServiceMock.On("GetCurrentVersion", decodeHex("0x1212")).Return(&generic.Generic{
					CoreDocument: cd,
					Data:         generic.Data{},
				}, nil)
				docServiceMock.On("CreateProofs", decodeHex("0x1212"), []string{"collaborators[0]"}).Return(proof, nil)
				invoiceUnpaidMock := &MockInvoiceUnpaid{}
				idServiceMock := testingcommons.MockIdentityService{}
				ethClientMock := ethereum.MockEthClient{}
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
				configMock.On("GetEthereumAccount", "main").Return(&config.AccountConfig{}, nil)
				configMock.On("GetEthereumContextWaitTimeout").Return(time.Second)
				configMock.On("GetReceiveEventNotificationEndpoint").Return("")
				configMock.On("GetP2PKeyPair").Return("", "")
				configMock.On("GetSigningKeyPair").Return("", "")
				configMock.On("GetPrecommitEnabled").Return(false)
				configMock.On("GetCentChainAccount").Return(config.CentChainAccount{}, nil).Once()
				dispatcher := new(jobs.MockDispatcher)
				dispatcher.On("Dispatch", mock.Anything, mock.Anything).Return(utils.RandomSlice(32), nil)
				return docServiceMock, invoiceUnpaidMock, idServiceMock, ethClientMock, configMock, dispatcher
			},
			MintNFTRequest{DocumentID: decodeHex("0x1212"), ProofFields: []string{"collaborators[0]"}, DepositAddress: common.HexToAddress("0xf72855759a39fb75fc7341139f5d7a3974d4da08")},
			nil,
			"",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// get mocks
			docService, paymentOb, idService, ethClient, mockCfg, dispatcher := test.mocker()
			// with below config the documentType has to be test.name to avoid conflicts since registry is a singleton
			service := newService(&ethClient, &docService, ethereum.BindContract, dispatcher, nil)
			ctxh := testingconfig.CreateAccountContext(t, &mockCfg)
			req := MintNFTRequest{
				DocumentID:      test.request.DocumentID,
				RegistryAddress: test.request.RegistryAddress,
				DepositAddress:  test.request.DepositAddress,
				ProofFields:     test.request.ProofFields,
			}
			_, err := service.MintNFT(ctxh, req)
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

func Test_getBundledHash(t *testing.T) {
	to := common.HexToAddress("0xf2bd5de8b57ebfc45dcee97524a7a08fccc80aef")
	props := [][]byte{
		common.FromHex("0x392614ecdd98ce9b86b6c82242ae1b85aaf53ebe6f52490ed44539c88215b17a"),
		common.FromHex("0x8db964a550ede5fea3f059ca6a74cf436890bb1d31a39c63ea0ccfbc8d8235fd"),
		common.FromHex("0xc437005805629feeb716f4ff329f62a4cf393f4cbfc7cd14fc0a64d8321a3e99"),
	}

	values := [][]byte{
		common.FromHex("0xd6ad85800460ea404f3289484f9300ed787dc951203cb3f0ef5fa0fa4db283cc"),
		common.FromHex("0x446bfed759680364b759d32d6d217e287df7aad0bf4c82816f124d7e03ab248f"),
		common.FromHex("0x443e4fa3d89952c9f24433d1112713a075d9205195dc9a16a12301caa1afb5d2"),
	}

	salts := [][32]byte{
		byteSliceToByteArray32(common.FromHex("0x34ea1aa3061dca2e1e23573c3b8866f80032d18fd85934d90339c8bafcab0408")),
		byteSliceToByteArray32(common.FromHex("0xe257b56611cf3244b2b63bfe486ea3072f10223d473285f8fea868aae2323b39")),
		byteSliceToByteArray32(common.FromHex("0xed58f4a0d0c76770c81d2b1cc035413edebb567f5c006160596dc73b9297a9cc")),
	}

	bh := common.FromHex("0xee49e1ca6aa1204cfb571094ce14ab254e5185005cbee3f26af9afd3140ac12d")
	got := getBundledHash(to, props, values, salts)
	assert.Equal(t, bh, got[:], "bundled hash mismatch")
}

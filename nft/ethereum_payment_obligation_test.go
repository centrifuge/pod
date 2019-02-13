// +build unit

package nft

import (
	"crypto/sha256"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/coredocument"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/invoice"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/nft"
	"github.com/centrifuge/go-centrifuge/testingutils"
	"github.com/centrifuge/go-centrifuge/testingutils/commons"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/testingutils/documents"
	"github.com/centrifuge/go-centrifuge/testingutils/testingtx"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/centrifuge/precise-proofs/proofs/proto"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/satori/go.uuid"
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
		mocker  func() (testingdocuments.MockService, *MockPaymentObligation, testingcommons.MockIDService, testingcommons.MockEthClient, testingconfig.MockConfig, *testingutils.MockQueue, *testingtx.MockTxManager)
		request *nftpb.NFTMintRequest
		err     error
		result  string
	}{
		{
			"happypath",
			func() (testingdocuments.MockService, *MockPaymentObligation, testingcommons.MockIDService, testingcommons.MockEthClient, testingconfig.MockConfig, *testingutils.MockQueue, *testingtx.MockTxManager) {
				coreDoc := coredocument.New()
				coreDoc.DocumentRoot = utils.RandomSlice(32)
				proof := getDummyProof(coreDoc)
				docServiceMock := testingdocuments.MockService{}
				docServiceMock.On("GetCurrentVersion", decodeHex("0x1212")).Return(&invoice.Invoice{InvoiceNumber: "1232", CoreDocument: coreDoc}, nil)
				docServiceMock.On("CreateProofs", decodeHex("0x1212"), []string{"collaborators[0]"}).Return(proof, nil)
				paymentObligationMock := &MockPaymentObligation{}
				idServiceMock := testingcommons.MockIDService{}
				ethClientMock := testingcommons.MockEthClient{}
				ethClientMock.On("GetTxOpts", "ethacc").Return(&bind.TransactOpts{}, nil)
				ethClientMock.On("SubmitTransactionWithRetries",
					mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything,
				).Return(&types.Transaction{}, nil)
				configMock := testingconfig.MockConfig{}
				configMock.On("GetEthereumDefaultAccountName").Return("ethacc")
				cid := identity.RandomCentID()
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
					mock.Anything, mock.Anything).Return(uuid.Nil, make(chan bool), nil)
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
				RegistryAddress: test.request.RegistryAddress,
				DepositAddress:  test.request.DepositAddress,
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

func Test_addNFT(t *testing.T) {
	cd := coredocument.New()
	registry := common.HexToAddress("0xf72855759a39fb75fc7341139f5d7a3974d4da08")
	registry2 := common.HexToAddress("0xf72855759a39fb75fc7341139f5d7a3974d4da02")
	tokenID := utils.RandomSlice(32)
	assert.Nil(t, cd.Nfts)

	addNFT(cd, registry.Bytes(), tokenID)
	assert.Len(t, cd.Nfts, 1)
	assert.Len(t, cd.Nfts[0].RegistryId, 32)
	assert.Equal(t, tokenID, getStoredNFT(cd.Nfts, registry.Bytes()).TokenId)
	assert.Nil(t, getStoredNFT(cd.Nfts, registry2.Bytes()))

	tokenID = utils.RandomSlice(32)
	addNFT(cd, registry.Bytes(), tokenID)
	assert.Len(t, cd.Nfts, 1)
	assert.Len(t, cd.Nfts[0].RegistryId, 32)
	assert.Equal(t, tokenID, getStoredNFT(cd.Nfts, registry.Bytes()).TokenId)
}

func Test_getRoleForAccount(t *testing.T) {
	cd := coredocument.New()
	roleName := "supplier"
	id := identity.RandomCentID()
	_, r := getRoleForAccount(cd, roleName, id)
	assert.Nil(t, r)

	// add role
	rk := sha256.Sum256([]byte(roleName))
	role := new(coredocumentpb.Role)
	role.RoleKey = rk[:]
	cd.Roles = append(cd.Roles, role)
	_, r = getRoleForAccount(cd, roleName, id)
	assert.Nil(t, r)

	// add id
	role.Collaborators = append(role.Collaborators, id[:])
	_, r = getRoleForAccount(cd, roleName, id)
	assert.NotNil(t, r)
	assert.Equal(t, r.RoleKey, rk[:])
	assert.Equal(t, r.Collaborators[0], id[:])
}

func Test_createTokenProof_error(t *testing.T) {
	cd, err := coredocument.NewWithCollaborators([]string{"0x010203040506"})
	assert.Nil(t, err)
	cd.EmbeddedData = &any.Any{
		Value:   utils.RandomSlice(32),
		TypeUrl: "some type",
	}

	cdTree, err := coredocument.GetCoreDocTree(cd)
	assert.Nil(t, err)

	registry := common.HexToAddress("0xf72855759a39fb75fc7341139f5d7a3974d4da08")

	// no nft registered yet
	_, err = createTokenProof(cd, cdTree, registry)
	assert.Error(t, err)
}

func Test_createTokenProof(t *testing.T) {
	cd := coredocument.New()
	registry := common.HexToAddress("0xf72855759a39fb75fc7341139f5d7a3974d4da08")
	tokenID := utils.RandomSlice(32)
	addNFT(cd, registry.Bytes(), tokenID)
	cd.EmbeddedData = &any.Any{
		Value:   utils.RandomSlice(32),
		TypeUrl: "some type",
	}
	assert.Nil(t, coredocument.FillSalts(cd))

	cdTree, err := coredocument.GetCoreDocTree(cd)
	assert.Nil(t, err)

	pf, err := createTokenProof(cd, cdTree, registry)
	assert.Nil(t, err)
	rk := hexutil.Encode(append(registry.Bytes(), make([]byte, 12, 12)...))
	assert.Equal(t, pf.GetReadableName(), fmt.Sprintf("nfts[%s]", rk))
	assert.Equal(t, pf.Value, hexutil.Encode(tokenID))
	valid, err := cdTree.ValidateProof(&pf)
	assert.NoError(t, err)
	assert.True(t, valid)
}

func Test_createNFTReadAccessProof_missing_nft(t *testing.T) {
	cd, err := coredocument.NewWithCollaborators([]string{"0x010203040506"})
	assert.Nil(t, err)
	cd.EmbeddedData = &any.Any{
		Value:   utils.RandomSlice(32),
		TypeUrl: "some type",
	}

	cdTree, err := coredocument.GetCoreDocTree(cd)
	assert.Nil(t, err)

	registry := common.HexToAddress("0xf72855759a39fb75fc7341139f5d7a3974d4da08")
	_, err = createNFTReadAccessProof(cd, cdTree, registry, utils.RandomSlice(32))
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(ErrNFTRoleMissing, err))
}

func Test_createNFTReadAccessProof(t *testing.T) {
	cd := coredocument.New()
	registry := common.HexToAddress("0xf72855759a39fb75fc7341139f5d7a3974d4da08")
	tokenID := utils.RandomSlice(32)
	cd.EmbeddedData = &any.Any{
		Value:   utils.RandomSlice(32),
		TypeUrl: "some type",
	}
	assert.NoError(t, coredocument.AddNFTToReadRules(cd, registry, tokenID))
	assert.Nil(t, coredocument.FillSalts(cd))

	cdTree, err := coredocument.GetCoreDocTree(cd)
	assert.Nil(t, err)

	pfs, err := createNFTReadAccessProof(cd, cdTree, registry, tokenID)
	assert.NoError(t, err)
	assert.Len(t, pfs, 3)

	rk := hexutil.Encode(make([]byte, 32, 32))
	assert.Equal(t, pfs[0].GetReadableName(), "read_rules[0].roles[0]")
	assert.Equal(t, pfs[0].Value, rk)
	valid, err := cdTree.ValidateProof(&pfs[0])
	assert.NoError(t, err)
	assert.True(t, valid)

	assert.Equal(t, pfs[1].GetReadableName(), fmt.Sprintf("roles[%s].nfts[0]", rk))
	enft, err := coredocument.ConstructNFT(registry, tokenID)
	assert.NoError(t, err)
	assert.Equal(t, pfs[1].Value, hexutil.Encode(enft))
	valid, err = cdTree.ValidateProof(&pfs[1])
	assert.NoError(t, err)
	assert.True(t, valid)

	assert.Equal(t, pfs[2].GetReadableName(), "read_rules[0].action")
	assert.Equal(t, pfs[2].Value, "1")
	valid, err = cdTree.ValidateProof(&pfs[2])
	assert.NoError(t, err)
	assert.True(t, valid)
}

func Test_createRoleProof_missing_role(t *testing.T) {
	cid := identity.RandomCentID()
	cd, err := coredocument.NewWithCollaborators([]string{cid.String()})
	assert.Nil(t, err)
	cd.EmbeddedData = &any.Any{
		Value:   utils.RandomSlice(32),
		TypeUrl: "some type",
	}

	cdTree, err := coredocument.GetCoreDocTree(cd)
	assert.Nil(t, err)

	_, err = createRoleProof(cd, cdTree, "Some key", cid)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(ErrNFTRoleMissing, err))
}

func Test_createRoleProof(t *testing.T) {
	cid := identity.RandomCentID()
	cd := coredocument.New()
	cd.EmbeddedData = &any.Any{
		Value:   utils.RandomSlice(32),
		TypeUrl: "some type",
	}

	role := new(coredocumentpb.Role)
	rk := sha256.Sum256([]byte("supplier"))
	role.RoleKey = rk[:]
	role.Collaborators = append(role.Collaborators, cid[:])
	cd.Roles = append(cd.Roles, role)

	assert.Nil(t, coredocument.FillSalts(cd))
	cdTree, err := coredocument.GetCoreDocTree(cd)
	assert.Nil(t, err)

	pf, err := createRoleProof(cd, cdTree, "supplier", cid)
	assert.NoError(t, err)

	pk := fmt.Sprintf("roles[%s].collaborators[0]", hexutil.Encode(rk[:]))
	assert.Equal(t, pf.GetReadableName(), pk)
	assert.Equal(t, pf.Value, cid.String())
}

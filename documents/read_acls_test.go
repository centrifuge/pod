// +build unit

package documents

import (
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	p2ppb "github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/commons"
	testingconfig "github.com/centrifuge/go-centrifuge/testingutils/config"
	testingidentity "github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/sha3"
)

func TestReadACLs_initReadRules(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	cd.initReadRules(nil)
	assert.Nil(t, cd.Document.Roles)
	assert.Nil(t, cd.Document.ReadRules)

	cs := []identity.DID{testingidentity.GenerateRandomDID()}
	cd.initReadRules(cs)
	assert.Len(t, cd.Document.ReadRules, 1)
	assert.Len(t, cd.Document.Roles, 1)

	cd.initReadRules(cs)
	assert.Len(t, cd.Document.ReadRules, 1)
	assert.Len(t, cd.Document.Roles, 1)
}

func TestReadAccessValidator_AccountCanRead(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	account := testingidentity.GenerateRandomDID()
	ncd, err := cd.PrepareNewVersion(nil, CollaboratorsAccess{ReadWriteCollaborators: []identity.DID{account}}, nil)
	assert.NoError(t, err)
	assert.NotNil(t, ncd.Document.ReadRules)
	assert.NotNil(t, ncd.Document.Roles)

	// account who cant access
	rcid := testingidentity.GenerateRandomDID()
	assert.False(t, ncd.AccountCanRead(rcid))

	// account can access
	assert.True(t, ncd.AccountCanRead(account))
}

type mockRegistry struct {
	mock.Mock
}

func (m mockRegistry) OwnerOf(registry common.Address, tokenID []byte) (common.Address, error) {
	args := m.Called(registry, tokenID)
	addr, _ := args.Get(0).(common.Address)
	return addr, args.Error(1)
}

func (m mockRegistry) OwnerOfWithRetrial(registry common.Address, tokenID []byte) (common.Address, error) {
	args := m.Called(registry, tokenID)
	addr, _ := args.Get(0).(common.Address)
	return addr, args.Error(1)
}

func (m mockRegistry) CurrentIndexOfToken(registry common.Address, tokenID []byte) (*big.Int, error) {
	args := m.Called(registry, tokenID)
	addr, _ := args.Get(0).(*big.Int)
	return addr, args.Error(1)
}

func TestCoreDocument_addNFTToReadRules(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)

	// wrong registry or token format
	registry := common.HexToAddress("0xf72855759a39fb75fc7341139f5d7a3974d4da08")
	tokenID := utils.RandomSlice(34)
	err = cd.addNFTToReadRules(registry, tokenID)
	assert.Error(t, err)
	assert.Nil(t, cd.Document.ReadRules)
	assert.Nil(t, cd.Document.Roles)

	tokenID = utils.RandomSlice(32)
	err = cd.addNFTToReadRules(registry, tokenID)
	assert.NoError(t, err)
	assert.Len(t, cd.Document.ReadRules, 1)
	assert.Equal(t, cd.Document.ReadRules[0].Action, coredocumentpb.Action_ACTION_READ)
	assert.Len(t, cd.Document.Roles, 1)
	enft, err := ConstructNFT(registry, tokenID)
	assert.NoError(t, err)
	assert.Equal(t, enft, cd.Document.Roles[0].Nfts[0])
}

func TestCoreDocument_NFTOwnerCanRead(t *testing.T) {
	account := testingidentity.GenerateRandomDID()
	cd, err := NewCoreDocument(nil, CollaboratorsAccess{ReadWriteCollaborators: []identity.DID{account}}, nil)
	assert.NoError(t, err)
	registry := common.HexToAddress("0xf72855759a39fb75fc7341139f5d7a3974d4da08")

	// account can read
	assert.NoError(t, cd.NFTOwnerCanRead(nil, registry, nil, account))

	// account not in read rules and nft missing
	account = testingidentity.GenerateRandomDID()
	tokenID := utils.RandomSlice(32)
	assert.Error(t, cd.NFTOwnerCanRead(nil, registry, tokenID, account))

	tr := mockRegistry{}
	tr.On("OwnerOf", registry, tokenID).Return(nil, errors.New("failed to get owner of")).Once()
	assert.NoError(t, cd.addNFTToReadRules(registry, tokenID))
	assert.Error(t, cd.NFTOwnerCanRead(tr, registry, tokenID, account))
	tr.AssertExpectations(t)

	// not the same owner
	owner := common.BytesToAddress(utils.RandomSlice(20))
	tr.On("OwnerOf", registry, tokenID).Return(owner, nil).Once()
	assert.Error(t, cd.NFTOwnerCanRead(tr, registry, tokenID, account))
	tr.AssertExpectations(t)

	// same owner
	owner = account.ToAddress()
	tr.On("OwnerOf", registry, tokenID).Return(owner, nil).Once()
	assert.NoError(t, cd.NFTOwnerCanRead(tr, registry, tokenID, account))
	tr.AssertExpectations(t)
}

func TestCoreDocumentModel_AddNFT(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	registry := common.HexToAddress("0xf72855759a39fb75fc7341139f5d7a3974d4da08")
	registry2 := common.HexToAddress("0xf72855759a39fb75fc7341139f5d7a3974d4da02")
	tokenID := utils.RandomSlice(32)
	assert.Nil(t, cd.Document.Nfts)
	assert.Nil(t, cd.Document.ReadRules)
	assert.Nil(t, cd.Document.Roles)

	cd, err = cd.AddNFT(true, registry, tokenID)
	assert.Nil(t, err)
	assert.Len(t, cd.Document.Nfts, 1)
	assert.Len(t, cd.Document.Nfts[0].RegistryId, 32)
	assert.Equal(t, tokenID, getStoredNFT(cd.Document.Nfts, registry.Bytes()).TokenId)
	assert.Nil(t, getStoredNFT(cd.Document.Nfts, registry2.Bytes()))
	assert.Len(t, cd.Document.ReadRules, 1)
	assert.Len(t, cd.Document.Roles, 1)
	assert.Len(t, cd.Document.Roles[0].Nfts, 1)

	tokenID = utils.RandomSlice(32)
	cd, err = cd.AddNFT(true, registry, tokenID)
	assert.Nil(t, err)
	assert.Len(t, cd.Document.Nfts, 1)
	assert.Len(t, cd.Document.Nfts[0].RegistryId, 32)
	assert.Equal(t, tokenID, getStoredNFT(cd.Document.Nfts, registry.Bytes()).TokenId)
	assert.Nil(t, getStoredNFT(cd.Document.Nfts, registry2.Bytes()))
	assert.Len(t, cd.Document.ReadRules, 2)
	assert.Len(t, cd.Document.Roles, 2)
	assert.Len(t, cd.Document.Roles[1].Nfts, 1)
}

func TestCoreDocument_IsNFTMinted(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	registry := common.HexToAddress("0xf72855759a39fb75fc7341139f5d7a3974d4da08")
	assert.False(t, cd.IsNFTMinted(nil, registry))

	tokenID := utils.RandomSlice(32)
	owner := common.HexToAddress("0xf72855759a39fb75fc7341139f5d7a3974d4da02")
	cd, err = cd.AddNFT(true, registry, tokenID)
	assert.Nil(t, err)

	tr := new(mockRegistry)
	tr.On("OwnerOf", registry, tokenID).Return(owner, nil).Once()
	assert.True(t, cd.IsNFTMinted(tr, registry))
	tr.AssertExpectations(t)
}

func TestCoreDocument_getReadAccessProofKeys(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	registry := common.HexToAddress("0xf72855759a39fb75fc7341139f5d7a3974d4da08")
	tokenID := utils.RandomSlice(32)

	pfs, err := getReadAccessProofKeys(cd.Document, registry, tokenID)
	assert.Error(t, err)
	assert.Nil(t, pfs)

	cd, err = cd.AddNFT(true, registry, tokenID)
	assert.NoError(t, err)
	assert.NotNil(t, cd)

	pfs, err = getReadAccessProofKeys(cd.Document, registry, tokenID)
	assert.NoError(t, err)
	assert.Len(t, pfs, 3)
	assert.Equal(t, CDTreePrefix+".read_rules[0].roles[0]", pfs[0])
	assert.Equal(t, CDTreePrefix+".read_rules[0].action", pfs[1])
	assert.Equal(t, fmt.Sprintf(CDTreePrefix+".roles[%s].nfts[0]", hexutil.Encode(cd.Document.Roles[0].RoleKey)), pfs[2])
}

func TestCoreDocument_getNFTUniqueProofKey(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	registry := common.HexToAddress("0xf72855759a39fb75fc7341139f5d7a3974d4da08")
	pf, err := getNFTUniqueProofKey(cd.Document.Nfts, registry)
	assert.Error(t, err)
	assert.Empty(t, pf)

	tokenID := utils.RandomSlice(32)
	cd, err = cd.AddNFT(false, registry, tokenID)
	assert.NoError(t, err)
	assert.NotNil(t, cd)

	pf, err = getNFTUniqueProofKey(cd.Document.Nfts, registry)
	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf(CDTreePrefix+".nfts[%s]", hexutil.Encode(append(registry.Bytes(), make([]byte, 12, 12)...))), pf)
}

func TestCoreDocumentModel_GetNFTProofs(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	testTree, err := cd.DefaultTreeWithPrefix("invoice", []byte{1, 0, 0, 0})
	assert.NoError(t, err)
	props := []proofs.Property{NewLeafProperty("invoice.sample_field", []byte{1, 0, 0, 0, 0, 0, 0, 200})}
	err = testTree.AddLeaf(proofs.LeafNode{Hash: utils.RandomSlice(32), Hashed: true, Property: props[0]})
	assert.NoError(t, err)
	cd.GetTestCoreDocWithReset()
	err = testTree.Generate()
	assert.NoError(t, err)
	cd.GetTestCoreDocWithReset().EmbeddedData = &any.Any{Value: utils.RandomSlice(32), TypeUrl: documenttypes.InvoiceDataTypeUrl}

	account := testingidentity.GenerateRandomDID()
	cd.initReadRules([]identity.DID{account})
	registry := common.HexToAddress("0xf72855759a39fb75fc7341139f5d7a3974d4da08")
	tokenID := utils.RandomSlice(32)
	cd, err = cd.AddNFT(true, registry, tokenID)
	assert.NoError(t, err)
	dataRoot := calculateBasicDataRoot(t, cd, documenttypes.InvoiceDataTypeUrl, testTree.GetLeaves())
	_, err = cd.CalculateDocumentRoot(documenttypes.InvoiceDataTypeUrl, testTree.GetLeaves())
	assert.NoError(t, err)

	tests := []struct {
		registry       common.Address
		tokenID        []byte
		nftReadAccess  bool
		nftUniqueProof bool
		error          bool
	}{

		// failed nft unique proof
		{
			nftUniqueProof: true,
			registry:       common.BytesToAddress(utils.RandomSlice(20)),
			error:          true,
		},

		// good nft unique proof
		{
			nftUniqueProof: true,
			registry:       registry,
		},

		// failed read access proof
		{
			nftReadAccess: true,
			registry:      registry,
			tokenID:       utils.RandomSlice(32),
			error:         true,
		},

		// good read access proof
		{
			nftReadAccess: true,
			registry:      registry,
			tokenID:       tokenID,
		},

		// all proofs
		{
			nftUniqueProof: true,
			registry:       registry,
			nftReadAccess:  true,
			tokenID:        tokenID,
		},
	}

	for _, c := range tests {
		pfs, err := cd.CreateNFTProofs(documenttypes.InvoiceDataTypeUrl, testTree.GetLeaves(), account, c.registry, c.tokenID, c.nftUniqueProof, c.nftReadAccess)
		if c.error {
			assert.Error(t, err)
			continue
		}

		assert.NoError(t, err)
		assert.True(t, len(pfs.FieldProofs) > 0)

		h, err := blake2b.New256(nil)
		assert.NoError(t, err)
		for _, pf := range pfs.FieldProofs {
			valid, err := ValidateProof(pf, dataRoot, h, sha3.NewLegacyKeccak256())
			assert.NoError(t, err)
			assert.True(t, valid)
		}
	}
}

func TestCoreDocumentModel_ATOwnerCanRead(t *testing.T) {
	ctx := testingconfig.CreateAccountContext(t, cfg)
	account, _ := contextutil.Account(ctx)
	srv := new(testingcommons.MockIdentityService)
	docSrv := new(MockService)
	id := account.GetIdentityID()
	granteeID, err := identity.NewDIDFromString("0xBAEb33a61f05e6F269f1c4b4CFF91A901B54DaF7")
	assert.NoError(t, err)
	granterID, err := identity.NewDIDFromBytes(id)
	assert.NoError(t, err)
	cd, err := NewCoreDocument(nil, CollaboratorsAccess{ReadWriteCollaborators: []identity.DID{granterID}}, nil)
	assert.NoError(t, err)
	payload := AccessTokenParams{
		Grantee:            hexutil.Encode(granteeID[:]),
		DocumentIdentifier: hexutil.Encode(cd.Document.DocumentIdentifier),
	}
	ncd, err := cd.AddAccessToken(ctx, payload)
	assert.NoError(t, err)
	at := ncd.Document.AccessTokens[0]
	assert.NotNil(t, at)
	// wrong token identifier
	tr := &p2ppb.AccessTokenRequest{
		DelegatingDocumentIdentifier: ncd.Document.DocumentIdentifier,
		AccessTokenId:                []byte("randomtokenID"),
	}
	dr := &p2ppb.GetDocumentRequest{
		DocumentIdentifier: ncd.Document.DocumentIdentifier,
		AccessType:         p2ppb.AccessType_ACCESS_TYPE_ACCESS_TOKEN_VERIFICATION,
		AccessTokenRequest: tr,
	}
	err = ncd.ATGranteeCanRead(ctx, docSrv, srv, dr.AccessTokenRequest.AccessTokenId, dr.DocumentIdentifier, granteeID)
	assert.Error(t, err, "access token not found")
	// invalid signing key
	tr = &p2ppb.AccessTokenRequest{
		DelegatingDocumentIdentifier: ncd.Document.DocumentIdentifier,
		AccessTokenId:                at.Identifier,
	}
	dr.AccessTokenRequest = tr
	srv.On("ValidateKey", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("key not linked to identity")).Once()
	m := new(mockModel)
	docSrv.On("GetVersion", mock.Anything, mock.Anything, mock.Anything).Return(m, nil)
	m.On("Timestamp").Return(time.Now(), nil)
	err = ncd.ATGranteeCanRead(ctx, docSrv, srv, dr.AccessTokenRequest.AccessTokenId, dr.DocumentIdentifier, granteeID)
	assert.Error(t, err)
	// valid key
	srv.On("ValidateKey", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	err = ncd.ATGranteeCanRead(ctx, docSrv, srv, dr.AccessTokenRequest.AccessTokenId, dr.DocumentIdentifier, granteeID)
	assert.NoError(t, err)
}

func TestCoreDocumentModel_AddAccessToken(t *testing.T) {
	m, err := newCoreDocument()
	assert.NoError(t, err)
	ctx := testingconfig.CreateAccountContext(t, cfg)
	account, err := contextutil.Account(ctx)
	assert.NoError(t, err)

	cd := m.Document
	assert.Len(t, cd.AccessTokens, 0)

	// invalid centID format
	payload := AccessTokenParams{
		// invalid grantee format
		Grantee:            "randomCentID",
		DocumentIdentifier: "randomDocID",
	}
	_, err = m.AddAccessToken(ctx, payload)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to construct access token: malformed address provided")
	// invalid centID length
	invalidCentID := utils.RandomSlice(25)
	payload = AccessTokenParams{
		Grantee:            hexutil.Encode(invalidCentID),
		DocumentIdentifier: hexutil.Encode(m.Document.DocumentIdentifier),
	}
	_, err = m.AddAccessToken(ctx, payload)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to construct access token: malformed address provided")
	// invalid docID length
	id := account.GetIdentityID()
	invalidDocID := utils.RandomSlice(33)
	payload = AccessTokenParams{
		Grantee:            hexutil.Encode(id),
		DocumentIdentifier: hexutil.Encode(invalidDocID),
	}

	_, err = m.AddAccessToken(ctx, payload)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to construct access token: invalid identifier length")
	// valid
	payload = AccessTokenParams{
		Grantee:            hexutil.Encode(id),
		DocumentIdentifier: hexutil.Encode(m.Document.DocumentIdentifier),
	}

	ncd, err := m.AddAccessToken(ctx, payload)
	assert.NoError(t, err)
	assert.Len(t, ncd.Document.AccessTokens, 1)
}

func TestCoreDocumentModel_DeleteAccessToken(t *testing.T) {
	m, err := newCoreDocument()
	assert.NoError(t, err)

	ctx := testingconfig.CreateAccountContext(t, cfg)
	account, err := contextutil.Account(ctx)
	assert.NoError(t, err)

	id, err := identity.NewDIDFromBytes(account.GetIdentityID())
	assert.NoError(t, err)
	cd := m.Document
	assert.Len(t, cd.AccessTokens, 0)

	payload := AccessTokenParams{
		Grantee:            id.String(),
		DocumentIdentifier: hexutil.Encode(m.Document.DocumentIdentifier),
	}

	// no access token found
	_, err = m.DeleteAccessToken(id)
	assert.Error(t, err)

	// add access token
	ncd, err := m.AddAccessToken(ctx, payload)
	assert.NoError(t, err)
	assert.Len(t, ncd.Document.AccessTokens, 1)

	// invalid granteeID
	_, err = ncd.DeleteAccessToken(testingidentity.GenerateRandomDID())
	assert.Error(t, err)
	assert.Len(t, ncd.Document.AccessTokens, 1)

	// add second access token, valid deletion
	did := testingidentity.GenerateRandomDID()
	payload.Grantee = hexutil.Encode(did[:])
	updated, err := ncd.AddAccessToken(ctx, payload)
	assert.NoError(t, err)
	assert.Len(t, updated.Document.AccessTokens, 2)

	final, err := updated.DeleteAccessToken(id)
	assert.NoError(t, err)
	assert.Len(t, final.Document.AccessTokens, 1)
	assert.Equal(t, final.Document.AccessTokens[0].Grantee, did[:])
}

func calculateBasicDataRoot(t *testing.T, cd *CoreDocument, docType string, dataLeaves []proofs.LeafNode) []byte {
	trees, _, err := cd.SigningDataTrees(docType, dataLeaves)
	assert.NoError(t, err)
	return trees[0].RootHash()
}

func calculateZKDataRoot(t *testing.T, cd *CoreDocument, docType string, dataLeaves []proofs.LeafNode) []byte {
	trees, _, err := cd.SigningDataTrees(docType, dataLeaves)
	assert.NoError(t, err)
	return trees[1].RootHash()
}

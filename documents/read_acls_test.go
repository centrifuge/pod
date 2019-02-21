// +build unit

package documents

import (
	"fmt"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/invoice"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestReadACLs_initReadRules(t *testing.T) {
	cd := newCoreDocument()
	cd.initReadRules(nil)
	assert.Nil(t, cd.document.Roles)
	assert.Nil(t, cd.document.ReadRules)

	cs := []identity.CentID{identity.RandomCentID()}
	cd.initReadRules(cs)
	assert.Len(t, cd.document.ReadRules, 1)
	assert.Len(t, cd.document.Roles, 1)

	cd.initReadRules(cs)
	assert.Len(t, cd.document.ReadRules, 1)
	assert.Len(t, cd.document.Roles, 1)
}

func TestReadAccessValidator_AccountCanRead(t *testing.T) {
	cd := newCoreDocument()
	account, err := identity.CentIDFromString("0x010203040506")
	assert.NoError(t, err)

	cd.document.DocumentRoot = utils.RandomSlice(32)
	ncd, err := cd.PrepareNewVersion([]string{account.String()}, false)
	assert.NoError(t, err)
	assert.NotNil(t, ncd.document.ReadRules)
	assert.NotNil(t, ncd.document.Roles)

	// account who cant access
	rcid := identity.RandomCentID()
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

func TestCoreDocument_addNFTToReadRules(t *testing.T) {
	cd := newCoreDocument()

	// wrong registry or token format
	registry := common.HexToAddress("0xf72855759a39fb75fc7341139f5d7a3974d4da08")
	tokenID := utils.RandomSlice(34)
	err := cd.addNFTToReadRules(registry, tokenID)
	assert.Error(t, err)
	assert.Nil(t, cd.document.CoredocumentSalts)
	assert.Nil(t, cd.document.ReadRules)
	assert.Nil(t, cd.document.Roles)

	tokenID = utils.RandomSlice(32)
	err = cd.addNFTToReadRules(registry, tokenID)
	assert.NoError(t, err)
	assert.NotNil(t, cd.document.CoredocumentSalts)
	assert.Len(t, cd.document.ReadRules, 1)
	assert.Equal(t, cd.document.ReadRules[0].Action, coredocumentpb.Action_ACTION_READ)
	assert.Len(t, cd.document.Roles, 1)
	enft, err := ConstructNFT(registry, tokenID)
	assert.NoError(t, err)
	assert.Equal(t, enft, cd.document.Roles[0].Nfts[0])
}

func TestCoreDocument_NFTOwnerCanRead(t *testing.T) {
	account, err := identity.CentIDFromString("0x010203040506")
	cd, err := NewCoreDocumentWithCollaborators([]string{account.String()})
	assert.NoError(t, err)
	registry := common.HexToAddress("0xf72855759a39fb75fc7341139f5d7a3974d4da08")

	// account can read
	assert.NoError(t, cd.NFTOwnerCanRead(nil, registry, nil, account))

	// account not in read rules and nft missing
	account, err = identity.CentIDFromString("0x010203040505")
	assert.NoError(t, err)
	tokenID := utils.RandomSlice(32)
	assert.Error(t, cd.NFTOwnerCanRead(nil, registry, tokenID, account))

	tr := mockRegistry{}
	tr.On("OwnerOf", registry, tokenID).Return(nil, errors.New("failed to get owner of")).Once()
	assert.NoError(t, cd.addNFTToReadRules(registry, tokenID))
	assert.Error(t, cd.NFTOwnerCanRead(tr, registry, tokenID, account))
	assert.Contains(t, err, "failed to get owner of")
	tr.AssertExpectations(t)

	// not the same owner
	owner := common.BytesToAddress(utils.RandomSlice(20))
	tr.On("OwnerOf", registry, tokenID).Return(owner, nil).Once()
	assert.Error(t, cd.NFTOwnerCanRead(tr, registry, tokenID, account))
	tr.AssertExpectations(t)

	// TODO(ved): add a successful test once identity v2 is complete
}

func TestCoreDocumentModel_AddNFT(t *testing.T) {
	cd := newCoreDocument()
	cd.document.DocumentRoot = utils.RandomSlice(32)
	registry := common.HexToAddress("0xf72855759a39fb75fc7341139f5d7a3974d4da08")
	registry2 := common.HexToAddress("0xf72855759a39fb75fc7341139f5d7a3974d4da02")
	tokenID := utils.RandomSlice(32)
	assert.Nil(t, cd.document.Nfts)
	assert.Nil(t, cd.document.ReadRules)
	assert.Nil(t, cd.document.Roles)

	cd, err := cd.AddNFT(true, registry, tokenID)
	assert.Nil(t, err)
	assert.Len(t, cd.document.Nfts, 1)
	assert.Len(t, cd.document.Nfts[0].RegistryId, 32)
	assert.Equal(t, tokenID, getStoredNFT(cd.document.Nfts, registry.Bytes()).TokenId)
	assert.Nil(t, getStoredNFT(cd.document.Nfts, registry2.Bytes()))
	assert.Len(t, cd.document.ReadRules, 1)
	assert.Len(t, cd.document.Roles, 1)
	assert.Len(t, cd.document.Roles[0].Nfts, 1)

	tokenID = utils.RandomSlice(32)
	cd.document.DocumentRoot = utils.RandomSlice(32)
	cd, err = cd.AddNFT(true, registry, tokenID)
	assert.Nil(t, err)
	assert.Len(t, cd.document.Nfts, 1)
	assert.Len(t, cd.document.Nfts[0].RegistryId, 32)
	assert.Equal(t, tokenID, getStoredNFT(cd.document.Nfts, registry.Bytes()).TokenId)
	assert.Nil(t, getStoredNFT(cd.document.Nfts, registry2.Bytes()))
	assert.Len(t, cd.document.ReadRules, 2)
	assert.Len(t, cd.document.Roles, 2)
	assert.Len(t, cd.document.Roles[1].Nfts, 1)
}

func TestCoreDocument_IsNFTMinted(t *testing.T) {
	cd := newCoreDocument()
	registry := common.HexToAddress("0xf72855759a39fb75fc7341139f5d7a3974d4da08")
	assert.False(t, cd.IsNFTMinted(nil, registry))

	cd.document.DocumentRoot = utils.RandomSlice(32)
	tokenID := utils.RandomSlice(32)
	owner := common.HexToAddress("0xf72855759a39fb75fc7341139f5d7a3974d4da02")
	cd, err := cd.AddNFT(true, registry, tokenID)
	assert.Nil(t, err)

	tr := new(mockRegistry)
	tr.On("OwnerOf", registry, tokenID).Return(owner, nil).Once()
	assert.True(t, cd.IsNFTMinted(tr, registry))
	tr.AssertExpectations(t)
}

func TestCoreDocument_getReadAccessProofKeys(t *testing.T) {
	cd := newCoreDocument()
	registry := common.HexToAddress("0xf72855759a39fb75fc7341139f5d7a3974d4da08")
	tokenID := utils.RandomSlice(32)

	pfs, err := getReadAccessProofKeys(cd.document, registry, tokenID)
	assert.Error(t, err)
	assert.Nil(t, pfs)

	cd.document.DocumentRoot = utils.RandomSlice(32)
	cd, err = cd.AddNFT(true, registry, tokenID)
	assert.NoError(t, err)
	assert.NotNil(t, cd)

	pfs, err = getReadAccessProofKeys(cd.document, registry, tokenID)
	assert.NoError(t, err)
	assert.Len(t, pfs, 3)
	assert.Equal(t, "read_rules[0].roles[0]", pfs[0])
	assert.Equal(t, fmt.Sprintf("roles[%s].nfts[0]", hexutil.Encode(make([]byte, 32, 32))), pfs[1])
	assert.Equal(t, "read_rules[0].action", pfs[2])
}

func TestCoreDocument_getNFTUniqueProofKey(t *testing.T) {
	cd := newCoreDocument()
	registry := common.HexToAddress("0xf72855759a39fb75fc7341139f5d7a3974d4da08")
	pf, err := getNFTUniqueProofKey(cd.document.Nfts, registry)
	assert.Error(t, err)
	assert.Empty(t, pf)

	cd.document.DocumentRoot = utils.RandomSlice(32)
	tokenID := utils.RandomSlice(32)
	cd, err = cd.AddNFT(false, registry, tokenID)
	assert.NoError(t, err)
	assert.NotNil(t, cd)

	pf, err = getNFTUniqueProofKey(cd.document.Nfts, registry)
	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("nfts[%s]", hexutil.Encode(append(registry.Bytes(), make([]byte, 12, 12)...))), pf)
}

func TestCoreDocument_getRoleProofKey(t *testing.T) {
	cd := newCoreDocument()
	roleKey := make([]byte, 32, 32)
	account := identity.RandomCentID()
	pf, err := getRoleProofKey(cd.document.Roles, roleKey, account)
	assert.Error(t, err)
	assert.Empty(t, pf)

	cd.initReadRules([]identity.CentID{account})
	pf, err = getRoleProofKey(cd.document.Roles, roleKey, identity.RandomCentID())
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(ErrNFTRoleMissing, err))
	assert.Empty(t, pf)

	pf, err = getRoleProofKey(cd.document.Roles, roleKey, account)
	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("roles[%s].collaborators[0]", hexutil.Encode(roleKey)), pf)
}

func TestCoreDocumentModel_GetNFTProofs(t *testing.T) {
	cd := newCoreDocument()
	invData := &invoicepb.InvoiceData{}
	dataSalts, err := GenerateNewSalts(invData, "invoice", []byte{1, 0, 0, 0})
	assert.NoError(t, err)

	cd.document.DataRoot = utils.RandomSlice(32)
	cd.document.EmbeddedData = &any.Any{Value: utils.RandomSlice(32), TypeUrl: documenttypes.InvoiceDataTypeUrl}
	account := identity.RandomCentID()
	cd.initReadRules([]identity.CentID{account})
	registry := common.HexToAddress("0xf72855759a39fb75fc7341139f5d7a3974d4da08")
	tokenID := utils.RandomSlice(32)
	cd, err = cd.AddNFT(true, registry, tokenID)
	assert.NoError(t, err)
	cd.document.EmbeddedDataSalts = ConvertToProtoSalts(dataSalts)
	assert.NoError(t, err)
	assert.NoError(t, cd.setSalts())
	assert.NoError(t, cd.calculateSigningRoot())
	assert.NoError(t, cd.calculateDocumentRoot())

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

	tree, err := cd.documentRootTree()
	assert.NoError(t, err)

	for _, c := range tests {
		pfs, err := cd.GenerateNFTProofs(account, c.registry, c.tokenID, c.nftUniqueProof, c.nftReadAccess)
		if c.error {
			assert.Error(t, err)
			continue
		}

		assert.NoError(t, err)
		assert.True(t, len(pfs) > 0)

		for _, pf := range pfs {
			valid, err := tree.ValidateProof(pf)
			assert.NoError(t, err)
			assert.True(t, valid)
		}
	}

}

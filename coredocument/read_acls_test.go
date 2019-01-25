// +build unit

package coredocument

import (
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestReadACLs_initReadRules(t *testing.T) {
	cd := New()
	err := initReadRules(cd, nil)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(ErrZeroCollaborators, err))

	cs := []identity.CentID{identity.RandomCentID()}
	err = initReadRules(cd, cs)
	assert.NoError(t, err)
	assert.Len(t, cd.ReadRules, 1)
	assert.Len(t, cd.Roles, 1)

	err = initReadRules(cd, cs)
	assert.NoError(t, err)
	assert.Len(t, cd.ReadRules, 1)
	assert.Len(t, cd.Roles, 1)
}

func TestReadAccessValidator_AccountCanRead(t *testing.T) {
	pv := accountValidator()
	account, err := identity.CentIDFromString("0x010203040506")
	assert.NoError(t, err)

	cd, err := NewWithCollaborators([]string{account.String()})
	assert.NoError(t, err)
	assert.NotNil(t, cd.ReadRules)
	assert.NotNil(t, cd.Roles)

	// account who cant access
	rcid := identity.RandomCentID()
	assert.False(t, pv.AccountCanRead(cd, rcid))

	// account can access
	assert.True(t, pv.AccountCanRead(cd, account))
}

func Test_addNFTToReadRules(t *testing.T) {
	// wrong registry or token format
	registry := common.HexToAddress("0xf72855759a39fb75fc7341139f5d7a3974d4da08")
	tokenID := utils.RandomSlice(34)

	err := AddNFTToReadRules(nil, registry, tokenID)
	assert.Error(t, err)

	cd, err := NewWithCollaborators([]string{"0x010203040506"})
	assert.NoError(t, err)
	assert.Len(t, cd.ReadRules, 1)
	assert.Equal(t, cd.ReadRules[0].Action, coredocumentpb.Action_ACTION_READ_SIGN)
	assert.Len(t, cd.Roles, 1)

	tokenID = utils.RandomSlice(32)
	err = AddNFTToReadRules(cd, registry, tokenID)
	assert.NoError(t, err)
	assert.Len(t, cd.ReadRules, 2)
	assert.Equal(t, cd.ReadRules[1].Action, coredocumentpb.Action_ACTION_READ)
	assert.Len(t, cd.Roles, 2)
}

type mockRegistry struct {
	mock.Mock
}

func (m mockRegistry) OwnerOf(registry common.Address, tokenID []byte) (common.Address, error) {
	args := m.Called(registry, tokenID)
	addr, _ := args.Get(0).(common.Address)
	return addr, args.Error(1)
}

func TestReadAccessValidator_NFTOwnerCanRead(t *testing.T) {
	account, err := identity.CentIDFromString("0x010203040506")
	assert.NoError(t, err)

	cd, err := NewWithCollaborators([]string{account.String()})
	assert.NoError(t, err)

	registry := common.HexToAddress("0xf72855759a39fb75fc7341139f5d7a3974d4da08")

	// account can read
	validator := nftValidator(nil)
	err = validator.NFTOwnerCanRead(cd, registry, nil, account)
	assert.NoError(t, err)

	// account not in read rules and nft missing
	account, err = identity.CentIDFromString("0x010203040505")
	assert.NoError(t, err)
	tokenID := utils.RandomSlice(32)
	err = validator.NFTOwnerCanRead(cd, registry, tokenID, account)
	assert.Error(t, err)

	tr := mockRegistry{}
	tr.On("OwnerOf", registry, tokenID).Return(nil, errors.New("failed to get owner of")).Once()
	AddNFTToReadRules(cd, registry, tokenID)
	validator = nftValidator(tr)
	err = validator.NFTOwnerCanRead(cd, registry, tokenID, account)
	assert.Error(t, err)
	assert.Contains(t, err, "failed to get owner of")
	tr.AssertExpectations(t)

	// not the same owner
	owner := common.BytesToAddress(utils.RandomSlice(20))
	tr = mockRegistry{}
	tr.On("OwnerOf", registry, tokenID).Return(owner, nil).Once()
	validator = nftValidator(tr)
	err = validator.NFTOwnerCanRead(cd, registry, tokenID, account)
	assert.Error(t, err)
	tr.AssertExpectations(t)
}

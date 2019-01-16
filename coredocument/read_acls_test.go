// +build unit

package coredocument

import (
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/crypto/secp256k1"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
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

func TestReadAccessValidator_PeerCanRead(t *testing.T) {
	pv := peerValidator()
	peer, err := identity.CentIDFromString("0x010203040506")
	assert.NoError(t, err)

	cd, err := NewWithCollaborators([]string{peer.String()})
	assert.NoError(t, err)
	assert.NotNil(t, cd.ReadRules)
	assert.NotNil(t, cd.Roles)

	// peer who cant access
	rcid := identity.RandomCentID()
	err = pv.PeerCanRead(cd, rcid)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(ErrPeerNotFound, err))

	// peer can access
	assert.NoError(t, pv.PeerCanRead(cd, peer))
}

func Test_addNFTToReadRules(t *testing.T) {
	// wrong registry or token format
	registry := common.HexToAddress("0xf72855759a39fb75fc7341139f5d7a3974d4da08")
	tokenID := utils.RandomSlice(34)

	err := addNFTToReadRules(nil, registry, tokenID)
	assert.Error(t, err)

	cd, err := NewWithCollaborators([]string{"0x010203040506"})
	assert.NoError(t, err)
	assert.Len(t, cd.ReadRules, 1)
	assert.Equal(t, cd.ReadRules[0].Action, coredocumentpb.Action_ACTION_READ_SIGN)
	assert.Len(t, cd.Roles, 1)

	tokenID = utils.RandomSlice(32)
	err = addNFTToReadRules(cd, registry, tokenID)
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
	peer, err := identity.CentIDFromString("0x010203040506")
	assert.NoError(t, err)

	cd, err := NewWithCollaborators([]string{peer.String()})
	assert.NoError(t, err)

	registry := common.HexToAddress("0xf72855759a39fb75fc7341139f5d7a3974d4da08")

	// peer can read
	validator := nftValidator(nil)
	err = validator.NFTOwnerCanRead(cd, registry, nil, "", peer)
	assert.NoError(t, err)

	// peer not in read rules and nft missing
	peer, err = identity.CentIDFromString("0x010203040505")
	assert.NoError(t, err)
	tokenID := utils.RandomSlice(32)
	err = validator.NFTOwnerCanRead(cd, registry, tokenID, "", peer)
	assert.Error(t, err)

	tr := mockRegistry{}
	tr.On("OwnerOf", registry, tokenID).Return(nil, errors.New("failed to get owner of")).Once()
	addNFTToReadRules(cd, registry, tokenID)
	validator = nftValidator(tr)
	err = validator.NFTOwnerCanRead(cd, registry, tokenID, "", peer)
	assert.Error(t, err)
	assert.Contains(t, err, "failed to get owner of")
	tr.AssertExpectations(t)

	c := ctx[bootstrap.BootstrappedConfig].(config.Configuration)
	acc, err := c.GetEthereumAccount("main")
	assert.NoError(t, err)
	key, err := keystore.DecryptKey([]byte(acc.Key), "")
	assert.NoError(t, err)
	msg, err := constructNFT(registry, tokenID)
	assert.NoError(t, err)
	sig, err := secp256k1.SignEthereum(msg, key.PrivateKey.D.Bytes())
	assert.NoError(t, err)
	tr = mockRegistry{}
	tr.On("OwnerOf", registry, tokenID).Return(key.Address, nil).Once()
	validator = nftValidator(tr)
	err = validator.NFTOwnerCanRead(cd, registry, tokenID, hexutil.Encode(sig), peer)
	assert.NoError(t, err)
}

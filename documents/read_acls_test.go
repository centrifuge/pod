//go:build unit

package documents

import (
	"bytes"
	"context"
	"crypto"
	cryptorand "crypto/rand"
	"fmt"
	"math/big"
	"math/rand"
	"testing"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	keystoreType "github.com/centrifuge/chain-custom-types/pkg/keystore"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types/codec"
	configMocks "github.com/centrifuge/pod/config"
	"github.com/centrifuge/pod/contextutil"
	"github.com/centrifuge/pod/crypto/ed25519"
	"github.com/centrifuge/pod/errors"
	v2 "github.com/centrifuge/pod/identity/v2"
	testingcommons "github.com/centrifuge/pod/testingutils/common"
	"github.com/centrifuge/pod/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestReadACLs_initReadRules(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)

	cd.initReadRules(nil)
	assert.Nil(t, cd.Document.Roles)
	assert.Nil(t, cd.Document.ReadRules)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	cs := []*types.AccountID{accountID}
	cd.initReadRules(cs)
	assert.Len(t, cd.Document.ReadRules, 1)
	assert.Len(t, cd.Document.Roles, 1)

	cd.initReadRules(cs)
	assert.Len(t, cd.Document.ReadRules, 1)
	assert.Len(t, cd.Document.Roles, 1)
}

func TestFindReadRole(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)

	signCollab1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	signCollab2, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	readCollab1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	readCollab2, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	signRoleKey := utils.RandomSlice(32)
	readRoleKey := utils.RandomSlice(32)

	readRules := []*coredocumentpb.ReadRule{
		{
			Roles: [][]byte{
				signRoleKey,
			},
			Action: coredocumentpb.Action_ACTION_READ_SIGN,
		},
		{
			Roles: [][]byte{
				readRoleKey,
			},
			Action: coredocumentpb.Action_ACTION_READ,
		},
	}

	readNft := utils.RandomSlice(32)
	signNft := utils.RandomSlice(32)

	readRoles := []*coredocumentpb.Role{
		{
			RoleKey: signRoleKey,
			Collaborators: [][]byte{
				signCollab1.ToBytes(),
				signCollab2.ToBytes(),
			},
			Nfts: [][]byte{
				signNft,
			},
		},
		{
			RoleKey: readRoleKey,
			Collaborators: [][]byte{
				readCollab1.ToBytes(),
				readCollab2.ToBytes(),
			},
			Nfts: [][]byte{
				readNft,
			},
		},
	}

	cd.Document.ReadRules = readRules
	cd.Document.Roles = readRoles

	onRoleFn := func(ruleIndex, roleIndex int, role *coredocumentpb.Role) bool {
		assert.Equal(t, readRoleKey, role.RoleKey)
		assert.Contains(t, role.Collaborators, readCollab1.ToBytes())
		assert.Contains(t, role.Collaborators, readCollab2.ToBytes())
		assert.Contains(t, role.Nfts, readNft)
		return true
	}

	res := findReadRole(
		cd.Document,
		onRoleFn,
		coredocumentpb.Action_ACTION_READ,
	)
	assert.True(t, res)

	onRoleFn = func(ruleIndex, roleIndex int, role *coredocumentpb.Role) bool {
		assert.Equal(t, signRoleKey, role.RoleKey)
		assert.Contains(t, role.Collaborators, signCollab1.ToBytes())
		assert.Contains(t, role.Collaborators, signCollab1.ToBytes())
		assert.Contains(t, role.Nfts, signNft)
		return true
	}

	res = findReadRole(
		cd.Document,
		onRoleFn,
		coredocumentpb.Action_ACTION_READ_SIGN,
	)
	assert.True(t, res)

	res = findReadRole(
		cd.Document,
		func(ruleIndex, roleIndex int, role *coredocumentpb.Role) bool {
			switch {
			case utils.IsSameByteSlice(role.GetRoleKey(), signRoleKey):
				assert.Contains(t, role.Collaborators, signCollab1.ToBytes())
				assert.Contains(t, role.Collaborators, signCollab1.ToBytes())
				assert.Contains(t, role.Nfts, signNft)
			case utils.IsSameByteSlice(role.GetRoleKey(), readRoleKey):
				assert.Contains(t, role.Collaborators, readCollab1.ToBytes())
				assert.Contains(t, role.Collaborators, readCollab2.ToBytes())
				assert.Contains(t, role.Nfts, readNft)
			default:
				t.Errorf("Invalid role found with key - %s", string(role.GetRoleKey()))
			}
			return false
		},
		coredocumentpb.Action_ACTION_READ,
		coredocumentpb.Action_ACTION_READ_SIGN,
	)
	assert.False(t, res)
}

func TestFindTransitionRule(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	assert.NotNil(t, cd)

	editCollab1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	editCollab2, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	computeCollab1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	computeCollab2, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	editRoleKey := utils.RandomSlice(32)
	computeRoleKey := utils.RandomSlice(32)

	transitionRule := []*coredocumentpb.TransitionRule{
		{
			RuleKey: utils.RandomSlice(32),
			Roles: [][]byte{
				editRoleKey,
			},
			MatchType: 0,
			Field:     utils.RandomSlice(32),
			Action:    coredocumentpb.TransitionAction_TRANSITION_ACTION_EDIT,
		},
		{
			RuleKey: utils.RandomSlice(32),
			Roles: [][]byte{
				computeRoleKey,
			},
			MatchType: 0,
			Field:     utils.RandomSlice(32),
			Action:    coredocumentpb.TransitionAction_TRANSITION_ACTION_COMPUTE,
		},
	}

	editRoleNFT := utils.RandomSlice(32)
	computeRoleNFT := utils.RandomSlice(32)

	editRoles := []*coredocumentpb.Role{
		{
			RoleKey: editRoleKey,
			Collaborators: [][]byte{
				editCollab1.ToBytes(),
				editCollab2.ToBytes(),
			},
			Nfts: [][]byte{
				editRoleNFT,
			},
		},
		{
			RoleKey: computeRoleKey,
			Collaborators: [][]byte{
				computeCollab1.ToBytes(),
				computeCollab2.ToBytes(),
			},
			Nfts: [][]byte{
				computeRoleNFT,
			},
		},
	}

	cd.Document.TransitionRules = transitionRule
	cd.Document.Roles = editRoles

	onRoleFn := func(rridx, ridx int, role *coredocumentpb.Role) bool {
		assert.Equal(t, editRoleKey, role.RoleKey)
		assert.Contains(t, role.Collaborators, editCollab1.ToBytes())
		assert.Contains(t, role.Collaborators, editCollab2.ToBytes())
		assert.Contains(t, role.Nfts, editRoleNFT)
		return true
	}

	res := findTransitionRole(
		cd.Document,
		onRoleFn,
		coredocumentpb.TransitionAction_TRANSITION_ACTION_EDIT,
	)
	assert.True(t, res)

	onRoleFn = func(rridx, ridx int, role *coredocumentpb.Role) bool {
		assert.Equal(t, computeRoleKey, role.RoleKey)
		assert.Contains(t, role.Collaborators, computeCollab1.ToBytes())
		assert.Contains(t, role.Collaborators, computeCollab2.ToBytes())
		assert.Contains(t, role.Nfts, computeRoleNFT)
		return true
	}

	res = findTransitionRole(
		cd.Document,
		onRoleFn,
		coredocumentpb.TransitionAction_TRANSITION_ACTION_COMPUTE,
	)
	assert.True(t, res)

	onRoleFn = func(rridx, ridx int, role *coredocumentpb.Role) bool {
		switch {
		case utils.IsSameByteSlice(editRoleKey, role.RoleKey):
			assert.Contains(t, role.Collaborators, editCollab1.ToBytes())
			assert.Contains(t, role.Collaborators, editCollab2.ToBytes())
			assert.Contains(t, role.Nfts, editRoleNFT)
		case utils.IsSameByteSlice(computeRoleKey, role.RoleKey):
			assert.Contains(t, role.Collaborators, computeCollab1.ToBytes())
			assert.Contains(t, role.Collaborators, computeCollab2.ToBytes())
			assert.Contains(t, role.Nfts, computeRoleNFT)
		default:
			t.Errorf("Invalid role found with key - %s", string(role.GetRoleKey()))
		}
		return false
	}

	res = findTransitionRole(
		cd.Document,
		onRoleFn,
		coredocumentpb.TransitionAction_TRANSITION_ACTION_EDIT,
		coredocumentpb.TransitionAction_TRANSITION_ACTION_COMPUTE,
	)
	assert.False(t, res)
}

func TestCoreDocument_NFTCanRead(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)

	readRoleKey1 := utils.RandomSlice(32)
	readRoleKey2 := utils.RandomSlice(32)

	readRules := []*coredocumentpb.ReadRule{
		{
			Roles: [][]byte{
				readRoleKey1,
				readRoleKey2,
			},
			Action: coredocumentpb.Action_ACTION_READ,
		},
	}

	nftRegistryID := utils.RandomSlice(8)
	nftTokenID := utils.RandomSlice(16)

	readNft := append(nftRegistryID, nftTokenID...)

	readRoles := []*coredocumentpb.Role{
		{
			RoleKey:       readRoleKey1,
			Collaborators: [][]byte{},
			Nfts: [][]byte{
				utils.RandomSlice(64),
			},
		},
		{
			RoleKey:       readRoleKey2,
			Collaborators: [][]byte{},
			Nfts: [][]byte{
				readNft,
			},
		},
	}

	cd.Document.ReadRules = readRules
	cd.Document.Roles = readRoles

	res := cd.NFTCanRead(utils.RandomSlice(8), utils.RandomSlice(16))
	assert.False(t, res)

	res = cd.NFTCanRead(nftRegistryID, nftTokenID)
	assert.True(t, res)
}

func TestCoreDocument_AccountCanRead(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)

	signCollab1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	signCollab2, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	readCollab1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	readCollab2, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	signRoleKey := utils.RandomSlice(32)
	readRoleKey := utils.RandomSlice(32)

	readRules := []*coredocumentpb.ReadRule{
		{
			Roles: [][]byte{
				signRoleKey,
			},
			Action: coredocumentpb.Action_ACTION_READ_SIGN,
		},
		{
			Roles: [][]byte{
				readRoleKey,
			},
			Action: coredocumentpb.Action_ACTION_READ,
		},
	}

	readNft := utils.RandomSlice(32)
	signNft := utils.RandomSlice(32)

	readRoles := []*coredocumentpb.Role{
		{
			RoleKey: signRoleKey,
			Collaborators: [][]byte{
				signCollab1.ToBytes(),
				signCollab2.ToBytes(),
			},
			Nfts: [][]byte{
				signNft,
			},
		},
		{
			RoleKey: readRoleKey,
			Collaborators: [][]byte{
				readCollab1.ToBytes(),
				readCollab2.ToBytes(),
			},
			Nfts: [][]byte{
				readNft,
			},
		},
	}

	cd.Document.ReadRules = readRules
	cd.Document.Roles = readRoles

	res := cd.AccountCanRead(signCollab1)
	assert.True(t, res)
	res = cd.AccountCanRead(signCollab2)
	assert.True(t, res)
	res = cd.AccountCanRead(readCollab1)
	assert.True(t, res)
	res = cd.AccountCanRead(readCollab2)
	assert.True(t, res)

	randomAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	res = cd.AccountCanRead(randomAccountID)
	assert.False(t, res)
}

func TestCoreDocument_AddNFT(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)

	collectionID := types.U64(rand.Uint64())
	itemID := types.NewU128(*big.NewInt(rand.Int63()))

	encodedCollectionID, err := codec.Encode(collectionID)
	encodedItemID, err := codec.Encode(itemID)

	expectedNFT := &coredocumentpb.NFT{
		CollectionId: encodedCollectionID,
		ItemId:       encodedItemID,
	}

	cd, err = cd.AddNFT(false, collectionID, itemID)
	assert.NoError(t, err)
	assert.Len(t, cd.Document.Nfts, 1)
	assert.Contains(t, cd.Document.Nfts, expectedNFT)
	assert.Len(t, cd.Document.ReadRules, 0)
	assert.Len(t, cd.Document.Roles, 0)

	cd, err = cd.AddNFT(false, collectionID, itemID)
	assert.NotNil(t, err)
	assert.Nil(t, cd)

	cd, err = newCoreDocument()
	assert.NoError(t, err)

	encodedNFT := append(encodedCollectionID, encodedItemID...)

	cd, err = cd.AddNFT(true, collectionID, itemID)
	assert.NoError(t, err)
	assert.Len(t, cd.Document.Nfts, 1)
	assert.Contains(t, cd.Document.Nfts, expectedNFT)
	assert.Len(t, cd.Document.ReadRules, 1)
	assert.Len(t, cd.Document.Roles, 1)
	assert.Len(t, cd.Document.Roles[0].Nfts, 1)
	assert.Contains(t, cd.Document.Roles[0].Nfts, encodedNFT)
}

func TestCoreDocument_AddNFT_ExistingNFT(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)

	collectionID := types.U64(rand.Uint64())
	itemID := types.NewU128(*big.NewInt(rand.Int63()))

	encodedCollectionID, err := codec.Encode(collectionID)
	encodedItemID, err := codec.Encode(itemID)

	cd.Document.Nfts = append(cd.Document.Nfts, &coredocumentpb.NFT{
		CollectionId: encodedCollectionID,
		ItemId:       utils.RandomSlice(16),
	})

	expectedNFT := &coredocumentpb.NFT{
		CollectionId: encodedCollectionID,
		ItemId:       encodedItemID,
	}

	res, err := cd.AddNFT(false, collectionID, itemID)
	assert.NoError(t, err)
	assert.Len(t, res.Document.Nfts, 1)
	assert.Contains(t, res.Document.Nfts, expectedNFT)
	assert.Len(t, res.Document.ReadRules, 0)
	assert.Len(t, res.Document.Roles, 0)

	cd, err = newCoreDocument()
	assert.NoError(t, err)

	cd.Document.Nfts = append(cd.Document.Nfts, &coredocumentpb.NFT{
		CollectionId: encodedCollectionID,
		ItemId:       utils.RandomSlice(16),
	})

	encodedNFT := append(encodedCollectionID, encodedItemID...)

	res, err = cd.AddNFT(true, collectionID, itemID)
	assert.NoError(t, err)
	assert.Len(t, res.Document.Nfts, 1)
	assert.Contains(t, res.Document.Nfts, expectedNFT)
	assert.Len(t, res.Document.ReadRules, 1)
	assert.Len(t, res.Document.Roles, 1)
	assert.Len(t, res.Document.Roles[0].Nfts, 1)
	assert.Contains(t, res.Document.Roles[0].Nfts, encodedNFT)
}

func TestCoreDocument_AddNFT_PrepareNewVersionError(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)

	readRoleKey := utils.RandomSlice(32)

	readRules := []*coredocumentpb.ReadRule{
		{
			Roles: [][]byte{
				readRoleKey,
			},
			Action: coredocumentpb.Action_ACTION_READ,
		},
	}

	readNft := utils.RandomSlice(24)

	readRoles := []*coredocumentpb.Role{
		{
			RoleKey: readRoleKey,
			Collaborators: [][]byte{
				[]byte("invalid-account-id-bytes"),
			},
			Nfts: [][]byte{
				readNft,
			},
		},
	}

	cd.Document.ReadRules = readRules
	cd.Document.Roles = readRoles

	collectionID := types.U64(rand.Uint64())
	itemID := types.NewU128(*big.NewInt(rand.Int63()))

	res, err := cd.AddNFT(false, collectionID, itemID)
	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestCoreDocument_NFTs(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)

	expectedNFTs := []*coredocumentpb.NFT{
		{
			CollectionId: utils.RandomSlice(nftCollectionIDByteCount),
			ItemId:       utils.RandomSlice(nftItemIDByteCount),
		},
		{
			CollectionId: utils.RandomSlice(nftCollectionIDByteCount),
			ItemId:       utils.RandomSlice(nftItemIDByteCount),
		},
	}

	cd.Document.Nfts = expectedNFTs

	assert.Equal(t, cd.NFTs(), expectedNFTs)
}

func TestCoreDocument_ATGranteeCanRead(t *testing.T) {
	ctx := context.Background()
	documentServiceMock := NewServiceMock(t)
	identityServiceMock := v2.NewServiceMock(t)

	cd, err := newCoreDocument()
	assert.NoError(t, err)

	pubKey, privKey, err := ed25519.GenerateSigningKeyPair()
	assert.NoError(t, err)

	requesterID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	granterID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	tokenIdentifier := utils.RandomSlice(32)
	documentIdentifier := utils.RandomSlice(32)
	documentVersion := utils.RandomSlice(32)

	token := &coredocumentpb.AccessToken{
		Identifier:         tokenIdentifier,
		Grantee:            requesterID.ToBytes(),
		Granter:            granterID.ToBytes(),
		RoleIdentifier:     utils.RandomSlice(32),
		DocumentIdentifier: documentIdentifier,
		DocumentVersion:    documentVersion,
		Key:                pubKey,
	}

	tokenMessage, err := AssembleTokenMessage(
		token.Identifier,
		granterID,
		requesterID,
		token.RoleIdentifier,
		token.DocumentIdentifier,
		token.DocumentVersion,
	)
	assert.NoError(t, err)

	signature, err := privKey.Sign(cryptorand.Reader, tokenMessage, crypto.Hash(0))
	assert.NoError(t, err)

	token.Signature = signature

	cd.Document.AccessTokens = []*coredocumentpb.AccessToken{token}

	readRoleKey := utils.RandomSlice(32)

	readRules := []*coredocumentpb.ReadRule{
		{
			Roles: [][]byte{
				readRoleKey,
			},
			Action: coredocumentpb.Action_ACTION_READ,
		},
	}

	readRoles := []*coredocumentpb.Role{
		{
			RoleKey: readRoleKey,
			Collaborators: [][]byte{
				granterID.ToBytes(),
			},
		},
	}

	cd.Document.ReadRules = readRules
	cd.Document.Roles = readRoles

	documentServiceMock.On(
		"GetVersion",
		ctx,
		cd.Document.DocumentIdentifier,
		token.DocumentVersion,
	).Return(nil, nil)

	timestamp := timestamppb.Now()
	cd.Document.Timestamp = timestamp

	identityServiceMock.On(
		"ValidateKey",
		granterID,
		token.Key,
		keystoreType.KeyPurposeP2PDocumentSigning,
		timestamp.AsTime(),
	).Return(nil)

	err = cd.ATGranteeCanRead(
		ctx,
		documentServiceMock,
		identityServiceMock,
		tokenIdentifier,
		documentIdentifier,
		requesterID,
	)
	assert.NoError(t, err)
}

func TestCoreDocument_ATGranteeCanRead_AccessTokenNotFound(t *testing.T) {
	ctx := context.Background()
	documentServiceMock := NewServiceMock(t)
	identityServiceMock := v2.NewServiceMock(t)

	cd, err := newCoreDocument()
	assert.NoError(t, err)

	requesterID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	tokenIdentifier := utils.RandomSlice(32)
	documentIdentifier := utils.RandomSlice(32)

	err = cd.ATGranteeCanRead(
		ctx,
		documentServiceMock,
		identityServiceMock,
		tokenIdentifier,
		documentIdentifier,
		requesterID,
	)
	assert.ErrorIs(t, err, ErrAccessTokenNotFound)
}

func TestCoreDocument_ATGranteeCanRead_GranterAccountIDError(t *testing.T) {
	ctx := context.Background()
	documentServiceMock := NewServiceMock(t)
	identityServiceMock := v2.NewServiceMock(t)

	cd, err := newCoreDocument()
	assert.NoError(t, err)

	pubKey, _, err := ed25519.GenerateSigningKeyPair()
	assert.NoError(t, err)

	requesterID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	tokenIdentifier := utils.RandomSlice(32)
	documentIdentifier := utils.RandomSlice(32)
	documentVersion := utils.RandomSlice(32)

	token := &coredocumentpb.AccessToken{
		Identifier:         tokenIdentifier,
		Grantee:            requesterID.ToBytes(),
		Granter:            []byte("invalid-account-id-bytes"),
		RoleIdentifier:     utils.RandomSlice(32),
		DocumentIdentifier: documentIdentifier,
		DocumentVersion:    documentVersion,
		Key:                pubKey,
	}

	cd.Document.AccessTokens = []*coredocumentpb.AccessToken{token}

	err = cd.ATGranteeCanRead(
		ctx,
		documentServiceMock,
		identityServiceMock,
		tokenIdentifier,
		documentIdentifier,
		requesterID,
	)
	assert.ErrorIs(t, err, ErrGranterInvalidAccountID)
}

func TestCoreDocument_ATGranteeCanRead_GranteeAccountIDError(t *testing.T) {
	ctx := context.Background()
	documentServiceMock := NewServiceMock(t)
	identityServiceMock := v2.NewServiceMock(t)

	cd, err := newCoreDocument()
	assert.NoError(t, err)

	pubKey, _, err := ed25519.GenerateSigningKeyPair()
	assert.NoError(t, err)

	requesterID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	granterID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	tokenIdentifier := utils.RandomSlice(32)
	documentIdentifier := utils.RandomSlice(32)
	documentVersion := utils.RandomSlice(32)

	token := &coredocumentpb.AccessToken{
		Identifier:         tokenIdentifier,
		Grantee:            []byte("invalid-account-id-bytes"),
		Granter:            granterID.ToBytes(),
		RoleIdentifier:     utils.RandomSlice(32),
		DocumentIdentifier: documentIdentifier,
		DocumentVersion:    documentVersion,
		Key:                pubKey,
	}

	cd.Document.AccessTokens = []*coredocumentpb.AccessToken{token}

	err = cd.ATGranteeCanRead(
		ctx,
		documentServiceMock,
		identityServiceMock,
		tokenIdentifier,
		documentIdentifier,
		requesterID,
	)
	assert.ErrorIs(t, err, ErrGranteeInvalidAccountID)
}

func TestCoreDocument_ATGranteeCanRead_RequesterNotGrantee(t *testing.T) {
	ctx := context.Background()
	documentServiceMock := NewServiceMock(t)
	identityServiceMock := v2.NewServiceMock(t)

	cd, err := newCoreDocument()
	assert.NoError(t, err)

	pubKey, _, err := ed25519.GenerateSigningKeyPair()
	assert.NoError(t, err)

	requesterID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	granterID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	tokenIdentifier := utils.RandomSlice(32)
	documentIdentifier := utils.RandomSlice(32)
	documentVersion := utils.RandomSlice(32)

	token := &coredocumentpb.AccessToken{
		Identifier: tokenIdentifier,
		// Grantee account ID is not the requester account ID.
		Grantee:            utils.RandomSlice(32),
		Granter:            granterID.ToBytes(),
		RoleIdentifier:     utils.RandomSlice(32),
		DocumentIdentifier: documentIdentifier,
		DocumentVersion:    documentVersion,
		Key:                pubKey,
	}

	cd.Document.AccessTokens = []*coredocumentpb.AccessToken{token}

	err = cd.ATGranteeCanRead(
		ctx,
		documentServiceMock,
		identityServiceMock,
		tokenIdentifier,
		documentIdentifier,
		requesterID,
	)
	assert.ErrorIs(t, err, ErrRequesterNotGrantee)
}

func TestCoreDocument_ATGranteeCanRead_GranterNotCollab(t *testing.T) {
	ctx := context.Background()
	documentServiceMock := NewServiceMock(t)
	identityServiceMock := v2.NewServiceMock(t)

	cd, err := newCoreDocument()
	assert.NoError(t, err)

	pubKey, _, err := ed25519.GenerateSigningKeyPair()
	assert.NoError(t, err)

	requesterID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	granterID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	tokenIdentifier := utils.RandomSlice(32)
	documentIdentifier := utils.RandomSlice(32)
	documentVersion := utils.RandomSlice(32)

	token := &coredocumentpb.AccessToken{
		Identifier:         tokenIdentifier,
		Grantee:            requesterID.ToBytes(),
		Granter:            granterID.ToBytes(),
		RoleIdentifier:     utils.RandomSlice(32),
		DocumentIdentifier: documentIdentifier,
		DocumentVersion:    documentVersion,
		Key:                pubKey,
	}

	cd.Document.AccessTokens = []*coredocumentpb.AccessToken{token}

	err = cd.ATGranteeCanRead(
		ctx,
		documentServiceMock,
		identityServiceMock,
		tokenIdentifier,
		documentIdentifier,
		requesterID,
	)
	assert.ErrorIs(t, err, ErrGranterNotCollab)
}

func TestCoreDocument_ATGranteeCanRead_DocumentIDMismatch(t *testing.T) {
	ctx := context.Background()
	documentServiceMock := NewServiceMock(t)
	identityServiceMock := v2.NewServiceMock(t)

	cd, err := newCoreDocument()
	assert.NoError(t, err)

	pubKey, _, err := ed25519.GenerateSigningKeyPair()
	assert.NoError(t, err)

	requesterID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	granterID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	tokenIdentifier := utils.RandomSlice(32)
	documentIdentifier := utils.RandomSlice(32)
	documentVersion := utils.RandomSlice(32)

	token := &coredocumentpb.AccessToken{
		Identifier:         tokenIdentifier,
		Grantee:            requesterID.ToBytes(),
		Granter:            granterID.ToBytes(),
		RoleIdentifier:     utils.RandomSlice(32),
		DocumentIdentifier: utils.RandomSlice(32),
		DocumentVersion:    documentVersion,
		Key:                pubKey,
	}

	cd.Document.AccessTokens = []*coredocumentpb.AccessToken{token}

	readRoleKey := utils.RandomSlice(32)

	readRules := []*coredocumentpb.ReadRule{
		{
			Roles: [][]byte{
				readRoleKey,
			},
			Action: coredocumentpb.Action_ACTION_READ,
		},
	}

	readRoles := []*coredocumentpb.Role{
		{
			RoleKey: readRoleKey,
			Collaborators: [][]byte{
				granterID.ToBytes(),
			},
		},
	}

	cd.Document.ReadRules = readRules
	cd.Document.Roles = readRoles

	err = cd.ATGranteeCanRead(
		ctx,
		documentServiceMock,
		identityServiceMock,
		tokenIdentifier,
		documentIdentifier,
		requesterID,
	)
	assert.ErrorIs(t, err, ErrReqDocNotMatch)
}

func TestCoreDocument_ATGranteeCanRead_DocumentServiceError(t *testing.T) {
	ctx := context.Background()
	documentServiceMock := NewServiceMock(t)
	identityServiceMock := v2.NewServiceMock(t)

	cd, err := newCoreDocument()
	assert.NoError(t, err)

	pubKey, privKey, err := ed25519.GenerateSigningKeyPair()
	assert.NoError(t, err)

	requesterID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	granterID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	tokenIdentifier := utils.RandomSlice(32)
	documentIdentifier := utils.RandomSlice(32)
	documentVersion := utils.RandomSlice(32)

	token := &coredocumentpb.AccessToken{
		Identifier:         tokenIdentifier,
		Grantee:            requesterID.ToBytes(),
		Granter:            granterID.ToBytes(),
		RoleIdentifier:     utils.RandomSlice(32),
		DocumentIdentifier: documentIdentifier,
		DocumentVersion:    documentVersion,
		Key:                pubKey,
	}

	tokenMessage, err := AssembleTokenMessage(
		token.Identifier,
		granterID,
		requesterID,
		token.RoleIdentifier,
		token.DocumentIdentifier,
		token.DocumentVersion,
	)
	assert.NoError(t, err)

	signature, err := privKey.Sign(cryptorand.Reader, tokenMessage, crypto.Hash(0))
	assert.NoError(t, err)

	token.Signature = signature

	cd.Document.AccessTokens = []*coredocumentpb.AccessToken{token}

	readRoleKey := utils.RandomSlice(32)

	readRules := []*coredocumentpb.ReadRule{
		{
			Roles: [][]byte{
				readRoleKey,
			},
			Action: coredocumentpb.Action_ACTION_READ,
		},
	}

	readRoles := []*coredocumentpb.Role{
		{
			RoleKey: readRoleKey,
			Collaborators: [][]byte{
				granterID.ToBytes(),
			},
		},
	}

	cd.Document.ReadRules = readRules
	cd.Document.Roles = readRoles

	documentServiceMock.On(
		"GetVersion",
		ctx,
		cd.Document.DocumentIdentifier,
		token.DocumentVersion,
	).Return(nil, errors.New("error"))

	err = cd.ATGranteeCanRead(
		ctx,
		documentServiceMock,
		identityServiceMock,
		tokenIdentifier,
		documentIdentifier,
		requesterID,
	)
	assert.ErrorIs(t, err, ErrDocumentRetrieval)
}

func TestCoreDocument_ATGranteeCanRead_DocumentTimestampError(t *testing.T) {
	ctx := context.Background()
	documentServiceMock := NewServiceMock(t)
	identityServiceMock := v2.NewServiceMock(t)

	cd, err := newCoreDocument()
	assert.NoError(t, err)

	pubKey, privKey, err := ed25519.GenerateSigningKeyPair()
	assert.NoError(t, err)

	requesterID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	granterID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	tokenIdentifier := utils.RandomSlice(32)
	documentIdentifier := utils.RandomSlice(32)
	documentVersion := utils.RandomSlice(32)

	token := &coredocumentpb.AccessToken{
		Identifier:         tokenIdentifier,
		Grantee:            requesterID.ToBytes(),
		Granter:            granterID.ToBytes(),
		RoleIdentifier:     utils.RandomSlice(32),
		DocumentIdentifier: documentIdentifier,
		DocumentVersion:    documentVersion,
		Key:                pubKey,
	}

	tokenMessage, err := AssembleTokenMessage(
		token.Identifier,
		granterID,
		requesterID,
		token.RoleIdentifier,
		token.DocumentIdentifier,
		token.DocumentVersion,
	)
	assert.NoError(t, err)

	signature, err := privKey.Sign(cryptorand.Reader, tokenMessage, crypto.Hash(0))
	assert.NoError(t, err)

	token.Signature = signature

	cd.Document.AccessTokens = []*coredocumentpb.AccessToken{token}

	readRoleKey := utils.RandomSlice(32)

	readRules := []*coredocumentpb.ReadRule{
		{
			Roles: [][]byte{
				readRoleKey,
			},
			Action: coredocumentpb.Action_ACTION_READ,
		},
	}

	readRoles := []*coredocumentpb.Role{
		{
			RoleKey: readRoleKey,
			Collaborators: [][]byte{
				granterID.ToBytes(),
			},
		},
	}

	cd.Document.ReadRules = readRules
	cd.Document.Roles = readRoles

	documentServiceMock.On(
		"GetVersion",
		ctx,
		cd.Document.DocumentIdentifier,
		token.DocumentVersion,
	).Return(nil, nil)

	err = cd.ATGranteeCanRead(
		ctx,
		documentServiceMock,
		identityServiceMock,
		tokenIdentifier,
		documentIdentifier,
		requesterID,
	)
	assert.ErrorIs(t, err, ErrDocumentTimestampRetrieval)
}

func TestCoreDocument_ATGranteeCanRead_IdentityServiceError(t *testing.T) {
	ctx := context.Background()
	documentServiceMock := NewServiceMock(t)
	identityServiceMock := v2.NewServiceMock(t)

	cd, err := newCoreDocument()
	assert.NoError(t, err)

	pubKey, privKey, err := ed25519.GenerateSigningKeyPair()
	assert.NoError(t, err)

	requesterID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	granterID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	tokenIdentifier := utils.RandomSlice(32)
	documentIdentifier := utils.RandomSlice(32)
	documentVersion := utils.RandomSlice(32)

	token := &coredocumentpb.AccessToken{
		Identifier:         tokenIdentifier,
		Grantee:            requesterID.ToBytes(),
		Granter:            granterID.ToBytes(),
		RoleIdentifier:     utils.RandomSlice(32),
		DocumentIdentifier: documentIdentifier,
		DocumentVersion:    documentVersion,
		Key:                pubKey,
	}

	tokenMessage, err := AssembleTokenMessage(
		token.Identifier,
		granterID,
		requesterID,
		token.RoleIdentifier,
		token.DocumentIdentifier,
		token.DocumentVersion,
	)
	assert.NoError(t, err)

	signature, err := privKey.Sign(cryptorand.Reader, tokenMessage, crypto.Hash(0))
	assert.NoError(t, err)

	token.Signature = signature

	cd.Document.AccessTokens = []*coredocumentpb.AccessToken{token}

	readRoleKey := utils.RandomSlice(32)

	readRules := []*coredocumentpb.ReadRule{
		{
			Roles: [][]byte{
				readRoleKey,
			},
			Action: coredocumentpb.Action_ACTION_READ,
		},
	}

	readRoles := []*coredocumentpb.Role{
		{
			RoleKey: readRoleKey,
			Collaborators: [][]byte{
				granterID.ToBytes(),
			},
		},
	}

	cd.Document.ReadRules = readRules
	cd.Document.Roles = readRoles

	documentServiceMock.On(
		"GetVersion",
		ctx,
		cd.Document.DocumentIdentifier,
		token.DocumentVersion,
	).Return(nil, nil)

	timestamp := timestamppb.Now()
	cd.Document.Timestamp = timestamp

	identityServiceMock.On(
		"ValidateKey",
		granterID,
		token.Key,
		keystoreType.KeyPurposeP2PDocumentSigning,
		timestamp.AsTime(),
	).Return(errors.New("error"))

	err = cd.ATGranteeCanRead(
		ctx,
		documentServiceMock,
		identityServiceMock,
		tokenIdentifier,
		documentIdentifier,
		requesterID,
	)
	assert.ErrorIs(t, err, ErrDocumentSigningKeyValidation)
}

func TestCoreDocument_ATGranteeCanRead_InvalidSignature(t *testing.T) {
	ctx := context.Background()
	documentServiceMock := NewServiceMock(t)
	identityServiceMock := v2.NewServiceMock(t)

	cd, err := newCoreDocument()
	assert.NoError(t, err)

	pubKey, _, err := ed25519.GenerateSigningKeyPair()
	assert.NoError(t, err)

	requesterID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	granterID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	tokenIdentifier := utils.RandomSlice(32)
	documentIdentifier := utils.RandomSlice(32)
	documentVersion := utils.RandomSlice(32)

	token := &coredocumentpb.AccessToken{
		Identifier:         tokenIdentifier,
		Grantee:            requesterID.ToBytes(),
		Granter:            granterID.ToBytes(),
		RoleIdentifier:     utils.RandomSlice(32),
		DocumentIdentifier: documentIdentifier,
		DocumentVersion:    documentVersion,
		Key:                pubKey,
		Signature:          []byte("invalid-signature"),
	}

	cd.Document.AccessTokens = []*coredocumentpb.AccessToken{token}

	readRoleKey := utils.RandomSlice(32)

	readRules := []*coredocumentpb.ReadRule{
		{
			Roles: [][]byte{
				readRoleKey,
			},
			Action: coredocumentpb.Action_ACTION_READ,
		},
	}

	readRoles := []*coredocumentpb.Role{
		{
			RoleKey: readRoleKey,
			Collaborators: [][]byte{
				granterID.ToBytes(),
			},
		},
	}

	cd.Document.ReadRules = readRules
	cd.Document.Roles = readRoles

	documentServiceMock.On(
		"GetVersion",
		ctx,
		cd.Document.DocumentIdentifier,
		token.DocumentVersion,
	).Return(nil, nil)

	timestamp := timestamppb.Now()
	cd.Document.Timestamp = timestamp

	identityServiceMock.On(
		"ValidateKey",
		granterID,
		token.Key,
		keystoreType.KeyPurposeP2PDocumentSigning,
		timestamp.AsTime(),
	).Return(nil)

	err = cd.ATGranteeCanRead(
		ctx,
		documentServiceMock,
		identityServiceMock,
		tokenIdentifier,
		documentIdentifier,
		requesterID,
	)
	assert.ErrorIs(t, err, ErrAccessTokenInvalid)
}

func TestCoreDocument_AddAccessToken(t *testing.T) {
	ctx := context.Background()

	accountMock := configMocks.NewAccountMock(t)

	granterAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	ctx = contextutil.WithAccount(ctx, accountMock)

	cd, err := newCoreDocument()
	assert.NoError(t, err)

	granteeAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	documentIdentifier := utils.RandomSlice(32)

	payload := AccessTokenParams{
		Grantee:            granteeAccountID.ToHexString(),
		DocumentIdentifier: hexutil.Encode(documentIdentifier),
	}

	accountMock.On("GetIdentity").
		Return(granterAccountID)

	signature := &coredocumentpb.Signature{
		SignatureId:         utils.RandomSlice(32),
		SignerId:            utils.RandomSlice(32),
		PublicKey:           utils.RandomSlice(32),
		Signature:           utils.RandomSlice(32),
		TransitionValidated: true,
	}

	accountMock.On("SignMsg", mock.Anything).
		Return(signature, nil)

	res, err := cd.AddAccessToken(ctx, payload)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Len(t, res.Document.AccessTokens, 1)
	assert.Len(t, res.Document.AccessTokens[0].Identifier, 32)
	assert.Len(t, res.Document.AccessTokens[0].RoleIdentifier, 32)
	assert.Equal(t, res.Document.AccessTokens[0].Granter, granterAccountID.ToBytes())
	assert.Equal(t, res.Document.AccessTokens[0].Grantee, granteeAccountID.ToBytes())
	assert.Equal(t, res.Document.AccessTokens[0].DocumentIdentifier, documentIdentifier)
	assert.Equal(t, res.Document.AccessTokens[0].Signature, signature.Signature)
	assert.Equal(t, res.Document.AccessTokens[0].Key, signature.PublicKey)
	assert.Equal(t, res.Document.AccessTokens[0].DocumentVersion, cd.CurrentVersion())
}

func TestCoreDocument_AddAccessToken_PrepareNewVersionError(t *testing.T) {
	ctx := context.Background()

	cd, err := newCoreDocument()
	assert.NoError(t, err)

	readRoleKey := utils.RandomSlice(32)

	readRules := []*coredocumentpb.ReadRule{
		{
			Roles: [][]byte{
				readRoleKey,
			},
			Action: coredocumentpb.Action_ACTION_READ,
		},
	}

	readNft := utils.RandomSlice(24)

	readRoles := []*coredocumentpb.Role{
		{
			RoleKey: readRoleKey,
			Collaborators: [][]byte{
				[]byte("invalid-account-id-bytes"),
			},
			Nfts: [][]byte{
				readNft,
			},
		},
	}

	cd.Document.ReadRules = readRules
	cd.Document.Roles = readRoles

	granteeAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	documentIdentifier := utils.RandomSlice(32)

	payload := AccessTokenParams{
		Grantee:            granteeAccountID.ToHexString(),
		DocumentIdentifier: hexutil.Encode(documentIdentifier),
	}

	res, err := cd.AddAccessToken(ctx, payload)
	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestCoreDocument_AddAccessToken_AssembleAccessTokenError(t *testing.T) {
	ctx := context.Background()

	cd, err := newCoreDocument()
	assert.NoError(t, err)

	granteeAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	documentIdentifier := utils.RandomSlice(32)

	payload := AccessTokenParams{
		Grantee:            granteeAccountID.ToHexString(),
		DocumentIdentifier: hexutil.Encode(documentIdentifier),
	}

	res, err := cd.AddAccessToken(ctx, payload)
	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestCoreDocument_DeleteAccessToken(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)

	granteeAccountID1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	granteeAccountID2, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	granteeAccountID3, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	cd.Document.AccessTokens = []*coredocumentpb.AccessToken{
		{
			Grantee: granteeAccountID1.ToBytes(),
		},
		{
			Grantee: granteeAccountID2.ToBytes(),
		},
	}

	res, err := cd.DeleteAccessToken(granteeAccountID1)
	assert.NoError(t, err)
	assert.True(t, res.Modified)
	assert.Len(t, res.Document.AccessTokens, 1)
	assert.Equal(t, res.Document.AccessTokens[0].Grantee, granteeAccountID2.ToBytes())

	res, err = res.DeleteAccessToken(granteeAccountID2)
	assert.NoError(t, err)
	assert.Len(t, res.Document.AccessTokens, 0)

	res, err = res.DeleteAccessToken(granteeAccountID3)
	assert.ErrorIs(t, err, ErrAccessTokenNotFound)
	assert.Nil(t, res)
}

func TestCoreDocument_DeleteAccessToken_PrepareNewVersionError(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)

	readRoleKey := utils.RandomSlice(32)

	readRules := []*coredocumentpb.ReadRule{
		{
			Roles: [][]byte{
				readRoleKey,
			},
			Action: coredocumentpb.Action_ACTION_READ,
		},
	}

	readNft := utils.RandomSlice(24)

	readRoles := []*coredocumentpb.Role{
		{
			RoleKey: readRoleKey,
			Collaborators: [][]byte{
				[]byte("invalid-account-id-bytes"),
			},
			Nfts: [][]byte{
				readNft,
			},
		},
	}

	cd.Document.ReadRules = readRules
	cd.Document.Roles = readRoles

	granteeAccountID1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	granteeAccountID2, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	cd.Document.AccessTokens = []*coredocumentpb.AccessToken{
		{
			Grantee: granteeAccountID1.ToBytes(),
		},
		{
			Grantee: granteeAccountID2.ToBytes(),
		},
	}

	res, err := cd.DeleteAccessToken(granteeAccountID1)
	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestCoreDocument_addNFTToReadRules(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)

	nftRegistryID := utils.RandomSlice(8)
	nftTokenID := utils.RandomSlice(16)

	err = cd.addNFTToReadRules(nftRegistryID, nftTokenID)
	assert.NoError(t, err)

	assert.Len(t, cd.Document.ReadRules, 1)
	assert.Len(t, cd.Document.Roles, 1)
	assert.Len(t, cd.Document.Roles[0].Nfts, 1)

	nftFound := findReadRole(
		cd.Document,
		func(ruleIndex, roleIndex int, role *coredocumentpb.Role) bool {
			for _, nft := range role.GetNfts() {
				if utils.IsSameByteSlice(nft, append(nftRegistryID, nftTokenID...)) {
					return true
				}
			}

			return false
		},
		coredocumentpb.Action_ACTION_READ,
	)

	assert.True(t, nftFound)

	res := cd.NFTCanRead(nftRegistryID, nftTokenID)
	assert.True(t, res)
}

func TestCoreDocument_addNFTToReadRules_ConstructNFTError(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)

	// Invalid byte slice sizes
	nftRegistryID := utils.RandomSlice(32)
	nftTokenID := utils.RandomSlice(32)

	err = cd.addNFTToReadRules(nftRegistryID, nftTokenID)
	assert.NotNil(t, err)
	assert.Len(t, cd.Document.ReadRules, 0)
	assert.Len(t, cd.Document.Roles, 0)
}

func TestCoreDocument_findAccessToken(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)

	tokenIdentifier1 := utils.RandomSlice(32)
	tokenIdentifier2 := utils.RandomSlice(32)

	token1 := &coredocumentpb.AccessToken{
		Identifier: tokenIdentifier1,
	}

	token2 := &coredocumentpb.AccessToken{
		Identifier: tokenIdentifier2,
	}

	cd.Document.AccessTokens = []*coredocumentpb.AccessToken{token1, token2}

	res, err := cd.findAccessToken(tokenIdentifier1)
	assert.NoError(t, err)
	assert.Equal(t, token1, res)

	res, err = cd.findAccessToken(tokenIdentifier2)
	assert.NoError(t, err)
	assert.Equal(t, token2, res)

	res, err = cd.findAccessToken(utils.RandomSlice(32))
	assert.ErrorIs(t, err, ErrAccessTokenNotFound)
	assert.Nil(t, res)
}

func TestConstructNFT(t *testing.T) {
	registryID := utils.RandomSlice(nftCollectionIDByteCount)
	tokenID := utils.RandomSlice(nftItemIDByteCount)

	expectedNFT := append(registryID, tokenID...)

	res, err := ConstructNFT(registryID, tokenID)
	assert.NoError(t, err)
	assert.Equal(t, expectedNFT, res)

	res, err = ConstructNFT(utils.RandomSlice(nftCollectionIDByteCount-1), utils.RandomSlice(nftItemIDByteCount))
	assert.True(t, errors.IsOfType(ErrNftByteLength, err))
	assert.Nil(t, res)

	res, err = ConstructNFT(utils.RandomSlice(nftCollectionIDByteCount), utils.RandomSlice(nftItemIDByteCount-1))
	assert.True(t, errors.IsOfType(ErrNftByteLength, err))
	assert.Nil(t, res)
}

func TestIsNFTInRole(t *testing.T) {
	encodedCollectionID := utils.RandomSlice(nftCollectionIDByteCount)
	encodedItemID := utils.RandomSlice(nftItemIDByteCount)

	nft, err := ConstructNFT(encodedCollectionID, encodedItemID)
	assert.NoError(t, err)

	role := &coredocumentpb.Role{
		Nfts: [][]byte{
			nft,
		},
	}

	res, found := isNFTInRole(role, encodedCollectionID, encodedItemID)
	assert.True(t, found)
	assert.Equal(t, res, 0)

	res, found = isNFTInRole(role, utils.RandomSlice(nftCollectionIDByteCount), utils.RandomSlice(nftItemIDByteCount))
	assert.False(t, found)
	assert.Equal(t, res, 0)
}

func TestGetStoredNFT(t *testing.T) {
	encodedCollectionID := utils.RandomSlice(nftCollectionIDByteCount)
	encodedItemID := utils.RandomSlice(nftItemIDByteCount)

	nft := &coredocumentpb.NFT{
		CollectionId: encodedCollectionID,
		ItemId:       encodedItemID,
	}

	nfts := []*coredocumentpb.NFT{
		nft,
	}

	res := getStoredNFT(nfts, encodedCollectionID)
	assert.Equal(t, res, nft)

	res = getStoredNFT(nil, encodedCollectionID)
	assert.Nil(t, res)
}

func TestGetReadAccessProofKeys(t *testing.T) {
	encodedCollectionID1 := utils.RandomSlice(nftCollectionIDByteCount)
	encodedItemID1 := utils.RandomSlice(nftItemIDByteCount)

	encodedCollectionID2 := utils.RandomSlice(nftCollectionIDByteCount)
	encodedItemID2 := utils.RandomSlice(nftItemIDByteCount)

	encodedCollectionID3 := utils.RandomSlice(nftCollectionIDByteCount)
	encodedItemID3 := utils.RandomSlice(nftItemIDByteCount)

	readRoleKey1 := utils.RandomSlice(32)
	readRoleKey2 := utils.RandomSlice(32)

	readRules := []*coredocumentpb.ReadRule{
		{
			Roles: [][]byte{
				readRoleKey2,
			},
			Action: coredocumentpb.Action_ACTION_READ,
		},
		{
			Roles: [][]byte{
				readRoleKey1,
			},
			Action: coredocumentpb.Action_ACTION_READ,
		},
	}

	readNft1 := append(encodedCollectionID1, encodedItemID1...)
	readNft2 := append(encodedCollectionID2, encodedItemID2...)
	readNft3 := append(encodedCollectionID3, encodedItemID3...)

	readRoles := []*coredocumentpb.Role{
		{
			RoleKey: readRoleKey1,
			Nfts: [][]byte{
				readNft1,
				readNft3,
			},
		},
		{
			RoleKey: readRoleKey2,
			Nfts: [][]byte{
				readNft2,
			},
		},
	}

	cd, err := newCoreDocument()
	assert.NoError(t, err)

	cd.Document.ReadRules = readRules
	cd.Document.Roles = readRoles

	expectedResult := []string{
		fmt.Sprintf(CDTreePrefix+".read_rules[%d].roles[%d]", 1, 0),
		fmt.Sprintf(CDTreePrefix+".read_rules[%d].action", 1),
		fmt.Sprintf(CDTreePrefix+".roles[%s].nfts[%d]", hexutil.Encode(readRoleKey1), 0),
	}

	res, err := getReadAccessProofKeys(cd.Document, encodedCollectionID1, encodedItemID1)
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, res)

	expectedResult = []string{
		fmt.Sprintf(CDTreePrefix+".read_rules[%d].roles[%d]", 0, 0),
		fmt.Sprintf(CDTreePrefix+".read_rules[%d].action", 0),
		fmt.Sprintf(CDTreePrefix+".roles[%s].nfts[%d]", hexutil.Encode(readRoleKey2), 0),
	}

	res, err = getReadAccessProofKeys(cd.Document, encodedCollectionID2, encodedItemID2)
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, res)

	expectedResult = []string{
		fmt.Sprintf(CDTreePrefix+".read_rules[%d].roles[%d]", 1, 0),
		fmt.Sprintf(CDTreePrefix+".read_rules[%d].action", 1),
		fmt.Sprintf(CDTreePrefix+".roles[%s].nfts[%d]", hexutil.Encode(readRoleKey1), 1),
	}

	res, err = getReadAccessProofKeys(cd.Document, encodedCollectionID3, encodedItemID3)
	assert.NoError(t, err)
	assert.Equal(t, expectedResult, res)
}

func TestGetReadAccessProofKeys_NFTRoleMissing(t *testing.T) {
	encodedCollectionID1 := utils.RandomSlice(nftCollectionIDByteCount)
	encodedItemID1 := utils.RandomSlice(nftItemIDByteCount)

	encodedCollectionID2 := utils.RandomSlice(nftCollectionIDByteCount)
	encodedItemID2 := utils.RandomSlice(nftItemIDByteCount)

	encodedCollectionID3 := utils.RandomSlice(nftCollectionIDByteCount)
	encodedItemID3 := utils.RandomSlice(nftItemIDByteCount)

	readRoleKey1 := utils.RandomSlice(32)
	readRoleKey2 := utils.RandomSlice(32)

	readRules := []*coredocumentpb.ReadRule{
		{
			Roles: [][]byte{
				readRoleKey2,
			},
			Action: coredocumentpb.Action_ACTION_READ,
		},
		{
			Roles: [][]byte{
				readRoleKey1,
			},
			Action: coredocumentpb.Action_ACTION_READ,
		},
	}

	readNft1 := append(encodedCollectionID1, encodedItemID1...)
	readNft2 := append(encodedCollectionID2, encodedItemID2...)

	readRoles := []*coredocumentpb.Role{
		{
			RoleKey: readRoleKey1,
			Nfts: [][]byte{
				readNft1,
			},
		},
		{
			RoleKey: readRoleKey2,
			Nfts: [][]byte{
				readNft2,
			},
		},
	}

	cd, err := newCoreDocument()
	assert.NoError(t, err)

	cd.Document.ReadRules = readRules
	cd.Document.Roles = readRoles

	res, err := getReadAccessProofKeys(cd.Document, encodedCollectionID3, encodedItemID3)
	assert.ErrorIs(t, err, ErrNFTRoleMissing)
	assert.Nil(t, res)
}

func TestIsAccountIDInRole(t *testing.T) {
	accountID1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)
	accountID2, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)
	accountID3, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	role := &coredocumentpb.Role{
		Collaborators: [][]byte{
			accountID1.ToBytes(),
			accountID2.ToBytes(),
		},
	}

	res, found := isAccountIDinRole(role, accountID1)
	assert.True(t, found)
	assert.Equal(t, 0, res)

	res, found = isAccountIDinRole(role, accountID2)
	assert.True(t, found)
	assert.Equal(t, 1, res)

	res, found = isAccountIDinRole(role, accountID3)
	assert.False(t, found)
	assert.Equal(t, 0, res)
}

func TestGetRole(t *testing.T) {
	roleKey1 := utils.RandomSlice(32)
	roleKey2 := utils.RandomSlice(32)
	roleKey3 := utils.RandomSlice(32)

	role1 := &coredocumentpb.Role{
		RoleKey: roleKey1,
	}

	role2 := &coredocumentpb.Role{
		RoleKey: roleKey2,
	}

	roles := []*coredocumentpb.Role{
		role1,
		role2,
	}

	res, err := getRole(roleKey1, roles)
	assert.NoError(t, err)
	assert.Equal(t, res, role1)

	res, err = getRole(roleKey2, roles)
	assert.NoError(t, err)
	assert.Equal(t, res, role2)

	res, err = getRole(roleKey3, roles)
	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestValidateAccessToken(t *testing.T) {
	pubKey, privKey, err := ed25519.GenerateSigningKeyPair()
	assert.NoError(t, err)

	requesterID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	granterID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	token := &coredocumentpb.AccessToken{
		Identifier:         utils.RandomSlice(32),
		Granter:            granterID.ToBytes(),
		RoleIdentifier:     utils.RandomSlice(32),
		DocumentIdentifier: utils.RandomSlice(32),
		Signature:          nil,
		DocumentVersion:    utils.RandomSlice(32),
	}

	tokenMessage, err := AssembleTokenMessage(
		token.Identifier,
		granterID,
		requesterID,
		token.RoleIdentifier,
		token.DocumentIdentifier,
		token.DocumentVersion,
	)
	assert.NoError(t, err)

	signature, err := privKey.Sign(cryptorand.Reader, tokenMessage, crypto.Hash(0))
	assert.NoError(t, err)

	token.Signature = signature

	err = validateAccessToken(pubKey, token, requesterID)
	assert.NoError(t, err)
}

func TestValidateAccessToken_GranterIDError(t *testing.T) {
	pubKey, _, err := ed25519.GenerateSigningKeyPair()
	assert.NoError(t, err)

	requesterID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	token := &coredocumentpb.AccessToken{
		Identifier: utils.RandomSlice(32),
		// Invalid byte slice length for granter ID.
		Granter:            utils.RandomSlice(11),
		RoleIdentifier:     utils.RandomSlice(32),
		DocumentIdentifier: utils.RandomSlice(32),
		Signature:          nil,
		DocumentVersion:    utils.RandomSlice(32),
	}

	err = validateAccessToken(pubKey, token, requesterID)
	assert.NotNil(t, err)
}

func TestValidateAccessToken_TokenMessageError(t *testing.T) {
	pubKey, _, err := ed25519.GenerateSigningKeyPair()
	assert.NoError(t, err)

	requesterID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	granterID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	token := &coredocumentpb.AccessToken{
		// Invalid byte slice length for identifier.
		Identifier:         utils.RandomSlice(11),
		Granter:            granterID.ToBytes(),
		RoleIdentifier:     utils.RandomSlice(32),
		DocumentIdentifier: utils.RandomSlice(32),
		Signature:          nil,
		DocumentVersion:    utils.RandomSlice(32),
	}

	err = validateAccessToken(pubKey, token, requesterID)
	assert.NotNil(t, err)
}

func TestValidateAccessToken_InvalidSignature(t *testing.T) {
	pubKey, _, err := ed25519.GenerateSigningKeyPair()
	assert.NoError(t, err)

	requesterID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	granterID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	token := &coredocumentpb.AccessToken{
		Identifier:         utils.RandomSlice(32),
		Granter:            granterID.ToBytes(),
		RoleIdentifier:     utils.RandomSlice(32),
		DocumentIdentifier: utils.RandomSlice(32),
		Signature:          []byte("invalid-signature"),
		DocumentVersion:    utils.RandomSlice(32),
	}

	err = validateAccessToken(pubKey, token, requesterID)
	assert.ErrorIs(t, err, ErrAccessTokenInvalid)
}

func TestRemoveTokenAtIndex(t *testing.T) {
	token1 := &coredocumentpb.AccessToken{
		Identifier: utils.RandomSlice(32),
	}
	token2 := &coredocumentpb.AccessToken{
		Identifier: utils.RandomSlice(32),
	}
	token3 := &coredocumentpb.AccessToken{
		Identifier: utils.RandomSlice(32),
	}

	tokens := []*coredocumentpb.AccessToken{
		token1,
		token2,
		token3,
	}

	res := removeTokenAtIndex(0, tokens)
	assert.Len(t, tokens, 3)
	assert.Len(t, res, 2)
	assert.Contains(t, res, token2)
	assert.Contains(t, res, token3)

	res = removeTokenAtIndex(1, tokens)
	assert.Len(t, tokens, 3)
	assert.Len(t, res, 2)
	assert.Contains(t, res, token1)
	assert.Contains(t, res, token3)

	res = removeTokenAtIndex(2, tokens)
	assert.Len(t, tokens, 3)
	assert.Len(t, res, 2)
	assert.Contains(t, res, token1)
	assert.Contains(t, res, token2)
}

func TestAssembleAccessToken(t *testing.T) {
	mockAccount := configMocks.NewAccountMock(t)

	ctx := contextutil.WithAccount(context.Background(), mockAccount)

	granterAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	mockAccount.On("GetIdentity").
		Return(granterAccountID)

	granteeAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	documentIdentifier := utils.RandomSlice(32)
	documentVersion := utils.RandomSlice(32)

	payload := AccessTokenParams{
		Grantee:            granteeAccountID.ToHexString(),
		DocumentIdentifier: hexutil.Encode(documentIdentifier),
	}

	publicKey := utils.RandomSlice(32)
	signatureBytes := utils.RandomSlice(32)

	signature := &coredocumentpb.Signature{
		SignatureId:         utils.RandomSlice(32),
		SignerId:            utils.RandomSlice(32),
		PublicKey:           publicKey,
		Signature:           signatureBytes,
		TransitionValidated: false,
	}

	mockAccount.On("SignMsg", mock.Anything).
		Return(signature, nil)

	res, err := assembleAccessToken(ctx, payload, documentVersion)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, granterAccountID.ToBytes(), res.GetGranter())
	assert.Equal(t, granteeAccountID.ToBytes(), res.GetGrantee())
	assert.Equal(t, documentIdentifier, res.GetDocumentIdentifier())
	assert.Equal(t, signatureBytes, res.GetSignature())
	assert.Equal(t, publicKey, res.GetKey())
	assert.Equal(t, documentVersion, res.GetDocumentVersion())
}

func TestAssembleAccessToken_ContextAccountError(t *testing.T) {
	granteeAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	documentIdentifier := utils.RandomSlice(32)
	documentVersion := utils.RandomSlice(32)

	payload := AccessTokenParams{
		Grantee:            granteeAccountID.ToHexString(),
		DocumentIdentifier: hexutil.Encode(documentIdentifier),
	}

	res, err := assembleAccessToken(context.Background(), payload, documentVersion)
	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestAssembleAccessToken_GranteeIDError(t *testing.T) {
	mockAccount := configMocks.NewAccountMock(t)

	ctx := contextutil.WithAccount(context.Background(), mockAccount)

	granterAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	mockAccount.On("GetIdentity").
		Return(granterAccountID)

	documentIdentifier := utils.RandomSlice(32)
	documentVersion := utils.RandomSlice(32)

	payload := AccessTokenParams{
		Grantee:            hexutil.Encode([]byte("invalid-account-id-bytes")),
		DocumentIdentifier: hexutil.Encode(documentIdentifier),
	}

	res, err := assembleAccessToken(ctx, payload, documentVersion)
	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestAssembleAccessToken_DocumentIdentifierError(t *testing.T) {
	mockAccount := configMocks.NewAccountMock(t)

	ctx := contextutil.WithAccount(context.Background(), mockAccount)

	granterAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	mockAccount.On("GetIdentity").
		Return(granterAccountID)

	granteeAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	documentVersion := utils.RandomSlice(32)

	payload := AccessTokenParams{
		Grantee:            granteeAccountID.ToHexString(),
		DocumentIdentifier: "invalid-document-identifier",
	}

	res, err := assembleAccessToken(ctx, payload, documentVersion)
	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestAssembleAccessToken_TokenMessageError(t *testing.T) {
	mockAccount := configMocks.NewAccountMock(t)

	ctx := contextutil.WithAccount(context.Background(), mockAccount)

	granterAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	mockAccount.On("GetIdentity").
		Return(granterAccountID)

	granteeAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	// Invalid document identifier byte slice length
	documentIdentifier := utils.RandomSlice(11)
	documentVersion := utils.RandomSlice(32)

	payload := AccessTokenParams{
		Grantee:            granteeAccountID.ToHexString(),
		DocumentIdentifier: hexutil.Encode(documentIdentifier),
	}

	res, err := assembleAccessToken(ctx, payload, documentVersion)
	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestAssembleAccessToken_SignError(t *testing.T) {
	mockAccount := configMocks.NewAccountMock(t)

	ctx := contextutil.WithAccount(context.Background(), mockAccount)

	granterAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	mockAccount.On("GetIdentity").
		Return(granterAccountID)

	granteeAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	documentIdentifier := utils.RandomSlice(32)
	documentVersion := utils.RandomSlice(32)

	payload := AccessTokenParams{
		Grantee:            granteeAccountID.ToHexString(),
		DocumentIdentifier: hexutil.Encode(documentIdentifier),
	}

	signError := errors.New("error")

	mockAccount.On("SignMsg", mock.Anything).
		Return(nil, signError)

	res, err := assembleAccessToken(ctx, payload, documentVersion)
	assert.ErrorIs(t, err, signError)
	assert.Nil(t, res)
}

func TestAssembleTokenMessage(t *testing.T) {
	granterAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)
	granteeAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	tokenIdentifier := utils.RandomSlice(idSize)
	roleIdentifier := utils.RandomSlice(idSize)
	documentIdentifier := utils.RandomSlice(idSize)
	documentVersion := utils.RandomSlice(idSize)

	res, err := AssembleTokenMessage(
		tokenIdentifier,
		granterAccountID,
		granteeAccountID,
		roleIdentifier,
		documentIdentifier,
		documentVersion,
	)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	tokenIdentifier = utils.RandomSlice(idSize - 1)

	res, err = AssembleTokenMessage(
		tokenIdentifier,
		granterAccountID,
		granteeAccountID,
		roleIdentifier,
		documentIdentifier,
		documentVersion,
	)
	assert.ErrorIs(t, err, ErrInvalidIDLength)
	assert.Nil(t, res)

	tokenIdentifier = utils.RandomSlice(idSize)
	roleIdentifier = utils.RandomSlice(idSize - 1)

	res, err = AssembleTokenMessage(
		tokenIdentifier,
		granterAccountID,
		granteeAccountID,
		roleIdentifier,
		documentIdentifier,
		documentVersion,
	)
	assert.ErrorIs(t, err, ErrInvalidIDLength)
	assert.Nil(t, res)

	roleIdentifier = utils.RandomSlice(idSize)
	documentIdentifier = utils.RandomSlice(idSize - 1)

	res, err = AssembleTokenMessage(
		tokenIdentifier,
		granterAccountID,
		granteeAccountID,
		roleIdentifier,
		documentIdentifier,
		documentVersion,
	)
	assert.ErrorIs(t, err, ErrInvalidIDLength)
	assert.Nil(t, res)
}

func getStoredNFT(nfts []*coredocumentpb.NFT, encodedCollectionID []byte) *coredocumentpb.NFT {
	for _, nft := range nfts {
		if bytes.Equal(nft.GetCollectionId(), encodedCollectionID) {
			return nft
		}
	}

	return nil
}

func getReadAccessProofKeys(
	cd *coredocumentpb.CoreDocument,
	encodedCollectionID []byte,
	encodedItemID []byte,
) (pks []string, err error) {
	var rridx int  // index of the read rules which contain the role
	var ridx int   // index of the role
	var nftIdx int // index of the NFT in the above role
	var rk []byte  // role key of the above role

	found := findReadRole(
		cd,
		func(i, j int, role *coredocumentpb.Role) bool {
			z, found := isNFTInRole(role, encodedCollectionID, encodedItemID)
			if found {
				rridx = i
				ridx = j
				rk = role.RoleKey
				nftIdx = z
			}

			return found
		},
		coredocumentpb.Action_ACTION_READ,
	)

	if !found {
		return nil, ErrNFTRoleMissing
	}

	return []string{
		fmt.Sprintf(CDTreePrefix+".read_rules[%d].roles[%d]", rridx, ridx),          // proof that a read rule exists with the nft role
		fmt.Sprintf(CDTreePrefix+".read_rules[%d].action", rridx),                   // proof that this read rule has read access
		fmt.Sprintf(CDTreePrefix+".roles[%s].nfts[%d]", hexutil.Encode(rk), nftIdx), // proof that role with nft exists
	}, nil
}

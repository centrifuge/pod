//go:build unit

package documents

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/centrifuge/go-centrifuge/utils/byteutils"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"golang.org/x/crypto/blake2b"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/centrifuge/precise-proofs/proofs"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/errors"
	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/commons"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewCoreDocument(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	assert.NotNil(t, cd)

	documentPrefix := utils.RandomSlice(32)

	readCollaborator, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	readWriteCollaborator, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collaboratorsAccess := CollaboratorsAccess{
		ReadCollaborators:      []*types.AccountID{readCollaborator},
		ReadWriteCollaborators: []*types.AccountID{readWriteCollaborator},
	}

	attrKey := utils.RandomByte32()

	attributes := map[AttrKey]Attribute{
		attrKey: {
			KeyLabel: "label",
			Key:      attrKey,
			Value: AttrVal{
				Type: AttrString,
				Str:  "test",
			},
		},
	}

	cd, err = NewCoreDocument(documentPrefix, collaboratorsAccess, attributes)
	assert.NoError(t, err)
	assert.NotNil(t, cd)

	invalidAttributes := map[AttrKey]Attribute{
		attrKey: {
			KeyLabel: "label",
			Key:      attrKey,
			Value:    AttrVal{},
		},
	}

	cd, err = NewCoreDocument(documentPrefix, collaboratorsAccess, invalidAttributes)
	assert.NotNil(t, err)
	assert.NotNil(t, cd)
}

func TestNewCoreDocumentFromProtobuf(t *testing.T) {
	cd := &coredocumentpb.CoreDocument{}
	res, err := NewCoreDocumentFromProtobuf(cd)
	assert.NoError(t, err)
	assert.Equal(t, cd, res.Document)

	cd.Attributes = []*coredocumentpb.Attribute{
		{
			// Invalid key length
			Key: utils.RandomSlice(33),
		},
	}

	res, err = NewCoreDocumentFromProtobuf(cd)
	assert.NotNil(t, err)
	assert.Equal(t, cd, res.Document)
}

func TestNewClonedDocument(t *testing.T) {
	transitionRules := []*coredocumentpb.TransitionRule{
		{
			RuleKey: utils.RandomSlice(32),
			Roles: [][]byte{
				utils.RandomSlice(32),
			},
			MatchType: 0,
			Field:     utils.RandomSlice(32),
			Action:    0,
			ComputeFields: [][]byte{
				utils.RandomSlice(32),
			},
			ComputeTargetField: utils.RandomSlice(32),
			ComputeCode:        utils.RandomSlice(32),
		},
	}

	readRules := []*coredocumentpb.ReadRule{
		{
			Roles: [][]byte{
				utils.RandomSlice(32),
			},
			Action: 0,
		},
	}

	roles := []*coredocumentpb.Role{
		{
			RoleKey: utils.RandomSlice(32),
			Collaborators: [][]byte{
				utils.RandomSlice(32),
			},
			Nfts: [][]byte{
				utils.RandomSlice(32),
			},
		},
	}

	attributes := []*coredocumentpb.Attribute{
		{
			Key:      utils.RandomSlice(32),
			KeyLabel: utils.RandomSlice(32),
			Type:     coredocumentpb.AttributeType_ATTRIBUTE_TYPE_STRING,
			Value: &coredocumentpb.Attribute_StrVal{
				StrVal: "test",
			},
		},
	}

	cd := &coredocumentpb.CoreDocument{
		TransitionRules: transitionRules,
		ReadRules:       readRules,
		Roles:           roles,
		Attributes:      attributes,
	}

	res, err := NewClonedDocument(cd)
	assert.NoError(t, err)
	assert.Equal(t, transitionRules, res.Document.TransitionRules)
	assert.Equal(t, readRules, res.Document.ReadRules)
	assert.Equal(t, roles, res.Document.Roles)
	assert.Equal(t, attributes, res.Document.Attributes)

	invalidAttributes := []*coredocumentpb.Attribute{
		{
			Key: utils.RandomSlice(33),
		},
	}

	cd = &coredocumentpb.CoreDocument{
		Attributes: invalidAttributes,
	}

	res, err = NewClonedDocument(cd)
	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestRemoveDuplicateAccountIDs(t *testing.T) {
	accountID1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)
	accountID2, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)
	accountID3, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountIDs := []*types.AccountID{
		accountID1,
		accountID2,
		accountID3,
		accountID1,
	}

	res := RemoveDuplicateAccountIDs(accountIDs)
	assert.Equal(t, 3, len(res))
	assert.Contains(t, res, accountID1)
	assert.Contains(t, res, accountID2)
	assert.Contains(t, res, accountID3)
}

func TestParseAccountIDStrings(t *testing.T) {
	accountID1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)
	accountID2, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)
	accountID3, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountIDStrings := []string{
		accountID1.ToHexString(),
		accountID2.ToHexString(),
		accountID3.ToHexString(),
	}

	res, err := ParseAccountIDStrings(accountIDStrings...)
	assert.NoError(t, err)
	assert.Contains(t, res, accountID1)
	assert.Contains(t, res, accountID2)
	assert.Contains(t, res, accountID3)

	accountIDStrings = append(accountIDStrings, "invalid account ID string")

	res, err = ParseAccountIDStrings(accountIDStrings...)
	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestNewCoreDocumentWithAccessToken(t *testing.T) {
	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	signature := &coredocumentpb.Signature{}

	accountMock.On("SignMsg", mock.Anything).
		Return(signature, nil)

	signingPubKey := utils.RandomSlice(32)

	accountMock.On("GetSigningPublicKey").
		Return(signingPubKey)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	documentPrefix := utils.RandomSlice(32)

	grantee, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	documentID := hexutil.Encode(utils.RandomSlice(32))

	accessTokenParams := AccessTokenParams{
		Grantee:            grantee.ToHexString(),
		DocumentIdentifier: documentID,
	}

	res, err := NewCoreDocumentWithAccessToken(ctx, documentPrefix, accessTokenParams)
	assert.NoError(t, err)
	assert.NotNil(t, res)
}

func TestNewCoreDocumentWithAccessToken_InvalidGrantee(t *testing.T) {
	documentPrefix := utils.RandomSlice(32)

	documentID := hexutil.Encode(utils.RandomSlice(32))

	accessTokenParams := AccessTokenParams{
		Grantee:            "invalid account ID hex",
		DocumentIdentifier: documentID,
	}

	res, err := NewCoreDocumentWithAccessToken(context.Background(), documentPrefix, accessTokenParams)
	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestNewCoreDocumentWithAccessToken_MissingIdentity(t *testing.T) {
	documentPrefix := utils.RandomSlice(32)

	grantee, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	documentID := hexutil.Encode(utils.RandomSlice(32))

	accessTokenParams := AccessTokenParams{
		Grantee:            grantee.ToHexString(),
		DocumentIdentifier: documentID,
	}

	res, err := NewCoreDocumentWithAccessToken(context.Background(), documentPrefix, accessTokenParams)
	assert.ErrorIs(t, err, ErrDocumentConfigAccount)
	assert.Nil(t, res)
}

func TestNewCoreDocumentWithAccessToken_AssembleAccessTokenError(t *testing.T) {
	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)
	accountMock.On("GetIdentity").
		Return(accountID)

	signError := errors.New("error")

	accountMock.On("SignMsg", mock.Anything).
		Return(nil, signError)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	documentPrefix := utils.RandomSlice(32)

	grantee, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	documentID := hexutil.Encode(utils.RandomSlice(32))

	accessTokenParams := AccessTokenParams{
		Grantee:            grantee.ToHexString(),
		DocumentIdentifier: documentID,
	}

	res, err := NewCoreDocumentWithAccessToken(ctx, documentPrefix, accessTokenParams)
	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestCoreDocument_SetStatus(t *testing.T) {
	cd := &CoreDocument{}

	err := cd.SetStatus(Pending)
	assert.NoError(t, err)

	err = cd.SetStatus(Committing)
	assert.NoError(t, err)

	err = cd.SetStatus(Committed)
	assert.NoError(t, err)

	err = cd.SetStatus(Committed)
	assert.NoError(t, err)

	err = cd.SetStatus(Committing)
	assert.ErrorIs(t, err, ErrCDStatus)
}

func TestCoreDocument_AppendSignatures(t *testing.T) {
	cd := &CoreDocument{
		Document: &coredocumentpb.CoreDocument{},
	}

	signatures := []*coredocumentpb.Signature{
		{
			SignatureId:         utils.RandomSlice(32),
			SignerId:            utils.RandomSlice(32),
			PublicKey:           utils.RandomSlice(32),
			Signature:           utils.RandomSlice(32),
			TransitionValidated: true,
		},
		{
			SignatureId:         utils.RandomSlice(32),
			SignerId:            utils.RandomSlice(32),
			PublicKey:           utils.RandomSlice(32),
			Signature:           utils.RandomSlice(32),
			TransitionValidated: true,
		},
	}

	cd.AppendSignatures(signatures...)

	assert.True(t, cd.Modified)
	assert.Equal(t, signatures, cd.Document.SignatureData.Signatures)
}

func TestCoreDocument_Patch(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	assert.NotNil(t, cd)

	documentPrefix := utils.RandomSlice(32)

	readCollaborator, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	readWriteCollaborator, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collaboratorsAccess := CollaboratorsAccess{
		ReadCollaborators:      []*types.AccountID{readCollaborator},
		ReadWriteCollaborators: []*types.AccountID{readWriteCollaborator},
	}

	attrKey := utils.RandomByte32()

	attributes := map[AttrKey]Attribute{
		attrKey: {
			KeyLabel: "label",
			Key:      attrKey,
			Value: AttrVal{
				Type: AttrString,
				Str:  "test",
			},
		},
	}

	res, err := cd.Patch(documentPrefix, collaboratorsAccess, attributes)
	assert.NoError(t, err)
	assert.NotNil(t, res)
}

func TestCoreDocument_Patch_StatusError(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	assert.NotNil(t, cd)

	documentPrefix := utils.RandomSlice(32)

	collaboratorsAccess := CollaboratorsAccess{}

	attributes := make(map[AttrKey]Attribute)

	cd.Status = Committing

	res, err := cd.Patch(documentPrefix, collaboratorsAccess, attributes)
	assert.ErrorIs(t, err, ErrDocumentNotInAllowedState)
	assert.Nil(t, res)

	cd.Status = Committed

	res, err = cd.Patch(documentPrefix, collaboratorsAccess, attributes)
	assert.ErrorIs(t, err, ErrDocumentNotInAllowedState)
	assert.Nil(t, res)
}

func TestCoreDocument_Patch_UpdateAttributesError(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	assert.NotNil(t, cd)

	documentPrefix := utils.RandomSlice(32)

	readCollaborator, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	readWriteCollaborator, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collaboratorsAccess := CollaboratorsAccess{
		ReadCollaborators:      []*types.AccountID{readCollaborator},
		ReadWriteCollaborators: []*types.AccountID{readWriteCollaborator},
	}

	attrKey := utils.RandomByte32()

	attributes := map[AttrKey]Attribute{
		attrKey: {
			KeyLabel: "label",
			Key:      attrKey,
		},
	}

	res, err := cd.Patch(documentPrefix, collaboratorsAccess, attributes)
	assert.True(t, errors.IsOfType(ErrCDNewVersion, err))
	assert.Nil(t, res)
}

func TestCoreDocument_PrepareNewVersion(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	assert.NotNil(t, cd)

	documentPrefix := utils.RandomSlice(32)

	readCollaborator, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	readWriteCollaborator, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collaboratorsAccess := CollaboratorsAccess{
		ReadCollaborators:      []*types.AccountID{readCollaborator},
		ReadWriteCollaborators: []*types.AccountID{readWriteCollaborator},
	}

	attrKey := utils.RandomByte32()

	attributes := map[AttrKey]Attribute{
		attrKey: {
			KeyLabel: "label",
			Key:      attrKey,
			Value: AttrVal{
				Type: AttrString,
				Str:  "test",
			},
		},
	}

	res, err := cd.PrepareNewVersion(documentPrefix, collaboratorsAccess, attributes)
	assert.NoError(t, err)
	assert.NotNil(t, res)
}

func TestCoreDocument_PrepareNewVersion_GetCollaboratorsError(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	assert.NotNil(t, cd)

	roleKey := utils.RandomSlice(32)

	roles := []*coredocumentpb.Role{
		{
			RoleKey: roleKey,
			Collaborators: [][]byte{
				[]byte{1}, // invalid account ID bytes
			},
		},
	}

	cd.Document.ReadRules = []*coredocumentpb.ReadRule{
		{
			Roles: [][]byte{
				roleKey,
			},
			Action: coredocumentpb.Action_ACTION_READ,
		},
	}

	cd.Document.Roles = roles

	documentPrefix := utils.RandomSlice(32)

	readCollaborator, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	readWriteCollaborator, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collaboratorsAccess := CollaboratorsAccess{
		ReadCollaborators:      []*types.AccountID{readCollaborator},
		ReadWriteCollaborators: []*types.AccountID{readWriteCollaborator},
	}

	attrKey := utils.RandomByte32()

	attributes := map[AttrKey]Attribute{
		attrKey: {
			KeyLabel: "label",
			Key:      attrKey,
			Value: AttrVal{
				Type: AttrString,
				Str:  "test",
			},
		},
	}

	res, err := cd.PrepareNewVersion(documentPrefix, collaboratorsAccess, attributes)
	assert.True(t, errors.IsOfType(ErrCDNewVersion, err))
	assert.Nil(t, res)
}

func TestCoreDocument_PrepareNewVersion_UpdateAttributesError(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	assert.NotNil(t, cd)

	documentPrefix := utils.RandomSlice(32)

	readCollaborator, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	readWriteCollaborator, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collaboratorsAccess := CollaboratorsAccess{
		ReadCollaborators:      []*types.AccountID{readCollaborator},
		ReadWriteCollaborators: []*types.AccountID{readWriteCollaborator},
	}

	attrKey := utils.RandomByte32()

	invalidAttributes := map[AttrKey]Attribute{
		attrKey: {
			KeyLabel: "label",
			Key:      attrKey,
			// missing value
		},
	}

	res, err := cd.PrepareNewVersion(documentPrefix, collaboratorsAccess, invalidAttributes)
	assert.True(t, errors.IsOfType(ErrCDNewVersion, err))
	assert.Nil(t, res)
}

func TestCoreDocument_UpdateAttributes_both(t *testing.T) {
	oldCAttrs := map[string]attribute{
		"time_test": {
			Type:  AttrTimestamp.String(),
			Value: time.Now().UTC().Format(time.RFC3339),
		},

		"string_test": {
			Type:  AttrString.String(),
			Value: "some string",
		},

		"bytes_test": {
			Type:  AttrBytes.String(),
			Value: hexutil.Encode([]byte("some bytes data")),
		},

		"int256_test": {
			Type:  AttrInt256.String(),
			Value: "1000000001",
		},

		"decimal_test": {
			Type:  AttrDecimal.String(),
			Value: "1000.000001",
		},
	}

	updates := map[string]attribute{
		"time_test": {
			Type:  AttrTimestamp.String(),
			Value: time.Now().Add(60 * time.Hour).UTC().Format(time.RFC3339),
		},

		"string_test": {
			Type:  AttrString.String(),
			Value: "new string",
		},

		"bytes_test": {
			Type:  AttrBytes.String(),
			Value: hexutil.Encode([]byte("new bytes data")),
		},

		"int256_test": {
			Type:  AttrInt256.String(),
			Value: "1000000002",
		},

		"decimal_test": {
			Type:  AttrDecimal.String(),
			Value: "1000.000002",
		},

		"decimal_test_1": {
			Type:  AttrDecimal.String(),
			Value: "1111.00012",
		},
	}

	oldAttrs := toAttrsMap(t, oldCAttrs)
	newAttrs := toAttrsMap(t, updates)

	newPattrs, err := toProtocolAttributes(newAttrs)
	assert.NoError(t, err)

	oldPattrs, err := toProtocolAttributes(oldAttrs)
	assert.NoError(t, err)

	upattrs, uattrs, err := updateAttributes(oldPattrs, newAttrs)
	assert.NoError(t, err)

	assert.Equal(t, upattrs, newPattrs)
	assert.Equal(t, newAttrs, uattrs)

	oldPattrs[0].Key = utils.RandomSlice(33)
	_, _, err = updateAttributes(oldPattrs, newAttrs)
	assert.Error(t, err)
}

func TestCoreDocument_UpdateAttributes_old_nil(t *testing.T) {
	updates := map[string]attribute{
		"time_test": {
			Type:  AttrTimestamp.String(),
			Value: time.Now().Add(60 * time.Hour).UTC().Format(time.RFC3339),
		},

		"string_test": {
			Type:  AttrString.String(),
			Value: "new string",
		},

		"bytes_test": {
			Type:  AttrBytes.String(),
			Value: hexutil.Encode([]byte("new bytes data")),
		},

		"int256_test": {
			Type:  AttrInt256.String(),
			Value: "1000000002",
		},

		"decimal_test": {
			Type:  AttrDecimal.String(),
			Value: "1000.000002",
		},

		"decimal_test_1": {
			Type:  AttrDecimal.String(),
			Value: "1111.00012",
		},
	}

	newAttrs := toAttrsMap(t, updates)
	newPattrs, err := toProtocolAttributes(newAttrs)
	assert.NoError(t, err)

	upattrs, uattrs, err := updateAttributes(nil, newAttrs)
	assert.NoError(t, err)

	assert.Equal(t, upattrs, newPattrs)
	assert.Equal(t, newAttrs, uattrs)
}

func TestCoreDocument_UpdateAttributes_updates_nil(t *testing.T) {
	oldCAttrs := map[string]attribute{
		"time_test": {
			Type:  AttrTimestamp.String(),
			Value: time.Now().UTC().Format(time.RFC3339),
		},

		"string_test": {
			Type:  AttrString.String(),
			Value: "some string",
		},

		"bytes_test": {
			Type:  AttrBytes.String(),
			Value: hexutil.Encode([]byte("some bytes data")),
		},

		"int256_test": {
			Type:  AttrInt256.String(),
			Value: "1000000001",
		},
	}

	oldAttrs := toAttrsMap(t, oldCAttrs)
	oldPattrs, err := toProtocolAttributes(oldAttrs)
	assert.NoError(t, err)

	upattrs, uattrs, err := updateAttributes(oldPattrs, nil)
	assert.NoError(t, err)

	assert.Equal(t, upattrs, oldPattrs)
	assert.Equal(t, oldAttrs, uattrs)
}

func TestCoreDocument_UpdateAttributes_both_nil(t *testing.T) {
	upattrs, uattrs, err := updateAttributes(nil, nil)
	assert.NoError(t, err)
	assert.Len(t, upattrs, 0)
	assert.Len(t, uattrs, 0)
}

func TestCoreDocument_newRoleWithCollaborators(t *testing.T) {
	accountID1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)
	accountID2, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	role := newRoleWithCollaborators(accountID1, accountID2)
	assert.Len(t, role.Collaborators, 2)
	assert.Equal(t, role.Collaborators[0], accountID1.ToBytes())
	assert.Equal(t, role.Collaborators[1], accountID2.ToBytes())
}

func TestCoreDocument_CreateProofs(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	assert.NotNil(t, cd)

	docType := "doc"

	dataLeaves := []proofs.LeafNode{
		{
			Property: proofs.Property{
				Parent:     nil,
				Text:       "name.test1",
				Compact:    utils.RandomSlice(32),
				NameFormat: "",
			},
			Value:  utils.RandomSlice(32),
			Salt:   utils.RandomSlice(32),
			Hash:   utils.RandomSlice(32),
			Hashed: true,
		},
		{
			Property: proofs.Property{
				Parent:     nil,
				Text:       "name.test2",
				Compact:    utils.RandomSlice(32),
				NameFormat: "",
			},
			Value:  utils.RandomSlice(32),
			Salt:   utils.RandomSlice(32),
			Hash:   utils.RandomSlice(32),
			Hashed: false,
		},
	}

	fields := []string{
		"cd_tree.author",
		"cd_tree.timestamp",
	}

	drTree, err := cd.DocumentRootTree(docType, dataLeaves)
	assert.NoError(t, err)

	signaturesTree, err := cd.GetSignaturesDataTree()
	assert.NoError(t, err)

	signingTree, err := cd.SigningDataTree(docType, dataLeaves)
	assert.NoError(t, err)

	dataPrefix, err := getDataTreePrefix(dataLeaves)
	assert.NoError(t, err)

	treeProofs := make(map[string]*proofs.DocumentTree, 4)

	treeProofs[dataPrefix] = signingTree
	treeProofs[CDTreePrefix] = treeProofs[dataPrefix]
	treeProofs[SignaturesTreePrefix] = signaturesTree
	treeProofs[DRTreePrefix] = drTree

	rawProofs, err := generateProofs(fields, treeProofs)
	assert.NoError(t, err)

	res, err := cd.CreateProofs(docType, dataLeaves, fields)
	assert.NoError(t, err)
	assert.Equal(t, res.FieldProofs, rawProofs)
	assert.Equal(t, res.SigningRoot, signingTree.RootHash())
	assert.Equal(t, res.SignaturesRoot, signaturesTree.RootHash())
}

func TestCoreDocument_CreateProofs_DocumentRootTreeError(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	assert.NotNil(t, cd)

	docType := "doc"

	fields := []string{
		"cd_tree.author",
		"cd_tree.timestamp",
	}

	res, err := cd.CreateProofs(docType, nil, fields)
	assert.True(t, errors.IsOfType(ErrCDTree, err))
	assert.Nil(t, res)
}

func TestCoreDocument_CreateProofs_DataPrefixError(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	assert.NotNil(t, cd)

	docType := "doc"

	dataLeaves := []proofs.LeafNode{
		{
			Property: proofs.Property{
				Parent: nil,
				// Invalid property text
				Text:       "invalidPropertyText",
				Compact:    utils.RandomSlice(32),
				NameFormat: "",
			},
			Value:  utils.RandomSlice(32),
			Salt:   utils.RandomSlice(32),
			Hash:   utils.RandomSlice(32),
			Hashed: true,
		},
	}

	fields := []string{
		"cd_tree.author",
		"cd_tree.timestamp",
	}

	res, err := cd.CreateProofs(docType, dataLeaves, fields)
	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestCoreDocument_CreateProofs_GenerateProofsError(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	assert.NotNil(t, cd)

	docType := "doc"

	dataLeaves := []proofs.LeafNode{
		{
			Property: proofs.Property{
				Parent:     nil,
				Text:       "name.test1",
				Compact:    utils.RandomSlice(32),
				NameFormat: "",
			},
			Value:  utils.RandomSlice(32),
			Salt:   utils.RandomSlice(32),
			Hash:   utils.RandomSlice(32),
			Hashed: true,
		},
		{
			Property: proofs.Property{
				Parent:     nil,
				Text:       "name.test2",
				Compact:    utils.RandomSlice(32),
				NameFormat: "",
			},
			Value:  utils.RandomSlice(32),
			Salt:   utils.RandomSlice(32),
			Hash:   utils.RandomSlice(32),
			Hashed: false,
		},
	}

	// Invalid field
	fields := []string{
		"invalidField",
	}

	res, err := cd.CreateProofs(docType, dataLeaves, fields)
	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestCoreDocument_CalculateTransitionRulesFingerprint(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	assert.NotNil(t, cd)

	res, err := cd.CalculateTransitionRulesFingerprint()
	assert.NoError(t, err)
	assert.Nil(t, res)

	roleKey := utils.RandomSlice(32)

	roles := []*coredocumentpb.Role{
		{
			RoleKey: roleKey,
			Collaborators: [][]byte{
				utils.RandomSlice(32),
			},
			Nfts: [][]byte{
				utils.RandomSlice(32),
			},
		},
	}

	transitionRules := []*coredocumentpb.TransitionRule{
		{
			RuleKey: utils.RandomSlice(32),
			Roles: [][]byte{
				roleKey,
			},
			MatchType: 0,
			Field:     utils.RandomSlice(32),
			Action:    0,
			ComputeFields: [][]byte{
				utils.RandomSlice(32),
			},
			ComputeTargetField: utils.RandomSlice(32),
			ComputeCode:        utils.RandomSlice(32),
		},
	}

	cd.Document.Roles = roles
	cd.Document.TransitionRules = transitionRules

	res, err = cd.CalculateTransitionRulesFingerprint()
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Len(t, res, 32)
}

func TestCoreDocument_CalculateSignaturesRoot(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	assert.NotNil(t, cd)

	res, err := cd.CalculateSignaturesRoot()
	assert.NoError(t, err)
	assert.NotNil(t, res)
}

func TestCoreDocument_GetSignaturesDataTree(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	assert.NotNil(t, cd)

	res, err := cd.GetSignaturesDataTree()
	assert.NoError(t, err)
	assert.NotNil(t, res)
}

func TestGenerateTransitionFingerprintHash(t *testing.T) {
	roleKey := utils.RandomSlice(32)

	roles := []*coredocumentpb.Role{
		{
			RoleKey: roleKey,
			Collaborators: [][]byte{
				utils.RandomSlice(32),
			},
			Nfts: [][]byte{
				utils.RandomSlice(32),
			},
		},
	}

	transitionRules := []*coredocumentpb.TransitionRule{
		{
			RuleKey: utils.RandomSlice(32),
			Roles: [][]byte{
				roleKey,
			},
			MatchType: 0,
			Field:     utils.RandomSlice(32),
			Action:    0,
			ComputeFields: [][]byte{
				utils.RandomSlice(32),
			},
			ComputeTargetField: utils.RandomSlice(32),
			ComputeCode:        utils.RandomSlice(32),
		},
	}

	fp := &coredocumentpb.TransitionRulesFingerprint{
		Roles:           roles,
		TransitionRules: transitionRules,
	}

	res, err := generateTransitionRulesFingerprintHash(fp)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Len(t, res, 32)

	res, err = generateTransitionRulesFingerprintHash(nil)
	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestGetDataTreePrefix(t *testing.T) {
	prefix := "prefix"

	dataLeaves := []proofs.LeafNode{
		{
			Property: proofs.Property{
				Parent:     nil,
				Text:       fmt.Sprintf("%s.%s", prefix, "test1"),
				Compact:    utils.RandomSlice(32),
				NameFormat: "",
			},
			Value:  utils.RandomSlice(32),
			Salt:   utils.RandomSlice(32),
			Hash:   utils.RandomSlice(32),
			Hashed: true,
		},
	}

	res, err := getDataTreePrefix(dataLeaves)
	assert.NoError(t, err)
	assert.Equal(t, prefix, res)

	dataLeaves = []proofs.LeafNode{
		{
			Property: proofs.Property{
				Parent:     nil,
				Text:       "invalidtext", // Invalid text
				Compact:    utils.RandomSlice(32),
				NameFormat: "",
			},
			Value:  utils.RandomSlice(32),
			Salt:   utils.RandomSlice(32),
			Hash:   utils.RandomSlice(32),
			Hashed: true,
		},
	}

	res, err = getDataTreePrefix(dataLeaves)
	assert.NotNil(t, err)
	assert.Equal(t, "", res)
}

func TestGenerateProofs(t *testing.T) {
	h, err := blake2b.New256(nil)
	assert.NoError(t, err)

	cd, err := newCoreDocument()
	assert.NoError(t, err)

	testTree, err := cd.DefaultTreeWithPrefix("prefix", []byte{1, 0, 0, 0})
	assert.NoError(t, err)

	props := []proofs.Property{NewLeafProperty("prefix.sample_field", []byte{1, 0, 0, 0, 0, 0, 0, 200}), NewLeafProperty("prefix.sample_field2", []byte{1, 0, 0, 0, 0, 0, 0, 202})}
	compactProps := [][]byte{props[0].Compact, props[1].Compact}

	err = testTree.AddLeaf(proofs.LeafNode{Hash: utils.RandomSlice(32), Hashed: true, Property: props[0]})
	assert.NoError(t, err)

	err = testTree.AddLeaf(proofs.LeafNode{Hash: utils.RandomSlice(32), Hashed: true, Property: props[1]})
	assert.NoError(t, err)

	err = testTree.Generate()
	assert.NoError(t, err)

	docAny := &anypb.Any{
		TypeUrl: documenttypes.InvoiceDataTypeUrl,
		Value:   []byte{},
	}

	cd, err = newCoreDocument()
	assert.NoError(t, err)

	cd.Modified = true
	cd.Document.EmbeddedData = docAny

	dataRoot, err := cd.SigningDataTree(documenttypes.InvoiceDataTypeUrl, testTree.GetLeaves())
	assert.NoError(t, err)

	cdTree, err := cd.coredocTree(documenttypes.InvoiceDataTypeUrl)
	assert.NoError(t, err)
	tests := []struct {
		fieldName   string
		fromCoreDoc bool
		proofLength int
	}{
		{
			"prefix.sample_field",
			false,
			5,
		},
		{
			CDTreePrefix + ".document_identifier",
			true,
			5,
		},
		{
			"prefix.sample_field2",
			false,
			5,
		},
		{
			CDTreePrefix + ".next_version",
			true,
			5,
		},
	}
	for _, test := range tests {
		t.Run(test.fieldName, func(t *testing.T) {
			p, err := cd.CreateProofs(documenttypes.InvoiceDataTypeUrl, testTree.GetLeaves(), []string{test.fieldName})
			assert.NoError(t, err)
			assert.Equal(t, test.proofLength, len(p.FieldProofs[0].SortedHashes))

			_, l := testTree.GetLeafByProperty(test.fieldName)

			if !test.fromCoreDoc {
				assert.Contains(t, compactProps, l.Property.CompactName())
			} else {
				_, l = cdTree.GetLeafByProperty(test.fieldName)
			}
			assert.NotNil(t, l)

			valid, err := proofs.ValidateProofSortedHashes(l.Hash, p.FieldProofs[0].SortedHashes, dataRoot.RootHash(), h)
			assert.NoError(t, err)
			assert.True(t, valid)

			docRoot, err := cd.CalculateDocumentRoot(documenttypes.InvoiceDataTypeUrl, testTree.GetLeaves())
			assert.NoError(t, err)

			signRoot, err := cd.CalculateSigningRoot(documenttypes.InvoiceDataTypeUrl, testTree.GetLeaves())
			assert.NoError(t, err)

			// Validate document root for basic data tree
			calcDocRoot := proofs.HashTwoValues(signRoot, p.SignaturesRoot, h)
			assert.Equal(t, docRoot, calcDocRoot)
		})
	}

	// Prefix test

	cd, err = newCoreDocument()
	assert.NoError(t, err)
	assert.NotNil(t, cd)

	docType := "doc"

	treePrefix := "tree_prefix"

	dataLeaves := []proofs.LeafNode{
		{
			Property: proofs.Property{
				Parent:     nil,
				Text:       fmt.Sprintf("%s.%s", treePrefix, "test1"),
				Compact:    utils.RandomSlice(32),
				NameFormat: "",
			},
			Value:  utils.RandomSlice(32),
			Salt:   utils.RandomSlice(32),
			Hash:   utils.RandomSlice(32),
			Hashed: true,
		},
		{
			Property: proofs.Property{
				Parent:     nil,
				Text:       fmt.Sprintf("%s.%s", treePrefix, "test2"),
				Compact:    utils.RandomSlice(32),
				NameFormat: "",
			},
			Value:  utils.RandomSlice(32),
			Salt:   utils.RandomSlice(32),
			Hash:   utils.RandomSlice(32),
			Hashed: false,
		},
	}

	tree, err := cd.SigningDataTree(docType, dataLeaves)
	assert.NoError(t, err)

	treeProofs := map[string]*proofs.DocumentTree{
		treePrefix: tree,
	}

	fields := []string{
		fmt.Sprintf("%s.%s", treePrefix, "test1"),
		fmt.Sprintf("%s.%s", treePrefix, "test2"),
	}

	res, err := generateProofs(fields, treeProofs)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	// Invalid prefix

	fields = []string{
		"invalidPrefix.test1",
	}

	res, err = generateProofs(fields, treeProofs)
	assert.NotNil(t, err)
	assert.Nil(t, res)

	// Invalid field

	fields = []string{
		fmt.Sprintf("%s.%s", treePrefix, "test3"),
	}

	res, err = generateProofs(fields, treeProofs)
	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestGenerateProofs_PrefixTest(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	assert.NotNil(t, cd)

	docType := "doc"

	treePrefix := "tree_prefix"

	dataLeaves := []proofs.LeafNode{
		{
			Property: proofs.Property{
				Parent:     nil,
				Text:       fmt.Sprintf("%s.%s", treePrefix, "test1"),
				Compact:    utils.RandomSlice(32),
				NameFormat: "",
			},
			Value:  utils.RandomSlice(32),
			Salt:   utils.RandomSlice(32),
			Hash:   utils.RandomSlice(32),
			Hashed: true,
		},
		{
			Property: proofs.Property{
				Parent:     nil,
				Text:       fmt.Sprintf("%s.%s", treePrefix, "test2"),
				Compact:    utils.RandomSlice(32),
				NameFormat: "",
			},
			Value:  utils.RandomSlice(32),
			Salt:   utils.RandomSlice(32),
			Hash:   utils.RandomSlice(32),
			Hashed: false,
		},
	}

	tree, err := cd.SigningDataTree(docType, dataLeaves)
	assert.NoError(t, err)

	treeProofs := map[string]*proofs.DocumentTree{
		treePrefix: tree,
	}

	fields := []string{
		fmt.Sprintf("%s.%s", treePrefix, "test1"),
		fmt.Sprintf("%s.%s", treePrefix, "test2"),
	}

	res, err := generateProofs(fields, treeProofs)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	// Invalid prefix

	fields = []string{
		"invalidPrefix.test1",
	}

	res, err = generateProofs(fields, treeProofs)
	assert.NotNil(t, err)
	assert.Nil(t, res)

	// Invalid field

	fields = []string{
		fmt.Sprintf("%s.%s", treePrefix, "test3"),
	}

	res, err = generateProofs(fields, treeProofs)
	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestGetSignaturesTree(t *testing.T) {
	docAny := &anypb.Any{
		TypeUrl: documenttypes.InvoiceDataTypeUrl,
		Value:   []byte{},
	}

	cd, err := newCoreDocument()
	assert.NoError(t, err)

	cd.Modified = true
	cd.Document.EmbeddedData = docAny
	sig := &coredocumentpb.Signature{
		SignerId:    utils.RandomSlice(32),
		PublicKey:   utils.RandomSlice(32),
		SignatureId: utils.RandomSlice(64),
		Signature:   utils.RandomSlice(32),
	}

	cd.Document.SignatureData.Signatures = []*coredocumentpb.Signature{sig}
	signatureTree, err := cd.GetSignaturesDataTree()

	signatureRoot, err := cd.CalculateSignaturesRoot()
	assert.NoError(t, err)
	assert.NoError(t, err)
	assert.NotNil(t, signatureTree)
	assert.Equal(t, signatureTree.RootHash(), signatureRoot)

	lengthIdx, lengthLeaf := signatureTree.GetLeafByProperty(SignaturesTreePrefix + ".signatures.length")
	assert.Equal(t, 0, lengthIdx)
	assert.NotNil(t, lengthLeaf)
	assert.Equal(t, SignaturesTreePrefix+".signatures.length", lengthLeaf.Property.ReadableName())
	assert.Equal(t, append(CompactProperties(SignaturesTreePrefix), []byte{0, 0, 0, 1}...), lengthLeaf.Property.CompactName())

	signerKey := hexutil.Encode(sig.SignatureId)
	_, signerLeaf := signatureTree.GetLeafByProperty(fmt.Sprintf("%s.signatures[%s]", SignaturesTreePrefix, signerKey))
	assert.NotNil(t, signerLeaf)
	assert.Equal(t, fmt.Sprintf("%s.signatures[%s]", SignaturesTreePrefix, signerKey), signerLeaf.Property.ReadableName())
	assert.Equal(t, append(CompactProperties(SignaturesTreePrefix), append([]byte{0, 0, 0, 1}, sig.SignatureId...)...), signerLeaf.Property.CompactName())
	assert.Equal(t, byteutils.AddZeroBytesSuffix(sig.Signature, 66), signerLeaf.Value)
}

//var ctx map[string]interface{}
//var cfg config.Configuration
//var did = testingidentity.GenerateRandomDID()
//
//func TestMain(m *testing.M) {
//	ctx = make(map[string]interface{})
//	ethClient := &ethereum.MockEthClient{}
//	ethClient.On("GetEthClient").Return(nil)
//	ctx[ethereum.BootstrappedEthereumClient] = ethClient
//	centChainClient := &centchain.MockAPI{}
//	ctx[centchain.BootstrappedCentChainClient] = centChainClient
//
//	ibootstappers := []bootstrap.TestBootstrapper{
//		&testlogging.TestLoggingBootstrapper{},
//		&config.Bootstrapper{},
//		&leveldb.Bootstrapper{},
//		jobs.Bootstrapper{},
//		&configstore.Bootstrapper{},
//		&anchors.Bootstrapper{},
//		&Bootstrapper{},
//	}
//	ctx[identity.BootstrappedDIDService] = &testingcommons.MockIdentityService{}
//	ctx[identity.BootstrappedDIDFactory] = &identity.MockFactory{}
//	bootstrap.RunTestBootstrappers(ibootstappers, ctx)
//	cfg = ctx[bootstrap.BootstrappedConfig].(config.Configuration)
//	cfg.Set("keys.p2p.publicKey", "../build/resources/p2pKey.pub.pem")
//	cfg.Set("keys.p2p.privateKey", "../build/resources/p2pKey.key.pem")
//	cfg.Set("keys.signing.publicKey", "../build/resources/signingKey.pub.pem")
//	cfg.Set("keys.signing.privateKey", "../build/resources/signingKey.key.pem")
//	cfg.Set("identityId", did.String())
//	result := m.Run()
//	bootstrap.RunTestTeardown(ibootstappers)
//	os.Exit(result)
//}
//
//func Test_fetchUniqueCollaborators(t *testing.T) {
//	o1 := testingidentity.GenerateRandomDID()
//	o2 := testingidentity.GenerateRandomDID()
//	n1 := testingidentity.GenerateRandomDID()
//	tests := []struct {
//		old    []identity.DID
//		new    []identity.DID
//		result []identity.DID
//	}{
//		// when old cs are nil
//		{
//			new: []identity.DID{n1},
//		},
//
//		{
//			old:    []identity.DID{o1, o2},
//			result: []identity.DID{o1, o2},
//		},
//
//		{
//			old:    []identity.DID{o1},
//			new:    []identity.DID{n1},
//			result: []identity.DID{o1},
//		},
//
//		{
//			old:    []identity.DID{o1, n1},
//			new:    []identity.DID{n1},
//			result: []identity.DID{o1},
//		},
//
//		{
//			old:    []identity.DID{o1, n1},
//			new:    []identity.DID{o2},
//			result: []identity.DID{o1, n1},
//		},
//	}
//
//	for _, c := range tests {
//		uc := filterCollaborators(c.old, c.new...)
//		assert.Equal(t, c.result, uc)
//	}
//}
//
//func TestCoreDocument_Author(t *testing.T) {
//	cd, err := newCoreDocument()
//	assert.NoError(t, err)
//
//	did := testingidentity.GenerateRandomDID()
//	cd.Document.Author = did[:]
//	a, err := cd.Author()
//	assert.NoError(t, err)
//
//	aID, err := identity.NewDIDFromBytes(cd.Document.Author)
//	assert.NoError(t, err)
//	assert.Equal(t, a, aID)
//}
//
//func TestCoreDocument_ID(t *testing.T) {
//	cd, err := newCoreDocument()
//	assert.NoError(t, err)
//
//	assert.Equal(t, cd.Document.DocumentIdentifier, cd.ID())
//}
//
//func TestNewCoreDocumentWithCollaborators(t *testing.T) {
//	did1 := testingidentity.GenerateRandomDID()
//	did2 := testingidentity.GenerateRandomDID()
//	c := &CollaboratorsAccess{
//		ReadCollaborators:      []identity.DID{did1},
//		ReadWriteCollaborators: []identity.DID{did2},
//	}
//	cd, err := NewCoreDocument([]byte("inv"), *c, nil)
//	assert.NoError(t, err)
//
//	collabs, err := cd.GetCollaborators(identity.DID{})
//	assert.NoError(t, err)
//	assert.Equal(t, did1, collabs.ReadCollaborators[0])
//	assert.Equal(t, did2, collabs.ReadWriteCollaborators[0])
//}
//
//
//func TestCoreDocument_AddUpdateLog(t *testing.T) {
//	did1 := testingidentity.GenerateRandomDID()
//	cd, err := newCoreDocument()
//	assert.NoError(t, err)
//
//	err = cd.AddUpdateLog(did1)
//	assert.NoError(t, err)
//	assert.Equal(t, cd.Document.Author, did1[:])
//	assert.True(t, cd.Modified)
//}
//
//func TestGetSigningProofHash(t *testing.T) {
//	docAny := &any.Any{
//		TypeUrl: documenttypes.InvoiceDataTypeUrl,
//		Value:   []byte{},
//	}
//
//	cd, err := newCoreDocument()
//	assert.NoError(t, err)
//	cd.GetTestCoreDocWithReset().EmbeddedData = docAny
//	testTree, err := cd.DefaultTreeWithPrefix("invoice", []byte{1, 0, 0, 0})
//	assert.NoError(t, err)
//	signingRoot, err := cd.CalculateSigningRoot(documenttypes.InvoiceDataTypeUrl, testTree.GetLeaves())
//	assert.Nil(t, err)
//
//	cd.GetTestCoreDocWithReset()
//	docRoot, err := cd.CalculateDocumentRoot(documenttypes.InvoiceDataTypeUrl, testTree.GetLeaves())
//	assert.Nil(t, err)
//
//	signatureTree, err := cd.GetSignaturesDataTree()
//	assert.Nil(t, err)
//	h, err := blake2b.New256(nil)
//	assert.NoError(t, err)
//	valid, err := proofs.ValidateProofHashes(signingRoot, []*proofspb.MerkleHash{{Right: signatureTree.RootHash()}}, docRoot, h)
//	assert.True(t, valid)
//	assert.Nil(t, err)
//}
//
//
//// TestGetDocumentRootTree tests that the document root tree is properly calculated
//func TestGetDocumentRootTree(t *testing.T) {
//	cd, err := newCoreDocument()
//	assert.NoError(t, err)
//
//	sig := &coredocumentpb.Signature{
//		SignerId:    utils.RandomSlice(identity.DIDLength),
//		PublicKey:   utils.RandomSlice(32),
//		SignatureId: utils.RandomSlice(52),
//		Signature:   utils.RandomSlice(32),
//	}
//	cd.GetTestCoreDocWithReset().SignatureData.Signatures = []*coredocumentpb.Signature{sig}
//	testTree, err := cd.DefaultTreeWithPrefix("invoice", []byte{1, 0, 0, 0})
//	assert.NoError(t, err)
//
//	// successful document root generation
//	signingRoot, err := cd.CalculateSigningRoot(documenttypes.InvoiceDataTypeUrl, testTree.GetLeaves())
//	assert.NoError(t, err)
//	tree, err := cd.DocumentRootTree(documenttypes.InvoiceDataTypeUrl, testTree.GetLeaves())
//	assert.NoError(t, err)
//	_, leaf := tree.GetLeafByProperty(fmt.Sprintf("%s.%s", DRTreePrefix, SigningRootField))
//	assert.NotNil(t, leaf)
//	assert.Equal(t, signingRoot, leaf.Hash)
//
//	// Get signaturesLeaf
//	_, signaturesLeaf := tree.GetLeafByProperty(fmt.Sprintf("%s.%s", DRTreePrefix, SignaturesRootField))
//	assert.NotNil(t, signaturesLeaf)
//	assert.Equal(t, fmt.Sprintf("%s.%s", DRTreePrefix, SignaturesRootField), signaturesLeaf.Property.ReadableName())
//	assert.Equal(t, append(CompactProperties(DRTreePrefix), CompactProperties(SignaturesRootField)...), signaturesLeaf.Property.CompactName())
//}
//
//func TestGetDataTreePrefix(t *testing.T) {
//	cds, err := newCoreDocument()
//	assert.NoError(t, err)
//	testTree, err := cds.DefaultTreeWithPrefix("prefix", []byte{1, 0, 0, 0})
//	assert.NoError(t, err)
//	props := []proofs.Property{NewLeafProperty("prefix.sample_field", []byte{1, 0, 0, 0, 0, 0, 0, 200}), NewLeafProperty("prefix.sample_field2", []byte{1, 0, 0, 0, 0, 0, 0, 202})}
//	//compactProps := [][]byte{props[0].Compact, props[1].Compact}
//	err = testTree.AddLeaf(proofs.LeafNode{Hash: utils.RandomSlice(32), Hashed: true, Property: props[0]})
//	assert.NoError(t, err)
//	err = testTree.AddLeaf(proofs.LeafNode{Hash: utils.RandomSlice(32), Hashed: true, Property: props[1]})
//	assert.NoError(t, err)
//	err = testTree.Generate()
//	assert.NoError(t, err)
//
//	// nil length leaves
//	prfx, err := getDataTreePrefix(nil)
//	assert.Error(t, err)
//
//	// zero length leaves
//	prfx, err = getDataTreePrefix([]proofs.LeafNode{})
//	assert.Error(t, err)
//
//	// success
//	prfx, err = getDataTreePrefix(testTree.GetLeaves())
//	assert.NoError(t, err)
//	assert.Equal(t, "prefix", prfx)
//
//	// non-prefixed tree error
//	testTree, err = cds.DefaultTreeWithPrefix("", []byte{1, 0, 0, 0})
//	assert.NoError(t, err)
//	props = []proofs.Property{NewLeafProperty("sample_field", []byte{1, 0, 0, 0, 0, 0, 0, 200}), NewLeafProperty("sample_field2", []byte{1, 0, 0, 0, 0, 0, 0, 202})}
//	//compactProps := [][]byte{props[0].Compact, props[1].Compact}
//	err = testTree.AddLeaf(proofs.LeafNode{Hash: utils.RandomSlice(32), Hashed: true, Property: props[0]})
//	assert.NoError(t, err)
//	err = testTree.AddLeaf(proofs.LeafNode{Hash: utils.RandomSlice(32), Hashed: true, Property: props[1]})
//	assert.NoError(t, err)
//	err = testTree.Generate()
//	assert.NoError(t, err)
//
//	prfx, err = getDataTreePrefix(testTree.GetLeaves())
//	assert.Error(t, err)
//}
//
//func TestCoreDocument_getReadCollaborators(t *testing.T) {
//	id1 := testingidentity.GenerateRandomDID()
//	id2 := testingidentity.GenerateRandomDID()
//	cas := CollaboratorsAccess{
//		ReadWriteCollaborators: []identity.DID{id1},
//	}
//	cd, err := NewCoreDocument(nil, cas, nil)
//	assert.NoError(t, err)
//	cs, err := cd.getReadCollaborators(coredocumentpb.Action_ACTION_READ_SIGN)
//	assert.NoError(t, err)
//	assert.Len(t, cs, 1)
//	assert.Equal(t, cs[0], id1)
//
//	cs, err = cd.getReadCollaborators(coredocumentpb.Action_ACTION_READ)
//	assert.NoError(t, err)
//	assert.Len(t, cs, 0)
//	role := newRoleWithCollaborators(id2)
//	cd.Document.Roles = append(cd.Document.Roles, role)
//	cd.addNewReadRule(role.RoleKey, coredocumentpb.Action_ACTION_READ)
//
//	cs, err = cd.getReadCollaborators(coredocumentpb.Action_ACTION_READ)
//	assert.NoError(t, err)
//	assert.Len(t, cs, 1)
//	assert.Equal(t, cs[0], id2)
//
//	cs, err = cd.getReadCollaborators(coredocumentpb.Action_ACTION_READ, coredocumentpb.Action_ACTION_READ_SIGN)
//	assert.NoError(t, err)
//	assert.Len(t, cs, 2)
//	assert.Contains(t, cs, id1)
//	assert.Contains(t, cs, id2)
//}
//
//func TestCoreDocument_getWriteCollaborators(t *testing.T) {
//	id1 := testingidentity.GenerateRandomDID()
//	id2 := testingidentity.GenerateRandomDID()
//	cas := CollaboratorsAccess{ReadWriteCollaborators: []identity.DID{id1}}
//	cd, err := NewCoreDocument([]byte("inv"), cas, nil)
//	assert.NoError(t, err)
//	cs, err := cd.getWriteCollaborators(coredocumentpb.TransitionAction_TRANSITION_ACTION_EDIT)
//	assert.NoError(t, err)
//	assert.Len(t, cs, 1)
//
//	role := newRoleWithCollaborators(id2)
//	cd.Document.Roles = append(cd.Document.Roles, role)
//	cd.addNewTransitionRule(role.RoleKey, coredocumentpb.FieldMatchType_FIELD_MATCH_TYPE_PREFIX, nil, coredocumentpb.TransitionAction_TRANSITION_ACTION_EDIT)
//
//	cs, err = cd.getWriteCollaborators(coredocumentpb.TransitionAction_TRANSITION_ACTION_EDIT)
//	assert.NoError(t, err)
//	assert.Len(t, cs, 2)
//	assert.Equal(t, cs[1], id2)
//}
//
//func TestCoreDocument_GetCollaborators(t *testing.T) {
//	id1 := testingidentity.GenerateRandomDID()
//	id2 := testingidentity.GenerateRandomDID()
//	id3 := testingidentity.GenerateRandomDID()
//	cas := CollaboratorsAccess{ReadWriteCollaborators: []identity.DID{id1}}
//	cd, err := NewCoreDocument(nil, cas, nil)
//	assert.NoError(t, err)
//	cs, err := cd.GetCollaborators()
//	assert.NoError(t, err)
//	assert.Len(t, cs.ReadCollaborators, 0)
//	assert.Len(t, cs.ReadWriteCollaborators, 1)
//	assert.Equal(t, cs.ReadWriteCollaborators[0], id1)
//
//	role := newRoleWithCollaborators(id2)
//	cd.Document.Roles = append(cd.Document.Roles, role)
//	cd.addNewReadRule(role.RoleKey, coredocumentpb.Action_ACTION_READ)
//
//	cs, err = cd.GetCollaborators()
//	assert.NoError(t, err)
//	assert.Len(t, cs.ReadCollaborators, 1)
//	assert.Contains(t, cs.ReadCollaborators, id2)
//	assert.Len(t, cs.ReadWriteCollaborators, 1)
//	assert.Contains(t, cs.ReadWriteCollaborators, id1)
//
//	cs, err = cd.GetCollaborators(id2)
//	assert.NoError(t, err)
//	assert.Len(t, cs.ReadCollaborators, 0)
//	assert.Len(t, cs.ReadWriteCollaborators, 1)
//	assert.Contains(t, cs.ReadWriteCollaborators, id1)
//
//	role2 := newRoleWithCollaborators(id3)
//	cd.Document.Roles = append(cd.Document.Roles, role2)
//	cd.addNewTransitionRule(role2.RoleKey, coredocumentpb.FieldMatchType_FIELD_MATCH_TYPE_PREFIX, nil, coredocumentpb.TransitionAction_TRANSITION_ACTION_EDIT)
//	cs, err = cd.GetCollaborators()
//	assert.NoError(t, err)
//	assert.Len(t, cs.ReadCollaborators, 1)
//	assert.Contains(t, cs.ReadCollaborators, id2)
//	assert.Len(t, cs.ReadWriteCollaborators, 2)
//	assert.Contains(t, cs.ReadWriteCollaborators, id1)
//	assert.Contains(t, cs.ReadWriteCollaborators, id3)
//}
//
//func TestCoreDocument_GetSignCollaborators(t *testing.T) {
//	id1 := testingidentity.GenerateRandomDID()
//	id2 := testingidentity.GenerateRandomDID()
//	cas := CollaboratorsAccess{ReadWriteCollaborators: []identity.DID{id1}}
//	cd, err := NewCoreDocument(nil, cas, nil)
//	assert.NoError(t, err)
//	cs, err := cd.GetSignerCollaborators()
//	assert.NoError(t, err)
//	assert.Len(t, cs, 1)
//	assert.Equal(t, cs[0], id1)
//
//	cs, err = cd.GetSignerCollaborators(id1)
//	assert.NoError(t, err)
//	assert.Len(t, cs, 0)
//
//	role := newRoleWithCollaborators(id2)
//	cd.Document.Roles = append(cd.Document.Roles, role)
//	cd.addNewReadRule(role.RoleKey, coredocumentpb.Action_ACTION_READ)
//
//	cs, err = cd.GetSignerCollaborators()
//	assert.NoError(t, err)
//	assert.Len(t, cs, 1)
//	assert.Contains(t, cs, id1)
//	assert.NotContains(t, cs, id2)
//
//	cs, err = cd.GetSignerCollaborators(id2)
//	assert.NoError(t, err)
//	assert.Len(t, cs, 1)
//	assert.Contains(t, cs, id1)
//}
//
//func TestCoreDocument_Attribute(t *testing.T) {
//	id1 := testingidentity.GenerateRandomDID()
//	cas := CollaboratorsAccess{ReadWriteCollaborators: []identity.DID{id1}}
//	cd, err := NewCoreDocument(nil, cas, nil)
//	assert.NoError(t, err)
//	cd.Attributes = nil
//	label := "com.basf.deliverynote.chemicalnumber"
//	value := "100"
//	key, err := AttrKeyFromLabel(label)
//	assert.NoError(t, err)
//
//	// failed get
//	assert.False(t, cd.AttributeExists(key))
//	_, err = cd.GetAttribute(key)
//	assert.Error(t, err)
//
//	// failed delete
//	_, err = cd.DeleteAttribute(key, true, nil)
//	assert.Error(t, err)
//
//	// success
//	attr, err := NewStringAttribute(label, AttrString, value)
//	assert.NoError(t, err)
//	cd, err = cd.AddAttributes(CollaboratorsAccess{}, true, nil, attr)
//	assert.NoError(t, err)
//	assert.Len(t, cd.Attributes, 1)
//	assert.Len(t, cd.GetAttributes(), 1)
//
//	// check
//	assert.True(t, cd.AttributeExists(key))
//	attr, err = cd.GetAttribute(key)
//	assert.NoError(t, err)
//	assert.Equal(t, key, attr.Key)
//	assert.Equal(t, label, attr.KeyLabel)
//	str, err := attr.Value.String()
//	assert.NoError(t, err)
//	assert.Equal(t, value, str)
//	assert.Equal(t, AttrString, attr.Value.Type)
//
//	// update
//	nvalue := "2000"
//	attr, err = NewStringAttribute(label, AttrDecimal, nvalue)
//	assert.NoError(t, err)
//	cd, err = cd.AddAttributes(CollaboratorsAccess{}, true, nil, attr)
//	assert.NoError(t, err)
//	assert.True(t, cd.AttributeExists(key))
//	attr, err = cd.GetAttribute(key)
//	assert.NoError(t, err)
//	assert.Len(t, cd.Attributes, 1)
//	assert.Len(t, cd.GetAttributes(), 1)
//	assert.Equal(t, key, attr.Key)
//	assert.Equal(t, label, attr.KeyLabel)
//	str, err = attr.Value.String()
//	assert.NoError(t, err)
//	assert.NotEqual(t, value, str)
//	assert.Equal(t, nvalue, str)
//	assert.Equal(t, AttrDecimal, attr.Value.Type)
//
//	// delete
//	cd, err = cd.DeleteAttribute(key, true, nil)
//	assert.NoError(t, err)
//	assert.Len(t, cd.Attributes, 0)
//	assert.Len(t, cd.GetAttributes(), 0)
//	assert.False(t, cd.AttributeExists(key))
//}
//
//func TestCoreDocument_SetUsedAnchorRepoAddress(t *testing.T) {
//	addr := testingidentity.GenerateRandomDID()
//	cd := new(CoreDocument)
//	cd.SetUsedAnchorRepoAddress(addr.ToAddress())
//	assert.Equal(t, addr.ToAddress().Bytes(), cd.AnchorRepoAddress().Bytes())
//}
//
//
//func TestCoreDocument_Status(t *testing.T) {
//	cd, err := NewCoreDocument(nil, CollaboratorsAccess{}, nil)
//	assert.NoError(t, err)
//	assert.Equal(t, cd.GetStatus(), Pending)
//
//	// set status to Committed
//	err = cd.SetStatus(Committed)
//	assert.NoError(t, err)
//	assert.Equal(t, cd.GetStatus(), Committed)
//
//	// try to update status to Committing
//	err = cd.SetStatus(Committing)
//	assert.Error(t, err)
//	assert.True(t, errors.IsOfType(ErrCDStatus, err))
//	assert.Equal(t, cd.GetStatus(), Committed)
//}
//
//func TestCoreDocument_RemoveCollaborators(t *testing.T) {
//	did1 := testingidentity.GenerateRandomDID()
//	did2 := testingidentity.GenerateRandomDID()
//	did3 := testingidentity.GenerateRandomDID() // missing
//	cd, err := NewCoreDocument(
//		nil,
//		CollaboratorsAccess{
//			ReadWriteCollaborators: []identity.DID{did1, did},
//			ReadCollaborators:      []identity.DID{did1, did2}},
//		nil)
//	assert.NoError(t, err)
//	assert.NoError(t, cd.RemoveCollaborators([]identity.DID{did1}))
//	found, err := cd.IsDIDCollaborator(did1)
//	assert.NoError(t, err)
//	assert.False(t, found)
//
//	found, err = cd.IsDIDCollaborator(did3)
//	assert.NoError(t, err)
//	assert.False(t, found)
//}
//
//func TestCoreDocument_AddRole(t *testing.T) {
//	key := hexutil.Encode(utils.RandomSlice(32))
//	tests := []struct {
//		key     string
//		collabs []identity.DID
//		roleKey []byte
//		err     error
//	}{
//		// empty string
//		{
//			err: ErrEmptyRoleKey,
//		},
//
//		// 30 byte hex
//		{
//			key:     hexutil.Encode(utils.RandomSlice(30)),
//			collabs: []identity.DID{testingidentity.GenerateRandomDID()},
//		},
//
//		// random string
//		{
//			key:     "role key 1",
//			collabs: []identity.DID{testingidentity.GenerateRandomDID()},
//		},
//
//		// missing collabs
//		{
//			key: hexutil.Encode(utils.RandomSlice(32)),
//			err: ErrEmptyCollaborators,
//		},
//
//		// 32 byte key
//		{
//			key:     key,
//			collabs: []identity.DID{testingidentity.GenerateRandomDID()},
//		},
//
//		// role exists
//		{
//			key:     key,
//			collabs: []identity.DID{testingidentity.GenerateRandomDID()},
//			err:     ErrRoleExist,
//		},
//	}
//
//	cd, err := newCoreDocument()
//	assert.NoError(t, err)
//	for _, c := range tests {
//		r, err := cd.AddRole(c.key, c.collabs)
//		if err != nil {
//			assert.Equal(t, err, c.err)
//			continue
//		}
//
//		assert.NoError(t, err)
//		assert.Len(t, r.RoleKey, idSize)
//	}
//}
//
//func TestCoreDocument_UpdateRole(t *testing.T) {
//	cd, err := newCoreDocument()
//	assert.NoError(t, err)
//
//	// invalid role key
//	key := utils.RandomSlice(30)
//	collabs := []identity.DID{testingidentity.GenerateRandomDID()}
//	_, err = cd.UpdateRole(key, collabs)
//	assert.Error(t, err)
//	assert.Equal(t, err, ErrInvalidRoleKey)
//
//	// missing role
//	key = utils.RandomSlice(32)
//	_, err = cd.UpdateRole(key, collabs)
//	assert.Error(t, err)
//	assert.Equal(t, err, ErrRoleNotExist)
//
//	// empty collabs
//	r, err := cd.AddRole(hexutil.Encode(key), []identity.DID{testingidentity.GenerateRandomDID()})
//	assert.NoError(t, err)
//	assert.Equal(t, r.RoleKey, key)
//	assert.Len(t, r.Collaborators, 1)
//	assert.NotEqual(t, r.Collaborators[0], collabs[0][:])
//	_, err = cd.UpdateRole(key, nil)
//	assert.Error(t, err)
//	assert.Equal(t, err, ErrEmptyCollaborators)
//
//	// success
//	r, err = cd.UpdateRole(key, collabs)
//	assert.NoError(t, err)
//	assert.Equal(t, r.RoleKey, key)
//	assert.Len(t, r.Collaborators, 1)
//	assert.Equal(t, r.Collaborators[0], collabs[0][:])
//	sr, err := cd.GetRole(key)
//	assert.NoError(t, err)
//	assert.Equal(t, r, sr)
//}
//
//func TestFingerprintGeneration(t *testing.T) {
//	id1 := testingidentity.GenerateRandomDID()
//	cd, err := NewCoreDocument([]byte("inv"), CollaboratorsAccess{}, nil)
//	assert.NoError(t, err)
//	role, err := cd.AddRole("test_r", []identity.DID{id1})
//	assert.NoError(t, err)
//	cd.Document.Roles = append(cd.Document.Roles, role)
//	assert.NoError(t, err)
//	cd.addNewTransitionRule(role.RoleKey, coredocumentpb.FieldMatchType_FIELD_MATCH_TYPE_PREFIX, nil, coredocumentpb.TransitionAction_TRANSITION_ACTION_EDIT)
//
//	// copy over transition rules and roles to generate fingerprint
//	f := coredocumentpb.TransitionRulesFingerprint{}
//	f.Roles = cd.Document.Roles
//	f.TransitionRules = cd.Document.TransitionRules
//	p, err := cd.CalculateTransitionRulesFingerprint()
//	assert.NoError(t, err)
//
//	// create second document with same roles and transition rules to check if generated fingerprint is the same
//	cd1, err := NewCoreDocument([]byte("inv"), CollaboratorsAccess{}, nil)
//	assert.NoError(t, err)
//	cd1.Document.Roles = cd.Document.Roles
//	cd1.Document.TransitionRules = cd.Document.TransitionRules
//
//	f1 := coredocumentpb.TransitionRulesFingerprint{}
//	f1.Roles = cd1.Document.Roles
//	f1.TransitionRules = cd1.Document.TransitionRules
//	p1, err := cd1.CalculateTransitionRulesFingerprint()
//	assert.NoError(t, err)
//	assert.True(t, bytes.Equal(p, p1))
//}

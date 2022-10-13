//go:build unit

package documents

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/crypto"
	"github.com/centrifuge/go-centrifuge/errors"
	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/common"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/go-centrifuge/utils/byteutils"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/precise-proofs/proofs"
	proofspb "github.com/centrifuge/precise-proofs/proofs/proto"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/blake2b"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
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
	assert.True(t, errors.IsOfType(ErrGranteeInvalidAccountID, err))
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
	assert.ErrorIs(t, err, ErrAccountNotFoundInContext)
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

func TestCoreDocument_ID(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)

	assert.Equal(t, cd.Document.DocumentIdentifier, cd.ID())
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
			SignatureId:         utils.RandomSlice(64),
			SignerId:            utils.RandomSlice(32),
			PublicKey:           utils.RandomSlice(32),
			Signature:           utils.RandomSlice(32),
			TransitionValidated: true,
		},
		{
			SignatureId:         utils.RandomSlice(64),
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
	readCollaborator1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	readWriteCollaborator1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	cd, err := NewCoreDocument(
		[]byte("prefix"),
		CollaboratorsAccess{
			ReadCollaborators:      []*types.AccountID{readCollaborator1},
			ReadWriteCollaborators: []*types.AccountID{readWriteCollaborator1},
		},
		nil,
	)
	assert.NoError(t, err)
	assert.NotNil(t, cd)

	documentPrefix := utils.RandomSlice(32)

	readCollaborator2, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	readWriteCollaborator2, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collaboratorsAccess := CollaboratorsAccess{
		ReadCollaborators:      []*types.AccountID{readCollaborator2},
		ReadWriteCollaborators: []*types.AccountID{readWriteCollaborator2},
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

	ca, err := res.GetCollaborators()
	assert.NoError(t, err)
	assert.Len(t, ca.ReadCollaborators, 1)
	assert.Contains(t, ca.ReadCollaborators, readCollaborator2)
	assert.Len(t, ca.ReadWriteCollaborators, 1)
	assert.Contains(t, ca.ReadWriteCollaborators, readWriteCollaborator2)
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

	// No data leaves will cause and error when retrieving the document tree.

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
	// New Document

	cd, err := newCoreDocument()
	assert.NoError(t, err)
	assert.NotNil(t, cd)

	res, err := cd.CalculateSignaturesRoot()
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Len(t, res, 32)

	// Document with signatures

	cd.Document.SignatureData.Signatures = []*coredocumentpb.Signature{
		{
			SignatureId:         utils.RandomSlice(64),
			SignerId:            utils.RandomSlice(32),
			PublicKey:           utils.RandomSlice(32),
			Signature:           utils.RandomSlice(32),
			TransitionValidated: true,
		},
		{
			SignatureId:         utils.RandomSlice(64),
			SignerId:            utils.RandomSlice(32),
			PublicKey:           utils.RandomSlice(32),
			Signature:           utils.RandomSlice(32),
			TransitionValidated: false,
		},
	}

	res, err = cd.CalculateSignaturesRoot()
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Len(t, res, 32)
}

func TestCoreDocument_CalculateSignaturesRoot_InvalidSignature(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	assert.NotNil(t, cd)

	cd.Document.SignatureData.Signatures = []*coredocumentpb.Signature{
		{
			// Invalid length for signature ID byte slice.
			SignatureId:         utils.RandomSlice(54),
			SignerId:            utils.RandomSlice(32),
			PublicKey:           utils.RandomSlice(32),
			Signature:           utils.RandomSlice(32),
			TransitionValidated: true,
		},
		{
			SignatureId:         utils.RandomSlice(64),
			SignerId:            utils.RandomSlice(32),
			PublicKey:           utils.RandomSlice(32),
			Signature:           utils.RandomSlice(32),
			TransitionValidated: false,
		},
	}

	res, err := cd.CalculateSignaturesRoot()
	assert.True(t, errors.IsOfType(ErrCDTree, err))
	assert.Nil(t, res)
}

func TestCoreDocument_GetSignaturesDataTree(t *testing.T) {
	// New Document

	cd, err := newCoreDocument()
	assert.NoError(t, err)
	assert.NotNil(t, cd)

	res, err := cd.GetSignaturesDataTree()
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Len(t, res.GetLeaves(), 1)

	// Document with signatures

	cd.Document.SignatureData.Signatures = []*coredocumentpb.Signature{
		{
			SignatureId:         utils.RandomSlice(64),
			SignerId:            utils.RandomSlice(32),
			PublicKey:           utils.RandomSlice(32),
			Signature:           utils.RandomSlice(32),
			TransitionValidated: true,
		},
		{
			SignatureId:         utils.RandomSlice(64),
			SignerId:            utils.RandomSlice(32),
			PublicKey:           utils.RandomSlice(32),
			Signature:           utils.RandomSlice(32),
			TransitionValidated: false,
		},
	}

	res, err = cd.GetSignaturesDataTree()
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Len(t, res.GetLeaves(), 3)
}

func TestCoreDocument_GetSignaturesDataTree_InvalidSignature(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	assert.NotNil(t, cd)

	cd.Document.SignatureData.Signatures = []*coredocumentpb.Signature{
		{
			// Invalid length for signature ID byte slice.
			SignatureId:         utils.RandomSlice(54),
			SignerId:            utils.RandomSlice(32),
			PublicKey:           utils.RandomSlice(32),
			Signature:           utils.RandomSlice(32),
			TransitionValidated: true,
		},
	}

	res, err := cd.GetSignaturesDataTree()
	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestCoreDocument_DocumentRootTree(t *testing.T) {
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

	res, err := cd.DocumentRootTree(docType, dataLeaves)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Len(t, res.GetLeaves(), 2)
}

func TestCoreDocument_DocumentRootTree_NoLeaves(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	assert.NotNil(t, cd)

	docType := "doc"

	res, err := cd.DocumentRootTree(docType, nil)
	assert.True(t, errors.IsOfType(ErrCDTree, err))
	assert.Nil(t, res)
}

func TestCoreDocument_DocumentRootTree_InvalidLeaves(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	assert.NotNil(t, cd)

	docType := "doc"

	// Having the same compact property will trigger an error when adding the leaves.

	compactProperty := utils.RandomSlice(32)

	dataLeaves := []proofs.LeafNode{
		{
			Property: proofs.Property{
				Parent:     nil,
				Text:       "name.test1",
				Compact:    compactProperty,
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
				Compact:    compactProperty,
				NameFormat: "",
			},
			Value:  utils.RandomSlice(32),
			Salt:   utils.RandomSlice(32),
			Hash:   utils.RandomSlice(32),
			Hashed: false,
		},
	}

	res, err := cd.DocumentRootTree(docType, dataLeaves)
	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestCoreDocument_DocumentRootTree_WithSignatures(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	assert.NotNil(t, cd)

	cd.Document.SignatureData.Signatures = []*coredocumentpb.Signature{
		{
			SignatureId:         utils.RandomSlice(64),
			SignerId:            utils.RandomSlice(32),
			PublicKey:           utils.RandomSlice(32),
			Signature:           utils.RandomSlice(32),
			TransitionValidated: true,
		},
		{
			SignatureId:         utils.RandomSlice(64),
			SignerId:            utils.RandomSlice(32),
			PublicKey:           utils.RandomSlice(32),
			Signature:           utils.RandomSlice(32),
			TransitionValidated: false,
		},
	}

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

	res, err := cd.DocumentRootTree(docType, dataLeaves)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Len(t, res.GetLeaves(), 2)
}

func TestCoreDocument_DocumentRootTree_InvalidSignature(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	assert.NotNil(t, cd)

	cd.Document.SignatureData.Signatures = []*coredocumentpb.Signature{
		{
			// Invalid signature ID byte slice length.
			SignatureId:         utils.RandomSlice(54),
			SignerId:            utils.RandomSlice(32),
			PublicKey:           utils.RandomSlice(32),
			Signature:           utils.RandomSlice(32),
			TransitionValidated: true,
		},
		{
			SignatureId:         utils.RandomSlice(64),
			SignerId:            utils.RandomSlice(32),
			PublicKey:           utils.RandomSlice(32),
			Signature:           utils.RandomSlice(32),
			TransitionValidated: false,
		},
	}

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

	res, err := cd.DocumentRootTree(docType, dataLeaves)
	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestCoreDocument_SigningDataTree(t *testing.T) {
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

	res, err := cd.SigningDataTree(docType, dataLeaves)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Len(t, res.GetLeaves(), 18)
}

func TestCoreDocument_SigningDataTree_InvalidLeaves(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	assert.NotNil(t, cd)

	docType := "doc"

	// No data leaves

	res, err := cd.SigningDataTree(docType, nil)
	assert.NotNil(t, err)
	assert.Nil(t, res)

	// Having the same compact property will trigger an error when adding the leaves.

	compactProperty := utils.RandomSlice(32)

	dataLeaves := []proofs.LeafNode{
		{
			Property: proofs.Property{
				Parent:     nil,
				Text:       "name.test1",
				Compact:    compactProperty,
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
				Compact:    compactProperty,
				NameFormat: "",
			},
			Value:  utils.RandomSlice(32),
			Salt:   utils.RandomSlice(32),
			Hash:   utils.RandomSlice(32),
			Hashed: false,
		},
	}

	res, err = cd.SigningDataTree(docType, dataLeaves)
	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestCoreDocument_GetSignerCollaborators(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	assert.NotNil(t, cd)

	res, err := cd.GetSignerCollaborators()
	assert.NoError(t, err)
	assert.Nil(t, res)

	signCollab1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	signCollab2, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	editCollab1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	editCollab2, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	signRoleKey := utils.RandomSlice(32)

	signRules := []*coredocumentpb.ReadRule{
		{
			Roles: [][]byte{
				signRoleKey,
			},
			Action: coredocumentpb.Action_ACTION_READ_SIGN,
		},
	}

	signRoles := []*coredocumentpb.Role{
		{
			RoleKey: signRoleKey,
			Collaborators: [][]byte{
				signCollab1.ToBytes(),
				signCollab2.ToBytes(),
				editCollab1.ToBytes(),
			},
			Nfts: [][]byte{
				utils.RandomSlice(32),
			},
		},
	}

	editRoleKey := utils.RandomSlice(32)

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
	}

	editRoles := []*coredocumentpb.Role{
		{
			RoleKey: editRoleKey,
			Collaborators: [][]byte{
				editCollab1.ToBytes(),
				editCollab2.ToBytes(),
				signCollab2.ToBytes(),
			},
			Nfts: [][]byte{
				utils.RandomSlice(32),
			},
		},
	}

	cd.Document.TransitionRules = transitionRule
	cd.Document.ReadRules = signRules
	cd.Document.Roles = append(signRoles, editRoles...)

	res, err = cd.GetSignerCollaborators()
	assert.NoError(t, err)
	assert.Len(t, res, 4)
	assert.Contains(t, res, signCollab1)
	assert.Contains(t, res, signCollab2)
	assert.Contains(t, res, editCollab1)
	assert.Contains(t, res, editCollab2)

	res, err = cd.GetSignerCollaborators(signCollab1, editCollab2)
	assert.NoError(t, err)
	assert.Len(t, res, 2)
	assert.Contains(t, res, signCollab2)
	assert.Contains(t, res, editCollab1)
}

func TestCoreDocument_GetSignerCollaborators_ReadCollaboratorError(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	assert.NotNil(t, cd)

	signCollab1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	signRoleKey := utils.RandomSlice(32)

	signRules := []*coredocumentpb.ReadRule{
		{
			Roles: [][]byte{
				signRoleKey,
			},
			Action: coredocumentpb.Action_ACTION_READ_SIGN,
		},
	}

	signRoles := []*coredocumentpb.Role{
		{
			RoleKey: signRoleKey,
			Collaborators: [][]byte{
				signCollab1.ToBytes(),
				[]byte("some-non-account-id-bytes"),
			},
			Nfts: [][]byte{
				utils.RandomSlice(32),
			},
		},
	}

	cd.Document.ReadRules = signRules
	cd.Document.Roles = signRoles

	res, err := cd.GetSignerCollaborators()
	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestCoreDocument_GetSignerCollaborators_WriteCollaboratorError(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	assert.NotNil(t, cd)

	editCollab1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	editRoleKey := utils.RandomSlice(32)

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
	}

	editRoles := []*coredocumentpb.Role{
		{
			RoleKey: editRoleKey,
			Collaborators: [][]byte{
				editCollab1.ToBytes(),
				[]byte("invalid-account-id-bytes"),
			},
			Nfts: [][]byte{
				utils.RandomSlice(32),
			},
		},
	}

	cd.Document.TransitionRules = transitionRule
	cd.Document.Roles = editRoles

	res, err := cd.GetSignerCollaborators()
	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestCoreDocument_GetCollaborators(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	assert.NotNil(t, cd)

	res, err := cd.GetCollaborators()
	assert.NoError(t, err)
	assert.Nil(t, res.ReadCollaborators)
	assert.Nil(t, res.ReadWriteCollaborators)

	signCollab1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	signCollab2, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	editCollab1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	editCollab2, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	readCollab1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	signRoleKey := utils.RandomSlice(32)
	readRoleKey := utils.RandomSlice(32)

	signRules := []*coredocumentpb.ReadRule{
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

	signRoles := []*coredocumentpb.Role{
		{
			RoleKey: signRoleKey,
			Collaborators: [][]byte{
				signCollab1.ToBytes(),
				signCollab2.ToBytes(),
				editCollab1.ToBytes(),
			},
			Nfts: [][]byte{
				utils.RandomSlice(32),
			},
		},
		{
			RoleKey: readRoleKey,
			Collaborators: [][]byte{
				signCollab1.ToBytes(),
				signCollab2.ToBytes(),
				editCollab1.ToBytes(),
				readCollab1.ToBytes(),
			},
			Nfts: [][]byte{
				utils.RandomSlice(32),
			},
		},
	}

	editRoleKey := utils.RandomSlice(32)

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
	}

	editRoles := []*coredocumentpb.Role{
		{
			RoleKey: editRoleKey,
			Collaborators: [][]byte{
				editCollab1.ToBytes(),
				editCollab2.ToBytes(),
				signCollab2.ToBytes(),
			},
			Nfts: [][]byte{
				utils.RandomSlice(32),
			},
		},
	}

	cd.Document.TransitionRules = transitionRule
	cd.Document.ReadRules = signRules
	cd.Document.Roles = append(signRoles, editRoles...)

	res, err = cd.GetCollaborators()
	assert.NoError(t, err)
	assert.Len(t, res.ReadCollaborators, 2)
	assert.Contains(t, res.ReadCollaborators, signCollab1)
	assert.Contains(t, res.ReadCollaborators, readCollab1)

	assert.Len(t, res.ReadWriteCollaborators, 3)
	assert.Contains(t, res.ReadWriteCollaborators, editCollab1)
	assert.Contains(t, res.ReadWriteCollaborators, editCollab2)
	assert.Contains(t, res.ReadWriteCollaborators, signCollab2)

	res, err = cd.GetCollaborators(readCollab1, editCollab2)
	assert.NoError(t, err)
	assert.Len(t, res.ReadCollaborators, 1)
	assert.Contains(t, res.ReadCollaborators, signCollab1)

	assert.Len(t, res.ReadWriteCollaborators, 2)
	assert.Contains(t, res.ReadWriteCollaborators, editCollab1)
	assert.Contains(t, res.ReadWriteCollaborators, signCollab2)
}

func TestCoreDocument_GetCollaborators_ReadCollaboratorError(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	assert.NotNil(t, cd)

	res, err := cd.GetCollaborators()
	assert.NoError(t, err)
	assert.Nil(t, res.ReadCollaborators)
	assert.Nil(t, res.ReadWriteCollaborators)

	signRoleKey := utils.RandomSlice(32)

	signRules := []*coredocumentpb.ReadRule{
		{
			Roles: [][]byte{
				signRoleKey,
			},
			Action: coredocumentpb.Action_ACTION_READ_SIGN,
		},
	}

	signRoles := []*coredocumentpb.Role{
		{
			RoleKey: signRoleKey,
			Collaborators: [][]byte{
				[]byte("invalid-account-id-bytes"),
			},
			Nfts: [][]byte{
				utils.RandomSlice(32),
			},
		},
	}

	cd.Document.ReadRules = signRules
	cd.Document.Roles = signRoles

	res, err = cd.GetCollaborators()
	assert.NotNil(t, err)
	assert.Nil(t, res.ReadCollaborators)
	assert.Nil(t, res.ReadWriteCollaborators)
}

func TestCoreDocument_GetCollaborators_WriteCollaboratorError(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	assert.NotNil(t, cd)

	res, err := cd.GetCollaborators()
	assert.NoError(t, err)
	assert.Nil(t, res.ReadCollaborators)
	assert.Nil(t, res.ReadWriteCollaborators)

	signCollab1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	signCollab2, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	editCollab1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	signRoleKey := utils.RandomSlice(32)

	signRules := []*coredocumentpb.ReadRule{
		{
			Roles: [][]byte{
				signRoleKey,
			},
			Action: coredocumentpb.Action_ACTION_READ_SIGN,
		},
	}

	signRoles := []*coredocumentpb.Role{
		{
			RoleKey: signRoleKey,
			Collaborators: [][]byte{
				signCollab1.ToBytes(),
				signCollab2.ToBytes(),
				editCollab1.ToBytes(),
			},
			Nfts: [][]byte{
				utils.RandomSlice(32),
			},
		},
	}

	editRoleKey := utils.RandomSlice(32)

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
	}

	editRoles := []*coredocumentpb.Role{
		{
			RoleKey: editRoleKey,
			Collaborators: [][]byte{
				[]byte("invalid-account-id-bytes"),
			},
			Nfts: [][]byte{
				utils.RandomSlice(32),
			},
		},
	}

	cd.Document.TransitionRules = transitionRule
	cd.Document.ReadRules = signRules
	cd.Document.Roles = append(signRoles, editRoles...)

	res, err = cd.GetCollaborators()
	assert.NotNil(t, err)
	assert.Nil(t, res.ReadCollaborators)
	assert.Nil(t, res.ReadWriteCollaborators)
}

func TestCoreDocument_CalculateDocumentRoot(t *testing.T) {
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

	tree, err := cd.DocumentRootTree(docType, dataLeaves)
	assert.NoError(t, err)

	res, err := cd.CalculateDocumentRoot(docType, dataLeaves)
	assert.NoError(t, err)
	assert.Equal(t, tree.RootHash(), res)
}

func TestCoreDocument_CalculateDocumentRoot_DocumentRootTreeError(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	assert.NotNil(t, cd)

	docType := "doc"

	res, err := cd.CalculateDocumentRoot(docType, nil)
	assert.True(t, errors.IsOfType(ErrCDTree, err))
	assert.Nil(t, res)
}

func TestCoredDocument_CalculateSigningRoot(t *testing.T) {
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

	tree, err := cd.SigningDataTree(docType, dataLeaves)
	assert.NoError(t, err)
	assert.NotNil(t, tree)

	res, err := cd.CalculateSigningRoot(docType, dataLeaves)
	assert.NoError(t, err)
	assert.Equal(t, tree.RootHash(), res)
}

func TestCoredDocument_CalculateSigningRoot_SigningDataTreeError(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	assert.NotNil(t, cd)

	docType := "doc"

	res, err := cd.CalculateSigningRoot(docType, nil)
	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestCoreDocument_PackCoreDocument(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	assert.NotNil(t, cd)

	data := &anypb.Any{
		TypeUrl: "type",
		Value:   utils.RandomSlice(32),
	}

	res := cd.PackCoreDocument(data)
	assert.NotEqual(t, cd, res)
	assert.Equal(t, data, res.EmbeddedData)
}

func TestCoreDocument_Signatures(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	assert.NotNil(t, cd)

	signatures := []*coredocumentpb.Signature{
		{
			SignatureId:         utils.RandomSlice(32),
			SignerId:            utils.RandomSlice(32),
			PublicKey:           utils.RandomSlice(32),
			Signature:           utils.RandomSlice(32),
			TransitionValidated: false,
		},
		{
			SignatureId:         utils.RandomSlice(32),
			SignerId:            utils.RandomSlice(32),
			PublicKey:           utils.RandomSlice(32),
			Signature:           utils.RandomSlice(32),
			TransitionValidated: true,
		},
	}

	cd.Document.SignatureData = &coredocumentpb.SignatureData{
		Signatures: signatures,
	}

	res := cd.Signatures()
	assert.Equal(t, res, signatures)
}

func TestCoreDocument_AddUpdateLog(t *testing.T) {
	cd := &coredocumentpb.CoreDocument{
		SignatureData: new(coredocumentpb.SignatureData),
	}

	doc := &CoreDocument{
		Document: cd,
	}

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	doc.AddUpdateLog(accountID)

	assert.Equal(t, doc.Document.Author, accountID.ToBytes())
	assert.True(t, doc.Modified)
}

func TestCoreDocument_Author(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)

	res, err := cd.Author()
	assert.NotNil(t, err)
	assert.Nil(t, res)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	cd.Document.Author = accountID.ToBytes()

	res, err = cd.Author()
	assert.NoError(t, err)
	assert.Equal(t, accountID, res)
}

func TestCoreDocument_Timestamp(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)

	res, err := cd.Timestamp()
	assert.NotNil(t, err)
	assert.True(t, res.IsZero())

	cd.Document.Timestamp = timestamppb.Now()

	res, err = cd.Timestamp()
	assert.NoError(t, err)
	assert.False(t, res.IsZero())
}

func TestCoreDocument_Attributes(t *testing.T) {
	accountID1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	cas := CollaboratorsAccess{ReadWriteCollaborators: []*types.AccountID{accountID1}}

	cd, err := NewCoreDocument(nil, cas, nil)
	assert.NoError(t, err)

	cd.Attributes = nil
	label := "com.basf.deliverynote.chemicalnumber"
	value := "100"
	key, err := AttrKeyFromLabel(label)
	assert.NoError(t, err)

	// failed get
	assert.False(t, cd.AttributeExists(key))

	_, err = cd.GetAttribute(key)
	assert.Error(t, err)

	// failed delete
	_, err = cd.DeleteAttribute(key, true, nil)
	assert.Error(t, err)

	// success
	attr, err := NewStringAttribute(label, AttrString, value)
	assert.NoError(t, err)

	cd, err = cd.AddAttributes(CollaboratorsAccess{}, true, nil, attr)
	assert.NoError(t, err)
	assert.Len(t, cd.Attributes, 1)
	assert.Len(t, cd.GetAttributes(), 1)

	// check
	assert.True(t, cd.AttributeExists(key))

	attr, err = cd.GetAttribute(key)

	assert.NoError(t, err)
	assert.Equal(t, key, attr.Key)
	assert.Equal(t, label, attr.KeyLabel)

	str, err := attr.Value.String()
	assert.NoError(t, err)
	assert.Equal(t, value, str)
	assert.Equal(t, AttrString, attr.Value.Type)

	// update
	nvalue := "2000"
	attr, err = NewStringAttribute(label, AttrDecimal, nvalue)
	assert.NoError(t, err)

	cd, err = cd.AddAttributes(CollaboratorsAccess{}, true, nil, attr)
	assert.NoError(t, err)
	assert.True(t, cd.AttributeExists(key))

	attr, err = cd.GetAttribute(key)
	assert.NoError(t, err)
	assert.Len(t, cd.Attributes, 1)
	assert.Len(t, cd.GetAttributes(), 1)
	assert.Equal(t, key, attr.Key)
	assert.Equal(t, label, attr.KeyLabel)

	str, err = attr.Value.String()
	assert.NoError(t, err)
	assert.NotEqual(t, value, str)
	assert.Equal(t, nvalue, str)
	assert.Equal(t, AttrDecimal, attr.Value.Type)

	// delete
	cd, err = cd.DeleteAttribute(key, true, nil)
	assert.NoError(t, err)
	assert.Len(t, cd.Attributes, 0)
	assert.Len(t, cd.GetAttributes(), 0)
	assert.False(t, cd.AttributeExists(key))
}

func TestCoreDocument_IsCollaborator(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	assert.NotNil(t, cd)

	signCollab1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	signCollab2, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	editCollab1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	editCollab2, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	readCollab1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	signRoleKey := utils.RandomSlice(32)

	signRules := []*coredocumentpb.ReadRule{
		{
			Roles: [][]byte{
				signRoleKey,
			},
			Action: coredocumentpb.Action_ACTION_READ_SIGN,
		},
	}

	signRoles := []*coredocumentpb.Role{
		{
			RoleKey: signRoleKey,
			Collaborators: [][]byte{
				signCollab1.ToBytes(),
				signCollab2.ToBytes(),
				editCollab1.ToBytes(),
			},
			Nfts: [][]byte{
				utils.RandomSlice(32),
			},
		},
	}

	editRoleKey := utils.RandomSlice(32)

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
	}

	editRoles := []*coredocumentpb.Role{
		{
			RoleKey: editRoleKey,
			Collaborators: [][]byte{
				editCollab1.ToBytes(),
				editCollab2.ToBytes(),
				signCollab2.ToBytes(),
			},
			Nfts: [][]byte{
				utils.RandomSlice(32),
			},
		},
	}

	cd.Document.TransitionRules = transitionRule
	cd.Document.ReadRules = signRules
	cd.Document.Roles = append(signRoles, editRoles...)

	res, err := cd.IsCollaborator(signCollab1)
	assert.NoError(t, err)
	assert.True(t, res)

	res, err = cd.IsCollaborator(signCollab2)
	assert.NoError(t, err)
	assert.True(t, res)

	res, err = cd.IsCollaborator(editCollab1)
	assert.NoError(t, err)
	assert.True(t, res)

	res, err = cd.IsCollaborator(editCollab2)
	assert.NoError(t, err)
	assert.True(t, res)

	res, err = cd.IsCollaborator(readCollab1)
	assert.NoError(t, err)
	assert.False(t, res)
}

func TestCoreDocument_IsCollaboratorError(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	assert.NotNil(t, cd)

	signCollab1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	signRoleKey := utils.RandomSlice(32)

	signRules := []*coredocumentpb.ReadRule{
		{
			Roles: [][]byte{
				signRoleKey,
			},
			Action: coredocumentpb.Action_ACTION_READ_SIGN,
		},
	}

	signRoles := []*coredocumentpb.Role{
		{
			RoleKey: signRoleKey,
			Collaborators: [][]byte{
				[]byte("invalid-account-id-bytes"),
			},
			Nfts: [][]byte{
				utils.RandomSlice(32),
			},
		},
	}

	cd.Document.ReadRules = signRules
	cd.Document.Roles = signRoles

	res, err := cd.IsCollaborator(signCollab1)
	assert.NotNil(t, err)
	assert.False(t, res)
}

func TestCoreDocument_GetAccessTokens(t *testing.T) {
	cd := coredocumentpb.CoreDocument{}

	accessTokens := []*coredocumentpb.AccessToken{
		{
			Identifier:         utils.RandomSlice(32),
			Granter:            utils.RandomSlice(32),
			Grantee:            utils.RandomSlice(32),
			RoleIdentifier:     utils.RandomSlice(32),
			DocumentIdentifier: utils.RandomSlice(32),
			Signature:          utils.RandomSlice(32),
			Key:                utils.RandomSlice(32),
			DocumentVersion:    utils.RandomSlice(32),
		},
		{
			Identifier:         utils.RandomSlice(32),
			Granter:            utils.RandomSlice(32),
			Grantee:            utils.RandomSlice(32),
			RoleIdentifier:     utils.RandomSlice(32),
			DocumentIdentifier: utils.RandomSlice(32),
			Signature:          utils.RandomSlice(32),
			Key:                utils.RandomSlice(32),
			DocumentVersion:    utils.RandomSlice(32),
		},
	}

	cd.AccessTokens = accessTokens

	assert.Equal(t, accessTokens, cd.GetAccessTokens())
}

func TestCoreDocument_RemoveCollaborators(t *testing.T) {
	accountID1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountID2, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountID3, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	cd, err := NewCoreDocument(
		nil,
		CollaboratorsAccess{
			ReadWriteCollaborators: []*types.AccountID{accountID1, accountID2},
			ReadCollaborators:      []*types.AccountID{accountID2, accountID3},
		},
		nil)
	assert.NoError(t, err)

	collaborators, err := cd.GetCollaborators()
	assert.NoError(t, err)
	assert.Len(t, collaborators.ReadWriteCollaborators, 2)
	assert.Len(t, collaborators.ReadCollaborators, 1)

	assert.NoError(t, cd.RemoveCollaborators([]*types.AccountID{accountID1}))

	found, err := cd.IsCollaborator(accountID1)
	assert.NoError(t, err)
	assert.False(t, found)
}

func TestCoreDocument_RemoveCollaborators_StatusError(t *testing.T) {
	accountID1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountID2, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountID3, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	cd, err := NewCoreDocument(
		nil,
		CollaboratorsAccess{
			ReadWriteCollaborators: []*types.AccountID{accountID1, accountID2},
			ReadCollaborators:      []*types.AccountID{accountID2, accountID3},
		},
		nil)
	assert.NoError(t, err)

	collaborators, err := cd.GetCollaborators()
	assert.NoError(t, err)
	assert.Len(t, collaborators.ReadWriteCollaborators, 2)
	assert.Len(t, collaborators.ReadCollaborators, 1)

	cd.Status = Committing

	err = cd.RemoveCollaborators([]*types.AccountID{accountID1})

	assert.ErrorIs(t, err, ErrDocumentNotInAllowedState)

	cd.Status = Committed

	err = cd.RemoveCollaborators([]*types.AccountID{accountID1})

	assert.ErrorIs(t, err, ErrDocumentNotInAllowedState)
}

func TestCoreDocument_GetRole(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	assert.NotNil(t, cd)

	signCollab1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	signCollab2, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	roleKey := utils.RandomSlice(32)

	role := &coredocumentpb.Role{
		RoleKey: roleKey,
		Collaborators: [][]byte{
			signCollab1.ToBytes(),
			signCollab2.ToBytes(),
		},
		Nfts: [][]byte{
			utils.RandomSlice(32),
		},
	}

	cd.Document.Roles = []*coredocumentpb.Role{role}

	res, err := cd.GetRole(roleKey)
	assert.NoError(t, err)
	assert.Equal(t, role, res)
}

func TestCoreDocument_GetRole_Errors(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	assert.NotNil(t, cd)

	res, err := cd.GetRole(utils.RandomSlice(idSize - 1))
	assert.ErrorIs(t, err, ErrInvalidRoleKey)
	assert.Nil(t, res)

	res, err = cd.GetRole(utils.RandomSlice(32))
	assert.ErrorIs(t, err, ErrRoleNotExist)
	assert.Nil(t, res)
}

func TestCoreDocument_AddRole(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	assert.NotNil(t, cd)

	signCollab1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	signCollab2, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	roleKey := utils.RandomSlice(32)

	role := &coredocumentpb.Role{
		RoleKey: roleKey,
		Collaborators: [][]byte{
			signCollab1.ToBytes(),
			signCollab2.ToBytes(),
		},
		Nfts: [][]byte{
			utils.RandomSlice(32),
		},
	}

	cd.Document.Roles = []*coredocumentpb.Role{role}

	newAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	newRoleKey := utils.RandomSlice(32)

	res, err := cd.AddRole(string(newRoleKey), []*types.AccountID{newAccountID})
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.True(t, cd.Modified)

	newRole, err := cd.GetRole(res.GetRoleKey())
	assert.NoError(t, err)
	assert.Equal(t, res, newRole)
}

func TestCoreDocument_AddRole_Errors(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	assert.NotNil(t, cd)

	signCollab1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	signCollab2, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	roleKey := utils.RandomSlice(32)

	role := &coredocumentpb.Role{
		RoleKey: roleKey,
		Collaborators: [][]byte{
			signCollab1.ToBytes(),
			signCollab2.ToBytes(),
		},
		Nfts: [][]byte{
			utils.RandomSlice(32),
		},
	}

	cd.Document.Roles = []*coredocumentpb.Role{role}

	newAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	// Empty key
	res, err := cd.AddRole("", []*types.AccountID{newAccountID})
	assert.ErrorIs(t, err, ErrEmptyRoleKey)
	assert.Nil(t, res)

	// Role exists
	res, err = cd.AddRole(hexutil.Encode(role.GetRoleKey()), []*types.AccountID{newAccountID})
	assert.ErrorIs(t, err, ErrRoleExist)
	assert.Nil(t, res)

	// Empty collaborators
	res, err = cd.AddRole(string(utils.RandomSlice(32)), nil)
	assert.ErrorIs(t, err, ErrEmptyCollaborators)
	assert.Nil(t, res)
}

func TestCoreDocument_UpdateRole(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	assert.NotNil(t, cd)

	signCollab1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	signCollab2, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	roleKey := utils.RandomSlice(32)

	role := &coredocumentpb.Role{
		RoleKey: roleKey,
		Collaborators: [][]byte{
			signCollab1.ToBytes(),
			signCollab2.ToBytes(),
		},
		Nfts: [][]byte{
			utils.RandomSlice(32),
		},
	}

	cd.Document.Roles = []*coredocumentpb.Role{role}

	newAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	res, err := cd.UpdateRole(roleKey, []*types.AccountID{newAccountID})
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Contains(t, res.Collaborators, newAccountID.ToBytes())
	assert.NotContains(t, res.Collaborators, signCollab1.ToBytes())
	assert.NotContains(t, res.Collaborators, signCollab2.ToBytes())
}

func TestCoreDocument_UpdateRole_Errors(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	assert.NotNil(t, cd)

	signCollab1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	signCollab2, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	roleKey := utils.RandomSlice(32)

	role := &coredocumentpb.Role{
		RoleKey: roleKey,
		Collaborators: [][]byte{
			signCollab1.ToBytes(),
			signCollab2.ToBytes(),
		},
		Nfts: [][]byte{
			utils.RandomSlice(32),
		},
	}

	cd.Document.Roles = []*coredocumentpb.Role{role}

	newAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	// Role does not exist
	res, err := cd.UpdateRole(utils.RandomSlice(32), []*types.AccountID{newAccountID})
	assert.ErrorIs(t, err, ErrRoleNotExist)
	assert.Nil(t, res)

	// Empty collaborators
	res, err = cd.UpdateRole(roleKey, nil)
	assert.ErrorIs(t, err, ErrEmptyCollaborators)
	assert.Nil(t, res)
}

func TestCoreDocument_Status(t *testing.T) {
	cd, err := NewCoreDocument(nil, CollaboratorsAccess{}, nil)
	assert.NoError(t, err)
	assert.Equal(t, cd.GetStatus(), Pending)

	// set status to Committed
	err = cd.SetStatus(Committed)
	assert.NoError(t, err)
	assert.Equal(t, cd.GetStatus(), Committed)

	// try to update status to Committing
	err = cd.SetStatus(Committing)
	assert.Error(t, err)
	assert.True(t, errors.IsOfType(ErrCDStatus, err))
	assert.Equal(t, cd.GetStatus(), Committed)
}

func TestCoreDocument_getReadCollaborators(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	assert.NotNil(t, cd)

	collaborators, err := cd.GetCollaborators()
	assert.NoError(t, err)
	assert.Nil(t, collaborators.ReadCollaborators)
	assert.Nil(t, collaborators.ReadWriteCollaborators)

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

	readRoles := []*coredocumentpb.Role{
		{
			RoleKey: signRoleKey,
			Collaborators: [][]byte{
				signCollab1.ToBytes(),
				signCollab2.ToBytes(),
			},
			Nfts: [][]byte{
				utils.RandomSlice(32),
			},
		},
		{
			RoleKey: readRoleKey,
			Collaborators: [][]byte{
				readCollab1.ToBytes(),
				readCollab2.ToBytes(),
			},
			Nfts: [][]byte{
				utils.RandomSlice(32),
			},
		},
	}

	cd.Document.ReadRules = readRules
	cd.Document.Roles = readRoles

	res, err := cd.getReadCollaborators(coredocumentpb.Action_ACTION_READ_SIGN, coredocumentpb.Action_ACTION_READ)
	assert.NoError(t, err)
	assert.Len(t, res, 4)
	assert.Contains(t, res, signCollab1)
	assert.Contains(t, res, signCollab2)
	assert.Contains(t, res, readCollab1)
	assert.Contains(t, res, readCollab2)

	res, err = cd.getReadCollaborators(coredocumentpb.Action_ACTION_READ_SIGN)
	assert.NoError(t, err)
	assert.Len(t, res, 2)
	assert.Contains(t, res, signCollab1)
	assert.Contains(t, res, signCollab2)

	res, err = cd.getReadCollaborators(coredocumentpb.Action_ACTION_READ)
	assert.NoError(t, err)
	assert.Len(t, res, 2)
	assert.Contains(t, res, readCollab1)
	assert.Contains(t, res, readCollab2)
}

func TestCoreDocument_getReadCollaborators_NoCollaborators(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	assert.NotNil(t, cd)

	collaborators, err := cd.GetCollaborators()
	assert.NoError(t, err)
	assert.Nil(t, collaborators.ReadCollaborators)
	assert.Nil(t, collaborators.ReadWriteCollaborators)

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

	readRoles := []*coredocumentpb.Role{
		{
			RoleKey:       signRoleKey,
			Collaborators: [][]byte{},
			Nfts: [][]byte{
				utils.RandomSlice(32),
			},
		},
		{
			RoleKey:       readRoleKey,
			Collaborators: [][]byte{},
			Nfts: [][]byte{
				utils.RandomSlice(32),
			},
		},
	}

	cd.Document.ReadRules = readRules
	cd.Document.Roles = readRoles

	res, err := cd.getReadCollaborators(coredocumentpb.Action_ACTION_READ_SIGN, coredocumentpb.Action_ACTION_READ)
	assert.NoError(t, err)
	assert.Len(t, res, 0)
}

func TestCoreDocument_getReadCollaborators_Error(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	assert.NotNil(t, cd)

	collaborators, err := cd.GetCollaborators()
	assert.NoError(t, err)
	assert.Nil(t, collaborators.ReadCollaborators)
	assert.Nil(t, collaborators.ReadWriteCollaborators)

	signCollab1, err := testingcommons.GetRandomAccountID()
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

	readRoles := []*coredocumentpb.Role{
		{
			RoleKey: signRoleKey,
			Collaborators: [][]byte{
				signCollab1.ToBytes(),
				[]byte("invalid-account-id-bytes"),
			},
			Nfts: [][]byte{
				utils.RandomSlice(32),
			},
		},
		{
			RoleKey: readRoleKey,
			Collaborators: [][]byte{
				[]byte("invalid-account-id-bytes"),
				readCollab2.ToBytes(),
			},
			Nfts: [][]byte{
				utils.RandomSlice(32),
			},
		},
	}

	cd.Document.ReadRules = readRules
	cd.Document.Roles = readRoles

	res, err := cd.getReadCollaborators(coredocumentpb.Action_ACTION_READ_SIGN, coredocumentpb.Action_ACTION_READ)
	assert.NotNil(t, err)
	assert.Len(t, res, 1)
	assert.Contains(t, res, signCollab1)

	res, err = cd.getReadCollaborators(coredocumentpb.Action_ACTION_READ_SIGN)
	assert.NotNil(t, err)
	assert.Len(t, res, 1)
	assert.Contains(t, res, signCollab1)

	res, err = cd.getReadCollaborators(coredocumentpb.Action_ACTION_READ)
	assert.NotNil(t, err)
	assert.Len(t, res, 0)
}

func TestCoreDocument_getWriteCollaborators(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	assert.NotNil(t, cd)

	collaborators, err := cd.GetCollaborators()
	assert.NoError(t, err)
	assert.Nil(t, collaborators.ReadCollaborators)
	assert.Nil(t, collaborators.ReadWriteCollaborators)

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

	editRoles := []*coredocumentpb.Role{
		{
			RoleKey: editRoleKey,
			Collaborators: [][]byte{
				editCollab1.ToBytes(),
				editCollab2.ToBytes(),
			},
			Nfts: [][]byte{
				utils.RandomSlice(32),
			},
		},
		{
			RoleKey: computeRoleKey,
			Collaborators: [][]byte{
				computeCollab1.ToBytes(),
				computeCollab2.ToBytes(),
			},
			Nfts: [][]byte{
				utils.RandomSlice(32),
			},
		},
	}

	cd.Document.TransitionRules = transitionRule
	cd.Document.Roles = editRoles

	res, err := cd.getWriteCollaborators(
		coredocumentpb.TransitionAction_TRANSITION_ACTION_EDIT,
		coredocumentpb.TransitionAction_TRANSITION_ACTION_COMPUTE,
	)
	assert.NoError(t, err)
	assert.Len(t, res, 4)
	assert.Contains(t, res, editCollab1)
	assert.Contains(t, res, editCollab2)
	assert.Contains(t, res, computeCollab1)
	assert.Contains(t, res, computeCollab2)

	res, err = cd.getWriteCollaborators(coredocumentpb.TransitionAction_TRANSITION_ACTION_EDIT)
	assert.NoError(t, err)
	assert.Len(t, res, 2)
	assert.Contains(t, res, editCollab1)
	assert.Contains(t, res, editCollab2)

	res, err = cd.getWriteCollaborators(coredocumentpb.TransitionAction_TRANSITION_ACTION_COMPUTE)
	assert.NoError(t, err)
	assert.Len(t, res, 2)
	assert.Contains(t, res, computeCollab1)
	assert.Contains(t, res, computeCollab2)
}

func TestCoreDocument_getWriteCollaborators_NoCollaborators(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	assert.NotNil(t, cd)

	collaborators, err := cd.GetCollaborators()
	assert.NoError(t, err)
	assert.Nil(t, collaborators.ReadCollaborators)
	assert.Nil(t, collaborators.ReadWriteCollaborators)

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

	editRoles := []*coredocumentpb.Role{
		{
			RoleKey:       editRoleKey,
			Collaborators: [][]byte{},
			Nfts: [][]byte{
				utils.RandomSlice(32),
			},
		},
		{
			RoleKey:       computeRoleKey,
			Collaborators: [][]byte{},
			Nfts: [][]byte{
				utils.RandomSlice(32),
			},
		},
	}

	cd.Document.TransitionRules = transitionRule
	cd.Document.Roles = editRoles

	res, err := cd.getWriteCollaborators(
		coredocumentpb.TransitionAction_TRANSITION_ACTION_EDIT,
		coredocumentpb.TransitionAction_TRANSITION_ACTION_COMPUTE,
	)
	assert.NoError(t, err)
	assert.Len(t, res, 0)
}

func TestCoreDocument_getWriteCollaborators_Error(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	assert.NotNil(t, cd)

	collaborators, err := cd.GetCollaborators()
	assert.NoError(t, err)
	assert.Nil(t, collaborators.ReadCollaborators)
	assert.Nil(t, collaborators.ReadWriteCollaborators)

	editCollab1, err := testingcommons.GetRandomAccountID()
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

	editRoles := []*coredocumentpb.Role{
		{
			RoleKey: editRoleKey,
			Collaborators: [][]byte{
				editCollab1.ToBytes(),
				[]byte("invalid-account-id-bytes"),
			},
			Nfts: [][]byte{
				utils.RandomSlice(32),
			},
		},
		{
			RoleKey: computeRoleKey,
			Collaborators: [][]byte{
				[]byte("invalid-account-id-bytes"),
				computeCollab2.ToBytes(),
			},
			Nfts: [][]byte{
				utils.RandomSlice(32),
			},
		},
	}

	cd.Document.TransitionRules = transitionRule
	cd.Document.Roles = editRoles

	res, err := cd.getWriteCollaborators(
		coredocumentpb.TransitionAction_TRANSITION_ACTION_EDIT,
		coredocumentpb.TransitionAction_TRANSITION_ACTION_COMPUTE,
	)
	assert.NotNil(t, err)
	assert.Len(t, res, 1)
	assert.Contains(t, res, editCollab1)

	res, err = cd.getWriteCollaborators(coredocumentpb.TransitionAction_TRANSITION_ACTION_EDIT)
	assert.NotNil(t, err)
	assert.Len(t, res, 1)
	assert.Contains(t, res, editCollab1)

	res, err = cd.getWriteCollaborators(coredocumentpb.TransitionAction_TRANSITION_ACTION_COMPUTE)
	assert.NotNil(t, err)
	assert.Len(t, res, 0)
}

func TestNewRoleWithCollaborators(t *testing.T) {
	accountID1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)
	accountID2, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	role := newRoleWithCollaborators(accountID1, accountID2)
	assert.Len(t, role.Collaborators, 2)
	assert.Equal(t, role.Collaborators[0], accountID1.ToBytes())
	assert.Equal(t, role.Collaborators[1], accountID2.ToBytes())
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
	cds, err := newCoreDocument()
	assert.NoError(t, err)

	testTree, err := cds.DefaultTreeWithPrefix("prefix", []byte{1, 0, 0, 0})
	assert.NoError(t, err)

	props := []proofs.Property{NewLeafProperty("prefix.sample_field", []byte{1, 0, 0, 0, 0, 0, 0, 200}), NewLeafProperty("prefix.sample_field2", []byte{1, 0, 0, 0, 0, 0, 0, 202})}
	err = testTree.AddLeaf(proofs.LeafNode{Hash: utils.RandomSlice(32), Hashed: true, Property: props[0]})
	assert.NoError(t, err)

	err = testTree.AddLeaf(proofs.LeafNode{Hash: utils.RandomSlice(32), Hashed: true, Property: props[1]})
	assert.NoError(t, err)

	err = testTree.Generate()
	assert.NoError(t, err)

	// nil length leaves
	prfx, err := getDataTreePrefix(nil)
	assert.Error(t, err)

	// zero length leaves
	prfx, err = getDataTreePrefix([]proofs.LeafNode{})
	assert.Error(t, err)

	// success
	prfx, err = getDataTreePrefix(testTree.GetLeaves())
	assert.NoError(t, err)
	assert.Equal(t, "prefix", prfx)

	// non-prefixed tree error
	testTree, err = cds.DefaultTreeWithPrefix("", []byte{1, 0, 0, 0})
	assert.NoError(t, err)

	props = []proofs.Property{NewLeafProperty("sample_field", []byte{1, 0, 0, 0, 0, 0, 0, 200}), NewLeafProperty("sample_field2", []byte{1, 0, 0, 0, 0, 0, 0, 202})}
	err = testTree.AddLeaf(proofs.LeafNode{Hash: utils.RandomSlice(32), Hashed: true, Property: props[0]})
	assert.NoError(t, err)

	err = testTree.AddLeaf(proofs.LeafNode{Hash: utils.RandomSlice(32), Hashed: true, Property: props[1]})
	assert.NoError(t, err)

	err = testTree.Generate()
	assert.NoError(t, err)

	prfx, err = getDataTreePrefix(testTree.GetLeaves())
	assert.Error(t, err)

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

func TestCoredocRawTree(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	assert.NotNil(t, cd)

	docType := "doc"

	res, err := cd.coredocRawTree(docType)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Len(t, res.GetLeaves(), 16)
}

func TestCoredocLeaves(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	assert.NotNil(t, cd)

	docType := "doc"

	res, err := cd.coredocLeaves(docType)
	assert.NoError(t, err)
	assert.Len(t, res, 16)
}

func TestCoredocTree(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)
	assert.NotNil(t, cd)

	docType := "doc"

	res, err := cd.coredocTree(docType)
	assert.NoError(t, err)
	assert.Len(t, res.GetLeaves(), 16)
}

func TestFilterCollaborator(t *testing.T) {
	accountID1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountID2, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountID3, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountID4, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	ids := []*types.AccountID{
		accountID1,
		accountID2,
		accountID3,
		accountID4,
	}

	res := filterCollaborators(ids)
	assert.Len(t, res, 4)
	assert.Contains(t, res, accountID1)
	assert.Contains(t, res, accountID2)
	assert.Contains(t, res, accountID3)
	assert.Contains(t, res, accountID4)

	res = filterCollaborators(ids, accountID3)
	assert.Len(t, res, 3)
	assert.Contains(t, res, accountID1)
	assert.Contains(t, res, accountID2)
	assert.Contains(t, res, accountID4)

	res = filterCollaborators(ids, accountID1, accountID2, accountID3, accountID4)
	assert.Len(t, res, 0)
}

func TestPopulateVersions(t *testing.T) {
	currentDoc := &coredocumentpb.CoreDocument{}

	previousDoc := &coredocumentpb.CoreDocument{
		CurrentVersion: utils.RandomSlice(idSize),
		NextVersion:    utils.RandomSlice(idSize),
		NextPreimage:   utils.RandomSlice(idSize),
	}

	err := populateVersions(currentDoc, previousDoc)
	assert.NoError(t, err)

	assert.Equal(t, previousDoc.CurrentVersion, currentDoc.PreviousVersion)
	assert.Equal(t, previousDoc.NextVersion, currentDoc.CurrentVersion)
	assert.Equal(t, previousDoc.NextPreimage, currentDoc.CurrentPreimage)
	assert.Len(t, currentDoc.NextPreimage, 32)
	assert.Len(t, currentDoc.NextVersion, 32)

	currentDoc = &coredocumentpb.CoreDocument{}

	err = populateVersions(currentDoc, nil)

	assert.Equal(t, currentDoc.CurrentVersion, currentDoc.DocumentIdentifier)
	assert.Len(t, currentDoc.CurrentPreimage, 32)
	assert.Len(t, currentDoc.CurrentVersion, 32)
	assert.Len(t, currentDoc.NextPreimage, 32)
	assert.Len(t, currentDoc.NextVersion, 32)
}

func TestFingerprintGeneration(t *testing.T) {
	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	cd, err := NewCoreDocument([]byte("inv"), CollaboratorsAccess{}, nil)
	assert.NoError(t, err)

	role, err := cd.AddRole("test_r", []*types.AccountID{accountID})
	assert.NoError(t, err)

	cd.Document.Roles = append(cd.Document.Roles, role)
	assert.NoError(t, err)

	cd.addNewTransitionRule(role.RoleKey, coredocumentpb.FieldMatchType_FIELD_MATCH_TYPE_PREFIX, nil, coredocumentpb.TransitionAction_TRANSITION_ACTION_EDIT)

	// copy over transition rules and roles to generate fingerprint
	f := coredocumentpb.TransitionRulesFingerprint{}
	f.Roles = cd.Document.Roles
	f.TransitionRules = cd.Document.TransitionRules

	p, err := cd.CalculateTransitionRulesFingerprint()
	assert.NoError(t, err)

	// create second document with same roles and transition rules to check if generated fingerprint is the same
	cd1, err := NewCoreDocument([]byte("inv"), CollaboratorsAccess{}, nil)
	assert.NoError(t, err)

	cd1.Document.Roles = cd.Document.Roles
	cd1.Document.TransitionRules = cd.Document.TransitionRules

	f1 := coredocumentpb.TransitionRulesFingerprint{}
	f1.Roles = cd1.Document.Roles
	f1.TransitionRules = cd1.Document.TransitionRules

	p1, err := cd1.CalculateTransitionRulesFingerprint()
	assert.NoError(t, err)
	assert.True(t, bytes.Equal(p, p1))
}

func TestGetSigningProofHash(t *testing.T) {
	docAny := &anypb.Any{
		TypeUrl: documenttypes.InvoiceDataTypeUrl,
		Value:   []byte{},
	}

	cd, err := newCoreDocument()
	assert.NoError(t, err)

	cd.Document.EmbeddedData = docAny
	testTree, err := cd.DefaultTreeWithPrefix("invoice", []byte{1, 0, 0, 0})
	assert.NoError(t, err)

	signingRoot, err := cd.CalculateSigningRoot(documenttypes.InvoiceDataTypeUrl, testTree.GetLeaves())
	assert.Nil(t, err)

	docRoot, err := cd.CalculateDocumentRoot(documenttypes.InvoiceDataTypeUrl, testTree.GetLeaves())
	assert.Nil(t, err)

	signatureTree, err := cd.GetSignaturesDataTree()
	assert.Nil(t, err)

	h, err := blake2b.New256(nil)
	assert.NoError(t, err)

	valid, err := proofs.ValidateProofHashes(signingRoot, []*proofspb.MerkleHash{{Right: signatureTree.RootHash()}}, docRoot, h)
	assert.True(t, valid)
	assert.Nil(t, err)
}

func TestGetDocumentRootTree(t *testing.T) {
	cd, err := newCoreDocument()
	assert.NoError(t, err)

	sig := &coredocumentpb.Signature{
		SignerId:    utils.RandomSlice(32),
		PublicKey:   utils.RandomSlice(32),
		SignatureId: utils.RandomSlice(64),
		Signature:   utils.RandomSlice(32),
	}
	cd.Document.SignatureData.Signatures = []*coredocumentpb.Signature{sig}

	testTree, err := cd.DefaultTreeWithPrefix("invoice", []byte{1, 0, 0, 0})
	assert.NoError(t, err)

	// successful document root generation
	signingRoot, err := cd.CalculateSigningRoot(documenttypes.InvoiceDataTypeUrl, testTree.GetLeaves())
	assert.NoError(t, err)

	tree, err := cd.DocumentRootTree(documenttypes.InvoiceDataTypeUrl, testTree.GetLeaves())
	assert.NoError(t, err)

	_, leaf := tree.GetLeafByProperty(fmt.Sprintf("%s.%s", DRTreePrefix, SigningRootField))
	assert.NotNil(t, leaf)
	assert.Equal(t, signingRoot, leaf.Hash)

	// Get signaturesLeaf
	_, signaturesLeaf := tree.GetLeafByProperty(fmt.Sprintf("%s.%s", DRTreePrefix, SignaturesRootField))
	assert.NotNil(t, signaturesLeaf)
	assert.Equal(t, fmt.Sprintf("%s.%s", DRTreePrefix, SignaturesRootField), signaturesLeaf.Property.ReadableName())
	assert.Equal(t, append(CompactProperties(DRTreePrefix), CompactProperties(SignaturesRootField)...), signaturesLeaf.Property.CompactName())
}

func TestGet32ByteKey(t *testing.T) {
	testKey1 := string(utils.RandomSlice(3))
	expectedResult1, err := crypto.Sha256Hash([]byte(testKey1))
	assert.NoError(t, err)

	testKey2 := hexutil.Encode(utils.RandomSlice(idSize - 1))
	expectedResult2, err := crypto.Sha256Hash([]byte(testKey2))
	assert.NoError(t, err)

	key := utils.RandomSlice(idSize)
	testKey3 := hexutil.Encode(key)

	tests := []struct {
		Name           string
		Key            string
		ExpectedResult []byte
		ExpectedError  error
	}{
		{
			Name:           fmt.Sprintf("Random key string - %s", testKey1),
			Key:            testKey1,
			ExpectedResult: expectedResult1,
		},
		{
			Name:           fmt.Sprintf("Hex encoded byte slice with size %d - %s", idSize-1, testKey2),
			Key:            testKey2,
			ExpectedResult: expectedResult2,
		},
		{
			Name:           fmt.Sprintf("Hex encoded byte slice with size %d - %s", idSize, testKey3),
			Key:            testKey3,
			ExpectedResult: key,
		},
		{
			Name:          "Empty key string",
			Key:           "",
			ExpectedError: ErrEmptyRoleKey,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			res, err := get32ByteKey(test.Key)

			if test.ExpectedError != nil {
				assert.ErrorIs(t, err, test.ExpectedError)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, test.ExpectedResult, res)
		})
	}
}

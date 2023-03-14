//go:build unit

package entityrelationship

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	entitypb "github.com/centrifuge/centrifuge-protobufs/gen/go/entity"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/pod/config"
	"github.com/centrifuge/pod/contextutil"
	"github.com/centrifuge/pod/documents"
	"github.com/centrifuge/pod/errors"
	testingcommons "github.com/centrifuge/pod/testingutils/common"
	"github.com/centrifuge/pod/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/blake2b"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

func TestEntityRelationship_PackCoreDocument(t *testing.T) {
	entityRelationship := getTestEntityRelationship(
		t,
		documents.CollaboratorsAccess{},
		nil,
	)

	cd, err := entityRelationship.PackCoreDocument()
	assert.NoError(t, err)
	assert.NotNil(t, cd)
	assert.NotNil(t, cd.EmbeddedData)

	data, err := proto.Marshal(entityRelationship.createP2PProtobuf())
	assert.NoError(t, err)

	embeddedData := &anypb.Any{
		TypeUrl: entityRelationship.DocumentType(),
		Value:   data,
	}

	assert.Equal(t, embeddedData, embeddedData)
}

func TestEntityRelationship_UnpackCoreDocument(t *testing.T) {
	entityRelationship := getTestEntityRelationship(
		t,
		documents.CollaboratorsAccess{},
		nil,
	)

	// No embedded data
	err := entityRelationship.UnpackCoreDocument(&coredocumentpb.CoreDocument{})
	assert.True(t, errors.IsOfType(documents.ErrDocumentConvertInvalidSchema, err))

	// Invalid embedded data schema
	err = entityRelationship.UnpackCoreDocument(&coredocumentpb.CoreDocument{EmbeddedData: new(anypb.Any)})
	assert.True(t, errors.IsOfType(documents.ErrDocumentConvertInvalidSchema, err))

	// Invalid embedded data
	err = entityRelationship.UnpackCoreDocument(
		&coredocumentpb.CoreDocument{
			EmbeddedData: &anypb.Any{
				TypeUrl: documenttypes.EntityRelationshipDataTypeUrl,
				Value:   utils.RandomSlice(32),
			},
		},
	)
	assert.True(t, errors.IsOfType(documents.ErrDocumentDataUnmarshalling, err))

	// Invalid attributes
	entityRelationshippb := getTestEntityRelationshipProto()

	data, err := proto.Marshal(entityRelationshippb)
	assert.NoError(t, err)

	err = entityRelationship.UnpackCoreDocument(
		&coredocumentpb.CoreDocument{
			EmbeddedData: &anypb.Any{
				TypeUrl: documenttypes.EntityRelationshipDataTypeUrl,
				Value:   data,
			},
			Attributes: []*coredocumentpb.Attribute{
				{
					// Invalid key.
					Key: utils.RandomSlice(31),
				},
			},
		},
	)
	assert.NotNil(t, err)

	// Invalid account ID bytes.
	entityRelationshippb.OwnerIdentity = utils.RandomSlice(31)

	data, err = proto.Marshal(entityRelationshippb)
	assert.NoError(t, err)

	err = entityRelationship.UnpackCoreDocument(
		&coredocumentpb.CoreDocument{
			EmbeddedData: &anypb.Any{
				TypeUrl: documenttypes.EntityRelationshipDataTypeUrl,
				Value:   data,
			},
		},
	)
	assert.True(t, errors.IsOfType(documents.ErrAccountIDBytesParsing, err))

	// Valid
	entityRelationshippb.OwnerIdentity = utils.RandomSlice(32)

	data, err = proto.Marshal(entityRelationshippb)
	assert.NoError(t, err)

	err = entityRelationship.UnpackCoreDocument(
		&coredocumentpb.CoreDocument{
			EmbeddedData: &anypb.Any{
				TypeUrl: documenttypes.EntityRelationshipDataTypeUrl,
				Value:   data,
			},
		},
	)
	assert.NoError(t, err)
}

func TestEntityRelationship_ToAndFromJSON(t *testing.T) {
	entityRelationship := getTestEntityRelationship(
		t,
		documents.CollaboratorsAccess{},
		nil,
	)

	b, err := entityRelationship.JSON()
	assert.NoError(t, err)

	newEntityRelationship := &EntityRelationship{}

	err = newEntityRelationship.FromJSON(b)
	assert.NoError(t, err)
}

func TestEntityRelationship_CreateProofs(t *testing.T) {
	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	entityRelationship := getTestEntityRelationship(
		t,
		documents.CollaboratorsAccess{ReadWriteCollaborators: []*types.AccountID{accountID}},
		nil,
	)

	rk := entityRelationship.Document.Roles[0].RoleKey
	pf := fmt.Sprintf(documents.CDTreePrefix+".roles[%s].collaborators[0]", hexutil.Encode(rk))

	proof, err := entityRelationship.CreateProofs(
		[]string{
			"entity_relationship.owner_identity",
			pf,
			documents.CDTreePrefix + ".document_type",
		},
	)
	assert.NoError(t, err)
	assert.NotNil(t, proof)

	dataRoot := calculateBasicDataRoot(t, entityRelationship)

	nodeAndLeafHash, err := blake2b.New256(nil)
	assert.NoError(t, err)

	// Validate entity_number
	valid, err := documents.ValidateProof(proof.FieldProofs[0], dataRoot, nodeAndLeafHash, nodeAndLeafHash)
	assert.Nil(t, err)
	assert.True(t, valid)

	// Validate roles
	valid, err = documents.ValidateProof(proof.FieldProofs[1], dataRoot, nodeAndLeafHash, nodeAndLeafHash)
	assert.Nil(t, err)
	assert.True(t, valid)

	// Validate []byte value
	acc, err := types.NewAccountID(proof.FieldProofs[1].Value)
	assert.NoError(t, err)
	assert.True(t, entityRelationship.AccountCanRead(acc))

	// Validate document_type
	valid, err = documents.ValidateProof(proof.FieldProofs[2], dataRoot, nodeAndLeafHash, nodeAndLeafHash)
	assert.Nil(t, err)
	assert.True(t, valid)

	// Non-existing field
	res, err := entityRelationship.CreateProofs([]string{"invalid-field"})
	assert.NotNil(t, err)
	assert.Nil(t, res)

	// Nil core documents
	entityRelationship.CoreDocument = nil
	res, err = entityRelationship.CreateProofs([]string{"entity_relationship.owner_identity"})
	assert.True(t, errors.IsOfType(documents.ErrDocumentProof, err))
	assert.Nil(t, res)
}

func TestEntityRelationship_DocumentType(t *testing.T) {
	entityRelationship := &EntityRelationship{}
	assert.Equal(t, documenttypes.EntityRelationshipDataTypeUrl, entityRelationship.DocumentType())
}

func TestEntityRelationship_AddNFT(t *testing.T) {
	entityRelationship := &EntityRelationship{}

	collectionID := types.U64(1111)
	itemID := types.NewU128(*big.NewInt(2222))

	err := entityRelationship.AddNFT(true, collectionID, itemID)
	assert.ErrorIs(t, err, documents.ErrNotImplemented)

	err = entityRelationship.AddNFT(false, collectionID, itemID)
	assert.ErrorIs(t, err, documents.ErrNotImplemented)
}

func TestEntityRelationship_CalculateSigningRoot(t *testing.T) {
	entityRelationship := getTestEntityRelationship(
		t,
		documents.CollaboratorsAccess{ReadWriteCollaborators: nil},
		nil,
	)

	b, err := entityRelationship.CalculateSigningRoot()
	assert.NoError(t, err)
	assert.NotNil(t, b)

	entityRelationship.CoreDocument = nil

	b, err = entityRelationship.CalculateSigningRoot()
	assert.True(t, errors.IsOfType(documents.ErrDataTree, err))
	assert.Nil(t, b)
}

func TestEntityRelationship_CalculateDocumentRoot(t *testing.T) {
	entityRelationship := getTestEntityRelationship(
		t,
		documents.CollaboratorsAccess{ReadWriteCollaborators: nil},
		nil,
	)

	b, err := entityRelationship.CalculateDocumentRoot()
	assert.NoError(t, err)
	assert.NotNil(t, b)

	entityRelationship.CoreDocument = nil

	b, err = entityRelationship.CalculateDocumentRoot()
	assert.True(t, errors.IsOfType(documents.ErrDataTree, err))
	assert.Nil(t, b)
}

func TestEntityRelationship_CollaboratorCanUpdate(t *testing.T) {
	accountID1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	entityRelationship1 := getTestEntityRelationship(
		t,
		documents.CollaboratorsAccess{ReadWriteCollaborators: nil},
		nil,
	)
	entityRelationship1.Data.OwnerIdentity = accountID1

	entityRelationship2 := getTestEntityRelationship(
		t,
		documents.CollaboratorsAccess{ReadWriteCollaborators: nil},
		nil,
	)

	entityRelationship2.Data.OwnerIdentity = accountID1

	err = entityRelationship1.CollaboratorCanUpdate(entityRelationship2, accountID1)
	assert.NoError(t, err)

	// Invalid doc type
	documentMock := documents.NewDocumentMock(t)

	err = entityRelationship1.CollaboratorCanUpdate(documentMock, accountID1)
	assert.True(t, errors.IsOfType(documents.ErrDocumentInvalidType, err))

	// Owner identity mismatch
	randomAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	err = entityRelationship1.CollaboratorCanUpdate(entityRelationship2, randomAccountID)
	assert.ErrorIs(t, err, documents.ErrIdentityNotOwner)

	// Update entityRelationship1 with the random account ID
	entityRelationship1.Data.OwnerIdentity = randomAccountID

	err = entityRelationship1.CollaboratorCanUpdate(entityRelationship2, randomAccountID)
	assert.ErrorIs(t, err, documents.ErrIdentityNotOwner)

	// Reset entityRelationship1 and update entityRelationship2 with the random account ID
	entityRelationship1.Data.OwnerIdentity = accountID1
	entityRelationship2.Data.OwnerIdentity = randomAccountID

	err = entityRelationship1.CollaboratorCanUpdate(entityRelationship2, randomAccountID)
	assert.ErrorIs(t, err, documents.ErrIdentityNotOwner)
}

func TestEntity_AddAndDeleteAttributes(t *testing.T) {
	entityRelationship := getTestEntityRelationship(
		t,
		documents.CollaboratorsAccess{},
		nil,
	)

	attr1Label := "test_attr_1"
	attr1Key := utils.RandomByte32()
	attr1Value := documents.AttrVal{
		Type: documents.AttrString,
		Str:  "test_attr_string",
	}
	attr1 := documents.Attribute{
		KeyLabel: attr1Label,
		Key:      attr1Key,
		Value:    attr1Value,
	}

	attr2Label := "test_attr_1"
	attr2Key := utils.RandomByte32()
	attr2Value := documents.AttrVal{
		Type:  documents.AttrBytes,
		Bytes: []byte("test_attr_bytes"),
	}
	attr2 := documents.Attribute{
		KeyLabel: attr2Label,
		Key:      attr2Key,
		Value:    attr2Value,
	}

	attrs := []documents.Attribute{attr1, attr2}

	err := entityRelationship.AddAttributes(documents.CollaboratorsAccess{}, false)
	assert.True(t, errors.IsOfType(documents.ErrCDAttribute, err))

	err = entityRelationship.AddAttributes(documents.CollaboratorsAccess{}, false, attrs...)
	assert.NoError(t, err)

	res, err := entityRelationship.GetAttribute(attr1Key)
	assert.NoError(t, err)
	assert.Equal(t, attr1, res)

	res, err = entityRelationship.GetAttribute(attr2Key)
	assert.NoError(t, err)
	assert.Equal(t, attr2, res)

	err = entityRelationship.DeleteAttribute(attr1Key, false)
	assert.NoError(t, err)

	err = entityRelationship.DeleteAttribute(attr2Key, false)
	assert.NoError(t, err)

	err = entityRelationship.DeleteAttribute(attr2Key, false)
	assert.Error(t, err)

	_, err = entityRelationship.GetAttribute(attr1Key)
	assert.Error(t, err)

	_, err = entityRelationship.GetAttribute(attr2Key)
	assert.Error(t, err)
}

func TestEntityRelationship_GetData(t *testing.T) {
	entityRelationship := getTestEntityRelationship(
		t,
		documents.CollaboratorsAccess{},
		nil,
	)

	assert.Equal(t, entityRelationship.Data, entityRelationship.GetData())
}

func TestEntityRelationship_Scheme(t *testing.T) {
	entityRelationship := &EntityRelationship{}

	assert.Equal(t, Scheme, entityRelationship.Scheme())
}

func TestEntityRelationship_DeriveFromCreatePayload(t *testing.T) {
	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)

	ctx := contextutil.WithAccount(context.Background(), accountMock)

	entityRelationship1 := getTestEntityRelationship(
		t,
		documents.CollaboratorsAccess{},
		nil,
	)

	b, err := json.Marshal(entityRelationship1.Data)
	assert.NoError(t, err)

	payload := documents.CreatePayload{
		Data: b,
	}

	entityRelationship2 := &EntityRelationship{}

	accountMock.On("GetIdentity").
		Return(accountID).
		Times(2)

	signature := &coredocumentpb.Signature{
		Signature: utils.RandomSlice(64),
		PublicKey: utils.RandomSlice(32),
	}

	accountMock.On("SignMsg", mock.Anything).
		Return(signature, nil).
		Once()

	err = entityRelationship2.DeriveFromCreatePayload(ctx, payload)
	assert.NoError(t, err)
	assert.True(t, entityRelationship2.AccountCanRead(accountID))
	assert.True(t, entityRelationship2.AccountCanRead(entityRelationship1.Data.TargetIdentity))
	assert.False(t, entityRelationship2.AccountCanRead(entityRelationship1.Data.OwnerIdentity))
	assert.Len(t, entityRelationship2.Document.AccessTokens, 1)
	assert.Equal(t, entityRelationship2.Document.AccessTokens[0].Signature, signature.GetSignature())
	assert.Equal(t, entityRelationship2.Document.AccessTokens[0].Key, signature.GetPublicKey())

	// Invalid data
	payload = documents.CreatePayload{
		Data: utils.RandomSlice(32),
	}

	err = entityRelationship2.DeriveFromCreatePayload(ctx, payload)
	assert.True(t, errors.IsOfType(ErrERInvalidData, err))

	// Invalid target identity
	entityRelationship1.Data.TargetIdentity = nil

	b, err = json.Marshal(entityRelationship1.Data)
	assert.NoError(t, err)

	payload = documents.CreatePayload{
		Data: b,
	}

	err = entityRelationship2.DeriveFromCreatePayload(ctx, payload)
	assert.True(t, errors.IsOfType(documents.ErrCDCreate, err))
}

func TestEntityRelationship_DeriveFromUpdatePayload(t *testing.T) {
	ctx := context.Background()

	entityRelationship1 := getTestEntityRelationship(
		t,
		documents.CollaboratorsAccess{},
		nil,
	)

	entityRelationship1.Document.AccessTokens = []*coredocumentpb.AccessToken{
		{
			Grantee: entityRelationship1.Data.TargetIdentity.ToBytes(),
		},
	}

	b, err := json.Marshal(entityRelationship1.Data)
	assert.NoError(t, err)

	payload := documents.UpdatePayload{
		CreatePayload: documents.CreatePayload{
			Data: b,
		},
	}

	res, err := entityRelationship1.DeriveFromUpdatePayload(ctx, payload)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Empty(t, res.GetAccessTokens())

	// Invalid data
	payload = documents.UpdatePayload{
		CreatePayload: documents.CreatePayload{
			Data: utils.RandomSlice(32),
		},
	}

	res, err = entityRelationship1.DeriveFromUpdatePayload(ctx, payload)
	assert.True(t, errors.IsOfType(ErrERInvalidData, err))
	assert.Nil(t, res)

	// No accesss tokens
	entityRelationship1.Document.AccessTokens = nil

	payload = documents.UpdatePayload{
		CreatePayload: documents.CreatePayload{
			Data: b,
		},
	}

	res, err = entityRelationship1.DeriveFromUpdatePayload(ctx, payload)
	assert.ErrorIs(t, err, documents.ErrAccessTokenNotFound)
	assert.Nil(t, res)
}

func TestEntityRelationship_DeriveFromClonePayload(t *testing.T) {
	entityRelationship1 := getTestEntityRelationship(t, documents.CollaboratorsAccess{}, nil)
	entityRelationship2 := getTestEntityRelationship(t, documents.CollaboratorsAccess{}, nil)

	ctx := context.Background()

	err := entityRelationship1.DeriveFromClonePayload(ctx, entityRelationship2)
	assert.NoError(t, err)

	documentMock := documents.NewDocumentMock(t)
	documentMock.On("PackCoreDocument").
		Return(nil, errors.New("error")).
		Once()

	err = entityRelationship1.DeriveFromClonePayload(ctx, documentMock)
	assert.True(t, errors.IsOfType(documents.ErrDocumentPackingCoreDocument, err))

	coreDoc := &coredocumentpb.CoreDocument{
		Attributes: []*coredocumentpb.Attribute{
			{
				// Invalid key length
				Key: utils.RandomSlice(31),
			},
		},
	}

	documentMock.On("PackCoreDocument").
		Return(coreDoc, nil).
		Once()

	err = entityRelationship1.DeriveFromClonePayload(ctx, documentMock)
	assert.True(t, errors.IsOfType(documents.ErrCDClone, err))
}

func TestEntityRelationship_Patch(t *testing.T) {
	entityRelationship1 := getTestEntityRelationship(t, documents.CollaboratorsAccess{}, nil)

	b, err := json.Marshal(entityRelationship1.Data)
	assert.NoError(t, err)

	payload := documents.UpdatePayload{
		CreatePayload: documents.CreatePayload{
			Data: b,
		},
	}

	entityRelationship2 := getTestEntityRelationship(t, documents.CollaboratorsAccess{}, nil)

	err = entityRelationship2.Patch(payload)
	assert.NoError(t, err)

	// Invalid data
	payload = documents.UpdatePayload{
		CreatePayload: documents.CreatePayload{
			Data: utils.RandomSlice(32),
		},
	}

	err = entityRelationship2.Patch(payload)
	assert.True(t, errors.IsOfType(ErrERInvalidData, err))

	// Invalid attributes
	attrKey := utils.RandomByte32()
	attr := documents.Attribute{
		KeyLabel: "labels",
		Key:      attrKey,
		Value: documents.AttrVal{
			Type: "invalid_type",
		},
	}

	payload = documents.UpdatePayload{
		CreatePayload: documents.CreatePayload{
			Data: b,
			Attributes: map[documents.AttrKey]documents.Attribute{
				attrKey: attr,
			},
		},
	}

	err = entityRelationship2.Patch(payload)
	assert.True(t, errors.IsOfType(documents.ErrDocumentPatch, err))
}

func TestEntityRelationship_revokeRelationship(t *testing.T) {
	entityRelationship1 := getTestEntityRelationship(t, documents.CollaboratorsAccess{}, nil)

	granteeAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	entityRelationship1.Document.AccessTokens = []*coredocumentpb.AccessToken{
		{
			Grantee: granteeAccountID.ToBytes(),
		},
	}

	entityRelationship2 := &EntityRelationship{}

	err = entityRelationship2.revokeRelationship(entityRelationship1, granteeAccountID)
	assert.NoError(t, err)
	assert.Equal(t, entityRelationship1.Data, entityRelationship2.Data)
}

func TestEntityRelationship_loadData(t *testing.T) {
	entityRelationship := getTestEntityRelationship(
		t,
		documents.CollaboratorsAccess{},
		nil,
	)

	b, err := json.Marshal(entityRelationship.Data)
	assert.NoError(t, err)

	var data Data

	err = loadData(b, &data)
	assert.NoError(t, err)

	assert.Equal(t, entityRelationship.Data, data)
}

func TestEntityRelationship_getDataLeaves(t *testing.T) {
	entityRelationship := getTestEntityRelationship(
		t,
		documents.CollaboratorsAccess{},
		nil,
	)

	res, err := entityRelationship.getDataLeaves()
	assert.NoError(t, err)
	assert.NotNil(t, res)

	entityRelationship.CoreDocument = nil

	res, err = entityRelationship.getDataLeaves()
	assert.True(t, errors.IsOfType(documents.ErrDataTree, err))
	assert.Nil(t, res)
}

func TestEntityRelationship_getRawDataTree(t *testing.T) {
	entityRelationship := getTestEntityRelationship(
		t,
		documents.CollaboratorsAccess{},
		nil,
	)

	res, err := entityRelationship.getRawDataTree()
	assert.NoError(t, err)
	assert.NotNil(t, res)

	entityRelationship.CoreDocument = nil

	res, err = entityRelationship.getRawDataTree()
	assert.ErrorIs(t, err, documents.ErrCoreDocumentNil)
	assert.Nil(t, res)
}

func TestEntityRelationship_createAndLoadFromP2PProtobuf(t *testing.T) {
	entityRelationship1 := getTestEntityRelationship(
		t,
		documents.CollaboratorsAccess{},
		nil,
	)

	entityRelationshippb := entityRelationship1.createP2PProtobuf()
	assert.NotNil(t, entityRelationshippb)

	entityRelationship2 := &EntityRelationship{}

	err := entityRelationship2.loadFromP2PProtobuf(entityRelationshippb)
	assert.NoError(t, err)

	// Invalid account ID bytes for target identity
	entityRelationshippb.TargetIdentity = utils.RandomSlice(31)

	err = entityRelationship2.loadFromP2PProtobuf(entityRelationshippb)
	assert.True(t, errors.IsOfType(documents.ErrAccountIDBytesParsing, err))
}

func getTestEntityRelationshipProto() *entitypb.EntityRelationship {
	return &entitypb.EntityRelationship{
		OwnerIdentity:    utils.RandomSlice(32),
		EntityIdentifier: utils.RandomSlice(32),
		TargetIdentity:   utils.RandomSlice(32),
	}
}

func getTestEntityRelationship(
	t *testing.T,
	collaboratorsAccess documents.CollaboratorsAccess,
	attributes map[documents.AttrKey]documents.Attribute,
) *EntityRelationship {
	cd, err := documents.NewCoreDocument(compactPrefix(), collaboratorsAccess, attributes)
	assert.NoError(t, err)

	ownerIdentity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	targetIdentity, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	entityIdentifier := utils.RandomSlice(32)

	return &EntityRelationship{
		CoreDocument: cd,
		Data: Data{
			OwnerIdentity:    ownerIdentity,
			EntityIdentifier: entityIdentifier,
			TargetIdentity:   targetIdentity,
		},
	}
}

func calculateBasicDataRoot(t *testing.T, e *EntityRelationship) []byte {
	dataLeaves, err := e.getDataLeaves()
	assert.NoError(t, err)

	tree, err := e.CoreDocument.SigningDataTree(e.DocumentType(), dataLeaves)
	assert.NoError(t, err)

	return tree.RootHash()
}

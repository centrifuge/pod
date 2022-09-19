//go:build unit

package generic

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"

	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/common"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"golang.org/x/crypto/blake2b"

	"github.com/centrifuge/go-centrifuge/errors"

	"github.com/centrifuge/go-centrifuge/utils"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/centrifuge/go-centrifuge/documents"

	"github.com/stretchr/testify/assert"
)

func TestGeneric_PackCoreDocument(t *testing.T) {
	genericDocument := getTestGeneric(t, documents.CollaboratorsAccess{}, nil)

	data, err := proto.Marshal(getProtoGenericData())
	assert.NoError(t, err)

	embedData := &anypb.Any{
		TypeUrl: genericDocument.DocumentType(),
		Value:   data,
	}

	res, err := genericDocument.PackCoreDocument()
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, embedData, res.EmbeddedData)
}

func TestGeneric_UnpackCoreDocument(t *testing.T) {
	generic := getTestGeneric(t, documents.CollaboratorsAccess{}, nil)

	// No embedded data
	err := generic.UnpackCoreDocument(&coredocumentpb.CoreDocument{})
	assert.ErrorIs(t, err, documents.ErrDocumentConvertInvalidSchema)

	// Invalid embedded data type
	err = generic.UnpackCoreDocument(&coredocumentpb.CoreDocument{EmbeddedData: new(anypb.Any)})
	assert.ErrorIs(t, err, documents.ErrDocumentConvertInvalidSchema)

	// Invalid attributes
	data, err := proto.Marshal(getProtoGenericData())
	assert.NoError(t, err)

	embedData := &anypb.Any{
		TypeUrl: generic.DocumentType(),
		Value:   data,
	}

	err = generic.UnpackCoreDocument(
		&coredocumentpb.CoreDocument{
			EmbeddedData: embedData,
			Attributes: []*coredocumentpb.Attribute{
				{
					// Invalid key.
					Key: utils.RandomSlice(31),
				},
			},
		},
	)
	assert.NotNil(t, err)

	// Valid
	err = generic.UnpackCoreDocument(&coredocumentpb.CoreDocument{EmbeddedData: embedData})
	assert.NoError(t, err)
}

func TestGeneric_ToAndFromJSON(t *testing.T) {
	genericDocument := getTestGeneric(t, documents.CollaboratorsAccess{}, nil)

	b, err := genericDocument.JSON()
	assert.NoError(t, err)
	assert.NotNil(t, b)

	generic := &Generic{}

	err = generic.FromJSON(b)
	assert.NoError(t, err)
}

func TestGeneric_CreateProofs(t *testing.T) {
	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	generic := getTestGeneric(
		t,
		documents.CollaboratorsAccess{
			ReadWriteCollaborators: []*types.AccountID{accountID},
		},
		nil,
	)

	rk := generic.Document.Roles[0].RoleKey
	pf := fmt.Sprintf(documents.CDTreePrefix+".roles[%s].collaborators[0]", hexutil.Encode(rk))

	res, err := generic.CreateProofs(
		[]string{
			"generic.scheme",
			pf,
			documents.CDTreePrefix + ".document_type",
		},
	)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	dataRoot := calculateBasicDataRoot(t, generic)
	assert.NoError(t, err)

	nodeAndLeafHash, err := blake2b.New256(nil)
	assert.NoError(t, err)

	// Validate entity_number
	valid, err := documents.ValidateProof(res.FieldProofs[0], dataRoot, nodeAndLeafHash, nodeAndLeafHash)
	assert.Nil(t, err)
	assert.True(t, valid)

	// Validate roles
	valid, err = documents.ValidateProof(res.FieldProofs[1], dataRoot, nodeAndLeafHash, nodeAndLeafHash)
	assert.Nil(t, err)
	assert.True(t, valid)

	// Validate []byte value
	acc, err := types.NewAccountID(res.FieldProofs[1].Value)
	assert.NoError(t, err)
	assert.True(t, generic.AccountCanRead(acc))

	// Validate document_type
	valid, err = documents.ValidateProof(res.FieldProofs[2], dataRoot, nodeAndLeafHash, nodeAndLeafHash)
	assert.Nil(t, err)
	assert.True(t, valid)

	// Non-existing field
	res, err = generic.CreateProofs([]string{"invalid-field"})
	assert.NotNil(t, err)
	assert.Nil(t, res)

	// Nil CoreDocument
	generic.CoreDocument = nil
	res, err = generic.CreateProofs([]string{"generic.scheme"})
	assert.True(t, errors.IsOfType(documents.ErrDocumentProof, err))
	assert.Nil(t, res)
}

func TestGeneric_DocumentType(t *testing.T) {
	generic := &Generic{}
	assert.Equal(t, documenttypes.GenericDataTypeUrl, generic.DocumentType())
}

func TestGeneric_AddNFT(t *testing.T) {
	generic := getTestGeneric(t, documents.CollaboratorsAccess{}, nil)

	collectionID := types.U64(1111)
	itemID := types.NewU128(*big.NewInt(2222))

	err := generic.AddNFT(true, collectionID, itemID)
	assert.NoError(t, err)

	collectionID = types.U64(3333)
	itemID = types.NewU128(*big.NewInt(4444))

	err = generic.AddNFT(false, collectionID, itemID)
	assert.NoError(t, err)

	err = generic.AddNFT(false, collectionID, itemID)
	assert.NotNil(t, err)
}

func TestGeneric_CalculateSigningRoot(t *testing.T) {
	generic := getTestGeneric(t, documents.CollaboratorsAccess{}, nil)

	res, err := generic.CalculateSigningRoot()
	assert.NoError(t, err)
	assert.NotNil(t, res)

	generic.CoreDocument = nil

	res, err = generic.CalculateSigningRoot()
	assert.True(t, errors.IsOfType(documents.ErrDataTree, err))
	assert.Nil(t, res)
}

func TestGeneric_CalculateDocumentRoot(t *testing.T) {
	generic := getTestGeneric(t, documents.CollaboratorsAccess{}, nil)

	res, err := generic.CalculateDocumentRoot()
	assert.NoError(t, err)
	assert.NotNil(t, res)

	generic.CoreDocument = nil

	res, err = generic.CalculateDocumentRoot()
	assert.True(t, errors.IsOfType(documents.ErrDataTree, err))
	assert.Nil(t, res)
}

func TestGeneric_CollaboratorCanUpdate(t *testing.T) {
	accountID1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountID2, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountID3, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	// Create generic where accountID1 is the only one with write access.
	generic1 := getTestGeneric(
		t,
		documents.CollaboratorsAccess{
			ReadWriteCollaborators: []*types.AccountID{accountID1},
		},
		nil,
	)

	documentMock := documents.NewDocumentMock(t)

	err = generic1.CollaboratorCanUpdate(documentMock, accountID1)
	assert.True(t, errors.IsOfType(documents.ErrDocumentInvalidType, err))

	generic2 := getTestGeneric(t, documents.CollaboratorsAccess{}, nil)

	err = generic1.CollaboratorCanUpdate(generic2, accountID1)
	assert.NoError(t, err)

	err = generic1.CollaboratorCanUpdate(generic2, accountID2)
	assert.Error(t, err)

	err = generic1.CollaboratorCanUpdate(generic2, accountID3)
	assert.Error(t, err)

	// Update generic to include accountID2 as write collaborator.

	b, err := json.Marshal(generic1.Data)
	assert.NoError(t, err)

	updatePayload := documents.UpdatePayload{
		CreatePayload: documents.CreatePayload{
			Scheme:        documenttypes.EntityDataTypeUrl,
			Collaborators: documents.CollaboratorsAccess{ReadWriteCollaborators: []*types.AccountID{accountID2}},
			Data:          b,
		},
		DocumentID: generic1.Document.DocumentIdentifier,
	}

	doc, err := generic1.DeriveFromUpdatePayload(context.Background(), updatePayload)
	assert.NoError(t, err)

	err = doc.CollaboratorCanUpdate(generic2, accountID1)
	assert.NoError(t, err)

	err = doc.CollaboratorCanUpdate(generic2, accountID2)
	assert.NoError(t, err)

	err = doc.CollaboratorCanUpdate(generic2, accountID3)
	assert.Error(t, err)
}

func TestGeneric_AddAndDeleteAttributes(t *testing.T) {
	generic := getTestGeneric(t, documents.CollaboratorsAccess{}, nil)

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

	err := generic.AddAttributes(documents.CollaboratorsAccess{}, false)
	assert.True(t, errors.IsOfType(documents.ErrCDAttribute, err))

	err = generic.AddAttributes(documents.CollaboratorsAccess{}, false, attrs...)
	assert.NoError(t, err)

	res, err := generic.GetAttribute(attr1Key)
	assert.NoError(t, err)
	assert.Equal(t, attr1, res)

	res, err = generic.GetAttribute(attr2Key)
	assert.NoError(t, err)
	assert.Equal(t, attr2, res)

	err = generic.DeleteAttribute(attr1Key, false)
	assert.NoError(t, err)

	err = generic.DeleteAttribute(attr2Key, false)
	assert.NoError(t, err)

	err = generic.DeleteAttribute(attr2Key, false)
	assert.Error(t, err)

	_, err = generic.GetAttribute(attr1Key)
	assert.Error(t, err)

	_, err = generic.GetAttribute(attr2Key)
	assert.Error(t, err)
}

func TestGeneric_DeriveFromCreatePayload(t *testing.T) {
	generic := &Generic{}

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	attrKey1 := utils.RandomByte32()

	attr1 := documents.Attribute{
		KeyLabel: "label",
		Key:      attrKey1,
		Value: documents.AttrVal{
			Type: documents.AttrString,
			Str:  "string",
		},
	}

	attrs := map[documents.AttrKey]documents.Attribute{
		attrKey1: attr1,
	}

	payload := documents.CreatePayload{
		Scheme: documenttypes.EntityDataTypeUrl,
		Collaborators: documents.CollaboratorsAccess{
			ReadWriteCollaborators: []*types.AccountID{accountID},
		},
		Attributes: attrs,
		Data:       utils.RandomSlice(32),
	}

	err = generic.DeriveFromCreatePayload(context.Background(), payload)
	assert.NoError(t, err)
	assert.True(t, generic.AccountCanRead(accountID))

	// Invalid attributes

	attrKey2 := utils.RandomByte32()
	attr2 := documents.Attribute{
		KeyLabel: "label_2",
		Key:      attrKey2,
		Value: documents.AttrVal{
			Type: "invalid_type",
		},
	}

	attrs[attrKey2] = attr2
	payload.Attributes = attrs

	err = generic.DeriveFromCreatePayload(context.Background(), payload)
	assert.True(t, errors.IsOfType(documents.ErrCDCreate, err))
}

func TestGeneric_DeriveFromClonePayload(t *testing.T) {
	generic1 := getTestGeneric(t, documents.CollaboratorsAccess{}, nil)
	generic2 := getTestGeneric(t, documents.CollaboratorsAccess{}, nil)

	ctx := context.Background()

	err := generic1.DeriveFromClonePayload(ctx, generic2)
	assert.NoError(t, err)

	documentMock := documents.NewDocumentMock(t)

	documentMock.On("PackCoreDocument").
		Return(nil, errors.New("error")).
		Once()

	err = generic1.DeriveFromClonePayload(ctx, documentMock)
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

	err = generic1.DeriveFromClonePayload(ctx, documentMock)
	assert.True(t, errors.IsOfType(documents.ErrCDClone, err))
}

func TestGeneric_DeriveFromUpdatePayload(t *testing.T) {
	generic := getTestGeneric(t, documents.CollaboratorsAccess{}, nil)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	attrKey1 := utils.RandomByte32()

	attr1 := documents.Attribute{
		KeyLabel: "label",
		Key:      attrKey1,
		Value: documents.AttrVal{
			Type: documents.AttrString,
			Str:  "string",
		},
	}

	attrs := map[documents.AttrKey]documents.Attribute{
		attrKey1: attr1,
	}

	documentID := utils.RandomSlice(32)

	payload := documents.UpdatePayload{
		CreatePayload: documents.CreatePayload{
			Scheme: documenttypes.EntityDataTypeUrl,
			Collaborators: documents.CollaboratorsAccess{
				ReadWriteCollaborators: []*types.AccountID{accountID},
			},
			Attributes: attrs,
			Data:       utils.RandomSlice(32),
		},
		DocumentID: documentID,
	}

	ctx := context.Background()

	res, err := generic.DeriveFromUpdatePayload(ctx, payload)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, Data{}, res.GetData())

	// Invalid attributes
	attrKey2 := utils.RandomByte32()
	attr2 := documents.Attribute{
		KeyLabel: "label_2",
		Key:      attrKey2,
		Value: documents.AttrVal{
			Type: "invalid_type",
		},
	}

	attrs[attrKey2] = attr2
	payload.Attributes = attrs

	res, err = generic.DeriveFromUpdatePayload(ctx, payload)
	assert.True(t, errors.IsOfType(documents.ErrCDNewVersion, err))
	assert.Nil(t, res)
}

func TestGeneric_Patch(t *testing.T) {
	generic := getTestGeneric(t, documents.CollaboratorsAccess{}, nil)

	// Invalid attributes
	attrKey := utils.RandomByte32()
	attr := documents.Attribute{
		KeyLabel: "labels",
		Key:      attrKey,
		Value: documents.AttrVal{
			Type: "invalid_type",
		},
	}

	payload := documents.UpdatePayload{
		CreatePayload: documents.CreatePayload{
			Attributes: map[documents.AttrKey]documents.Attribute{
				attrKey: attr,
			},
		},
	}

	err := generic.Patch(payload)
	assert.True(t, errors.IsOfType(documents.ErrDocumentPatch, err))
}

func TestGeneric_Scheme(t *testing.T) {
	generic := &Generic{}
	assert.Equal(t, Scheme, generic.Scheme())
}

func TestGeneric_GetData(t *testing.T) {
	generic := getTestGeneric(t, documents.CollaboratorsAccess{}, nil)

	assert.Equal(t, Data{}, generic.GetData())
}

func TestGeneric_getDataLeaves(t *testing.T) {
	generic := getTestGeneric(t, documents.CollaboratorsAccess{}, nil)

	res, err := generic.getDataLeaves()
	assert.NoError(t, err)
	assert.NotNil(t, res)

	generic.CoreDocument = nil

	res, err = generic.getDataLeaves()
	assert.True(t, errors.IsOfType(documents.ErrDataTree, err))
	assert.Nil(t, res)
}

func TestGeneric_getRawDataTree(t *testing.T) {
	generic := getTestGeneric(
		t,
		documents.CollaboratorsAccess{},
		nil,
	)

	res, err := generic.getRawDataTree()
	assert.NoError(t, err)
	assert.NotNil(t, res)

	generic.CoreDocument = nil

	res, err = generic.getRawDataTree()
	assert.ErrorIs(t, err, documents.ErrCoreDocumentNil)
	assert.Nil(t, res)
}

func TestGeneric_getDocumentDataTree(t *testing.T) {
	generic := getTestGeneric(t, documents.CollaboratorsAccess{}, nil)

	res, err := generic.getDocumentDataTree()
	assert.NoError(t, err)
	assert.NotNil(t, res)

	_, leaf := res.GetLeafByProperty("generic.scheme")
	assert.NotNil(t, leaf)
	assert.Equal(t, "generic.scheme", leaf.Property.ReadableName())

	generic.CoreDocument = nil

	res, err = generic.getDocumentDataTree()
	assert.ErrorIs(t, err, documents.ErrCoreDocumentNil)
	assert.Nil(t, res)
}

func getTestGeneric(t *testing.T, collaboratorAccess documents.CollaboratorsAccess, attrs map[documents.AttrKey]documents.Attribute) *Generic {
	cd, err := documents.NewCoreDocument(compactPrefix(), collaboratorAccess, attrs)
	assert.NoError(t, err)

	return &Generic{
		CoreDocument: cd,
		Data:         Data{},
	}
}

func calculateBasicDataRoot(t *testing.T, g *Generic) []byte {
	dataLeaves, err := g.getDataLeaves()
	assert.NoError(t, err)

	tree, err := g.CoreDocument.SigningDataTree(g.DocumentType(), dataLeaves)
	assert.NoError(t, err)

	return tree.RootHash()
}

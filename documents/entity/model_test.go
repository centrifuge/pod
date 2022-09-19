//go:build unit

package entity

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	entitypb "github.com/centrifuge/centrifuge-protobufs/gen/go/entity"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/common"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/blake2b"
)

func TestEntity_PackCoreDocument(t *testing.T) {
	entity := getTestEntity(t, documents.CollaboratorsAccess{}, nil)

	cd, err := entity.PackCoreDocument()
	assert.NoError(t, err)
	assert.NotNil(t, cd)
	assert.NotNil(t, cd.EmbeddedData)

	data, err := proto.Marshal(entity.createP2PProtobuf())
	assert.NoError(t, err)

	embedData := &anypb.Any{
		TypeUrl: entity.DocumentType(),
		Value:   data,
	}
	assert.Equal(t, embedData, cd.EmbeddedData)
}

func TestEntityModel_UnpackCoreDocument(t *testing.T) {
	entity := getTestEntity(t, documents.CollaboratorsAccess{}, nil)

	// No embedded data
	err := entity.UnpackCoreDocument(&coredocumentpb.CoreDocument{})
	assert.ErrorIs(t, err, documents.ErrDocumentConvertInvalidSchema)

	// Invalid embedded data type
	err = entity.UnpackCoreDocument(&coredocumentpb.CoreDocument{EmbeddedData: new(anypb.Any)})
	assert.ErrorIs(t, err, documents.ErrDocumentConvertInvalidSchema)

	// Invalid embedded data
	err = entity.UnpackCoreDocument(&coredocumentpb.CoreDocument{
		EmbeddedData: &anypb.Any{
			Value:   utils.RandomSlice(32),
			TypeUrl: documenttypes.EntityDataTypeUrl,
		},
	})
	assert.True(t, errors.IsOfType(documents.ErrDocumentDataUnmarshalling, err))

	// Invalid document attributes
	entityData := getTestEntityProto()

	b, err := proto.Marshal(entityData)
	assert.NoError(t, err)

	err = entity.UnpackCoreDocument(&coredocumentpb.CoreDocument{
		EmbeddedData: &anypb.Any{
			Value:   b,
			TypeUrl: documenttypes.EntityDataTypeUrl,
		},
		Attributes: []*coredocumentpb.Attribute{
			{
				// Invalid key.
				Key: utils.RandomSlice(31),
			},
		},
	})
	assert.NotNil(t, err)

	// Invalid account ID bytes
	entityData.Identity = utils.RandomSlice(31)

	err = entity.UnpackCoreDocument(&coredocumentpb.CoreDocument{
		EmbeddedData: &anypb.Any{
			Value:   b,
			TypeUrl: documenttypes.EntityDataTypeUrl,
		},
	})
	assert.True(t, errors.IsOfType(documents.ErrAccountIDBytesParsing, err))

	// Valid
	entityData.Identity = utils.RandomSlice(32)

	b, err = proto.Marshal(entityData)
	assert.NoError(t, err)

	err = entity.UnpackCoreDocument(&coredocumentpb.CoreDocument{
		EmbeddedData: &anypb.Any{
			Value:   b,
			TypeUrl: documenttypes.EntityDataTypeUrl,
		},
	})
	assert.NoError(t, err)
}

func TestEntity_ToAndFromJSON(t *testing.T) {
	entity := getTestEntity(t, documents.CollaboratorsAccess{}, nil)

	b, err := entity.JSON()
	assert.NoError(t, err)

	newEntity := &Entity{}
	err = newEntity.FromJSON(b)
	assert.NoError(t, err)
}

func TestEntity_CreateProofs(t *testing.T) {
	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	entity := getTestEntity(
		t,
		documents.CollaboratorsAccess{
			ReadWriteCollaborators: []*types.AccountID{accountID},
		},
		nil,
	)

	rk := entity.Document.Roles[0].RoleKey
	pf := fmt.Sprintf(documents.CDTreePrefix+".roles[%s].collaborators[0]", hexutil.Encode(rk))

	res, err := entity.CreateProofs(
		[]string{
			"entity.legal_name",
			pf,
			documents.CDTreePrefix + ".document_type",
		},
	)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	dataRoot := calculateBasicDataRoot(t, entity)
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
	assert.True(t, entity.AccountCanRead(acc))

	// Validate document_type
	valid, err = documents.ValidateProof(res.FieldProofs[2], dataRoot, nodeAndLeafHash, nodeAndLeafHash)
	assert.Nil(t, err)
	assert.True(t, valid)

	// Non-existing field
	res, err = entity.CreateProofs([]string{"invalid-field"})
	assert.NotNil(t, err)
	assert.Nil(t, res)

	// Nil CoreDocument
	entity.CoreDocument = nil
	res, err = entity.CreateProofs([]string{"entity.legal_name"})
	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func TestEntity_DocumentType(t *testing.T) {
	entity := &Entity{}
	assert.Equal(t, documenttypes.EntityDataTypeUrl, entity.DocumentType())
}

func TestEntity_AddNFT(t *testing.T) {
	entity := getTestEntity(t, documents.CollaboratorsAccess{}, nil)

	collectionID := types.U64(1111)
	itemID := types.NewU128(*big.NewInt(2222))

	err := entity.AddNFT(true, collectionID, itemID)
	assert.NoError(t, err)

	collectionID = types.U64(3333)
	itemID = types.NewU128(*big.NewInt(4444))

	err = entity.AddNFT(false, collectionID, itemID)
	assert.NoError(t, err)

	err = entity.AddNFT(false, collectionID, itemID)
	assert.NotNil(t, err)
}

func TestEntity_CalculateSigningRoot(t *testing.T) {
	entity := getTestEntity(t, documents.CollaboratorsAccess{}, nil)

	res, err := entity.CalculateSigningRoot()
	assert.NoError(t, err)
	assert.NotNil(t, res)

	entity.CoreDocument = nil

	res, err = entity.CalculateSigningRoot()
	assert.True(t, errors.IsOfType(documents.ErrDataTree, err))
	assert.Nil(t, res)
}

func TestEntity_CalculateDocumentRoot(t *testing.T) {
	entity := getTestEntity(t, documents.CollaboratorsAccess{}, nil)

	res, err := entity.CalculateDocumentRoot()
	assert.NoError(t, err)
	assert.NotNil(t, res)

	entity.CoreDocument = nil

	res, err = entity.CalculateDocumentRoot()
	assert.True(t, errors.IsOfType(documents.ErrDataTree, err))
	assert.Nil(t, res)
}

func TestEntity_CollaboratorCanUpdate(t *testing.T) {
	accountID1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountID2, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	accountID3, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	// Create entity where accountID1 is the only one with write access.
	entity1 := getTestEntity(
		t,
		documents.CollaboratorsAccess{
			ReadWriteCollaborators: []*types.AccountID{accountID1},
		},
		nil,
	)

	documentMock := documents.NewDocumentMock(t)

	err = entity1.CollaboratorCanUpdate(documentMock, accountID1)
	assert.True(t, errors.IsOfType(documents.ErrDocumentInvalidType, err))

	entity2 := getTestEntity(t, documents.CollaboratorsAccess{}, nil)

	err = entity1.CollaboratorCanUpdate(entity2, accountID1)
	assert.NoError(t, err)

	err = entity1.CollaboratorCanUpdate(entity2, accountID2)
	assert.Error(t, err)

	err = entity1.CollaboratorCanUpdate(entity2, accountID3)
	assert.Error(t, err)

	// Update entity to include accountID2 as write collaborator.

	b, err := json.Marshal(entity1.Data)
	assert.NoError(t, err)

	updatePayload := documents.UpdatePayload{
		CreatePayload: documents.CreatePayload{
			Scheme:        documenttypes.EntityDataTypeUrl,
			Collaborators: documents.CollaboratorsAccess{ReadWriteCollaborators: []*types.AccountID{accountID2}},
			Data:          b,
		},
		DocumentID: entity1.Document.DocumentIdentifier,
	}

	doc, err := entity1.DeriveFromUpdatePayload(context.Background(), updatePayload)
	assert.NoError(t, err)

	err = doc.CollaboratorCanUpdate(entity2, accountID1)
	assert.NoError(t, err)

	err = doc.CollaboratorCanUpdate(entity2, accountID2)
	assert.NoError(t, err)

	err = doc.CollaboratorCanUpdate(entity2, accountID3)
	assert.Error(t, err)

	// Add transition rules and roles to give accountID3 write access for core document and
	// legal name field.
	// Note that write access for core document is required since getTestEntity creates a new core document.

	entity3 := doc.(*Entity)

	legalNameEditRoleKey := utils.RandomSlice(32)
	coreDocEditRoleKey := utils.RandomSlice(32)

	transitionRules := []*coredocumentpb.TransitionRule{
		{
			// transition rule for legal name field
			RuleKey: utils.RandomSlice(32),
			Roles: [][]byte{
				legalNameEditRoleKey,
			},
			MatchType: coredocumentpb.FieldMatchType_FIELD_MATCH_TYPE_EXACT,
			Field:     append(compactPrefix(), 0, 0, 0, 2),
			Action:    coredocumentpb.TransitionAction_TRANSITION_ACTION_EDIT,
		},
		{
			// transition rule for core document
			RuleKey: utils.RandomSlice(32),
			Roles: [][]byte{
				coreDocEditRoleKey,
			},
			MatchType: coredocumentpb.FieldMatchType_FIELD_MATCH_TYPE_PREFIX,
			Field:     documents.CompactProperties(documents.CDTreePrefix),
			Action:    coredocumentpb.TransitionAction_TRANSITION_ACTION_EDIT,
		},
	}

	editRoles := []*coredocumentpb.Role{
		{
			RoleKey: legalNameEditRoleKey,
			Collaborators: [][]byte{
				accountID3.ToBytes(),
			},
		},
		{
			RoleKey: coreDocEditRoleKey,
			Collaborators: [][]byte{
				accountID3.ToBytes(),
			},
		},
	}

	entity3.Document.TransitionRules = append(entity3.Document.TransitionRules, transitionRules...)
	entity3.Document.Roles = append(entity3.Document.Roles, editRoles...)

	// Ensure that entity 2 has the same identity and bank details identifier.
	entity2.Data.Identity = entity3.Data.Identity
	entity2.Data.PaymentDetails[0].BankPaymentMethod.Identifier = entity3.Data.PaymentDetails[0].BankPaymentMethod.Identifier

	entity2.Data.LegalName = "new_legal_name"

	err = entity3.CollaboratorCanUpdate(entity2, accountID3)
	assert.NoError(t, err)

	// Confirm that accountID 3 cannot change other fields.
	entity2.Data.Addresses = nil

	err = entity3.CollaboratorCanUpdate(entity2, accountID3)
	assert.Error(t, err)

	entity2.Data.Addresses = entity3.Data.Addresses
	entity2.Data.PaymentDetails = nil

	err = entity3.CollaboratorCanUpdate(entity2, accountID3)
	assert.Error(t, err)

	entity2.Data.PaymentDetails = entity3.Data.PaymentDetails
	entity2.Data.Contacts = nil

	err = entity3.CollaboratorCanUpdate(entity2, accountID3)
	assert.Error(t, err)

	randomAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	entity2.Data.Contacts = entity3.Data.Contacts
	entity2.Data.Identity = randomAccountID

	err = entity3.CollaboratorCanUpdate(entity2, accountID3)
	assert.Error(t, err)
}

func TestEntity_AddAndDeleteAttributes(t *testing.T) {
	entity := getTestEntity(t, documents.CollaboratorsAccess{}, nil)

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

	err := entity.AddAttributes(documents.CollaboratorsAccess{}, false)
	assert.True(t, errors.IsOfType(documents.ErrCDAttribute, err))

	err = entity.AddAttributes(documents.CollaboratorsAccess{}, false, attrs...)
	assert.NoError(t, err)

	res, err := entity.GetAttribute(attr1Key)
	assert.NoError(t, err)
	assert.Equal(t, attr1, res)

	res, err = entity.GetAttribute(attr2Key)
	assert.NoError(t, err)
	assert.Equal(t, attr2, res)

	err = entity.DeleteAttribute(attr1Key, false)
	assert.NoError(t, err)

	err = entity.DeleteAttribute(attr2Key, false)
	assert.NoError(t, err)

	err = entity.DeleteAttribute(attr2Key, false)
	assert.Error(t, err)

	_, err = entity.GetAttribute(attr1Key)
	assert.Error(t, err)

	_, err = entity.GetAttribute(attr2Key)
	assert.Error(t, err)
}

func TestEntity_GetData(t *testing.T) {
	entity := getTestEntity(t, documents.CollaboratorsAccess{}, nil)

	assert.Equal(t, entity.Data, entity.GetData())
}

func TestEntity_DeriveFromCreatePayload(t *testing.T) {
	entity := &Entity{}

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	testData := getTestData(accountID)

	b, err := json.Marshal(testData)
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
		Data:       b,
	}

	err = entity.DeriveFromCreatePayload(context.Background(), payload)
	assert.NoError(t, err)
	assert.True(t, entity.AccountCanRead(accountID))

	// Invalid data
	testData.PaymentDetails[0].CryptoPaymentMethod = &CryptoPaymentMethod{}

	b, err = json.Marshal(testData)
	assert.NoError(t, err)

	payload = documents.CreatePayload{
		Scheme: documenttypes.EntityDataTypeUrl,
		Collaborators: documents.CollaboratorsAccess{
			ReadWriteCollaborators: []*types.AccountID{accountID},
		},
		Attributes: attrs,
		Data:       b,
	}

	err = entity.DeriveFromCreatePayload(context.Background(), payload)
	assert.True(t, errors.IsOfType(ErrEntityInvalidData, err))

	// Reset test data
	testData.PaymentDetails[0].CryptoPaymentMethod = nil

	// Invalid attributes

	attrKey2 := utils.RandomByte32()
	attr2 := documents.Attribute{
		KeyLabel: "label_2",
		Key:      attrKey2,
		Value: documents.AttrVal{
			Type: "invalid_type",
		},
	}

	b, err = json.Marshal(testData)
	assert.NoError(t, err)

	attrs[attrKey2] = attr2
	payload.Attributes = attrs
	payload.Data = b

	err = entity.DeriveFromCreatePayload(context.Background(), payload)
	assert.True(t, errors.IsOfType(documents.ErrCDCreate, err))
}

func TestEntity_DeriveFromClonePayload(t *testing.T) {
	entity1 := getTestEntity(t, documents.CollaboratorsAccess{}, nil)
	entity2 := getTestEntity(t, documents.CollaboratorsAccess{}, nil)

	ctx := context.Background()

	err := entity1.DeriveFromClonePayload(ctx, entity2)
	assert.NoError(t, err)

	documentMock := documents.NewDocumentMock(t)

	documentMock.On("PackCoreDocument").
		Return(nil, errors.New("error")).
		Once()

	err = entity1.DeriveFromClonePayload(ctx, documentMock)
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

	err = entity1.DeriveFromClonePayload(ctx, documentMock)
	assert.True(t, errors.IsOfType(documents.ErrCDClone, err))
}

func TestEntity_DeriveFromUpdatePayload(t *testing.T) {
	entity := getTestEntity(t, documents.CollaboratorsAccess{}, nil)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	data := entity.Data
	data.LegalName = "new_name"

	b, err := json.Marshal(data)
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
			Data:       b,
		},
		DocumentID: documentID,
	}

	ctx := context.Background()

	res, err := entity.DeriveFromUpdatePayload(ctx, payload)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, data, res.GetData())

	// Invalid data
	data.PaymentDetails[0].CryptoPaymentMethod = &CryptoPaymentMethod{}

	b, err = json.Marshal(data)
	assert.NoError(t, err)

	payload.Data = b

	res, err = entity.DeriveFromUpdatePayload(ctx, payload)
	assert.True(t, errors.IsOfType(ErrEntityInvalidData, err))
	assert.Nil(t, res)

	// Reset data
	data.PaymentDetails[0].CryptoPaymentMethod = nil

	// Invalid attributes
	attrKey2 := utils.RandomByte32()
	attr2 := documents.Attribute{
		KeyLabel: "label_2",
		Key:      attrKey2,
		Value: documents.AttrVal{
			Type: "invalid_type",
		},
	}

	b, err = json.Marshal(data)
	assert.NoError(t, err)

	attrs[attrKey2] = attr2
	payload.Attributes = attrs
	payload.Data = b

	res, err = entity.DeriveFromUpdatePayload(ctx, payload)
	assert.True(t, errors.IsOfType(documents.ErrCDNewVersion, err))
	assert.Nil(t, res)
}

func TestEntity_Patch(t *testing.T) {
	entity := getTestEntity(t, documents.CollaboratorsAccess{}, nil)

	data := entity.Data

	b, err := json.Marshal(data)
	assert.NoError(t, err)

	payload := documents.UpdatePayload{
		CreatePayload: documents.CreatePayload{
			Data: b,
		},
	}

	err = entity.Patch(payload)
	assert.NoError(t, err)

	// Invalid data
	data.PaymentDetails[0].CryptoPaymentMethod = &CryptoPaymentMethod{}

	b, err = json.Marshal(data)
	assert.NoError(t, err)

	payload = documents.UpdatePayload{
		CreatePayload: documents.CreatePayload{
			Data: b,
		},
	}

	err = entity.Patch(payload)
	assert.True(t, errors.IsOfType(ErrEntityInvalidData, err))

	// Reset data
	data.PaymentDetails[0].CryptoPaymentMethod = nil

	// Invalid attributes
	attrKey := utils.RandomByte32()
	attr := documents.Attribute{
		KeyLabel: "labels",
		Key:      attrKey,
		Value: documents.AttrVal{
			Type: "invalid_type",
		},
	}

	b, err = json.Marshal(data)
	assert.NoError(t, err)

	payload = documents.UpdatePayload{
		CreatePayload: documents.CreatePayload{
			Data: b,
			Attributes: map[documents.AttrKey]documents.Attribute{
				attrKey: attr,
			},
		},
	}

	err = entity.Patch(payload)
	assert.True(t, errors.IsOfType(documents.ErrDocumentPatch, err))
}

func TestEntity_Scheme(t *testing.T) {
	entity := getTestEntity(t, documents.CollaboratorsAccess{}, nil)
	assert.Equal(t, Scheme, entity.Scheme())
}

func TestEntity_createAndLoadFromP2PProtobuf(t *testing.T) {
	entity1 := getTestEntity(t, documents.CollaboratorsAccess{}, nil)

	res := entity1.createP2PProtobuf()
	assert.NotNil(t, res)
	assert.Equal(t, entity1.Data.Identity.ToBytes(), res.Identity)
	assert.Equal(t, entity1.Data.LegalName, res.LegalName)
	assert.Equal(t, toProtoAddresses(entity1.Data.Addresses), res.Addresses)
	assert.Equal(t, toProtoPaymentDetails(entity1.Data.PaymentDetails), res.PaymentDetails)
	assert.Equal(t, toProtoContacts(entity1.Data.Contacts), res.Contacts)

	entity2 := &Entity{}
	err := entity2.loadFromP2PProtobuf(res)
	assert.NoError(t, err)
	assert.Equal(t, entity1.Data, entity2.Data)

	// Invalid account ID bytes
	res.Identity = utils.RandomSlice(31)

	err = entity2.loadFromP2PProtobuf(res)
	assert.True(t, errors.IsOfType(documents.ErrAccountIDBytesParsing, err))
}

func TestEntity_getDataLeaves(t *testing.T) {
	entity := getTestEntity(t, documents.CollaboratorsAccess{}, nil)

	res, err := entity.getDataLeaves()
	assert.NoError(t, err)
	assert.NotNil(t, res)

	entity.CoreDocument = nil

	res, err = entity.getDataLeaves()
	assert.True(t, errors.IsOfType(documents.ErrDataTree, err))
	assert.Nil(t, res)
}

func TestEntity_getRawDataTree(t *testing.T) {
	entity := getTestEntity(t, documents.CollaboratorsAccess{}, nil)

	res, err := entity.getRawDataTree()
	assert.NoError(t, err)
	assert.NotNil(t, res)

	entity.CoreDocument = nil

	res, err = entity.getRawDataTree()
	assert.ErrorIs(t, err, documents.ErrCoreDocumentNil)
	assert.Nil(t, res)
}

func TestEntity_getDocumentDataTree(t *testing.T) {
	entity := getTestEntity(t, documents.CollaboratorsAccess{}, nil)

	res, err := entity.getDocumentDataTree()
	assert.NoError(t, err)
	assert.NotNil(t, res)

	_, leaf := res.GetLeafByProperty("entity.legal_name")
	assert.NotNil(t, leaf)
	assert.Equal(t, "entity.legal_name", leaf.Property.ReadableName())

	entity.CoreDocument = nil

	res, err = entity.getDocumentDataTree()
	assert.ErrorIs(t, err, documents.ErrCoreDocumentNil)
	assert.Nil(t, res)
}

func TestEntity_IsOnlyOneSet(t *testing.T) {
	paymentDetail := PaymentDetail{}

	err := isOnlyOneSet(paymentDetail.BankPaymentMethod, paymentDetail.CryptoPaymentMethod, paymentDetail.OtherPaymentMethod)
	assert.ErrorIs(t, err, ErrNoPaymentMethodSet)

	paymentDetail.BankPaymentMethod = &BankPaymentMethod{}

	err = isOnlyOneSet(paymentDetail.BankPaymentMethod, paymentDetail.CryptoPaymentMethod, paymentDetail.OtherPaymentMethod)
	assert.NoError(t, err)

	paymentDetail.CryptoPaymentMethod = &CryptoPaymentMethod{}

	err = isOnlyOneSet(paymentDetail.BankPaymentMethod, paymentDetail.CryptoPaymentMethod, paymentDetail.OtherPaymentMethod)
	assert.ErrorIs(t, err, ErrMultiplePaymentMethodsSet)
}

func TestEntity_loadData(t *testing.T) {
	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	testData := getTestData(accountID)

	b, err := json.Marshal(testData)
	assert.NoError(t, err)

	var data Data

	err = loadData(b, &data)
	assert.NoError(t, err)

	testData.PaymentDetails[0].CryptoPaymentMethod = &CryptoPaymentMethod{}

	b, err = json.Marshal(testData)
	assert.NoError(t, err)

	err = loadData(b, &data)
	assert.Error(t, err)
}

func TestEntity_patch(t *testing.T) {
	entity := getTestEntity(t, documents.CollaboratorsAccess{}, nil)

	data := entity.Data

	b, err := json.Marshal(data)
	assert.NoError(t, err)

	payload := documents.UpdatePayload{
		CreatePayload: documents.CreatePayload{
			Data: b,
		},
	}

	res, err := entity.patch(payload)
	assert.NoError(t, err)
	assert.Equal(t, data, res)

	// Invalid data
	data.PaymentDetails[0].CryptoPaymentMethod = &CryptoPaymentMethod{}

	b, err = json.Marshal(data)
	assert.NoError(t, err)

	payload = documents.UpdatePayload{
		CreatePayload: documents.CreatePayload{
			Data: b,
		},
	}

	_, err = entity.patch(payload)
	assert.True(t, errors.IsOfType(ErrEntityInvalidData, err))
}

func getTestEntityProto() *entitypb.Entity {
	return &entitypb.Entity{
		Identity:  utils.RandomSlice(32),
		LegalName: "legal_name",
		Addresses: []*entitypb.Address{
			{
				IsShipTo:      true,
				Label:         "label",
				Zip:           "zip",
				State:         "state",
				Country:       "country",
				AddressLine1:  "address_line1",
				AddressLine2:  "address_line1",
				ContactPerson: "address_line1",
			},
		},
		PaymentDetails: []*entitypb.PaymentDetail{
			{
				Predefined: true,
				PaymentMethod: &entitypb.PaymentDetail_BankPaymentMethod{
					BankPaymentMethod: &entitypb.BankPaymentMethod{
						Identifier: utils.RandomSlice(32),
						Address: &entitypb.Address{
							IsShipTo:      true,
							Label:         "payment_label",
							Zip:           "payment_zip",
							State:         "payment_state",
							Country:       "payment_country",
							AddressLine1:  "payment_address_line1",
							AddressLine2:  "payment_address_line1",
							ContactPerson: "payment_address_line1",
						},
						HolderName:        "holder_name",
						BankKey:           "bank_key",
						BankAccountNumber: "bank_account_number",
						SupportedCurrency: "supported_currency",
					},
				},
			},
		},
		Contacts: []*entitypb.Contact{
			{
				Name:  "name",
				Title: "title",
				Email: "email",
				Phone: "phone",
				Fax:   "fax",
			},
		},
	}
}

func getTestEntity(t *testing.T, collaboratorsAccess documents.CollaboratorsAccess, attributes map[documents.AttrKey]documents.Attribute) *Entity {
	cd, err := documents.NewCoreDocument(compactPrefix(), collaboratorsAccess, attributes)
	assert.NoError(t, err)
	assert.NotNil(t, cd)

	accountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	return &Entity{
		CoreDocument: cd,
		Data:         *getTestData(accountID),
	}
}

func getTestData(identity *types.AccountID) *Data {
	return &Data{
		Identity:  identity,
		LegalName: "legal_name",
		Addresses: []Address{
			{
				IsMain:        true,
				Label:         "label",
				Zip:           "zip",
				State:         "state",
				Country:       "country",
				AddressLine1:  "addr_line1",
				AddressLine2:  "addr_line1",
				ContactPerson: "person",
			},
		},
		PaymentDetails: []PaymentDetail{
			{
				Predefined: true,
				BankPaymentMethod: &BankPaymentMethod{
					Identifier: utils.RandomSlice(32),
					Address: Address{
						IsPayTo:       true,
						Label:         "payment_label",
						Zip:           "payment_zip",
						State:         "payment_state",
						Country:       "payment_country",
						AddressLine1:  "payment_addr1",
						AddressLine2:  "payment_addr2",
						ContactPerson: "payment_contact_person",
					},
					HolderName:        "holder_name",
					BankKey:           "bank_key",
					BankAccountNumber: "bank_account_number",
					SupportedCurrency: "supported_currency",
				},
			},
		},
		Contacts: []Contact{
			{
				Name:  "name",
				Title: "title",
				Email: "email",
				Phone: "phone",
				Fax:   "fax",
			},
		},
	}
}

func calculateBasicDataRoot(t *testing.T, e *Entity) []byte {
	dataLeaves, err := e.getDataLeaves()
	assert.NoError(t, err)

	tree, err := e.CoreDocument.SigningDataTree(e.DocumentType(), dataLeaves)
	assert.NoError(t, err)

	return tree.RootHash()
}

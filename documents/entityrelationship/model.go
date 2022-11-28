package entityrelationship

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	entitypb "github.com/centrifuge/centrifuge-protobufs/gen/go/entity"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/utils/byteutils"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/jinzhu/copier"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

const (
	prefix string = "entity_relationship"

	// Scheme to identify entity relationship
	Scheme = prefix
)

// tree prefixes for specific documents use the second byte of a 4 byte slice by convention
func compactPrefix() []byte { return []byte{0, 4, 0, 0} }

// Data represents entity relationship data
type Data struct {
	// Owner of the relationship
	OwnerIdentity *types.AccountID `json:"owner_identity" swaggertype:"primitive,string"`
	// Entity identifier
	EntityIdentifier byteutils.HexBytes `json:"entity_identifier" swaggertype:"primitive,string"`
	// Identity which will be granted access
	TargetIdentity *types.AccountID `json:"target_identity" swaggertype:"primitive,string"`
}

// EntityRelationship implements the documents.Document and keeps track of entity-relationship related fields and state.
type EntityRelationship struct {
	*documents.CoreDocument

	Data Data `json:"data"`
}

// createP2PProtobuf returns Centrifuge protobuf-specific RelationshipData.
func (e *EntityRelationship) createP2PProtobuf() *entitypb.EntityRelationship {
	d := e.Data
	accountIDByteSlices := documents.AccountIDsToBytesSlice(d.OwnerIdentity, d.TargetIdentity)
	return &entitypb.EntityRelationship{
		OwnerIdentity:    accountIDByteSlices[0],
		TargetIdentity:   accountIDByteSlices[1],
		EntityIdentifier: d.EntityIdentifier,
	}
}

// loadFromP2PProtobuf loads the Entity Relationship from Centrifuge protobuf.
func (e *EntityRelationship) loadFromP2PProtobuf(entityRelationship *entitypb.EntityRelationship) error {
	accountIDs, err := documents.ParseAccountIDBytes(entityRelationship.OwnerIdentity, entityRelationship.TargetIdentity)
	if err != nil {
		return err
	}
	var d Data
	d.OwnerIdentity = accountIDs[0]
	d.TargetIdentity = accountIDs[1]
	d.EntityIdentifier = entityRelationship.EntityIdentifier
	e.Data = d
	return nil
}

// PackCoreDocument packs the EntityRelationship into a CoreDocument.
func (e *EntityRelationship) PackCoreDocument() (cd *coredocumentpb.CoreDocument, err error) {
	entityRelationship := e.createP2PProtobuf()

	data, err := proto.Marshal(entityRelationship)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentDataMarshalling, err)
	}

	embedData := &anypb.Any{
		TypeUrl: e.DocumentType(),
		Value:   data,
	}

	return e.CoreDocument.PackCoreDocument(embedData), nil
}

// UnpackCoreDocument unpacks the core document into an EntityRelationship.
func (e *EntityRelationship) UnpackCoreDocument(cd *coredocumentpb.CoreDocument) error {
	if cd.EmbeddedData == nil ||
		cd.EmbeddedData.TypeUrl != e.DocumentType() {
		return documents.ErrDocumentConvertInvalidSchema
	}

	entityRelationship := new(entitypb.EntityRelationship)
	err := proto.Unmarshal(cd.EmbeddedData.Value, entityRelationship)
	if err != nil {
		return errors.NewTypedError(documents.ErrDocumentDataUnmarshalling, err)
	}

	err = e.loadFromP2PProtobuf(entityRelationship)
	if err != nil {
		return err
	}
	e.CoreDocument, err = documents.NewCoreDocumentFromProtobuf(cd)
	return err
}

// JSON marshals EntityRelationship into a json bytes
func (e *EntityRelationship) JSON() ([]byte, error) {
	return e.CoreDocument.MarshalJSON(e)
}

// FromJSON unmarshals the json bytes into EntityRelationship
func (e *EntityRelationship) FromJSON(jsonData []byte) error {
	if e.CoreDocument == nil {
		e.CoreDocument = new(documents.CoreDocument)
	}

	return e.CoreDocument.UnmarshalJSON(jsonData, e)
}

// Type gives the EntityRelationship type.
func (e *EntityRelationship) Type() reflect.Type {
	return reflect.TypeOf(e)
}

func (e *EntityRelationship) getDataLeaves() ([]proofs.LeafNode, error) {
	t, err := e.getRawDataTree()
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDataTree, err)
	}
	return t.GetLeaves(), nil
}

func (e *EntityRelationship) getRawDataTree() (*proofs.DocumentTree, error) {
	entityProto := e.createP2PProtobuf()
	if e.CoreDocument == nil {
		return nil, documents.ErrCoreDocumentNil
	}

	t, err := e.CoreDocument.DefaultTreeWithPrefix(prefix, compactPrefix())
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDataTree, err)
	}

	err = t.AddLeavesFromDocument(entityProto)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDataTree, err)
	}
	return t, nil
}

// CreateProofs generates proofs for given fields.
func (e *EntityRelationship) CreateProofs(fields []string) (prf *documents.DocumentProof, err error) {
	dataLeaves, err := e.getDataLeaves()
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentProof, err)
	}

	return e.CoreDocument.CreateProofs(e.DocumentType(), dataLeaves, fields)
}

// DocumentType returns the entity relationship document type.
func (*EntityRelationship) DocumentType() string {
	return documenttypes.EntityRelationshipDataTypeUrl
}

// AddNFT is not implemented for EntityRelationship
func (e *EntityRelationship) AddNFT(_ bool, _ types.U64, _ types.U128) error {
	return documents.ErrNotImplemented
}

// CalculateSigningRoot calculates the signing root of the document.
func (e *EntityRelationship) CalculateSigningRoot() ([]byte, error) {
	dataLeaves, err := e.getDataLeaves()
	if err != nil {
		return nil, err
	}
	return e.CoreDocument.CalculateSigningRoot(e.DocumentType(), dataLeaves)
}

// CalculateDocumentRoot calculates the document root.
func (e *EntityRelationship) CalculateDocumentRoot() ([]byte, error) {
	dataLeaves, err := e.getDataLeaves()
	if err != nil {
		return nil, err
	}
	return e.CoreDocument.CalculateDocumentRoot(e.DocumentType(), dataLeaves)
}

// CollaboratorCanUpdate checks that the identity attempting to update the document is the identity which owns the document.
func (e *EntityRelationship) CollaboratorCanUpdate(updated documents.Document, identity *types.AccountID) error {
	newEntityRelationship, ok := updated.(*EntityRelationship)
	if !ok {
		return errors.NewTypedError(documents.ErrDocumentInvalidType, errors.New("expecting an entity relationship but got %T", updated))
	}

	if !e.Data.OwnerIdentity.Equal(identity) || !newEntityRelationship.Data.OwnerIdentity.Equal(identity) {
		return documents.ErrIdentityNotOwner
	}
	return nil
}

// AddAttributes adds attributes to the EntityRelationship model.
func (e *EntityRelationship) AddAttributes(ca documents.CollaboratorsAccess, prepareNewVersion bool, attrs ...documents.Attribute) error {
	ncd, err := e.CoreDocument.AddAttributes(ca, prepareNewVersion, compactPrefix(), attrs...)
	if err != nil {
		return errors.NewTypedError(documents.ErrCDAttribute, err)
	}

	e.CoreDocument = ncd
	return nil
}

// DeleteAttribute deletes the attribute from the model.
func (e *EntityRelationship) DeleteAttribute(key documents.AttrKey, prepareNewVersion bool) error {
	ncd, err := e.CoreDocument.DeleteAttribute(key, prepareNewVersion, compactPrefix())
	if err != nil {
		return errors.NewTypedError(documents.ErrCDAttribute, err)
	}

	e.CoreDocument = ncd
	return nil
}

// GetData returns entity relationship data
func (e *EntityRelationship) GetData() interface{} {
	return e.Data
}

// Scheme returns the entity relationship scheme.
func (e *EntityRelationship) Scheme() string {
	return Scheme
}

// loadData unmarshals json blob to Data.
func loadData(data []byte, d *Data) error {
	err := json.Unmarshal(data, d)
	if err != nil {
		return err
	}

	return nil
}

// DeriveFromCreatePayload unpacks the entity relationship data from the Payload.
func (e *EntityRelationship) DeriveFromCreatePayload(ctx context.Context, payload documents.CreatePayload) error {
	var d Data
	if err := loadData(payload.Data, &d); err != nil {
		return errors.NewTypedError(ErrERInvalidData, err)
	}

	params := documents.AccessTokenParams{
		Grantee:            d.TargetIdentity.ToHexString(),
		DocumentIdentifier: d.EntityIdentifier.String(),
	}

	cd, err := documents.NewCoreDocumentWithAccessToken(ctx, compactPrefix(), params)
	if err != nil {
		return errors.NewTypedError(documents.ErrCDCreate, err)
	}

	e.CoreDocument = cd
	e.Data = d
	return nil
}

// DeriveFromUpdatePayload removes any access tokens assigned to target did
func (e *EntityRelationship) DeriveFromUpdatePayload(_ context.Context, payload documents.UpdatePayload) (documents.Document, error) {
	var d Data
	if err := loadData(payload.Data, &d); err != nil {
		return nil, errors.NewTypedError(ErrERInvalidData, err)
	}

	ne := new(EntityRelationship)
	err := ne.revokeRelationship(e, d.TargetIdentity)
	if err != nil {
		return nil, err
	}

	return ne, nil
}

// DeriveFromClonePayload clones a new document.
func (e *EntityRelationship) DeriveFromClonePayload(_ context.Context, doc documents.Document) error {
	cd, err := doc.PackCoreDocument()
	if err != nil {
		return errors.NewTypedError(documents.ErrDocumentPackingCoreDocument, err)
	}

	e.CoreDocument, err = documents.NewClonedDocument(cd)
	if err != nil {
		return errors.NewTypedError(documents.ErrCDClone, err)
	}

	return nil
}

// Patch merges payload data into Document.
func (e *EntityRelationship) Patch(payload documents.UpdatePayload) error {
	var d Data
	err := copier.Copy(&d, &e.Data)
	if err != nil {
		return err
	}

	if err := loadData(payload.Data, &d); err != nil {
		return errors.NewTypedError(ErrERInvalidData, err)
	}

	ncd, err := e.CoreDocument.Patch(compactPrefix(), payload.Collaborators, payload.Attributes)
	if err != nil {
		return errors.NewTypedError(documents.ErrDocumentPatch, err)
	}

	e.Data = d
	e.CoreDocument = ncd
	return nil
}

// revokeRelationship revokes a relationship by deleting the access token in the Entity
func (e *EntityRelationship) revokeRelationship(old *EntityRelationship, grantee *types.AccountID) error {
	e.Data = old.Data
	cd, err := old.DeleteAccessToken(grantee)
	if err != nil {
		return err
	}

	e.CoreDocument = cd
	return nil
}

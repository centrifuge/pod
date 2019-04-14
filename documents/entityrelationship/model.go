package entityrelationship

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/protobufs/gen/go/document"
	entitypb2 "github.com/centrifuge/go-centrifuge/protobufs/gen/go/entity"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/centrifuge/precise-proofs/proofs/proto"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
)

const prefix string = "entity_relationship"

// tree prefixes for specific documents use the second byte of a 4 byte slice by convention
func compactPrefix() []byte { return []byte{0, 4, 0, 0} }

// EntityRelationship implements the documents.Model and keeps track of entity-relationship related fields and state.
type EntityRelationship struct {
	*documents.CoreDocument

	// owner of the relationship
	OwnerIdentity *identity.DID
	// Entity identifier
	EntityIdentifier []byte
	// identity which will be granted access
	TargetIdentity *identity.DID
}

// getRelationshipData returns the entity relationship data from the entity relationship model
func (e *EntityRelationship) getRelationshipData() *entitypb2.RelationshipData {
	dids := identity.DIDsToStrings(e.OwnerIdentity, e.TargetIdentity)
	eID := hexutil.Encode(e.EntityIdentifier)
	return &entitypb2.RelationshipData{
		OwnerIdentity:    dids[0],
		TargetIdentity:   dids[1],
		EntityIdentifier: eID,
	}
}

// createP2PProtobuf returns Centrifuge protobuf-specific RelationshipData.
func (e *EntityRelationship) createP2PProtobuf() *entitypb.EntityRelationship {
	dids := identity.DIDsToBytes(e.OwnerIdentity, e.TargetIdentity)
	return &entitypb.EntityRelationship{
		OwnerIdentity:    dids[0],
		TargetIdentity:   dids[1],
		EntityIdentifier: e.EntityIdentifier,
	}
}

// InitEntityRelationshipInput initialize the model based on the received parameters from the rest api call
func (e *EntityRelationship) InitEntityRelationshipInput(ctx context.Context, entityID string, data *entitypb2.RelationshipData) error {
	if err := e.initEntityRelationshipFromData(data); err != nil {
		return err
	}

	params := documentpb.AccessTokenParams{
		Grantee:            data.TargetIdentity,
		DocumentIdentifier: entityID,
	}

	cd, err := documents.NewCoreDocumentWithAccessToken(ctx, compactPrefix(), params)
	if err != nil {
		return errors.New("failed to init core document: %v", err)
	}

	e.CoreDocument = cd
	return nil
}

// PrepareNewVersion prepares new version from the old entity.
func (e *EntityRelationship) PrepareNewVersion(old documents.Model, data *entitypb2.RelationshipData, collaborators []string) error {
	err := e.initEntityRelationshipFromData(data)
	if err != nil {
		return err
	}

	oldCD := old.(*EntityRelationship).CoreDocument
	e.CoreDocument, err = oldCD.PrepareNewVersion(compactPrefix(), documents.CollaboratorsAccess{})
	if err != nil {
		return err
	}

	return nil
}

// initEntityRelationshipFromData initialises an EntityRelationship from RelationshipData.
func (e *EntityRelationship) initEntityRelationshipFromData(data *entitypb2.RelationshipData) error {
	dids, err := identity.StringsToDIDs(data.OwnerIdentity, data.TargetIdentity)
	if err != nil {
		return err
	}
	eID, err := hexutil.Decode(data.EntityIdentifier)
	if err != nil {
		return err
	}
	e.OwnerIdentity = dids[0]
	e.TargetIdentity = dids[1]
	e.EntityIdentifier = eID
	return nil
}

// loadFromP2PProtobuf loads the Entity Relationship from Centrifuge protobuf.
func (e *EntityRelationship) loadFromP2PProtobuf(entityRelationship *entitypb.EntityRelationship) error {
	dids, err := identity.BytesToDIDs(entityRelationship.OwnerIdentity, entityRelationship.TargetIdentity)
	if err != nil {
		return err
	}
	e.OwnerIdentity = dids[0]
	e.TargetIdentity = dids[1]
	e.EntityIdentifier = entityRelationship.EntityIdentifier
	return nil
}

// PackCoreDocument packs the EntityRelationship into a CoreDocument.
func (e *EntityRelationship) PackCoreDocument() (cd coredocumentpb.CoreDocument, err error) {
	entityRelationship := e.createP2PProtobuf()
	data, err := proto.Marshal(entityRelationship)
	if err != nil {
		return cd, errors.New("couldn't serialise EntityData: %v", err)
	}

	embedData := &any.Any{
		TypeUrl: e.DocumentType(),
		Value:   data,
	}

	return e.CoreDocument.PackCoreDocument(embedData), nil
}

// UnpackCoreDocument unpacks the core document into an EntityRelationship.
func (e *EntityRelationship) UnpackCoreDocument(cd coredocumentpb.CoreDocument) error {
	if cd.EmbeddedData == nil ||
		cd.EmbeddedData.TypeUrl != e.DocumentType() {
		return errors.New("trying to convert document with incorrect schema")
	}

	entityRelationship := new(entitypb.EntityRelationship)
	err := proto.Unmarshal(cd.EmbeddedData.Value, entityRelationship)
	if err != nil {
		return err
	}

	err = e.loadFromP2PProtobuf(entityRelationship)
	if err != nil {
		return err
	}
	e.CoreDocument = documents.NewCoreDocumentFromProtobuf(cd)
	return nil
}

// JSON marshals EntityRelationship into a json bytes
func (e *EntityRelationship) JSON() ([]byte, error) {
	return json.Marshal(e)
}

// FromJSON unmarshals the json bytes into EntityRelationship
func (e *EntityRelationship) FromJSON(jsonData []byte) error {
	return json.Unmarshal(jsonData, e)
}

// Type gives the EntityRelationship type.
func (e *EntityRelationship) Type() reflect.Type {
	return reflect.TypeOf(e)
}

// CalculateDataRoot calculates the data root.
func (e *EntityRelationship) CalculateDataRoot() ([]byte, error) {
	t, err := e.getDocumentDataTree()
	if err != nil {
		return nil, errors.New("failed to get data tree: %v", err)
	}

	dr := t.RootHash()
	return dr, nil
}

// getDocumentDataTree creates precise-proofs data tree for the model
func (e *EntityRelationship) getDocumentDataTree() (tree *proofs.DocumentTree, err error) {
	eProto := e.createP2PProtobuf()
	if err != nil {
		return nil, err
	}
	if e.CoreDocument == nil {
		return nil, errors.New("getDocumentDataTree error CoreDocument not set")
	}
	t := e.CoreDocument.DefaultTreeWithPrefix(prefix, compactPrefix())
	if err := t.AddLeavesFromDocument(eProto); err != nil {
		return nil, errors.New("getDocumentDataTree error %v", err)
	}
	if err := t.Generate(); err != nil {
		return nil, errors.New("getDocumentDataTree error %v", err)
	}

	return t, nil
}

// CreateNFTProofs is not implemented for EntityRelationship.
func (e *EntityRelationship) CreateNFTProofs(
	account identity.DID,
	registry common.Address,
	tokenID []byte,
	nftUniqueProof, readAccessProof bool) (proofs []*proofspb.Proof, err error) {
	return nil, documents.ErrNotImplemented
}

// CreateProofs generates proofs for given fields.
func (e *EntityRelationship) CreateProofs(fields []string) (proofs []*proofspb.Proof, err error) {
	tree, err := e.getDocumentDataTree()
	if err != nil {
		return nil, errors.New("createProofs error %v", err)
	}

	return e.CoreDocument.CreateProofs(e.DocumentType(), tree, fields)
}

// DocumentType returns the entity relationship document type.
func (*EntityRelationship) DocumentType() string {
	return documenttypes.EntityRelationshipDataTypeUrl
}

// AddNFT is not implemented for EntityRelationship
func (e *EntityRelationship) AddNFT(grantReadAccess bool, registry common.Address, tokenID []byte) error {
	return documents.ErrNotImplemented
}

// CalculateSigningRoot calculates the signing root of the document.
func (e *EntityRelationship) CalculateSigningRoot() ([]byte, error) {
	dr, err := e.CalculateDataRoot()
	if err != nil {
		return dr, err
	}
	return e.CoreDocument.CalculateSigningRoot(e.DocumentType(), dr)
}

// CalculateDocumentRoot calculates the document root.
func (e *EntityRelationship) CalculateDocumentRoot() ([]byte, error) {
	dr, err := e.CalculateDataRoot()
	if err != nil {
		return dr, err
	}
	return e.CoreDocument.CalculateDocumentRoot(e.DocumentType(), dr)
}

// DocumentRootTree creates and returns the document root tree.
func (e *EntityRelationship) DocumentRootTree() (tree *proofs.DocumentTree, err error) {
	dr, err := e.CalculateDataRoot()
	if err != nil {
		return nil, err
	}
	return e.CoreDocument.DocumentRootTree(e.DocumentType(), dr)
}

// CollaboratorCanUpdate checks that the identity attempting to update the document is the identity which owns the document.
func (e *EntityRelationship) CollaboratorCanUpdate(updated documents.Model, identity identity.DID) error {
	newEntityRelationship, ok := updated.(*EntityRelationship)
	if !ok {
		return errors.NewTypedError(documents.ErrDocumentInvalidType, errors.New("expecting an entity relationship but got %T", updated))
	}

	if !e.OwnerIdentity.Equal(identity) || !newEntityRelationship.OwnerIdentity.Equal(identity) {
		return documents.ErrIdentityNotOwner
	}
	return nil
}

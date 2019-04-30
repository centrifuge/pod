package entity

import (
	"encoding/json"
	"reflect"
	"strings"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/entity"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	cliententitypb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/entity"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/centrifuge/precise-proofs/proofs/proto"
	"github.com/ethereum/go-ethereum/common"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
)

const prefix string = "entity"

// tree prefixes for specific to documents use the second byte of a 4 byte slice by convention
func compactPrefix() []byte { return []byte{0, 3, 0, 0} }

// Entity implements the documents.Model keeps track of entity related fields and state
type Entity struct {
	*documents.CoreDocument

	Identity  *identity.DID
	LegalName string
	// address
	Addresses []*entitypb.Address
	// tax information
	PaymentDetails []*entitypb.PaymentDetail
	// Entity contact list
	Contacts []*entitypb.Contact
}

// getClientData returns the client data from the entity model
func (e *Entity) getClientData() (*cliententitypb.EntityData, error) {
	dids := identity.DIDsToStrings(e.Identity)
	attrs, err := documents.ToClientAttributes(e.Attributes)
	if err != nil {
		return nil, err
	}

	return &cliententitypb.EntityData{
		Identity:       dids[0],
		LegalName:      e.LegalName,
		Addresses:      e.Addresses,
		PaymentDetails: e.PaymentDetails,
		Contacts:       e.Contacts,
		Attributes:     attrs,
	}, nil
}

// createP2PProtobuf returns centrifuge protobuf specific entityData
func (e *Entity) createP2PProtobuf() *entitypb.Entity {
	dids := identity.DIDsToBytes(e.Identity)
	return &entitypb.Entity{
		Identity:       dids[0],
		LegalName:      e.LegalName,
		Addresses:      e.Addresses,
		PaymentDetails: e.PaymentDetails,
		Contacts:       e.Contacts,
	}
}

// InitEntityInput initialize the model based on the received parameters from the rest api call
func (e *Entity) InitEntityInput(payload *cliententitypb.EntityCreatePayload, self identity.DID) error {
	err := e.initEntityFromData(payload.Data)
	if err != nil {
		return err
	}

	ca, err := documents.FromClientCollaboratorAccess(payload.ReadAccess, payload.WriteAccess)
	if err != nil {
		return errors.New("failed to decode collaborator: %v", err)
	}

	ca.ReadWriteCollaborators = append(ca.ReadWriteCollaborators, self)
	attrs, err := documents.FromClientAttributes(payload.Data.Attributes)
	if err != nil {
		return err
	}

	cd, err := documents.NewCoreDocument(compactPrefix(), ca, attrs)
	if err != nil {
		return errors.New("failed to init core document: %v", err)
	}

	e.CoreDocument = cd
	return nil
}

// initEntityFromData initialises entity from entityData
func (e *Entity) initEntityFromData(data *cliententitypb.EntityData) error {
	data.Identity = strings.TrimSpace(data.Identity)
	if data.Identity == "" {
		return identity.ErrMalformedAddress
	}

	dids, err := identity.StringsToDIDs(data.Identity)
	if err != nil {
		return errors.NewTypedError(identity.ErrMalformedAddress, err)
	}

	e.Identity = dids[0]
	e.LegalName = data.LegalName
	e.Addresses = data.Addresses
	e.PaymentDetails = data.PaymentDetails
	e.Contacts = data.Contacts
	return nil
}

// loadFromP2PProtobuf  loads the entity from centrifuge protobuf entity data
func (e *Entity) loadFromP2PProtobuf(data *entitypb.Entity) error {
	dids, err := identity.BytesToDIDs(data.Identity)
	if err != nil {
		return err
	}

	e.Identity = dids[0]
	e.LegalName = data.LegalName
	e.Addresses = data.Addresses
	e.PaymentDetails = data.PaymentDetails
	e.Contacts = data.Contacts
	return nil
}

// PackCoreDocument packs the Entity into a CoreDocument.
func (e *Entity) PackCoreDocument() (cd coredocumentpb.CoreDocument, err error) {
	entityData := e.createP2PProtobuf()
	data, err := proto.Marshal(entityData)
	if err != nil {
		return cd, errors.New("couldn't serialise EntityData: %v", err)
	}

	embedData := &any.Any{
		TypeUrl: e.DocumentType(),
		Value:   data,
	}

	return e.CoreDocument.PackCoreDocument(embedData), nil
}

// UnpackCoreDocument unpacks the core document into Entity.
func (e *Entity) UnpackCoreDocument(cd coredocumentpb.CoreDocument) error {
	if cd.EmbeddedData == nil ||
		cd.EmbeddedData.TypeUrl != e.DocumentType() {
		return errors.New("trying to convert document with incorrect schema")
	}

	entityData := new(entitypb.Entity)
	err := proto.Unmarshal(cd.EmbeddedData.Value, entityData)
	if err != nil {
		return err
	}

	err = e.loadFromP2PProtobuf(entityData)
	if err != nil {
		return err
	}

	e.CoreDocument = documents.NewCoreDocumentFromProtobuf(cd)
	return nil
}

// JSON marshals Entity into a json bytes
func (e *Entity) JSON() ([]byte, error) {
	return json.Marshal(e)
}

// FromJSON unmarshals the json bytes into Entity
func (e *Entity) FromJSON(jsonData []byte) error {
	return json.Unmarshal(jsonData, e)
}

// Type gives the Entity type
func (e *Entity) Type() reflect.Type {
	return reflect.TypeOf(e)
}

// CalculateDataRoot calculates the data root and sets the root to core document.
func (e *Entity) CalculateDataRoot() ([]byte, error) {
	t, err := e.getDocumentDataTree()
	if err != nil {
		return nil, errors.New("failed to get data tree: %v", err)
	}

	dr := t.RootHash()
	return dr, nil
}

// getDocumentDataTree creates precise-proofs data tree for the model
func (e *Entity) getDocumentDataTree() (tree *proofs.DocumentTree, err error) {
	eProto := e.createP2PProtobuf()
	if e.CoreDocument == nil {
		return nil, errors.New("getDocumentDataTree error CoreDocument not set")
	}
	t := e.CoreDocument.DefaultTreeWithPrefix(prefix, compactPrefix())
	err = t.AddLeavesFromDocument(eProto)
	if err != nil {
		return nil, errors.New("getDocumentDataTree error %v", err)
	}
	err = t.Generate()
	if err != nil {
		return nil, errors.New("getDocumentDataTree error %v", err)
	}

	return t, nil
}

// CreateNFTProofs creates proofs specific to NFT minting.
func (e *Entity) CreateNFTProofs(
	account identity.DID,
	registry common.Address,
	tokenID []byte,
	nftUniqueProof, readAccessProof bool) (proofs []*proofspb.Proof, err error) {

	tree, err := e.getDocumentDataTree()
	if err != nil {
		return nil, err
	}

	return e.CoreDocument.CreateNFTProofs(
		e.DocumentType(),
		tree,
		account, registry, tokenID, nftUniqueProof, readAccessProof)
}

// CreateProofs generates proofs for given fields.
func (e *Entity) CreateProofs(fields []string) (proofs []*proofspb.Proof, err error) {
	tree, err := e.getDocumentDataTree()
	if err != nil {
		return nil, errors.New("createProofs error %v", err)
	}

	return e.CoreDocument.CreateProofs(e.DocumentType(), tree, fields)
}

// DocumentType returns the entity document type.
func (*Entity) DocumentType() string {
	return documenttypes.EntityDataTypeUrl
}

// PrepareNewVersion prepares new version from the old entity.
func (e *Entity) PrepareNewVersion(old documents.Model, data *cliententitypb.EntityData, collaborators documents.CollaboratorsAccess) error {
	err := e.initEntityFromData(data)
	if err != nil {
		return err
	}

	attrs, err := documents.FromClientAttributes(data.Attributes)
	if err != nil {
		return err
	}

	oldCD := old.(*Entity).CoreDocument
	e.CoreDocument, err = oldCD.PrepareNewVersion(compactPrefix(), collaborators, attrs)
	if err != nil {
		return err
	}

	return nil
}

// AddNFT adds NFT to the Entity.
func (e *Entity) AddNFT(grantReadAccess bool, registry common.Address, tokenID []byte) error {
	cd, err := e.CoreDocument.AddNFT(grantReadAccess, registry, tokenID)
	if err != nil {
		return err
	}

	e.CoreDocument = cd
	return nil
}

// CalculateSigningRoot calculates the signing root of the document.
func (e *Entity) CalculateSigningRoot() ([]byte, error) {
	dr, err := e.CalculateDataRoot()
	if err != nil {
		return dr, err
	}
	return e.CoreDocument.CalculateSigningRoot(e.DocumentType(), dr)
}

// CalculateDocumentRoot calculates the document root
func (e *Entity) CalculateDocumentRoot() ([]byte, error) {
	dr, err := e.CalculateDataRoot()
	if err != nil {
		return dr, err
	}
	return e.CoreDocument.CalculateDocumentRoot(e.DocumentType(), dr)
}

// DocumentRootTree creates and returns the document root tree
func (e *Entity) DocumentRootTree() (tree *proofs.DocumentTree, err error) {
	dr, err := e.CalculateDataRoot()
	if err != nil {
		return nil, err
	}
	return e.CoreDocument.DocumentRootTree(e.DocumentType(), dr)
}

// CollaboratorCanUpdate checks if the collaborator can update the document.
func (e *Entity) CollaboratorCanUpdate(updated documents.Model, collaborator identity.DID) error {
	newEntity, ok := updated.(*Entity)
	if !ok {
		return errors.NewTypedError(documents.ErrDocumentInvalidType, errors.New("expecting an entity but got %T", updated))
	}

	// check the core document changes
	err := e.CoreDocument.CollaboratorCanUpdate(newEntity.CoreDocument, collaborator, e.DocumentType())
	if err != nil {
		return err
	}

	// check entity specific changes
	oldTree, err := e.getDocumentDataTree()
	if err != nil {
		return err
	}

	newTree, err := newEntity.getDocumentDataTree()
	if err != nil {
		return err
	}

	rules := e.CoreDocument.TransitionRulesFor(collaborator)
	cf := documents.GetChangedFields(oldTree, newTree)
	return documents.ValidateTransitions(rules, cf)
}

package entity

import (
	"encoding/json"
	"reflect"

	cliententitypb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/entity"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/entity"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/centrifuge/precise-proofs/proofs/proto"
	"github.com/ethereum/go-ethereum/common"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
)

const prefix string = "entity"

// tree prefixes for specific to documents use the second byte of a 4 byte slice by convention
func compactPrefix() []byte { return []byte{0, 1, 0, 0} }

// Entity implements the documents.Model keeps track of entity related fields and state
type Entity struct {
	*documents.CoreDocument

	Identity  []byte
	LegalName string
	// address
	Addresses []*entitypb.Address
	// tax information
	PaymentDetails []*entitypb.PaymentDetail
	// Entity contact list
	Contacts             []*entitypb.Contact

	EntitySalts *proofs.Salts
}


// getClientData returns the client data from the entity model
func (e *Entity) getClientData() *cliententitypb.EntityData {
	return &cliententitypb.EntityData{
		Identity: e.Identity,
		LegalName:e.LegalName,
		Addresses:e.Addresses,
		PaymentDetails:e.PaymentDetails,
		Contacts:e.Contacts,

	}

}

// createP2PProtobuf returns centrifuge protobuf specific entityData
func (e *Entity) createP2PProtobuf() *entitypb.Entity {
	return &entitypb.Entity{
	Identity: e.Identity,
	LegalName: e.LegalName,
	Addresses: e.Addresses,
	PaymentDetails: e.PaymentDetails,
	Contacts: e.Contacts,
	}
}



// InitEntityInput initialize the model based on the received parameters from the rest api call
func (e *Entity) InitEntityInput(payload *cliententitypb.EntityCreatePayload, self string) error {
	err := e.initEntityFromData(payload.Data)
	if err != nil {
		return err
	}

	collaborators := append([]string{self}, payload.Collaborators...)
	cd, err := documents.NewCoreDocumentWithCollaborators(collaborators, compactPrefix())
	if err != nil {
		return errors.New("failed to init core document: %v", err)
	}

	e.CoreDocument = cd
	return nil
}

// initEntityFromData initialises entity from entityData
func (e *Entity) initEntityFromData(data *cliententitypb.EntityData) error {
	e.Identity = data.Identity
	return nil
}


// loadFromP2PProtobuf  loads the entity from centrifuge protobuf entity data
func (e *Entity) loadFromP2PProtobuf(entityData *entitypb.Entity) {
 e.Identity = entityData.Identity
 e.LegalName = entityData.LegalName
 e.Addresses = entityData.Addresses
 e.PaymentDetails = entityData.PaymentDetails
 e.Contacts = entityData.Contacts
}

// getEntitySalts returns the entity salts. Initialises if not present
func (e *Entity) getEntitySalts(entityData *entitypb.Entity) (*proofs.Salts, error) {
	if e.EntitySalts == nil {
		entitySalts, err := documents.GenerateNewSalts(entityData, prefix, compactPrefix())
		if err != nil {
			return nil, errors.New("getEntitySalts error %v", err)
		}
		e.EntitySalts = entitySalts
	}

	return e.EntitySalts, nil
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

	salts, err := e.getEntitySalts(entityData)
	if err != nil {
		return cd, errors.New("couldn't get EntitySalts: %v", err)
	}

	return e.CoreDocument.PackCoreDocument(embedData, documents.ConvertToProtoSalts(salts)), nil
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

	e.loadFromP2PProtobuf(entityData)
	if cd.EmbeddedDataSalts == nil {
		e.EntitySalts, err = e.getEntitySalts(entityData)
		if err != nil {
			return err
		}
	} else {
		e.EntitySalts = documents.ConvertToProofSalts(cd.EmbeddedDataSalts)
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
	e.CoreDocument.SetDataRoot(dr)
	return dr, nil
}

// getDocumentDataTree creates precise-proofs data tree for the model
func (e *Entity) getDocumentDataTree() (tree *proofs.DocumentTree, err error) {
	entityProto := e.createP2PProtobuf()
	salts, err := e.getEntitySalts(entityProto)
	if err != nil {
		return nil, err
	}
	t := documents.NewDefaultTreeWithPrefix(salts, prefix, compactPrefix())
	err = t.AddLeavesFromDocument(entityProto)
	if err != nil {
		return nil, errors.New("getDocumentDataTree error %v", err)
	}
	err = t.Generate()
	if err != nil {
		return nil, errors.New("getDocumentDataTree error %v", err)
	}
	return t, nil
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
func (e *Entity) PrepareNewVersion(old documents.Model, data *cliententitypb.EntityData, collaborators []string) error {
	err := e.initEntityFromData(data)
	if err != nil {
		return err
	}

	oldCD := old.(*Entity).CoreDocument
	e.CoreDocument, err = oldCD.PrepareNewVersion(collaborators, true, compactPrefix())
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
	return e.CoreDocument.CalculateSigningRoot(e.DocumentType())
}

// CreateNFTProofs creates proofs specific to NFT minting.
func (e *Entity) CreateNFTProofs(
	account identity.DID,
	registry common.Address,
	tokenID []byte,
	nftUniqueProof, readAccessProof bool) (proofs []*proofspb.Proof, err error) {
	return e.CoreDocument.CreateNFTProofs(
		e.DocumentType(),
		account, registry, tokenID, nftUniqueProof, readAccessProof)
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
	cf := documents.GetChangedFields(oldTree, newTree, proofs.DefaultSaltsLengthSuffix)
	return documents.ValidateTransitions(rules, cf)
}

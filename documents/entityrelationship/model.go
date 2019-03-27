package entityrelationship
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

const prefix string = "entity_relationship"

// tree prefixes for specific to documents use the second byte of a 4 byte slice by convention
func compactPrefix() []byte { return []byte{0, 4, 0, 0} }

// Entity implements the documents.Model keeps track of entity related fields and state
type EntityRelationship struct {
	*documents.CoreDocument

	// owner of the relationship
	OwnerIdentity *identity.DID
	// document identifier
	Label []byte
	// identity which will be granted access
	TargetIdentity *identity.DID
}


// createP2PProtobuf returns centrifuge protobuf specific entityData
func (e *EntityRelationship) createP2PProtobuf() *entitypb.EntityRelationship {
	var didByte []byte
	var tidByte []byte
	if e.OwnerIdentity != nil {
		didByte = e.OwnerIdentity[:]
	}
	if e.TargetIdentity != nil {
		tidByte = e.TargetIdentity[:]
	}

	return &entitypb.EntityRelationship{
		OwnerIdentity:       didByte,
		Label:      e.Label,
		TargetIdentity: tidByte,
	}
}

// initEntityFromData initialises entity from entityData
func (e *EntityRelationship) initEntityRelationshipFromData(data *cliententitypb.EntityRelationshipData) error {
	if data.Identity != "" {
		if did, err := identity.NewDIDFromString(data.Identity); err == nil {
			e.Identity = &did
		} else {
			return err
		}
	} else {
		return identity.ErrMalformedAddress
	}
	e.LegalName = data.LegalName
	//e.Addresses = data.Addresses //TODO precise proofs not supporting fields yet
	e.PaymentDetails = data.PaymentDetails
	e.Contacts = data.Contacts

	return nil
}

// loadFromP2PProtobuf  loads the entity from centrifuge protobuf entity data
func (e *EntityRelationship) loadFromP2PProtobuf(entityRelationship *entitypb.EntityRelationship) error {
	if entityRelationship.OwnerIdentity != nil {
		did, err := identity.NewDIDFromBytes(entityRelationship.OwnerIdentity)
		if err != nil {
			return err
		}
		e.OwnerIdentity = &did
	}

	if entityRelationship.TargetIdentity != nil {
		tid, err := identity.NewDIDFromBytes(entityRelationship.TargetIdentity)
		if err != nil {
			return err
		}
		e.TargetIdentity = &tid
	}

	e.Label = entityRelationship.Label

	return nil
}

// PackCoreDocument packs the Entity into a CoreDocument.
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

// UnpackCoreDocument unpacks the core document into Entity.
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

	e.loadFromP2PProtobuf(entityRelationship)
	e.CoreDocument = documents.NewCoreDocumentFromProtobuf(cd)
	return nil
}

// JSON marshals Entity into a json bytes
func (e *EntityRelationship) JSON() ([]byte, error) {
	return json.Marshal(e)
}

// FromJSON unmarshals the json bytes into Entity
func (e *EntityRelationship) FromJSON(jsonData []byte) error {
	return json.Unmarshal(jsonData, e)
}

// Type gives the Entity type
func (e *EntityRelationship) Type() reflect.Type {
	return reflect.TypeOf(e)
}

// CalculateDataRoot calculates the data root and sets the root to core document.
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
func (e *EntityRelationship) CreateNFTProofs(
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
func (e *EntityRelationship) CreateProofs(fields []string) (proofs []*proofspb.Proof, err error) {
	tree, err := e.getDocumentDataTree()
	if err != nil {
		return nil, errors.New("createProofs error %v", err)
	}

	return e.CoreDocument.CreateProofs(e.DocumentType(), tree, fields)
}

// DocumentType returns the entity document type.
func (*EntityRelationship) DocumentType() string {
	return documenttypes.EntityRelationshipDocumentTypeUrl
}

// PrepareNewVersion prepares new version from the old entity.
func (e *EntityRelationship) PrepareNewVersion(old documents.Model, data *cliententitypb.EntityData, collaborators []string) error {
	err := e.initEntityRelationshipFromData(data)
	if err != nil {
		return err
	}

	oldCD := old.(*EntityRelationship).CoreDocument
	e.CoreDocument, err = oldCD.PrepareNewVersion(compactPrefix(), collaborators...)
	if err != nil {
		return err
	}

	return nil
}

// AddNFT adds NFT to the Entity.
func (e *EntityRelationship) AddNFT(grantReadAccess bool, registry common.Address, tokenID []byte) error {
	cd, err := e.CoreDocument.AddNFT(grantReadAccess, registry, tokenID)
	if err != nil {
		return err
	}

	e.CoreDocument = cd
	return nil
}

// CalculateSigningRoot calculates the signing root of the document.
func (e *EntityRelationship) CalculateSigningRoot() ([]byte, error) {
	dr, err := e.CalculateDataRoot()
	if err != nil {
		return dr, err
	}
	return e.CoreDocument.CalculateSigningRoot(e.DocumentType(), dr)
}

// CalculateDocumentRoot calculates the document root
func (e *EntityRelationship) CalculateDocumentRoot() ([]byte, error) {
	dr, err := e.CalculateDataRoot()
	if err != nil {
		return dr, err
	}
	return e.CoreDocument.CalculateDocumentRoot(e.DocumentType(), dr)
}

// DocumentRootTree creates and returns the document root tree
func (e *EntityRelationship) DocumentRootTree() (tree *proofs.DocumentTree, err error) {
	dr, err := e.CalculateDataRoot()
	if err != nil {
		return nil, err
	}
	return e.CoreDocument.DocumentRootTree(e.DocumentType(), dr)
}

// CollaboratorCanUpdate checks if the collaborator can update the document.
func (e *EntityRelationship) CollaboratorCanUpdate(updated documents.Model, collaborator identity.DID) error {
	newEntity, ok := updated.(*EntityRelationship)
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
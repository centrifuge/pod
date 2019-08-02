package generic

import (
	"encoding/json"
	"reflect"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	genericpb "github.com/centrifuge/centrifuge-protobufs/gen/go/generic"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/centrifuge/precise-proofs/proofs/proto"
	"github.com/ethereum/go-ethereum/common"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
)

const (
	prefix string = "generic"

	// Scheme to identify generic document
	Scheme = prefix
)

// tree prefixes for specific to documents use the second byte of a 4 byte slice by convention
func compactPrefix() []byte { return []byte{0, 5, 0, 0} }

// Data is a empty  structure.
type Data struct{}

// Generic implements the documents.Model for Generic documents
type Generic struct {
	*documents.CoreDocument
	Data Data
}

func getProtoGenericData() *genericpb.GenericData {
	return &genericpb.GenericData{
		Scheme: []byte(Scheme),
	}
}

// PackCoreDocument packs the Generic into a CoreDocument.
func (g *Generic) PackCoreDocument() (cd coredocumentpb.CoreDocument, err error) {
	data, err := proto.Marshal(getProtoGenericData())
	if err != nil {
		return cd, errors.New("couldn't serialise GenericData: %v", err)
	}

	embedData := &any.Any{
		TypeUrl: g.DocumentType(),
		Value:   data,
	}
	return g.CoreDocument.PackCoreDocument(embedData), nil
}

// UnpackCoreDocument unpacks the core document into Generic.
func (g *Generic) UnpackCoreDocument(cd coredocumentpb.CoreDocument) (err error) {
	if cd.EmbeddedData == nil ||
		cd.EmbeddedData.TypeUrl != g.DocumentType() {
		return errors.New("trying to convert document with incorrect schema")
	}

	g.Data = Data{}
	g.CoreDocument, err = documents.NewCoreDocumentFromProtobuf(cd)
	return err
}

// JSON marshals Generic into a json bytes
func (g *Generic) JSON() ([]byte, error) {
	return g.CoreDocument.MarshalJSON(g)
}

// FromJSON unmarshals the json bytes into Generic
func (g *Generic) FromJSON(jsonData []byte) error {
	if g.CoreDocument == nil {
		g.CoreDocument = new(documents.CoreDocument)
	}

	return g.CoreDocument.UnmarshalJSON(jsonData, g)
}

// Type gives the Generic type
func (g *Generic) Type() reflect.Type {
	return reflect.TypeOf(g)
}

// CalculateDataRoot calculates the data root and sets the root to core document.
func (g *Generic) CalculateDataRoot() ([]byte, error) {
	t, err := g.getDocumentDataTree()
	if err != nil {
		return nil, errors.New("failed to get data tree: %v", err)
	}

	return t.RootHash(), nil
}

// getDocumentDataTree creates precise-proofs data tree for the model
func (g *Generic) getDocumentDataTree() (tree *proofs.DocumentTree, err error) {
	if g.CoreDocument == nil {
		return nil, errors.New("getDocumentDataTree error CoreDocument not set")
	}

	t := g.CoreDocument.DefaultTreeWithPrefix(prefix, compactPrefix())
	err = t.AddLeavesFromDocument(getProtoGenericData())
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
func (g *Generic) CreateProofs(fields []string) (proofs []*proofspb.Proof, err error) {
	tree, err := g.getDocumentDataTree()
	if err != nil {
		return nil, errors.New("createProofs error %v", err)
	}

	return g.CoreDocument.CreateProofs(g.DocumentType(), tree, fields)
}

// DocumentType returns the generic document type.
func (*Generic) DocumentType() string {
	return documenttypes.GenericDataTypeUrl
}

// PrepareNewVersion prepares new version from the old generic.
func (g *Generic) PrepareNewVersion(old documents.Model, collaborators documents.CollaboratorsAccess, attrs map[documents.AttrKey]documents.Attribute) (err error) {
	oldCD := old.(*Generic).CoreDocument
	g.CoreDocument, err = oldCD.PrepareNewVersion(compactPrefix(), collaborators, attrs)
	if err != nil {
		return err
	}

	return nil
}

// AddNFT adds NFT to the Generic.
func (g *Generic) AddNFT(grantReadAccess bool, registry common.Address, tokenID []byte) error {
	cd, err := g.CoreDocument.AddNFT(grantReadAccess, registry, tokenID)
	if err != nil {
		return err
	}

	g.CoreDocument = cd
	return nil
}

// CalculateSigningRoot calculates the signing root of the document.
func (g *Generic) CalculateSigningRoot() ([]byte, error) {
	dr, err := g.CalculateDataRoot()
	if err != nil {
		return dr, err
	}
	return g.CoreDocument.CalculateSigningRoot(g.DocumentType(), dr)
}

// CalculateDocumentRoot calculates the document root
func (g *Generic) CalculateDocumentRoot() ([]byte, error) {
	dr, err := g.CalculateDataRoot()
	if err != nil {
		return dr, err
	}
	return g.CoreDocument.CalculateDocumentRoot(g.DocumentType(), dr)
}

// DocumentRootTree creates and returns the document root tree
func (g *Generic) DocumentRootTree() (tree *proofs.DocumentTree, err error) {
	dr, err := g.CalculateDataRoot()
	if err != nil {
		return nil, err
	}
	return g.CoreDocument.DocumentRootTree(g.DocumentType(), dr)
}

// CreateNFTProofs creates proofs specific to NFT minting.
func (g *Generic) CreateNFTProofs(
	account identity.DID,
	registry common.Address,
	tokenID []byte,
	nftUniqueProof, readAccessProof bool) (proofs []*proofspb.Proof, err error) {

	tree, err := g.getDocumentDataTree()
	if err != nil {
		return nil, err
	}

	return g.CoreDocument.CreateNFTProofs(
		g.DocumentType(),
		tree,
		account, registry, tokenID, nftUniqueProof, readAccessProof)
}

// CollaboratorCanUpdate checks if the collaborator can update the document.
func (g *Generic) CollaboratorCanUpdate(updated documents.Model, collaborator identity.DID) error {
	newGeneric, ok := updated.(*Generic)
	if !ok {
		return errors.NewTypedError(documents.ErrDocumentInvalidType, errors.New("expecting an generic but got %T", updated))
	}

	// check the core document changes
	err := g.CoreDocument.CollaboratorCanUpdate(newGeneric.CoreDocument, collaborator, g.DocumentType())
	if err != nil {
		return err
	}

	// check generic doc specific changes
	oldTree, err := g.getDocumentDataTree()
	if err != nil {
		return err
	}

	newTree, err := newGeneric.getDocumentDataTree()
	if err != nil {
		return err
	}

	rules := g.CoreDocument.TransitionRulesFor(collaborator)
	cf := documents.GetChangedFields(oldTree, newTree)
	return documents.ValidateTransitions(rules, cf)
}

// AddAttributes adds attributes to the Generic model.
func (g *Generic) AddAttributes(ca documents.CollaboratorsAccess, prepareNewVersion bool, attrs ...documents.Attribute) error {
	ncd, err := g.CoreDocument.AddAttributes(ca, prepareNewVersion, compactPrefix(), attrs...)
	if err != nil {
		return errors.NewTypedError(documents.ErrCDAttribute, err)
	}

	g.CoreDocument = ncd
	return nil
}

// DeleteAttribute deletes the attribute from the model.
func (g *Generic) DeleteAttribute(key documents.AttrKey, prepareNewVersion bool) error {
	ncd, err := g.CoreDocument.DeleteAttribute(key, prepareNewVersion, compactPrefix())
	if err != nil {
		return errors.NewTypedError(documents.ErrCDAttribute, err)
	}

	g.CoreDocument = ncd
	return nil
}

// GetData returns Generic Data.
func (g *Generic) GetData() interface{} {
	return g.Data
}

// loadData unmarshals json blob to Data.
func (g *Generic) loadData(data []byte) error {
	var d Data
	err := json.Unmarshal(data, &d)
	if err != nil {
		return err
	}

	g.Data = d
	return nil
}

// unpackFromCreatePayload unpacks the invoice data from the Payload.
func (g *Generic) unpackFromCreatePayload(did identity.DID, payload documents.CreatePayload) error {
	payload.Collaborators.ReadWriteCollaborators = append(payload.Collaborators.ReadWriteCollaborators, did)
	cd, err := documents.NewCoreDocument(compactPrefix(), payload.Collaborators, payload.Attributes)
	if err != nil {
		return errors.NewTypedError(documents.ErrCDCreate, err)
	}

	g.CoreDocument = cd
	return nil
}

// unpackFromUpdatePayload unpacks the update payload and prepares a new version.
func (g *Generic) unpackFromUpdatePayload(old *Generic, payload documents.UpdatePayload) error {
	ncd, err := old.CoreDocument.PrepareNewVersion(compactPrefix(), payload.Collaborators, payload.Attributes)
	if err != nil {
		return err
	}

	g.CoreDocument = ncd
	return nil
}

// Patch merges payload data into model
func (g *Generic) Patch(payload documents.UpdatePayload) error {
	return documents.ErrNotImplemented
}

// Scheme returns the invoice Scheme.
func (g *Generic) Scheme() string {
	return Scheme
}

package generic

import (
	"context"
	"reflect"

	"github.com/centrifuge/centrifuge-protobufs/documenttypes"
	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	genericpb "github.com/centrifuge/centrifuge-protobufs/gen/go/generic"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/precise-proofs/proofs"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

const (
	prefix string = "generic"

	// Scheme to identify generic document
	Scheme = prefix
)

// tree prefixes for specific to documents use the second byte of a 4 byte slice by convention
func compactPrefix() []byte { return []byte{0, 5, 0, 0} }

// Data is an empty structure.
type Data struct{}

// Generic implements the documents.Document for Generic documents
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
func (g *Generic) PackCoreDocument() (cd *coredocumentpb.CoreDocument, err error) {
	data, err := proto.Marshal(getProtoGenericData())
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentDataMarshalling, err)
	}

	embedData := &anypb.Any{
		TypeUrl: g.DocumentType(),
		Value:   data,
	}
	return g.CoreDocument.PackCoreDocument(embedData), nil
}

// UnpackCoreDocument unpacks the core document into Generic.
func (g *Generic) UnpackCoreDocument(cd *coredocumentpb.CoreDocument) (err error) {
	if cd.EmbeddedData == nil ||
		cd.EmbeddedData.TypeUrl != g.DocumentType() {
		return documents.ErrDocumentConvertInvalidSchema
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

func (g *Generic) getDataLeaves() ([]proofs.LeafNode, error) {
	t, err := g.getRawDataTree()
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDataTree, err)
	}
	return t.GetLeaves(), nil
}

func (g *Generic) getRawDataTree() (*proofs.DocumentTree, error) {
	if g.CoreDocument == nil {
		return nil, documents.ErrCoreDocumentNil
	}

	t, err := g.CoreDocument.DefaultTreeWithPrefix(prefix, compactPrefix())
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDataTree, err)
	}

	err = t.AddLeavesFromDocument(getProtoGenericData())
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDataTree, err)
	}
	return t, nil
}

// getDocumentDataTree creates precise-proofs data tree for the model
func (g *Generic) getDocumentDataTree() (tree *proofs.DocumentTree, err error) {
	if g.CoreDocument == nil {
		return nil, documents.ErrCoreDocumentNil
	}

	t, err := g.CoreDocument.DefaultTreeWithPrefix(prefix, compactPrefix())
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDataTree, err)
	}

	err = t.AddLeavesFromDocument(getProtoGenericData())
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDataTree, err)
	}

	err = t.Generate()
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDataTree, err)
	}

	return t, nil
}

// CreateProofs generates proofs for given fields.
func (g *Generic) CreateProofs(fields []string) (prf *documents.DocumentProof, err error) {
	dataLeaves, err := g.getDataLeaves()
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentProof, err)
	}

	return g.CoreDocument.CreateProofs(g.DocumentType(), dataLeaves, fields)
}

// DocumentType returns the generic document type.
func (*Generic) DocumentType() string {
	return documenttypes.GenericDataTypeUrl
}

// AddNFT adds NFT to the Generic.
func (g *Generic) AddNFT(grantReadAccess bool, collectionID types.U64, itemID types.U128) error {
	cd, err := g.CoreDocument.AddNFT(grantReadAccess, collectionID, itemID)
	if err != nil {
		return err
	}

	g.CoreDocument = cd
	return nil
}

// CalculateSigningRoot calculates the signing root of the document.
func (g *Generic) CalculateSigningRoot() ([]byte, error) {
	dataLeaves, err := g.getDataLeaves()
	if err != nil {
		return nil, err
	}
	return g.CoreDocument.CalculateSigningRoot(g.DocumentType(), dataLeaves)
}

// CalculateDocumentRoot calculates the document root
func (g *Generic) CalculateDocumentRoot() ([]byte, error) {
	dataLeaves, err := g.getDataLeaves()
	if err != nil {
		return nil, err
	}
	return g.CoreDocument.CalculateDocumentRoot(g.DocumentType(), dataLeaves)
}

// CollaboratorCanUpdate checks if the collaborator can update the document.
func (g *Generic) CollaboratorCanUpdate(updated documents.Document, collaborator *types.AccountID) error {
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

// DeriveFromCreatePayload unpacks the invoice data from the Payload.
func (g *Generic) DeriveFromCreatePayload(_ context.Context, payload documents.CreatePayload) error {
	cd, err := documents.NewCoreDocument(compactPrefix(), payload.Collaborators, payload.Attributes)
	if err != nil {
		return errors.NewTypedError(documents.ErrCDCreate, err)
	}

	g.CoreDocument = cd
	return nil
}

// DeriveFromClonePayload unpacks the generic data from the Payload
// This method clones the  transition rules and roles from a template document.
func (g *Generic) DeriveFromClonePayload(_ context.Context, m documents.Document) error {
	d, err := m.PackCoreDocument()
	if err != nil {
		return errors.NewTypedError(documents.ErrDocumentPackingCoreDocument, err)
	}

	cd, err := documents.NewClonedDocument(d)
	if err != nil {
		return errors.NewTypedError(documents.ErrCDClone, err)
	}

	g.CoreDocument = cd
	return nil
}

// DeriveFromUpdatePayload unpacks the update payload and prepares a new version.
func (g *Generic) DeriveFromUpdatePayload(_ context.Context, payload documents.UpdatePayload) (documents.Document, error) {
	ncd, err := g.CoreDocument.PrepareNewVersion(compactPrefix(), payload.Collaborators, payload.Attributes)
	if err != nil {
		return nil, err
	}

	return &Generic{
		CoreDocument: ncd,
	}, nil
}

// Patch merges payload data into model
func (g *Generic) Patch(payload documents.UpdatePayload) error {
	ncd, err := g.CoreDocument.Patch(compactPrefix(), payload.Collaborators, payload.Attributes)
	if err != nil {
		return errors.NewTypedError(documents.ErrDocumentPatch, err)
	}

	g.CoreDocument = ncd
	return nil
}

// Scheme returns the invoice Scheme.
func (g *Generic) Scheme() string {
	return Scheme
}

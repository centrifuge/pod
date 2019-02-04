package documents

import (
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/precise-proofs/proofs"
)

// Model is an interface to abstract away model specificness like invoice or purchaseOrder
// The interface can cast into the type specified by the model if required
// It should only handle protocol-level Document actions
// TODO: rename to something indicating the over the wire nature of this interface
type Model interface {
	storage.Model
	Document
	Permitter

	// Get the ID of the document represented by this model
	//ID() ([]byte, error)

	// PackCoreDocument packs the implementing document into a core document
	// should create the identifiers for the core document if not present
	PackCoreDocument() (*coredocumentpb.CoreDocument, error)

	// UnpackCoreDocument must return the document.Model
	// assumes that core document has valid identifiers set
	UnpackCoreDocument(cd *coredocumentpb.CoreDocument) error

	// CreateProofs creates precise-proofs for given fields
	//CreateProofs(fields []string) (coreDoc *coredocumentpb.CoreDocument, proofs []*proofspb.Proof, err error)
}


// The CoreDocument interface handles model-level Document interactions
// These encompass interactions which create, mutate, or read data from a Document model
type Document interface {
	New() *DocumentModel
	NewWithCollaborators(collaborators []string) (*DocumentModel, error)
	FillSalts(doc *DocumentModel) error
 	GetExternalCollaborators(selfCentID identity.CentID)  ([][]byte, error)
	GetTypeURL() (string, error)
}

type DocumentModel struct {
	document *coredocumentpb.CoreDocument
}

// The Permitter interface handles Document interactions around ACL functionality
type Permitter interface {

}


// New returns a new core document
// Note: collaborators and salts are to be filled by the caller
func (m *DocumentModel) New() *DocumentModel {
	id := utils.RandomSlice(32)
	cd := &coredocumentpb.CoreDocument{
		DocumentIdentifier: id,
		CurrentVersion:     id,
		NextVersion:        utils.RandomSlice(32),
	}
	return &DocumentModel{
		cd,
	}
}

// NewWithCollaborators generates new core document, adds collaborators, adds read rules and fills salts
func (m *DocumentModel) NewWithCollaborators(collaborators []string) (*DocumentModel, error) {
	model :=  m.New()
	ids, err := identity.CentIDsFromStrings(collaborators)
	if err != nil {
		return nil, errors.New("failed to decode collaborator: %v", err)
	}

	for i := range ids {
		model.document.Collaborators = append(model.document.Collaborators, ids[i][:])
	}

	err = initReadRules(model, ids)
	if err != nil {
		return nil, errors.New("failed to init read rules: %v", err)
	}

	err = m.FillSalts(model)
	if err != nil {
		return nil, err
	}

	return model, nil
}


// FillSalts creates a new coredocument.Salts and fills it
func (m *DocumentModel) FillSalts(model *DocumentModel) error {
	salts := new(coredocumentpb.CoreDocumentSalts)
	err := proofs.FillSalts(model.document, salts)
	if err != nil {
		return errors.New("failed to fill coredocument salts: %v", err)
	}

	model.document.CoredocumentSalts = salts
	return nil
}


// GetExternalCollaborators returns collaborators of a document without the own centID.
func (m *DocumentModel) GetExternalCollaborators(selfCentID identity.CentID)  ([][]byte, error) {
	var collabs [][]byte

	for _, collab := range m.document.Collaborators {
		collabID, err := identity.ToCentID(collab)
		if err != nil {
			return nil, errors.New("failed to convert to CentID: %v", err)
		}
		if !selfCentID.Equal(collabID) {
			collabs = append(collabs, collab)
		}
	}

	return collabs, nil
}


// GetTypeURL returns the type of the embedded document
func (m *DocumentModel) GetTypeURL() (string, error) {

	if m.document == nil {
		return "", errors.New("core document is nil")
	}

	if m.document.EmbeddedData == nil {
		return "", errors.New("core document doesn't have embedded data")
	}

	if m.document.EmbeddedData.TypeUrl == "" {
		return "", errors.New("typeUrl not set properly")
	}
	return m.document.EmbeddedData.TypeUrl, nil
}


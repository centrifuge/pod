package documents

import (
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/centrifuge/precise-proofs/proofs/proto"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// Model is an interface to abstract away model specificness like invoice or purchaseOrder
// The interface can cast into the type specified by the model if required
// It should only handle protocol-level Document actions
type Model interface {
	storage.Model

	// Get the ID of the document represented by this model
	ID() ([]byte, error)
	
	// PackCoreDocument packs the implementing document into a core document
	// should create the identifiers for the core document if not present
	PackCoreDocument() (*coredocumentpb.CoreDocument, error)

	// UnpackCoreDocument must return the document.Model
	// assumes that core document has valid identifiers set
	UnpackCoreDocument(cd *coredocumentpb.CoreDocument) error

	// CreateProofs creates precise-proofs for given fields
	CreateProofs(fields []string) (coreDoc *coredocumentpb.CoreDocument, proofs []*proofspb.Proof, err error)
}

// CoreDocumentModel contains methods which handle all interactions mutating or reading from a core document
// Access to a core document should always go through this model
type CoreDocumentModel struct {
	document *coredocumentpb.CoreDocument
}

// NewCoreDocModel returns a new CoreDocumentModel
// Note: collaborators and salts are to be filled by the caller
func NewCoreDocModel() *CoreDocumentModel {
	id := utils.RandomSlice(32)
	cd := &coredocumentpb.CoreDocument{
		DocumentIdentifier: id,
		CurrentVersion:     id,
		NextVersion:        utils.RandomSlice(32),
	}
	return &CoreDocumentModel{
		cd,
	}
}

// GetDocument returns the coredocument from the CoreDocumentModel
func (m *CoreDocumentModel) GetDocument() (*coredocumentpb.CoreDocument, error) {
	cd := m.document
	if cd == nil {
		return nil, errors.New("error getting document in CoreDocModel")
	}
	return m.document, nil
}

// SetDocument set the coredocument in the CoreDocumentModel to an updated coredocument
func (m *CoreDocumentModel) SetDocument(ncd *coredocumentpb.CoreDocument) *CoreDocumentModel {
	m.document = ncd
	return m
}

// PrepareNewVersion creates a new CoreDocumentModel with the version fields updated
// Adds collaborators and fills salts
// Note: new collaborators are added to the list with old collaborators.
//TODO: this will change when collaborators are moved down to next level
func (m *CoreDocumentModel) PrepareNewVersion(collaborators []string) (*CoreDocumentModel, error) {
	ndm := NewCoreDocModel()
	ncd, err := ndm.GetDocument()
	if err != nil {
		return nil, err
	}
	ocd, err := m.GetDocument()
	if err != nil {
		return nil, err
	}
	ucs, err := fetchUniqueCollaborators(ocd.Collaborators, collaborators)
	if err != nil {
		return nil, errors.New("failed to decode collaborator: %v", err)
	}

	cs := ncd.Collaborators
	for _, c := range ucs {
		c := c
		cs = append(cs, c[:])
	}

	ncd.Collaborators = cs

	// copy read rules and roles
	ncd.Roles = m.document.Roles
	ncd.ReadRules = m.document.ReadRules
	addCollaboratorsToReadSignRules(ncd, ucs)

	err = ndm.fillSalts()
	if err != nil {
		return nil, err
	}

	cd, err := m.GetDocument()
	if err != nil {
		return nil, err
	}

	ncd.CurrentVersion = cd.NextVersion
	ncd.NextVersion = utils.RandomSlice(32)
	if cd.DocumentRoot == nil {
		return nil, errors.New("DocumentRoot is nil")
	}
	ncd.PreviousRoot = cd.DocumentRoot
	dm := ndm.SetDocument(ncd)

	return dm, nil
}

// FillSalts creates a new coredocument.Salts and fills it
func (m *CoreDocumentModel) fillSalts() error {
	salts := new(coredocumentpb.CoreDocumentSalts)
	cd, err := m.GetDocument()
	if err != nil {
		return err
	}
	err = proofs.FillSalts(cd, salts)
	if err != nil {
		return errors.New("failed to fill coredocument salts: %v", err)
	}

	cd.CoredocumentSalts = salts
	return nil
}

func fetchUniqueCollaborators(oldCollabs [][]byte, newCollabs []string) (ids []identity.CentID, err error) {
	ocsm := make(map[string]struct{})
	for _, c := range oldCollabs {
		ocsm[hexutil.Encode(c)] = struct{}{}
	}

	var uc []string
	for _, c := range newCollabs {
		if _, ok := ocsm[c]; ok {
			continue
		}

		uc = append(uc, c)
	}

	for _, c := range uc {
		id, err := identity.CentIDFromString(c)
		if err != nil {
			return nil, err
		}

		ids = append(ids, id)
	}

	return ids, nil
}

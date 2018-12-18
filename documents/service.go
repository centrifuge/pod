package documents

import (
	"bytes"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/header"
	"github.com/centrifuge/precise-proofs/proofs/proto"
)

// Config specified configs required by documents package
type Config interface {

	// GetIdentityID retrieves the centID(TenentID) configured
	GetIdentityID() ([]byte, error)
}

// DocumentProof is a value to represent a document and its field proofs
type DocumentProof struct {
	DocumentID  []byte
	VersionID   []byte
	State       string
	FieldProofs []*proofspb.Proof
}

// Service provides an interface for functions common to all document types
type Service interface {

	// GetCurrentVersion reads a document from the database
	GetCurrentVersion(documentID []byte) (Model, error)

	// Exists checks if a document exists
	Exists(documentID []byte) bool

	// GetVersion reads a document from the database
	GetVersion(documentID []byte, version []byte) (Model, error)

	// DeriveFromCoreDocument derives a model given the core document
	DeriveFromCoreDocument(cd *coredocumentpb.CoreDocument) (Model, error)

	// CreateProofs creates proofs for the latest version document given the fields
	CreateProofs(documentID []byte, fields []string) (*DocumentProof, error)

	// CreateProofsForVersion creates proofs for a particular version of the document given the fields
	CreateProofsForVersion(documentID, version []byte, fields []string) (*DocumentProof, error)

	// RequestDocumentSignature Validates and Signs document received over the p2p layer
	RequestDocumentSignature(contextHeader *header.ContextHeader, model Model) (*coredocumentpb.Signature, error)

	// ReceiveAnchoredDocument receives a new anchored document over the p2p layer, validates and updates the document in DB
	ReceiveAnchoredDocument(model Model, headers *p2ppb.CentrifugeHeader) error
}

// service implements Service
type service struct {
	config           Config
	repo             Repository
}


// DefaultService returns the default implementation of the service
func DefaultService(config config.Configuration, repo Repository) Service {
	return service{repo:repo,config:config}
}

func getIds(model Model) ([]byte,[]byte, error) {
	cd, err := model.PackCoreDocument()

	if err != nil {
		return nil,nil, err
	}

	return cd.DocumentIdentifier, cd.NextVersion, nil
}

func (s service) searchVersion(m Model) (Model, error) {

	id, next, err := getIds(m)

	nm, err := s.getVersion(id, next)

	if err != nil {
		// here the err is returned as nil because it is expected that the nextVersion
		// is not available in the db at some stage of the iteration
		return m, nil
	}

	if next != nil {
		return s.searchVersion(nm)
	}

	return nm, nil
}

func (s service) GetCurrentVersion(documentID []byte) (Model, error) {
	model, err := s.getVersion(documentID, documentID)
	if err != nil {
		return nil, errors.NewTypedError(ErrDocumentNotFound, err)
	}
	return s.searchVersion(model)

}

func (s service) GetVersion(documentID []byte, version []byte) (Model, error) {
	return s.getVersion(documentID, version)
}


func (s service) CreateProofs(documentID []byte, fields []string) (*DocumentProof, error) {
	return nil, nil
}

func (s service) CreateProofsForVersion(documentID, version []byte, fields []string) (*DocumentProof, error) {
	return nil, nil
}

func (s service) DeriveFromCoreDocument(cd *coredocumentpb.CoreDocument) (Model, error) {
	return nil, nil
}

func (s service) RequestDocumentSignature(contextHeader *header.ContextHeader, model Model) (*coredocumentpb.Signature, error) {
	return nil, nil
}

func (s service) ReceiveAnchoredDocument(model Model, headers *p2ppb.CentrifugeHeader) error {
	return nil
}

func (s service) Exists(documentID []byte) bool {
	// get tenant ID
	tenantID, err := s.config.GetIdentityID()
	if err != nil {
		return false
	}
	return s.repo.Exists(tenantID, documentID)
}


func (s service) getVersion(documentID, version []byte) (Model, error) {
	// get tenant ID
	tenantID, err := s.config.GetIdentityID()
	if err != nil {
		return nil, errors.NewTypedError(ErrDocumentConfigTenantID, err)
	}
	model, err := s.repo.Get(tenantID, version)
	if err != nil {
		return nil, errors.NewTypedError(ErrDocumentVersionNotFound, err)
	}

	cd, err := model.PackCoreDocument()
	if err != nil {
		return nil, err
	}

	if !bytes.Equal(cd.DocumentIdentifier, documentID) {
		return nil, errors.NewTypedError(ErrDocumentVersionNotFound, errors.New("version is not valid for this identifier"))
	}
	return model, nil
}






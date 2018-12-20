package genericdoc

import (
	"bytes"

	"github.com/centrifuge/go-centrifuge/coredocument"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/header"
	"github.com/centrifuge/go-centrifuge/identity"
)

// service implements Service
type service struct {
	config           documents.Config
	repo             documents.Repository
	identityService  identity.Service
	anchorRepository anchors.AnchorRepository
}

// DefaultService returns the default implementation of the service
func DefaultService(config config.Configuration, repo documents.Repository,
	anchorRepo anchors.AnchorRepository, idService identity.Service) documents.Service {
	return service{repo: repo,
		config:           config,
		anchorRepository: anchorRepo,
		identityService:  idService}
}

func getIDs(model documents.Model) ([]byte, []byte, error) {
	cd, err := model.PackCoreDocument()

	if err != nil {
		return nil, nil, err
	}

	return cd.DocumentIdentifier, cd.NextVersion, nil
}

func (s service) searchVersion(m documents.Model) (documents.Model, error) {
	id, next, err := getIDs(m)

	if err != nil {
		return nil, err
	}

	if s.Exists(next) {
		nm, err := s.getVersion(id, next)
		if err != nil {

			return nil, err
		}
		return s.searchVersion(nm)
	}

	return m, nil

}

func (s service) GetCurrentVersion(documentID []byte) (documents.Model, error) {
	model, err := s.getVersion(documentID, documentID)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentNotFound, err)
	}
	return s.searchVersion(model)

}

func (s service) GetVersion(documentID []byte, version []byte) (documents.Model, error) {
	return s.getVersion(documentID, version)
}

func (s service) CreateProofs(documentID []byte, fields []string) (*documents.DocumentProof, error) {
	model, err := s.GetCurrentVersion(documentID)
	if err != nil {
		return nil, err
	}
	return s.createProofs(model, fields)

}

func (s service) createProofs(model documents.Model, fields []string) (*documents.DocumentProof, error) {
	if err := coredocument.PostAnchoredValidator(s.identityService, s.anchorRepository).Validate(nil, model); err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentInvalid, err)
	}
	coreDoc, proofs, err := model.CreateProofs(fields)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentProof, err)
	}
	return &documents.DocumentProof{
		DocumentID:  coreDoc.DocumentIdentifier,
		VersionID:   coreDoc.CurrentVersion,
		FieldProofs: proofs,
	}, nil

}

func (s service) CreateProofsForVersion(documentID, version []byte, fields []string) (*documents.DocumentProof, error) {
	model, err := s.getVersion(documentID, version)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentNotFound, err)
	}
	return s.createProofs(model, fields)
}

func (s service) DeriveFromCoreDocument(cd *coredocumentpb.CoreDocument) (documents.Model, error) {
	return nil, nil
}

func (s service) RequestDocumentSignature(contextHeader *header.ContextHeader, model documents.Model) (*coredocumentpb.Signature, error) {
	return nil, nil
}

func (s service) ReceiveAnchoredDocument(model documents.Model, headers *p2ppb.CentrifugeHeader) error {
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

func (s service) getVersion(documentID, version []byte) (documents.Model, error) {
	// get tenant ID
	tenantID, err := s.config.GetIdentityID()
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentConfigTenantID, err)
	}
	model, err := s.repo.Get(tenantID, version)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentVersionNotFound, err)
	}

	cd, err := model.PackCoreDocument()
	if err != nil {
		return nil, err
	}

	if !bytes.Equal(cd.DocumentIdentifier, documentID) {
		return nil, errors.NewTypedError(documents.ErrDocumentVersionNotFound, errors.New("version is not valid for this identifier"))
	}
	return model, nil
}

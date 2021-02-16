package generic

import (
	"context"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
)

// service implements Service and handles all entity related persistence and validations
// service always returns errors of type `errors.Error` or `errors.TypedError`
type service struct {
	documents.Service
	repo      documents.Repository
	anchorSrv anchors.Service
}

// DefaultService returns the default implementation of the service.
func DefaultService(
	srv documents.Service,
	repo documents.Repository,
	anchorSrv anchors.Service,
) documents.Service {
	return service{
		repo:      repo,
		Service:   srv,
		anchorSrv: anchorSrv,
	}
}

// DeriveFromCoreDocument takes a core document model and returns an Generic Doc
func (s service) DeriveFromCoreDocument(cd coredocumentpb.CoreDocument) (documents.Document, error) {
	g := new(Generic)
	err := g.UnpackCoreDocument(cd)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentUnPackingCoreDocument, err)
	}

	return g, nil
}

// New returns a new uninitialised Generic document.
func (s service) New(_ string) (documents.Document, error) {
	return new(Generic), nil
}

// Validate takes care of document validation
func (s service) Validate(ctx context.Context, model documents.Document, old documents.Document) error {
	return nil
}

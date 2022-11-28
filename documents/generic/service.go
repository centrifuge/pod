package generic

import (
	"context"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
)

// service implements Service and handles all entity related persistence and validations
// service always returns errors of type `errors.Error` or `errors.TypedError`
type service struct {
	documents.Service
}

// NewService returns the default implementation of the service.
func NewService(
	srv documents.Service,
) documents.Service {
	return service{
		Service: srv,
	}
}

// DeriveFromCoreDocument takes a core document model and returns an Generic Doc
func (s service) DeriveFromCoreDocument(cd *coredocumentpb.CoreDocument) (documents.Document, error) {
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
func (s service) Validate(_ context.Context, _ documents.Document, _ documents.Document) error {
	return nil
}

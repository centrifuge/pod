package v2

import (
	"context"

	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/pending"
)

// Service is the entry point for all the V2 APIs.
type Service struct {
	pendingDocSrv pending.Service
	tokenRegistry documents.TokenRegistry
}

// CreateDocument creates a pending document from the given payload.
// if the document_id is provided, next version of the document is created.
func (s Service) CreateDocument(ctx context.Context, req documents.UpdatePayload) (documents.Model, error) {
	return s.pendingDocSrv.Create(ctx, req)
}

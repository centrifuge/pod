package coreapi

import (
	"context"

	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
)

// Service provides functionality for Core APIs.
type Service struct {
	docService  documents.Service
	jobsService jobs.Manager
}

// CreateDocument creates the document from the payload and anchors it.
func (s Service) CreateDocument(ctx context.Context, payload documents.CreatePayload) (documents.Model, jobs.JobID, error) {
	return s.docService.CreateModel(ctx, payload)
}

// UpdateDocument updates the document from the payload and anchors the next version.
func (s Service) UpdateDocument(ctx context.Context, payload documents.UpdatePayload) (documents.Model, jobs.JobID, error) {
	return s.docService.UpdateModel(ctx, payload)
}

// GetJobStatus returns the job status.
func (s Service) GetJobStatus(account identity.DID, id jobs.JobID) (jobs.StatusResponse, error) {
	return s.jobsService.GetJobStatus(account, id)
}

// GetDocument returns the latest version of the document.
func (s Service) GetDocument(ctx context.Context, docID []byte) (documents.Model, error) {
	return s.docService.GetCurrentVersion(ctx, docID)
}

// GetDocumentVersion returns the specific version of the document
func (s Service) GetDocumentVersion(ctx context.Context, docID, versionID []byte) (documents.Model, error) {
	return s.docService.GetVersion(ctx, docID, versionID)
}

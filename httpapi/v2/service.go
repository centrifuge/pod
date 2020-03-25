package v2

import (
	"context"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
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

// UpdateDocument updates a pending document with the given payload
func (s Service) UpdateDocument(ctx context.Context, req documents.UpdatePayload) (documents.Model, error) {
	return s.pendingDocSrv.Update(ctx, req)
}

// Commit creates a document out of a pending document.
func (s Service) Commit(ctx context.Context, docID []byte) (documents.Model, jobs.JobID, error) {
	return s.pendingDocSrv.Commit(ctx, docID)
}

// GetDocument returns the document associated with docID and status.
func (s Service) GetDocument(ctx context.Context, docID []byte, status documents.Status) (documents.Model, error) {
	return s.pendingDocSrv.Get(ctx, docID, status)
}

// GetDocumentVersion returns the specific version of the document.
func (s Service) GetDocumentVersion(ctx context.Context, docID, versionID []byte) (documents.Model, error) {
	return s.pendingDocSrv.GetVersion(ctx, docID, versionID)
}

// AddSignedAttribute signs the payload with acc signing key and add it the document associated with docID.
func (s Service) AddSignedAttribute(ctx context.Context, docID []byte, label string, payload []byte) (documents.Model, error) {
	return s.pendingDocSrv.AddSignedAttribute(ctx, docID, label, payload)
}

// RemoveCollaborators removes collaborators from the document.
func (s Service) RemoveCollaborators(ctx context.Context, docID []byte, dids []identity.DID) (documents.Model, error) {
	return s.pendingDocSrv.RemoveCollaborators(ctx, docID, dids)
}

// AddRole adds a new role to the document
func (s Service) AddRole(ctx context.Context, docID []byte, roleKey string, dids []identity.DID) (*coredocumentpb.Role, error) {
	return s.pendingDocSrv.AddRole(ctx, docID, roleKey, dids)
}

// GetRole gets the role from the document
func (s Service) GetRole(ctx context.Context, docID, roleID []byte) (*coredocumentpb.Role, error) {
	return s.pendingDocSrv.GetRole(ctx, docID, roleID)
}

// UpdateRole updates the role in the document
func (s Service) UpdateRole(ctx context.Context, docID, roleID []byte, dids []identity.DID) (*coredocumentpb.Role, error) {
	return s.pendingDocSrv.UpdateRole(ctx, docID, roleID, dids)
}

// AddTransitionRules adds new rules to the document
func (s Service) AddTransitionRules(
	ctx context.Context, docID []byte, addRules pending.AddTransitionRules) ([]*coredocumentpb.TransitionRule, error) {
	return s.pendingDocSrv.AddTransitionRules(ctx, docID, addRules)
}

// GetTransitionRule returns the transition rule associated with ruleID in the document.
func (s Service) GetTransitionRule(ctx context.Context, docID, ruleID []byte) (*coredocumentpb.TransitionRule, error) {
	return s.pendingDocSrv.GetTransitionRule(ctx, docID, ruleID)
}

// DeleteTransitionRule deletes the transition rule associated with ruleID from the document.
func (s Service) DeleteTransitionRule(ctx context.Context, docID, ruleID []byte) error {
	return s.pendingDocSrv.DeleteTransitionRule(ctx, docID, ruleID)
}

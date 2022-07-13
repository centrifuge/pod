package v2

import (
	"context"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/entity"
	"github.com/centrifuge/go-centrifuge/documents/entityrelationship"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/pending"
	"github.com/centrifuge/go-centrifuge/utils/byteutils"
	"github.com/centrifuge/gocelery/v2"
)

// Service is the entry point for all the V2 APIs.
type Service struct {
	pendingDocSrv pending.Service
	dispatcher    jobs.Dispatcher
	accountSrv    config.Service
	entitySrv     entity.Service
	erSrv         entityrelationship.Service
	docSrv        documents.Service
}

// CreateDocument creates a pending document from the given payload.
// if the document_id is provided, next version of the document is created.
func (s Service) CreateDocument(ctx context.Context, req documents.UpdatePayload) (documents.Document, error) {
	return s.pendingDocSrv.Create(ctx, req)
}

// CloneDocument creates a new cloned document from the template (docID specified in payload).
func (s Service) CloneDocument(ctx context.Context, payload documents.ClonePayload) (documents.Document, error) {
	return s.pendingDocSrv.Clone(ctx, payload)
}

// UpdateDocument updates a pending document with the given payload
func (s Service) UpdateDocument(ctx context.Context, req documents.UpdatePayload) (documents.Document, error) {
	return s.pendingDocSrv.Update(ctx, req)
}

// Commit creates a document out of a pending document.
func (s Service) Commit(ctx context.Context, docID []byte) (documents.Document, gocelery.JobID, error) {
	return s.pendingDocSrv.Commit(ctx, docID)
}

// GetDocument returns the document associated with docID and status.
func (s Service) GetDocument(ctx context.Context, docID []byte, status documents.Status) (documents.Document, error) {
	return s.pendingDocSrv.Get(ctx, docID, status)
}

// GetDocumentVersion returns the specific version of the document.
func (s Service) GetDocumentVersion(ctx context.Context, docID, versionID []byte) (documents.Document, error) {
	return s.pendingDocSrv.GetVersion(ctx, docID, versionID)
}

// AddSignedAttribute signs the payload with acc signing key and add it the document associated with docID.
func (s Service) AddSignedAttribute(ctx context.Context, docID []byte, label string, payload []byte, valType documents.AttributeType) (documents.Document, error) {
	return s.pendingDocSrv.AddSignedAttribute(ctx, docID, label, payload, valType)
}

// RemoveCollaborators removes collaborators from the document.
func (s Service) RemoveCollaborators(ctx context.Context, docID []byte, dids []identity.DID) (documents.Document, error) {
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

// AddAttributes add attributes to pending document
func (s Service) AddAttributes(ctx context.Context, docID []byte, attrs []documents.Attribute) (documents.Document, error) {
	return s.pendingDocSrv.AddAttributes(ctx, docID, attrs)
}

// DeleteAttribute deletes attribute on a pending document
func (s Service) DeleteAttribute(ctx context.Context, docID []byte, key documents.AttrKey) (documents.Document, error) {
	return s.pendingDocSrv.DeleteAttribute(ctx, docID, key)
}

// Job returns the job details
func (s Service) Job(accID identity.DID, jobID []byte) (*gocelery.Job, error) {
	return s.dispatcher.Job(accID, jobID)
}

// GenerateAccount generates a new account
func (s Service) GenerateAccount(acc config.CentChainAccount) (did, jobID byteutils.HexBytes, err error) {
	return s.accountSrv.GenerateAccountAsync(acc)
}

// SignPayload uses the accountID's secret key to sign the payload and returns the signature
func (s Service) SignPayload(accountID, payload []byte) (*coredocumentpb.Signature, error) {
	return s.accountSrv.Sign(accountID, payload)
}

// GetEntityByRelationship returns an entity through a relationship ID.
func (s Service) GetEntityByRelationship(ctx context.Context, docID []byte) (documents.Document, error) {
	return s.entitySrv.GetEntityByRelationship(ctx, docID)
}

// GetEntityRelationShips returns the entity relationships under the given entity
func (s Service) GetEntityRelationShips(ctx context.Context, entityID []byte) ([]documents.Document, error) {
	return s.erSrv.GetEntityRelationships(ctx, entityID)
}

// GetAccount returns the Account associated with accountID
func (s Service) GetAccount(accountID []byte) (config.Account, error) {
	return s.accountSrv.GetAccount(accountID)
}

// GetAccounts returns all the accounts.
func (s Service) GetAccounts() ([]config.Account, error) {
	return s.accountSrv.GetAccounts()
}

// GenerateProofs returns the proofs for the latest version of the document.
func (s Service) GenerateProofs(ctx context.Context, docID []byte, fields []string) (*documents.DocumentProof, error) {
	return s.docSrv.CreateProofs(ctx, docID, fields)
}

// GenerateProofsForVersion returns the proofs for the specific version of the document.
func (s Service) GenerateProofsForVersion(ctx context.Context, docID, versionID []byte, fields []string) (*documents.DocumentProof, error) {
	return s.docSrv.CreateProofsForVersion(ctx, docID, versionID, fields)
}

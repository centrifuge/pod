package userapi

import (
	"context"
	"fmt"

	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/entity"
	"github.com/centrifuge/go-centrifuge/documents/entityrelationship"
	"github.com/centrifuge/go-centrifuge/extensions/funding"
	"github.com/centrifuge/go-centrifuge/extensions/transferdetails"
	"github.com/centrifuge/go-centrifuge/httpapi/coreapi"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// Service provides functionality for User APIs.
type Service struct {
	coreAPISrv             coreapi.Service
	transferDetailsService transferdetails.Service
	entityRelationshipSrv  entityrelationship.Service
	entitySrv              entity.Service
	fundingSrv             funding.Service
	config                 config.Service
}

// TODO: this can be refactored into a generic Service which handles all kinds of custom attributes

// CreateTransferDetail creates and anchors a Transfer Detail
func (s Service) CreateTransferDetail(ctx context.Context, req transferdetails.CreateTransferDetailRequest) (documents.Model, jobs.JobID, error) {
	return s.transferDetailsService.CreateTransferDetail(ctx, req)
}

// UpdateTransferDetail updates and anchors a Transfer Detail
func (s Service) UpdateTransferDetail(ctx context.Context, req transferdetails.UpdateTransferDetailRequest) (documents.Model, jobs.JobID, error) {
	return s.transferDetailsService.UpdateTransferDetail(ctx, req)
}

// GetCurrentTransferDetail returns the current version on a Transfer Detail
func (s Service) GetCurrentTransferDetail(ctx context.Context, docID, transferID []byte) (*transferdetails.TransferDetail, documents.Model, error) {
	model, err := s.coreAPISrv.GetDocument(ctx, docID)
	if err != nil {
		return nil, nil, err
	}
	data, model, err := s.transferDetailsService.DeriveTransferDetail(ctx, model, transferID)
	if err != nil {
		return nil, nil, err
	}

	return data, model, nil
}

// GetCurrentTransferDetailsList returns a list of Transfer Details on the current version of a document
func (s Service) GetCurrentTransferDetailsList(ctx context.Context, docID []byte) (*transferdetails.TransferDetailList, documents.Model, error) {
	model, err := s.coreAPISrv.GetDocument(ctx, docID)
	if err != nil {
		return nil, nil, err
	}

	data, model, err := s.transferDetailsService.DeriveTransferList(ctx, model)
	if err != nil {
		return nil, nil, err
	}

	return data, model, nil
}

// GetVersionTransferDetail returns a Transfer Detail on a particular version of a Document
func (s Service) GetVersionTransferDetail(ctx context.Context, docID, versionID, transferID []byte) (*transferdetails.TransferDetail, documents.Model, error) {
	model, err := s.coreAPISrv.GetDocumentVersion(ctx, docID, versionID)
	if err != nil {
		return nil, nil, err
	}

	data, model, err := s.transferDetailsService.DeriveTransferDetail(ctx, model, transferID)
	if err != nil {
		return nil, nil, err
	}

	return data, model, nil
}

// GetVersionTransferDetailsList returns a list of Transfer Details on a particular version of a Document
func (s Service) GetVersionTransferDetailsList(ctx context.Context, docID, versionID []byte) (*transferdetails.TransferDetailList, documents.Model, error) {
	model, err := s.coreAPISrv.GetDocumentVersion(ctx, docID, versionID)
	if err != nil {
		return nil, nil, err
	}

	data, model, err := s.transferDetailsService.DeriveTransferList(ctx, model)
	if err != nil {
		return nil, nil, err
	}

	return data, model, nil
}

func convertEntityRequest(req CreateEntityRequest) (documents.CreatePayload, error) {
	coreAPIReq := coreapi.CreateDocumentRequest{
		Scheme:      entity.Scheme,
		WriteAccess: req.WriteAccess,
		ReadAccess:  req.ReadAccess,
		Data:        req.Data,
		Attributes:  req.Attributes,
	}

	return coreapi.ToDocumentsCreatePayload(coreAPIReq)
}

// CreateEntity creates Entity document and anchors it.
func (s Service) CreateEntity(ctx context.Context, req CreateEntityRequest) (documents.Model, jobs.JobID, error) {
	docReq, err := convertEntityRequest(req)
	if err != nil {
		return nil, jobs.NilJobID(), err
	}

	return s.coreAPISrv.CreateDocument(ctx, docReq)
}

// UpdateEntity updates existing entity associated with docID  with provided data and anchors it.
func (s Service) UpdateEntity(ctx context.Context, docID []byte, req CreateEntityRequest) (documents.Model, jobs.JobID, error) {
	docReq, err := convertEntityRequest(req)
	if err != nil {
		return nil, jobs.NilJobID(), err
	}

	return s.coreAPISrv.UpdateDocument(ctx, documents.UpdatePayload{
		DocumentID:    docID,
		CreatePayload: docReq,
	})
}

// GetEntity returns the Entity associated with docID.
func (s Service) GetEntity(ctx context.Context, docID []byte) (documents.Model, error) {
	return s.coreAPISrv.GetDocument(ctx, docID)
}

// ShareEntity shares an entity relationship document with target identity.
func (s Service) ShareEntity(ctx context.Context, docID []byte, req ShareEntityRequest) (documents.Model, jobs.JobID, error) {
	r, err := convertShareEntityRequest(ctx, docID, req.TargetIdentity)
	if err != nil {
		return nil, jobs.NilJobID(), err
	}

	return s.coreAPISrv.CreateDocument(ctx, r)
}

// RevokeRelationship revokes target_identity's access to entity.
func (s Service) RevokeRelationship(ctx context.Context, docID []byte, req ShareEntityRequest) (documents.Model, jobs.JobID, error) {
	r, err := convertShareEntityRequest(ctx, docID, req.TargetIdentity)
	if err != nil {
		return nil, jobs.NilJobID(), err
	}

	return s.coreAPISrv.UpdateDocument(ctx, documents.UpdatePayload{
		DocumentID:    docID,
		CreatePayload: r,
	})
}

// GetEntityByRelationship returns an entity through a relationship ID.
func (s Service) GetEntityByRelationship(ctx context.Context, docID []byte) (documents.Model, error) {
	return s.entitySrv.GetEntityByRelationship(ctx, docID)
}

// getRequiredInvoiceUnpaidProofFields returns required proof fields for an unpaid invoice mint
func getRequiredInvoiceUnpaidProofFields(ctx context.Context) ([]string, error) {
	var proofFields []string

	acc, err := contextutil.Account(ctx)
	if err != nil {
		return nil, err
	}
	accDIDBytes := acc.GetIdentityID()
	keys, err := acc.GetKeys()
	if err != nil {
		return nil, err
	}

	signingRoot := fmt.Sprintf("%s.%s", documents.DRTreePrefix, documents.SigningRootField)
	signerID := hexutil.Encode(append(accDIDBytes, keys[identity.KeyPurposeSigning.Name].PublicKey...))
	signatureSender := fmt.Sprintf("%s.signatures[%s]", documents.SignaturesTreePrefix, signerID)
	proofFields = []string{"invoice.gross_amount", "invoice.currency", "invoice.date_due", "invoice.sender", "invoice.status", signingRoot, signatureSender, documents.CDTreePrefix + ".next_version"}
	return proofFields, nil
}

// CreateFundingAgreement creates a new funding agreement on a document and anchors the document.
func (s Service) CreateFundingAgreement(ctx context.Context, docID []byte, data *funding.Data) (documents.Model, jobs.JobID, error) {
	return s.fundingSrv.CreateFundingAgreement(ctx, docID, data)
}

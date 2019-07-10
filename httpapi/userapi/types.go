// nolint
package userapi

import (
	"context"

	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/entity"
	"github.com/centrifuge/go-centrifuge/documents/entityrelationship"
	"github.com/centrifuge/go-centrifuge/documents/invoice"
	"github.com/centrifuge/go-centrifuge/documents/purchaseorder"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/extensions/transferdetails"
	"github.com/centrifuge/go-centrifuge/httpapi/coreapi"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
)

// TODO: think: generic custom attribute set creation?

//CreateTransferDetailRequest is the request body for creating a Transfer Detail
type CreateTransferDetailRequest struct {
	DocumentID string               `json:"document_id"`
	Data       transferdetails.Data `json:"data"`
}

// UpdateTransferDetailRequest is the request body for updating a Transfer Detail
type UpdateTransferDetailRequest struct {
	DocumentID string               `json:"document_id"`
	TransferID string               `json:"transfer_id"`
	Data       transferdetails.Data `json:"data"`
}

// TransferDetailResponse is the response body when fetching a Transfer Detail
type TransferDetailResponse struct {
	Header coreapi.ResponseHeader `json:"header"`
	Data   transferdetails.Data   `json:"data"`
}

// TransferDetailListResponse is the response body when fetching a list of Transfer Details
type TransferDetailListResponse struct {
	Header coreapi.ResponseHeader `json:"header"`
	Data   []transferdetails.Data `json:"data"`
}

func toTransferDetailCreatePayload(request CreateTransferDetailRequest) (*transferdetails.CreateTransferDetailRequest, error) {
	payload := transferdetails.CreateTransferDetailRequest{
		DocumentID: request.DocumentID,
		Data:       request.Data,
	}
	return &payload, nil
}

func toTransferDetailUpdatePayload(request UpdateTransferDetailRequest) (*transferdetails.UpdateTransferDetailRequest, error) {
	payload := transferdetails.UpdateTransferDetailRequest{
		DocumentID: request.DocumentID,
		Data:       request.Data,
		TransferID: request.TransferID,
	}
	return &payload, nil
}

// CreateInvoiceRequest defines the payload for creating documents.
type CreateInvoiceRequest struct {
	ReadAccess  []identity.DID              `json:"read_access" swaggertype:"array,string"`
	WriteAccess []identity.DID              `json:"write_access" swaggertype:"array,string"`
	Data        invoice.Data                `json:"data"`
	Attributes  coreapi.AttributeMapRequest `json:"attributes"`
}

// InvoiceResponse represents the invoice in client API format.
type InvoiceResponse struct {
	Header     coreapi.ResponseHeader       `json:"header"`
	Data       invoice.Data                 `json:"data"`
	Attributes coreapi.AttributeMapResponse `json:"attributes"`
}

func toInvoiceResponse(model documents.Model, tokenRegistry documents.TokenRegistry, jobID jobs.JobID) (resp InvoiceResponse, err error) {
	docResp, err := coreapi.GetDocumentResponse(model, tokenRegistry, jobID)
	if err != nil {
		return resp, err
	}

	return InvoiceResponse{
		Header:     docResp.Header,
		Attributes: docResp.Attributes,
		Data:       docResp.Data.(invoice.Data),
	}, nil
}

// CreatePurchaseOrderRequest holds details for creating Purchase order Document.
type CreatePurchaseOrderRequest struct {
	ReadAccess  []identity.DID              `json:"read_access" swaggertype:"array,string"`
	WriteAccess []identity.DID              `json:"write_access" swaggertype:"array,string"`
	Data        purchaseorder.Data          `json:"data"`
	Attributes  coreapi.AttributeMapRequest `json:"attributes"`
}

// PurchaseOrderResponse represents the purchase order in client API format.
type PurchaseOrderResponse struct {
	Header     coreapi.ResponseHeader       `json:"header"`
	Data       purchaseorder.Data           `json:"data"`
	Attributes coreapi.AttributeMapResponse `json:"attributes"`
}

func toPurchaseOrderResponse(model documents.Model, tokenRegistry documents.TokenRegistry, jobID jobs.JobID) (resp PurchaseOrderResponse, err error) {
	docResp, err := coreapi.GetDocumentResponse(model, tokenRegistry, jobID)
	if err != nil {
		return resp, err
	}

	return PurchaseOrderResponse{
		Header:     docResp.Header,
		Attributes: docResp.Attributes,
		Data:       docResp.Data.(purchaseorder.Data),
	}, nil
}

// CreateEntityRequest holds details for creating Entity Document.
type CreateEntityRequest struct {
	ReadAccess  []identity.DID              `json:"read_access" swaggertype:"array,string"`
	WriteAccess []identity.DID              `json:"write_access" swaggertype:"array,string"`
	Data        entity.Data                 `json:"data"`
	Attributes  coreapi.AttributeMapRequest `json:"attributes"`
}

// EntityResponse represents the entity in client API format.
type EntityResponse struct {
	Header     coreapi.ResponseHeader       `json:"header"`
	Data       EntityDataResponse           `json:"data"`
	Attributes coreapi.AttributeMapResponse `json:"attributes"`
}

// EntityDataResponse holds the entity data and Relationships
type EntityDataResponse struct {
	Entity        entity.Data    `json:"entity"`
	Relationships []Relationship `json:"relationships"`
}

// Relationship holds the identity and status of the relationship
type Relationship struct {
	Identity identity.DID `json:"identity" swaggertype:"primitive,string"`
	Active   bool         `json:"active"`
}

func getEntityRelationships(ctx context.Context, erSrv entityrelationship.Service, entity documents.Model) (relationships []Relationship, err error) {
	selfDID, err := contextutil.DIDFromContext(ctx)
	if err != nil {
		return nil, errors.New("failed to get self ID")
	}

	isCollaborator, err := entity.IsDIDCollaborator(selfDID)
	if err != nil {
		return nil, err
	}

	if !isCollaborator {
		return nil, nil
	}

	rs, err := erSrv.GetEntityRelationships(ctx, entity.ID())
	if err != nil {
		return nil, err
	}

	//list the relationships associated with the entity
	for _, r := range rs {
		tokens, err := r.GetAccessTokens()
		if err != nil {
			return nil, err
		}

		targetDID := r.(*entityrelationship.EntityRelationship).Data.TargetIdentity
		relationships = append(relationships, Relationship{
			Identity: *targetDID,
			Active:   len(tokens) != 0,
		})
	}

	return relationships, nil
}

func toEntityResponse(ctx context.Context, erSrv entityrelationship.Service, model documents.Model, tokenRegistry documents.TokenRegistry, jobID jobs.JobID) (resp EntityResponse, err error) {
	docResp, err := coreapi.GetDocumentResponse(model, tokenRegistry, jobID)
	if err != nil {
		return resp, err
	}

	rs, err := getEntityRelationships(ctx, erSrv, model)
	if err != nil {
		return resp, err
	}

	return EntityResponse{
		Header:     docResp.Header,
		Attributes: docResp.Attributes,
		Data: EntityDataResponse{
			Entity:        docResp.Data.(entity.Data),
			Relationships: rs,
		},
	}, nil
}

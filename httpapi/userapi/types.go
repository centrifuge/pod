// nolint
package userapi

import (
	"context"
	"encoding/json"

	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/entity"
	"github.com/centrifuge/go-centrifuge/documents/entityrelationship"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/httpapi/coreapi"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/utils/byteutils"
	"github.com/ethereum/go-ethereum/common"
)

// NFTMintInvoiceUnpaidRequest is the request for minting an NFT for an unpaid NFT
type NFTMintInvoiceUnpaidRequest struct {
	// Deposit address for NFT Token created
	DepositAddress common.Address `json:"deposit_address" swaggertype:"primitive,string"`
}

// ResponseHeader header with job id
type ResponseHeader struct {
	JobID string `json:"job_id"`
}

// NFTMintResponse is response from user api NFT minting
type NFTMintResponse struct {
	Header *ResponseHeader `json:"header"`
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
	TargetIdentity   identity.DID       `json:"target_identity" swaggertype:"primitive,string"`
	OwnerIdentity    identity.DID       `json:"owner_identity" swaggertype:"primitive,string"`
	EntityIdentifier byteutils.HexBytes `json:"entity_identifier" swaggertype:"primitive,string"`
	Active           bool               `json:"active"`
}

// ShareEntityRequest holds the documentID and target identity to share entity with.
type ShareEntityRequest struct {
	TargetIdentity identity.DID `json:"target_identity" swaggertype:"primitive,string"`
}

// ShareEntityResponse holds the response for entity share.
type ShareEntityResponse struct {
	Header       coreapi.ResponseHeader `json:"header"`
	Relationship Relationship           `json:"relationship"`
}

func toEntityShareResponse(model documents.Document, tokenRegistry documents.TokenRegistry, jobID jobs.JobID) (resp ShareEntityResponse, err error) {
	header, err := coreapi.DeriveResponseHeader(tokenRegistry, model, jobID)
	if err != nil {
		return resp, err
	}

	d := model.GetData().(entityrelationship.Data)
	return ShareEntityResponse{
		Header: header,
		Relationship: Relationship{
			TargetIdentity:   *d.TargetIdentity,
			EntityIdentifier: d.EntityIdentifier,
			OwnerIdentity:    *d.OwnerIdentity,
			Active:           true,
		},
	}, nil
}

func convertShareEntityRequest(ctx context.Context, docID byteutils.HexBytes, targetID identity.DID) (req documents.CreatePayload, err error) {
	self, err := contextutil.DIDFromContext(ctx)
	if err != nil {
		return req, err
	}

	d, err := json.Marshal(entityrelationship.Data{
		TargetIdentity:   &targetID,
		OwnerIdentity:    &self,
		EntityIdentifier: docID,
	})
	if err != nil {
		return req, err
	}

	return documents.CreatePayload{
		Scheme: entityrelationship.Scheme,
		Data:   d,
	}, nil
}

func getEntityRelationships(ctx context.Context, erSrv entityrelationship.Service, entity documents.Document) (relationships []Relationship, err error) {
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

		d := r.GetData().(entityrelationship.Data)
		relationships = append(relationships, Relationship{
			TargetIdentity:   *d.TargetIdentity,
			OwnerIdentity:    *d.OwnerIdentity,
			EntityIdentifier: d.EntityIdentifier,
			Active:           len(tokens) != 0,
		})
	}

	return relationships, nil
}

func toEntityResponse(ctx context.Context, erSrv entityrelationship.Service, model documents.Document, tokenRegistry documents.TokenRegistry, jobID jobs.JobID) (resp EntityResponse, err error) {
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

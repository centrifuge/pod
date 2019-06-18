package transferdetails

import (
	"context"
	"reflect"

	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/extensions"
	"github.com/centrifuge/go-centrifuge/identity"
	logging "github.com/ipfs/go-log"
)

// Service defines specific functions for extension funding
type Service interface {
	documents.Service

	// DeriveFromUpdatePayload derives Funding from clientUpdatePayload
	DeriveFromUpdatePayload(ctx context.Context, req UpdateTransferDetailRequest, identifier []byte) (documents.Model, error)

	// DeriveFromPayload derives Funding from clientPayload
	DeriveFromPayload(ctx context.Context, req CreateTransferDetailRequest, identifier []byte) (documents.Model, error)

	// DeriveFundingResponse returns a funding in client format
	DeriveTransferResponse(ctx context.Context, model documents.Model, transferID string) (*TransferDetailResponse, error)

	// DeriveFundingListResponse returns a funding list in client format
	DeriveTransferListResponse(ctx context.Context, model documents.Model) (*TransferDetailListResponse, error)
}

// service implements Service and handles all funding related persistence and validations
type service struct {
	documents.Service
	tokenRegistry documents.TokenRegistry
	idSrv         identity.Service
}

const (
	transfersLabel    = "transfer_details"
	transfersFieldKey = "transfer_details[{IDX}]."
	idxKey            = "{IDX}"
	transferIDLabel   = "transfer_id"
)

// DefaultService returns the default implementation of the service.
func DefaultService(
	srv documents.Service,
	tokenRegistry documents.TokenRegistry,
) Service {
	return service{
		Service:       srv,
		tokenRegistry: tokenRegistry,
	}
}

var transfersLog = logging.Logger("tranfers-api")

func deriveDIDs(data *TransferDetailData) ([]identity.DID, error) {
	var c []identity.DID
	for _, id := range []string{data.SenderId, data.RecipientId} {
		if id != "" {
			did, err := identity.NewDIDFromString(id)
			if err != nil {
				return nil, err
			}
			c = append(c, did)
		}
	}

	return c, nil
}

func (s service) DeriveFromPayload(ctx context.Context, req CreateTransferDetailRequest, identifier []byte) (documents.Model, error) {
	model, err := s.GetCurrentVersion(ctx, identifier)
	if err != nil {
		transfersLog.Error(err)
		return nil, documents.ErrDocumentNotFound
	}
	attributes, err := extensions.CreateAttributesList(model, *req.Data, transfersFieldKey, transfersLabel)
	if err != nil {
		return nil, err
	}

	c, err := deriveDIDs(req.Data)
	if err != nil {
		return nil, err
	}

	err = model.AddAttributes(
		documents.CollaboratorsAccess{
			ReadWriteCollaborators: c,
		},
		true,
		attributes...,
	)
	if err != nil {
		return nil, err
	}

	validator := CreateValidator()
	err = validator.Validate(nil, model)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentInvalid, err)
	}

	return model, nil
}

// DeriveFromUpdatePayload derives Funding from clientUpdatePayload
func (s service) DeriveFromUpdatePayload(ctx context.Context, req UpdateTransferDetailRequest, identifier []byte) (documents.Model, error) {
	model, err := s.GetCurrentVersion(ctx, identifier)
	if err != nil {
		transfersLog.Error(err)
		return nil, documents.ErrDocumentNotFound
	}

	idx, err := extensions.FindAttributeSetIDX(model, req.TransferId, transfersLabel, transferIDLabel, transfersFieldKey)
	if err != nil {
		return nil, err
	}

	c, err := deriveDIDs(req.Data)
	if err != nil {
		return nil, err
	}

	// overwriting is not enough because it is not required that
	// the funding payload contains all funding attributes
	model, err = extensions.DeleteAttributesSet(model, TransferDetailData{}, idx, transfersFieldKey)
	if err != nil {
		return nil, err
	}

	attributes, err := extensions.FillAttributeList(*req.Data, idx, transfersFieldKey)
	if err != nil {
		return nil, err
	}

	err = model.AddAttributes(
		documents.CollaboratorsAccess{
			ReadWriteCollaborators: c,
		},
		true,
		attributes...,
	)
	if err != nil {
		return nil, err
	}

	validator := CreateValidator()
	err = validator.Validate(nil, model)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentInvalid, err)
	}

	return model, nil
}

// TODO: generic?
func (s service) findTransfer(model documents.Model, transferID string) (*TransferDetailData, error) {
	idx, err := extensions.FindAttributeSetIDX(model, transferID, transfersLabel, transferIDLabel, transfersFieldKey)
	if err != nil {
		return nil, err
	}
	return s.deriveTransferData(model, idx)
}

// TODO: generic?
func (s service) deriveTransferData(model documents.Model, idx string) (*TransferDetailData, error) {
	data := new(TransferDetailData)

	types := reflect.TypeOf(*data)
	for i := 0; i < types.NumField(); i++ {
		// generate attr key
		jsonKey := types.Field(i).Tag.Get("json")
		label := extensions.LabelFromJSONTag(idx, jsonKey, transfersFieldKey)

		attrKey, err := documents.AttrKeyFromLabel(label)
		if err != nil {
			return nil, err
		}

		if model.AttributeExists(attrKey) {
			attr, err := model.GetAttribute(attrKey)
			if err != nil {
				return nil, err
			}

			// set field in data
			n := types.Field(i).Name

			v, err := attr.Value.String()
			if err != nil {
				return nil, err
			}

			reflect.ValueOf(data).Elem().FieldByName(n).SetString(v)

		}

	}
	return data, nil
}

// DeriveTransferResponse returns create response from the added transfer detail
func (s service) DeriveTransferResponse(ctx context.Context, model documents.Model, transferID string) (*TransferDetailResponse, error) {
	idx, err := extensions.FindAttributeSetIDX(model, transferID, transfersLabel, transferIDLabel, transfersFieldKey)
	if err != nil {
		return nil, err
	}

	h, err := documents.DeriveResponseHeader(s.tokenRegistry, model)
	if err != nil {
		return nil, errors.New("failed to derive response: %v", err)
	}
	data, err := s.deriveTransferData(model, idx)
	if err != nil {
		return nil, err
	}

	return &TransferDetailResponse{
		Header: h,
		Data:   data,
	}, nil

}

// DeriveTransfersListResponse returns a transfers list
func (s service) DeriveTransferListResponse(ctx context.Context, model documents.Model) (*TransferDetailListResponse, error) {
	response := new(TransferDetailListResponse)

	h, err := documents.DeriveResponseHeader(s.tokenRegistry, model)
	if err != nil {
		return nil, errors.New("failed to derive response: %v", err)
	}
	response.Header = h

	fl, err := documents.AttrKeyFromLabel(transfersLabel)
	if err != nil {
		return nil, err
	}

	if !model.AttributeExists(fl) {
		return response, nil
	}

	lastIdx, err := extensions.GetArrayLatestIDX(model, transfersLabel)
	if err != nil {
		return nil, err
	}

	i, err := documents.NewInt256("0")
	if err != nil {
		return nil, err
	}

	for i.Cmp(lastIdx) != 1 {
		transfer, err := s.deriveTransferData(model, i.String())
		if err != nil {
			continue
		}

		response.Data = append(response.Data, transfer)
		i, err = i.Inc()

		if err != nil {
			return nil, err
		}

	}
	return response, nil
}

package transferdetails

import (
	"context"
	"reflect"

	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/extensions"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/ethereum/go-ethereum/common/hexutil"
	logging "github.com/ipfs/go-log"
)

// Service defines specific functions for extension funding
type Service interface {

	// DeriveFromUpdatePayload derives TransferDetail from clientUpdatePayload
	DeriveFromUpdatePayload(ctx context.Context, req UpdateTransferDetailRequest) (documents.Model, error)

	// DeriveFromPayload derives TransferDetail from clientPayload
	DeriveFromPayload(ctx context.Context, req CreateTransferDetailRequest) (documents.Model, error)

	// DeriveFundingResponse returns a TransferDetail in client format
	DeriveTransferResponse(ctx context.Context, model documents.Model, transferID string) (*TransferDetailResponse, error)

	// DeriveFundingListResponse returns a TransferDetail list in client format
	DeriveTransferListResponse(ctx context.Context, model documents.Model) (*TransferDetailListResponse, error)
}

// service implements Service and handles all funding related persistence and validations
type service struct {
	srv documents.Service
	tokenRegistry documents.TokenRegistry
}

const (
	transfersLabel    = "transfer_details"
	transfersFieldKey = "transfer_details[{IDX}]."
	transferIDLabel   = "transfer_id"
)

// DefaultService returns the default implementation of the service.
func DefaultService(
	srv documents.Service,
	tokenRegistry documents.TokenRegistry,
) Service {
	return service{
		srv:       srv,
		tokenRegistry: tokenRegistry,
	}
}

var log = logging.Logger("transferdetail-api")

// TODO: get rid of this or make generic
func deriveDIDs(data *TransferDetailData) ([]identity.DID, error) {
	var c []identity.DID
	for _, id := range []string{data.SenderID, data.RecipientID} {
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

// DeriveFromPayload derives a new TransferDetail from a CreateTransferDetailRequest
func (s service) DeriveFromPayload(ctx context.Context, req CreateTransferDetailRequest) (model documents.Model, err error) {
	var docID []byte

	if req.DocumentID == "" {
		return nil, documents.ErrDocumentIdentifier
	}

	docID, err = hexutil.Decode(req.DocumentID)
	if err != nil {
		return nil, err
	}

	model, err = s.srv.GetCurrentVersion(ctx, docID)
	if err != nil {
		log.Error(err)
		return nil, documents.ErrDocumentNotFound
	}
	attributes, err := extensions.CreateAttributesList(model, *req.Data, transfersFieldKey, transfersLabel)
	if err != nil {
		return nil, err
	}

	//TODO: StringsToDIDS
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

// DeriveFromUpdatePayload derives an updated TransferDetail from an UpdateTransferDetailRequest
func (s service) DeriveFromUpdatePayload(ctx context.Context, req UpdateTransferDetailRequest) (model documents.Model, err error) {
	var docID []byte
	if req.DocumentID == "" {
		return nil, documents.ErrDocumentIdentifier
	}

	docID, err = hexutil.Decode(req.DocumentID)
	if err != nil {
		return nil, err
	}

	model, err = s.srv.GetCurrentVersion(ctx, docID)
	if err != nil {
		log.Error(err)
		return nil, documents.ErrDocumentNotFound
	}

	idx, err := extensions.FindAttributeSetIDX(model, req.TransferID, transfersLabel, transferIDLabel, transfersFieldKey)
	if err != nil {
		return nil, err
	}

	//TODO: StringsToDIDS
	c, err := deriveDIDs(req.Data)
	if err != nil {
		return nil, err
	}

	// overwriting is not enough because it is not required that
	// the TransferDetail payload contains all TransferDetail attributes
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

// TODO: move to generic function in attribute utils
func (s service) findTransfer(model documents.Model, transferID string) (*TransferDetailData, error) {
	idx, err := extensions.FindAttributeSetIDX(model, transferID, transfersLabel, transferIDLabel, transfersFieldKey)
	if err != nil {
		return nil, err
	}
	return s.deriveTransferData(model, idx)
}

// TODO: move to generic function in attribute utils
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

// DeriveTransferResponse returns create response from the added TransferDetail
func (s service) DeriveTransferResponse(ctx context.Context, model documents.Model, transferID string) (*TransferDetailResponse, error) {
	idx, err := extensions.FindAttributeSetIDX(model, transferID, transfersLabel, transferIDLabel, transfersFieldKey)
	if err != nil {
		return nil, err
	}

	data, err := s.deriveTransferData(model, idx)
	if err != nil {
		return nil, err
	}

	// TODO: derive response header in userapi
	return &TransferDetailResponse{
		Data: data,
	}, nil
}

// DeriveTransfersListResponse returns a transfers list
func (s service) DeriveTransferListResponse(ctx context.Context, model documents.Model) (*TransferDetailListResponse, error) {
	response := new(TransferDetailListResponse)
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
	// TODO: derive response header in userapi
	return response, nil
}

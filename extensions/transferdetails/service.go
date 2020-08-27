package transferdetails

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/extensions"
	"github.com/centrifuge/go-centrifuge/httpapi/coreapi"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/ethereum/go-ethereum/common/hexutil"
	logging "github.com/ipfs/go-log"
)

// Service defines specific functions for extension funding
type Service interface {

	// UpdateTransferDetail updates a TransferDetail
	UpdateTransferDetail(ctx context.Context, req UpdateTransferDetailRequest) (documents.Model, jobs.JobID, error)

	// CreateTransferDetail derives a TransferDetail from a request payload
	CreateTransferDetail(ctx context.Context, req CreateTransferDetailRequest) (documents.Model, jobs.JobID, error)

	// DeriveFundingResponse returns a TransferDetail in client format
	DeriveTransferDetail(ctx context.Context, model documents.Model, transferID []byte) (*TransferDetail, documents.Model, error)

	// DeriveFundingListResponse returns a TransferDetail list in client format
	DeriveTransferList(ctx context.Context, model documents.Model) (*TransferDetailList, documents.Model, error)
}

// service implements Service and handles all funding related persistence and validations
type service struct {
	coreAPISrv    coreapi.Service
	tokenRegistry documents.TokenRegistry
}

const (
	transfersLabel    = "transfer_details"
	transfersFieldKey = "transfer_details[{IDX}]."
	transferIDLabel   = "transfer_id"
)

// DefaultService returns the default implementation of the service.
func DefaultService(
	srv coreapi.Service,
	tokenRegistry documents.TokenRegistry,
) Service {
	return service{
		coreAPISrv:    srv,
		tokenRegistry: tokenRegistry,
	}
}

var log = logging.Logger("transferdetail-api")

// TODO: get rid of this or make generic
func deriveDIDs(data *Data) ([]identity.DID, error) {
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

func (s service) updateModel(ctx context.Context, model documents.Model) (documents.Model, jobs.JobID, error) {
	cs, err := model.GetCollaborators()
	if err != nil {
		return nil, jobs.NilJobID(), err
	}

	d, err := json.Marshal(model.GetData())
	if err != nil {
		return nil, jobs.NilJobID(), err
	}

	a := model.GetAttributes()
	attr := extensions.ToMapAttributes(a)

	payload := documents.UpdatePayload{
		DocumentID: model.ID(),
		CreatePayload: documents.CreatePayload{
			Scheme:        model.Scheme(),
			Collaborators: cs,
			Attributes:    attr,
			Data:          d,
		},
	}

	updated, jobID, err := s.coreAPISrv.UpdateDocument(ctx, payload)
	if err != nil {
		return nil, jobID, err
	}

	return updated, jobID, err
}

// CreateTransferDetail creates and anchors a TransferDetail
func (s service) CreateTransferDetail(ctx context.Context, req CreateTransferDetailRequest) (documents.Model, jobs.JobID, error) {
	model, err := s.deriveFromPayload(ctx, req)
	if err != nil {
		return nil, jobs.NilJobID(), err
	}

	updated, jobID, err := s.updateModel(ctx, model)
	if err != nil {
		return nil, jobID, err
	}

	return updated, jobID, nil
}

// deriveFromPayload derives a new TransferDetail from a CreateTransferDetailRequest
func (s service) deriveFromPayload(ctx context.Context, req CreateTransferDetailRequest) (model documents.Model, err error) {
	if req.DocumentID == "" {
		return nil, documents.ErrDocumentIdentifier
	}

	docID, err := hexutil.Decode(req.DocumentID)
	if err != nil {
		return nil, err
	}

	model, err = s.coreAPISrv.GetDocument(ctx, docID)
	if err != nil {
		log.Error(err)
		return nil, documents.ErrDocumentNotFound
	}
	attributes, err := extensions.CreateAttributesList(model, req.Data, transfersFieldKey, transfersLabel)
	if err != nil {
		return nil, err
	}

	//TODO: StringsToDIDS
	c, err := deriveDIDs(&req.Data)
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

// UpdateTransferDetail updates and anchors a TransferDetail
func (s service) UpdateTransferDetail(ctx context.Context, req UpdateTransferDetailRequest) (documents.Model, jobs.JobID, error) {
	model, err := s.deriveFromUpdatePayload(ctx, req)
	if err != nil {
		return nil, jobs.NilJobID(), err
	}

	updated, jobID, err := s.updateModel(ctx, model)
	if err != nil {
		return nil, jobID, err
	}

	return updated, jobID, nil
}

// deriveFromUpdatePayload derives an updated TransferDetail from an UpdateTransferDetailRequest
func (s service) deriveFromUpdatePayload(ctx context.Context, req UpdateTransferDetailRequest) (model documents.Model, err error) {
	var docID []byte
	if req.DocumentID == "" {
		return nil, documents.ErrDocumentIdentifier
	}

	docID, err = hexutil.Decode(req.DocumentID)
	if err != nil {
		return nil, err
	}

	model, err = s.coreAPISrv.GetDocument(ctx, docID)
	if err != nil {
		log.Error(err)
		return nil, documents.ErrDocumentNotFound
	}

	idx, err := extensions.FindAttributeSetIDX(model, req.TransferID, transfersLabel, transferIDLabel, transfersFieldKey)
	if err != nil {
		return nil, err
	}

	//TODO: StringsToDIDS
	c, err := deriveDIDs(&req.Data)
	if err != nil {
		return nil, err
	}

	// overwriting is not enough because it is not required that
	// the TransferDetail payload contains all TransferDetail attributes
	model, err = extensions.DeleteAttributesSet(model, Data{}, idx, transfersFieldKey)
	if err != nil {
		return nil, err
	}

	attributes, err := extensions.FillAttributeList(req.Data, idx, transfersFieldKey)
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
func (s service) findTransfer(model documents.Model, transferID string) (*Data, error) {
	idx, err := extensions.FindAttributeSetIDX(model, transferID, transfersLabel, transferIDLabel, transfersFieldKey)
	if err != nil {
		return nil, err
	}
	return s.deriveTransferData(model, idx)
}

// TODO: move to generic function in attribute utils
func (s service) deriveTransferData(model documents.Model, idx string) (*Data, error) {
	data := new(Data)

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
func (s service) DeriveTransferDetail(ctx context.Context, model documents.Model, transferID []byte) (*TransferDetail, documents.Model, error) {
	tID := hexutil.Encode(transferID)
	idx, err := extensions.FindAttributeSetIDX(model, tID, transfersLabel, transferIDLabel, transfersFieldKey)
	if err != nil {
		return nil, nil, err
	}

	data, err := s.deriveTransferData(model, idx)
	if err != nil {
		return nil, nil, err
	}

	return &TransferDetail{
		Data: *data,
	}, model, nil
}

// DeriveTransfersListResponse returns a transfers list
func (s service) DeriveTransferList(ctx context.Context, model documents.Model) (*TransferDetailList, documents.Model, error) {
	list := new(TransferDetailList)
	fl, err := documents.AttrKeyFromLabel(transfersLabel)
	if err != nil {
		return nil, nil, err
	}

	if !model.AttributeExists(fl) {
		return &TransferDetailList{
			Data: nil,
		}, model, nil
	}

	lastIdx, err := extensions.GetArrayLatestIDX(model, transfersLabel)
	if err != nil {
		return nil, nil, err
	}

	i, err := documents.NewInt256("0")
	if err != nil {
		return nil, nil, err
	}

	for i.Cmp(lastIdx) != 1 {
		transfer, err := s.deriveTransferData(model, i.String())
		if err != nil {
			continue
		}

		list.Data = append(list.Data, *transfer)
		i, err = i.Inc()

		if err != nil {
			return nil, nil, err
		}
	}

	return &TransferDetailList{
		Data: list.Data,
	}, model, nil
}

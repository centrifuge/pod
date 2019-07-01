package transferdetails

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/extensions"
	"github.com/centrifuge/go-centrifuge/httpapi"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/ethereum/go-ethereum/common/hexutil"
	logging "github.com/ipfs/go-log"
)

// service implements TransferDetailService and handles all funding related persistence and validations
type service struct {
	// coreSrvProvider is a lazy initiated service on the API router
	coreSrvProvider func() httpapi.CoreService
	tokenRegistry   documents.TokenRegistry
}

const (
	transfersLabel    = "transfer_details"
	transfersFieldKey = "transfer_details[{IDX}]."
	transferIDLabel   = "transfer_id"
)

// DefaultService returns the default implementation of the service.
func DefaultService(
	srv func() httpapi.CoreService,
	tokenRegistry documents.TokenRegistry,
) extensions.TransferDetailService {
	return service{
		coreSrvProvider: srv,
		tokenRegistry:   tokenRegistry,
	}
}

var log = logging.Logger("transferdetail-api")

// TODO: get rid of this or make generic
func deriveDIDs(data *extensions.Data) ([]identity.DID, error) {
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

	updated, jobID, err := s.coreSrvProvider().UpdateDocument(ctx, payload)
	if err != nil {
		return nil, jobID, err
	}

	return updated, jobID, err
}

// CreateTransferDetail creates and anchors a TransferDetail
func (s service) CreateTransferDetail(ctx context.Context, req extensions.CreateTransferDetailRequest) (documents.Model, jobs.JobID, error) {
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
func (s service) deriveFromPayload(ctx context.Context, req extensions.CreateTransferDetailRequest) (model documents.Model, err error) {
	if req.DocumentID == "" {
		return nil, documents.ErrDocumentIdentifier
	}

	docID, err := hexutil.Decode(req.DocumentID)
	if err != nil {
		return nil, err
	}

	model, err = s.coreSrvProvider().GetDocument(ctx, docID)
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
func (s service) UpdateTransferDetail(ctx context.Context, req extensions.UpdateTransferDetailRequest) (documents.Model, jobs.JobID, error) {
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
func (s service) deriveFromUpdatePayload(ctx context.Context, req extensions.UpdateTransferDetailRequest) (model documents.Model, err error) {
	var docID []byte
	if req.DocumentID == "" {
		return nil, documents.ErrDocumentIdentifier
	}

	docID, err = hexutil.Decode(req.DocumentID)
	if err != nil {
		return nil, err
	}

	model, err = s.coreSrvProvider().GetDocument(ctx, docID)
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
	model, err = extensions.DeleteAttributesSet(model, extensions.Data{}, idx, transfersFieldKey)
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
func (s service) findTransfer(model documents.Model, transferID string) (*extensions.Data, error) {
	idx, err := extensions.FindAttributeSetIDX(model, transferID, transfersLabel, transferIDLabel, transfersFieldKey)
	if err != nil {
		return nil, err
	}
	return s.deriveTransferData(model, idx)
}

// TODO: move to generic function in attribute utils
func (s service) deriveTransferData(model documents.Model, idx string) (*extensions.Data, error) {
	data := new(extensions.Data)

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
func (s service) DeriveTransferDetail(ctx context.Context, model documents.Model, transferID []byte) (*extensions.TransferDetail, documents.Model, error) {
	tID := hexutil.Encode(transferID)
	idx, err := extensions.FindAttributeSetIDX(model, tID, transfersLabel, transferIDLabel, transfersFieldKey)
	if err != nil {
		return nil, nil, err
	}

	data, err := s.deriveTransferData(model, idx)
	if err != nil {
		return nil, nil, err
	}

	return &extensions.TransferDetail{
		Data: *data,
	}, model, nil
}

// DeriveTransfersListResponse returns a transfers list
func (s service) DeriveTransferList(ctx context.Context, model documents.Model) (*extensions.TransferDetailList, documents.Model, error) {
	list := new(extensions.TransferDetailList)
	fl, err := documents.AttrKeyFromLabel(transfersLabel)
	if err != nil {
		return nil, nil, err
	}

	if !model.AttributeExists(fl) {
		return &extensions.TransferDetailList{
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

	return &extensions.TransferDetailList{
		Data: list.Data,
	}, model, nil
}

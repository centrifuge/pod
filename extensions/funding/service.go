package funding

import (
	"context"
	"reflect"

	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/extensions"
	"github.com/centrifuge/go-centrifuge/identity"
	clientfunpb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/funding"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// Service defines specific functions for extension funding
type Service interface {
	documents.Service

	// Sign adds a signature to an existing document
	Sign(ctx context.Context, fundingID string, identifier []byte) (documents.Model, error)

	// DeriveFromUpdatePayload derives Funding from clientUpdatePayload
	DeriveFromUpdatePayload(ctx context.Context, req *clientfunpb.FundingUpdatePayload) (documents.Model, error)

	// DeriveFromPayload derives Funding from clientPayload
	DeriveFromPayload(ctx context.Context, req *clientfunpb.FundingCreatePayload) (documents.Model, error)

	// DeriveFundingResponse returns a funding in client format
	DeriveFundingResponse(ctx context.Context, model documents.Model, fundingID string) (*clientfunpb.FundingResponse, error)

	// DeriveFundingListResponse returns a funding list in client format
	DeriveFundingListResponse(ctx context.Context, model documents.Model) (*clientfunpb.FundingListResponse, error)
}

// service implements TransferDetailService and handles all funding related persistence and validations
type service struct {
	documents.Service
	tokenRegistry documents.TokenRegistry
	idSrv         identity.Service
}

const (
	fundingLabel              = "funding_agreement"
	fundingFieldKey           = "funding_agreement[{IDX}]."
	agreementIDLabel          = "agreement_id"
	fundingSignatures         = "signatures"
	fundingSignaturesFieldKey = "signatures[{IDX}]"
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

func deriveDIDs(data *clientfunpb.FundingData) ([]identity.DID, error) {
	var c []identity.DID
	for _, id := range []string{data.BorrowerId, data.FunderId} {
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

func (s service) DeriveFromPayload(ctx context.Context, req *clientfunpb.FundingCreatePayload) (model documents.Model, err error) {
	var fd Data
	fd.initFundingFromData(req.Data)

	var docID []byte
	if req.DocumentId == "" {
		return nil, documents.ErrDocumentIdentifier
	}

	docID, err = hexutil.Decode(req.DocumentId)
	if err != nil {
		return nil, err
	}

	model, err = s.GetCurrentVersion(ctx, docID)
	if err != nil {
		apiLog.Error(err)
		return nil, documents.ErrDocumentNotFound
	}
	attributes, err := extensions.CreateAttributesList(model, fd, fundingFieldKey, fundingLabel)
	if err != nil {
		return nil, err
	}

	//TODO: use StringsToDIDS here
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
func (s service) DeriveFromUpdatePayload(ctx context.Context, req *clientfunpb.FundingUpdatePayload) (model documents.Model, err error) {
	var fd Data
	fd.initFundingFromData(req.Data)

	var docID []byte
	if req.DocumentId == "" {
		return nil, documents.ErrDocumentIdentifier
	}

	docID, err = hexutil.Decode(req.DocumentId)
	if err != nil {
		return nil, err
	}

	model, err = s.GetCurrentVersion(ctx, docID)
	if err != nil {
		apiLog.Error(err)
		return nil, documents.ErrDocumentNotFound
	}

	fd.AgreementId = req.AgreementId
	idx, err := extensions.FindAttributeSetIDX(model, fd.AgreementId, fundingLabel, agreementIDLabel, fundingFieldKey)
	if err != nil {
		return nil, err
	}

	//TODO: use StringsToDIDS here
	c, err := deriveDIDs(req.Data)
	if err != nil {
		return nil, err
	}

	// overwriting is not enough because it is not required that
	// the funding payload contains all funding attributes
	model, err = extensions.DeleteAttributesSet(model, Data{}, idx, fundingFieldKey)
	if err != nil {
		return nil, err
	}

	attributes, err := extensions.FillAttributeList(fd, idx, fundingFieldKey)
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

// TODO: Move to attribute utils
func (s service) findFunding(model documents.Model, fundingID string) (*Data, error) {
	idx, err := extensions.FindAttributeSetIDX(model, fundingID, fundingLabel, agreementIDLabel, fundingFieldKey)
	if err != nil {
		return nil, err
	}
	return s.deriveFundingData(model, idx)
}

// TODO: Move to attribute utils
func (s service) deriveFundingData(model documents.Model, idx string) (*Data, error) {
	data := new(Data)

	types := reflect.TypeOf(*data)
	for i := 0; i < types.NumField(); i++ {
		// generate attr key
		jsonKey := types.Field(i).Tag.Get("json")
		label := extensions.LabelFromJSONTag(idx, jsonKey, fundingFieldKey)

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

// DeriveFundingResponse returns create response from the added funding
func (s service) DeriveFundingResponse(ctx context.Context, model documents.Model, fundingID string) (*clientfunpb.FundingResponse, error) {
	idx, err := extensions.FindAttributeSetIDX(model, fundingID, fundingLabel, agreementIDLabel, fundingFieldKey)
	if err != nil {
		return nil, err
	}

	h, err := documents.DeriveResponseHeader(s.tokenRegistry, model)
	if err != nil {
		return nil, errors.New("failed to derive response: %v", err)
	}
	data, err := s.deriveFundingData(model, idx)
	if err != nil {
		return nil, err
	}

	signatures, err := s.deriveFundingSignatures(ctx, model, data, idx)
	if err != nil {
		return nil, errors.NewTypedError(extensions.ErrAttrSetSignature, err)
	}

	return &clientfunpb.FundingResponse{
		Header: h,
		Data:   &clientfunpb.FundingResponseData{Funding: data.getClientData(), Signatures: signatures},
	}, nil

}

// DeriveFundingListResponse returns a funding list in client format
func (s service) DeriveFundingListResponse(ctx context.Context, model documents.Model) (*clientfunpb.FundingListResponse, error) {
	response := new(clientfunpb.FundingListResponse)

	h, err := documents.DeriveResponseHeader(s.tokenRegistry, model)
	if err != nil {
		return nil, errors.New("failed to derive response: %v", err)
	}
	response.Header = h

	fl, err := documents.AttrKeyFromLabel(fundingLabel)
	if err != nil {
		return nil, err
	}

	if !model.AttributeExists(fl) {
		return response, nil
	}

	lastIdx, err := extensions.GetArrayLatestIDX(model, fundingLabel)
	if err != nil {
		return nil, err
	}

	i, err := documents.NewInt256("0")
	if err != nil {
		return nil, err
	}

	for i.Cmp(lastIdx) != 1 {
		funding, err := s.deriveFundingData(model, i.String())
		if err != nil {
			continue
		}

		signatures, err := s.deriveFundingSignatures(ctx, model, funding, i.String())
		if err != nil {
			return nil, errors.NewTypedError(extensions.ErrAttrSetSignature, err)
		}

		response.Data = append(response.Data, &clientfunpb.FundingResponseData{Funding: funding.getClientData(), Signatures: signatures})
		i, err = i.Inc()

		if err != nil {
			return nil, err
		}

	}
	return response, nil
}

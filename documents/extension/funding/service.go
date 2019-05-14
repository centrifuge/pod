package funding

import (
	"context"
	"reflect"
	"strings"

	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	clientfundingpb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/funding"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// Service defines specific functions for extension funding
type Service interface {
	documents.Service

	// Sign adds a signature to an existing document
	Sign(ctx context.Context, fundingID string, identifier []byte) (documents.Model, error)

	// DeriveFromUpdatePayload derives Funding from clientUpdatePayload
	DeriveFromUpdatePayload(ctx context.Context, req *clientfundingpb.FundingUpdatePayload, identifier []byte) (documents.Model, error)

	// DeriveFromPayload derives Funding from clientPayload
	DeriveFromPayload(ctx context.Context, req *clientfundingpb.FundingCreatePayload, identifier []byte) (documents.Model, error)

	// DeriveFundingResponse returns a funding in client format
	DeriveFundingResponse(ctx context.Context, model documents.Model, fundingID string) (*clientfundingpb.FundingResponse, error)

	// DeriveFundingListResponse returns a funding list in client format
	DeriveFundingListResponse(ctx context.Context, model documents.Model) (*clientfundingpb.FundingListResponse, error)
}

// service implements Service and handles all funding related persistence and validations
type service struct {
	documents.Service
	tokenRegistry documents.TokenRegistry
	idSrv         identity.Service
}

const (
	fundingLabel              = "funding_agreement"
	fundingFieldKey           = "funding_agreement[{IDX}]."
	idxKey                    = "{IDX}"
	fundingIDLabel            = "funding_id"
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

func newFundingID() string {
	return hexutil.Encode(utils.RandomSlice(32))
}

func generateLabel(field, idx, fieldName string) string {
	return strings.Replace(field, idxKey, idx, -1) + fieldName
}

func labelFromJSONTag(idx, jsonTag string) string {
	jsonKeyIdx := 0
	// example `json:"days,omitempty"`
	jsonParts := strings.Split(jsonTag, ",")
	return generateLabel(fundingFieldKey, idx, jsonParts[jsonKeyIdx])
}

func getArrayLatestIDX(model documents.Model, arrayLabel string) (idx *documents.Int256, err error) {
	key, err := documents.AttrKeyFromLabel(arrayLabel)
	if err != nil {
		return nil, err
	}

	attr, err := model.GetAttribute(key)
	if err != nil {
		return nil, err
	}

	idx = attr.Value.Int256

	z, err := documents.NewInt256("0")
	if err != nil {
		return nil, err
	}

	// idx < 0
	if idx.Cmp(z) == -1 {
		return nil, ErrFundingIndex
	}

	return idx, nil

}

func incrementArrayAttrIDX(model documents.Model, arrayLabel string) (attr documents.Attribute, err error) {
	key, err := documents.AttrKeyFromLabel(arrayLabel)
	if err != nil {
		return attr, err
	}

	if !model.AttributeExists(key) {
		return documents.NewAttribute(arrayLabel, documents.AttrInt256, "0")
	}

	idx, err := getArrayLatestIDX(model, arrayLabel)
	if err != nil {
		return attr, err
	}

	// increment idx
	newIdx, err := idx.Inc()

	if err != nil {
		return attr, err
	}

	return documents.NewAttribute(arrayLabel, documents.AttrInt256, newIdx.String())
}

func fillAttributeList(data Data, idx string) ([]documents.Attribute, error) {
	var attributes []documents.Attribute

	types := reflect.TypeOf(data)
	values := reflect.ValueOf(data)
	for i := 0; i < types.NumField(); i++ {

		value := values.Field(i).Interface().(string)
		if value != "" {
			jsonKey := types.Field(i).Tag.Get("json")
			label := labelFromJSONTag(idx, jsonKey)

			attrType := types.Field(i).Tag.Get("attr")
			attr, err := documents.NewAttribute(label, documents.AttributeType(attrType), value)
			if err != nil {
				return nil, err
			}

			attributes = append(attributes, attr)
		}

	}
	return attributes, nil
}

func createAttributesList(current documents.Model, data Data) ([]documents.Attribute, error) {
	var attributes []documents.Attribute

	idx, err := incrementArrayAttrIDX(current, fundingLabel)
	if err != nil {
		return nil, err
	}

	attributes, err = fillAttributeList(data, idx.Value.Int256.String())
	if err != nil {
		return nil, err
	}

	// add updated idx
	attributes = append(attributes, idx)

	return attributes, nil
}

func (s service) DeriveFromPayload(ctx context.Context, req *clientfundingpb.FundingCreatePayload, identifier []byte) (documents.Model, error) {
	var fd Data
	fd.initFundingFromData(req.Data)

	model, err := s.GetCurrentVersion(ctx, identifier)
	if err != nil {
		apiLog.Error(err)
		return nil, documents.ErrDocumentNotFound
	}

	attributes, err := createAttributesList(model, fd)
	if err != nil {
		return nil, err
	}

	err = model.AddAttributes(attributes...)
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

func (s service) deleteFunding(model documents.Model, idx string) (documents.Model, error) {
	var data Data
	types := reflect.TypeOf(data)
	for i := 0; i < types.NumField(); i++ {
		jsonKey := types.Field(i).Tag.Get("json")
		key, err := documents.AttrKeyFromLabel(labelFromJSONTag(idx, jsonKey))
		if err != nil {
			continue
		}

		if model.AttributeExists(key) {
			err := model.DeleteAttribute(key)
			if err != nil {
				return nil, err
			}
		}

	}
	return model, nil
}

// DeriveFromUpdatePayload derives Funding from clientUpdatePayload
func (s service) DeriveFromUpdatePayload(ctx context.Context, req *clientfundingpb.FundingUpdatePayload, identifier []byte) (documents.Model, error) {
	var fd Data
	fd.initFundingFromData(req.Data)

	model, err := s.GetCurrentVersion(ctx, identifier)
	if err != nil {
		apiLog.Error(err)
		return nil, documents.ErrDocumentNotFound
	}

	fd.FundingId = req.FundingId
	idx, err := s.findFundingIDX(model, fd.FundingId)
	if err != nil {
		return nil, err
	}

	// overwriting is not enough because it is not required that
	// the funding payload contains all funding attributes
	model, err = s.deleteFunding(model, idx)
	if err != nil {
		return nil, err
	}

	attributes, err := fillAttributeList(fd, idx)
	if err != nil {
		return nil, err
	}

	err = model.AddAttributes(attributes...)
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

func (s service) findFunding(model documents.Model, fundingID string) (*Data, error) {
	idx, err := s.findFundingIDX(model, fundingID)
	if err != nil {
		return nil, err
	}
	return s.deriveFundingData(model, idx)
}

func (s service) findFundingIDX(model documents.Model, fundingID string) (idx string, err error) {
	lastIdx, err := getArrayLatestIDX(model, fundingLabel)
	if err != nil {
		return idx, err
	}

	i, err := documents.NewInt256("0")
	if err != nil {
		return idx, err
	}

	for i.Cmp(lastIdx) != 1 {
		label := generateLabel(fundingFieldKey, i.String(), fundingIDLabel)
		k, err := documents.AttrKeyFromLabel(label)
		if err != nil {
			return idx, err
		}

		attr, err := model.GetAttribute(k)
		if err != nil {
			return idx, err
		}

		attrFundingID, err := attr.Value.String()
		if err != nil {
			return idx, err
		}
		if attrFundingID == fundingID {
			return i.String(), nil
		}
		i, err = i.Inc()

		if err != nil {
			return idx, err
		}

	}

	return idx, ErrFundingNotFound
}

func (s service) deriveFundingData(model documents.Model, idx string) (*Data, error) {
	data := new(Data)

	types := reflect.TypeOf(*data)
	for i := 0; i < types.NumField(); i++ {
		// generate attr key
		jsonKey := types.Field(i).Tag.Get("json")
		label := labelFromJSONTag(idx, jsonKey)

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
func (s service) DeriveFundingResponse(ctx context.Context, model documents.Model, fundingID string) (*clientfundingpb.FundingResponse, error) {
	idx, err := s.findFundingIDX(model, fundingID)
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
		return nil, errors.NewTypedError(ErrFundingSignature, err)
	}

	return &clientfundingpb.FundingResponse{
		Header: h,
		Data:   &clientfundingpb.FundingResponseData{Funding: data.getClientData(), Signatures: signatures},
	}, nil

}

// DeriveFundingListResponse returns a funding list in client format
func (s service) DeriveFundingListResponse(ctx context.Context, model documents.Model) (*clientfundingpb.FundingListResponse, error) {
	response := new(clientfundingpb.FundingListResponse)

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

	lastIdx, err := getArrayLatestIDX(model, fundingLabel)
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
			return nil, errors.NewTypedError(ErrFundingSignature, err)
		}

		response.List = append(response.List, &clientfundingpb.FundingResponseData{Funding: funding.getClientData(), Signatures: signatures})
		i, err = i.Inc()

		if err != nil {
			return nil, err
		}

	}
	return response, nil
}

package funding

import (
	"context"
	"reflect"
	"strings"

	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/centrifuge/go-centrifuge/centerrors"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	clientfundingpb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/funding"
)

// Service defines specific functions for extension funding
type Service interface {
	documents.Service
	// DeriveFromPayload derives Funding from clientPayload
	DeriveFromPayload(ctx context.Context, req *clientfundingpb.FundingCreatePayload, identifier []byte) (documents.Model, error)

	// DeriveFundingResponse returns a funding in client format
	DeriveFundingResponse(model documents.Model, fundingId string) (*clientfundingpb.FundingResponse, error)
}

// service implements Service and handles all funding related persistence and validations
type service struct {
	documents.Service
	tokenRegFinder func() documents.TokenRegistry
}

const (
	fundingLabel = "centrifuge_funding"
	fundingFieldKey = "centrifuge_funding[{IDX}]."
	fundingIdx = "{IDX}"
	fundingIdLabel = "funding_id"
)

// DefaultService returns the default implementation of the service.
func DefaultService(
	srv documents.Service,
	tokenRegFinder func() documents.TokenRegistry,
) Service {
	return service{
		Service:        srv,
		tokenRegFinder: tokenRegFinder,
	}
}

func newFundingId() string {
	return hexutil.Encode(utils.RandomSlice(32))
}

func generateLabel(idx, fieldName string) string {
	return strings.Replace(fundingFieldKey, fundingIdx, idx, -1) + fieldName
}

func keyFromJSONTag(idx, jsonTag string) (key string, err error) {
	correctJSONParts := 2
	jsonKeyIdx := 0

	// example `json:"days,omitempty"`
	jsonParts := strings.Split(jsonTag, ",")
	if len(jsonParts) == correctJSONParts {
		return generateLabel(idx, jsonParts[jsonKeyIdx]), nil

	}
	return key, ErrNoFundingField

}

func getFundingsLatestIdx(model documents.Model) (idx *documents.Int256, err error){
	key, err := documents.AttrKeyFromLabel(fundingLabel)
	if err != nil {
		return idx, err
	}

	attr, err := model.GetAttribute(key)
	if err != nil {
		return idx, err
	}

	idx = attr.Value.Int256

	z, err := documents.NewInt256("0")
	if err != nil {
		return idx, err
	}

	// idx < 0
	if idx.Cmp(z) == -1 {
		return idx, ErrFundingIndex
	}

	return idx, nil

}

func defineFundingAttrIdx(model documents.Model) (attr documents.Attribute, err error) {
	key, err := documents.AttrKeyFromLabel(fundingLabel)
	if err != nil {
		return attr, err
	}

	if !model.AttributeExists(key) {
		return documents.NewAttribute(fundingLabel, documents.AttrInt256, "0")
	}

	idx, err := getFundingsLatestIdx(model)
	if err != nil {
		return attr, err
	}

	// increment idx
	newIdx := idx.Inc()

	if err != nil {
		return attr, err
	}

	return documents.NewAttribute(fundingLabel, documents.AttrInt256, newIdx.String())

}

func createAttributesList(current documents.Model, data FundingData) ([]documents.Attribute, error) {
	var attributes []documents.Attribute

	idx, err := defineFundingAttrIdx(current)
	if err != nil {
		return nil, err
	}

	// add updated idx
	attributes = append(attributes, idx)

	types := reflect.TypeOf(data)
	values := reflect.ValueOf(data)
	for i := 0; i < types.NumField(); i++ {
		jsonKey := types.Field(i).Tag.Get("json")
		key, err := keyFromJSONTag(idx.Value.Int256.String(), jsonKey)
		if err != nil {
			continue
		}

		value := values.Field(i).Interface().(string)
		attrType := types.Field(i).Tag.Get("attr")

		attr, err := documents.NewAttribute(key, documents.AttributeType(attrType), value)
		if err != nil {
			return nil, err
		}

		attributes = append(attributes, attr)

	}

	return attributes, nil
}

func (s service) DeriveFromPayload(ctx context.Context, req *clientfundingpb.FundingCreatePayload, identifier []byte) (documents.Model, error) {
	fd := FundingData{}
	fd.initFundingFromData(req.Data)

	model, err := s.GetCurrentVersion(ctx, identifier)
	if err != nil {
		apiLog.Error(err)
		return nil, centerrors.Wrap(err, "document not found")
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

func (s service) findFunding(model documents.Model, fundingId string) (idx string ,err error) {
	lastIdx, err := getFundingsLatestIdx(model)
	if err != nil {
		return idx, err
	}

	i, err := documents.NewInt256("0")
	if err != nil {
		return idx, err
	}

	r := 0
	for r != 1 {
		label := generateLabel(i.String(),fundingIdLabel)
		k, err := documents.AttrKeyFromLabel(label)
		if err != nil {
			return idx, err
		}

		attr, err := model.GetAttribute(k)
		if err != nil {
			return idx, err
		}

		if  attr.Value.Str == fundingId {
			return i.String(), nil
		}

		i.Inc()
		r = i.Cmp(lastIdx)

	}

	return idx, ErrFundingNotFound
}


func (s service) deriveFundingData(model documents.Model, idx string) (*clientfundingpb.FundingData, error) {
	data := &clientfundingpb.FundingData{}
	fd := FundingData{}

	types := reflect.TypeOf(fd)
	for i := 0; i < types.NumField(); i++ {
		// generate attr key
		jsonKey := types.Field(i).Tag.Get("json")
		label, err := keyFromJSONTag(idx, jsonKey)
		if err != nil {
			continue
		}

		attrKey, err := documents.AttrKeyFromLabel(label)
		if err != nil {
			return nil, err
		}

		attr, err := model.GetAttribute(attrKey)

		if err != nil {
			return nil, err
		}

		// set field in client data
		n := types.Field(i).Name
		reflect.ValueOf(data).Elem().FieldByName(n).SetString(attr.Value.Str)

	}
	return data, nil

}

// DeriveFundingResponse returns create response from the added funding
func (s service) DeriveFundingResponse(model documents.Model,fundingId string) (*clientfundingpb.FundingResponse, error) {
	idx, err := s.findFunding(model, fundingId)
	if err != nil {
		return nil, err
	}

	h, err := documents.DeriveResponseHeader(s.tokenRegFinder(), model)
	if err != nil {
		return nil, errors.New("failed to derive response: %v", err)
	}

	data, err := s.deriveFundingData(model, idx)
	if err != nil {
		return nil, err
	}

	return &clientfundingpb.FundingResponse{
		Header: h,
		Data:   data,
	}, nil

}

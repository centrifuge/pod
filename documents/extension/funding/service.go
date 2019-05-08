package funding

import (
	"context"
	"reflect"
	"strings"

	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	clientfundingpb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/funding"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// Service defines specific functions for extension funding
type Service interface {
	documents.Service
	// DeriveFromPayload derives Funding from clientPayload
	DeriveFromPayload(ctx context.Context, req *clientfundingpb.FundingCreatePayload, identifier []byte) (documents.Model, error)

	// DeriveFundingResponse returns a funding in client format
	DeriveFundingResponse(model documents.Model, fundingID string) (*clientfundingpb.FundingResponse, error)

	// DeriveFundingListResponse returns a funding list in client format
	DeriveFundingListResponse(model documents.Model) (*clientfundingpb.FundingListResponse, error)
}

// service implements Service and handles all funding related persistence and validations
type service struct {
	documents.Service
	tokenRegistry documents.TokenRegistry
}

const (
	fundingLabel    = "funding_agreement"
	fundingFieldKey = "funding_agreement[{IDX}]."
	fundingIDx      = "{IDX}"
	fundingIDLabel  = "funding_id"
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

func generateLabel(idx, fieldName string) string {
	return strings.Replace(fundingFieldKey, fundingIDx, idx, -1) + fieldName
}

func labelFromJSONTag(idx, jsonTag string) string {
	jsonKeyIdx := 0
	// example `json:"days,omitempty"`
	jsonParts := strings.Split(jsonTag, ",")
	return generateLabel(idx, jsonParts[jsonKeyIdx])
}

func getFundingsLatestIDX(model documents.Model) (idx *documents.Int256, err error) {
	key, err := documents.AttrKeyFromLabel(fundingLabel)
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

func incrementFundingAttrIDX(model documents.Model) (attr documents.Attribute, err error) {
	key, err := documents.AttrKeyFromLabel(fundingLabel)
	if err != nil {
		return attr, err
	}

	if !model.AttributeExists(key) {
		return documents.NewAttribute(fundingLabel, documents.AttrInt256, "0")
	}

	idx, err := getFundingsLatestIDX(model)
	if err != nil {
		return attr, err
	}

	// increment idx
	newIdx, err := idx.Inc()

	if err != nil {
		return attr, err
	}

	return documents.NewAttribute(fundingLabel, documents.AttrInt256, newIdx.String())
}

func createAttributesList(current documents.Model, data Data) ([]documents.Attribute, error) {
	var attributes []documents.Attribute

	idx, err := incrementFundingAttrIDX(current)
	if err != nil {
		return nil, err
	}

	// add updated idx
	attributes = append(attributes, idx)

	types := reflect.TypeOf(data)
	values := reflect.ValueOf(data)
	for i := 0; i < types.NumField(); i++ {

		value := values.Field(i).Interface().(string)
		if value != "" {
			jsonKey := types.Field(i).Tag.Get("json")
			label := labelFromJSONTag(idx.Value.Int256.String(), jsonKey)

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

func (s service) findFunding(model documents.Model, fundingID string) (idx string, err error) {
	lastIdx, err := getFundingsLatestIDX(model)
	if err != nil {
		return idx, err
	}

	i, err := documents.NewInt256("0")
	if err != nil {
		return idx, err
	}

	for i.Cmp(lastIdx) != 1 {
		label := generateLabel(i.String(), fundingIDLabel)
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

func (s service) deriveFundingData(model documents.Model, idx string) (*clientfundingpb.FundingData, error) {
	data := new(clientfundingpb.FundingData)
	var fd Data

	types := reflect.TypeOf(fd)
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

			// set field in client data
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
func (s service) DeriveFundingResponse(model documents.Model, fundingID string) (*clientfundingpb.FundingResponse, error) {
	idx, err := s.findFunding(model, fundingID)
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

	return &clientfundingpb.FundingResponse{
		Header: h,
		Data:   data,
	}, nil

}

// DeriveFundingListResponse returns a funding list in client format
func (s service) DeriveFundingListResponse(model documents.Model) (*clientfundingpb.FundingListResponse, error) {
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

	lastIdx, err := getFundingsLatestIDX(model)
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
		response.List = append(response.List, funding)
		i, err = i.Inc()

		if err != nil {
			return nil, err
		}

	}
	return response, nil
}

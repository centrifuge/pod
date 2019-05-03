package funding

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
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

const fundingLabel = "centrifuge_funding"
const fundingFieldKey = "centrifuge_funding[{IDX}]."
const fundingIdx = "{IDX}"
const fundingIdLabel = "funding_id"

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

func getFundingsLatestIdx(model documents.Model) (idx int, err error){
	key, err := documents.AttrKeyFromLabel(fundingLabel)
	if err != nil {
		return idx, err
	}

	attr, err := model.GetAttribute(key)
	if err != nil {
		return idx, err
	}

	idx, err = strconv.Atoi(attr.Value.Str)
	if err != nil {
		return idx, err
	}

	if idx < 0 {
		return idx, ErrFundingIndex
	}

	return idx, nil

}

func defineFundingIdx(model documents.Model) (attr documents.Attribute, err error) {
	key, err := documents.AttrKeyFromLabel(fundingLabel)
	if err != nil {
		return attr, err
	}

	if !model.AttributeExists(key) {
		return documents.NewAttribute(fundingLabel, documents.AttrString, "0")

	}

	idx, err := getFundingsLatestIdx(model)
	if err != nil {
		return attr, err
	}

	newIdx, err := documents.AttrValFromString(documents.AttrString, strconv.Itoa(idx+1))
	if err != nil {
		return attr, err
	}

	return documents.NewAttribute(fundingLabel, documents.AttrString, newIdx.Str)

}

func createAttributesList(current documents.Model, data FundingData) ([]documents.Attribute, error) {
	var attributes []documents.Attribute

	idx, err := defineFundingIdx(current)
	if err != nil {
		return nil, err
	}

	// add updated idx
	attributes = append(attributes, idx)

	types := reflect.TypeOf(data)
	values := reflect.ValueOf(data)
	for i := 0; i < types.NumField(); i++ {
		jsonKey := types.Field(i).Tag.Get("json")
		key, err := keyFromJSONTag(idx.Value.Str, jsonKey)
		if err != nil {
			continue
		}

		value := values.Field(i).Interface().(string)
		attrType := types.Field(i).Type.String()

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

	fmt.Println(attributes)

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

func (s service) findFunding(model documents.Model, fundingId string) (idx int,err error) {
	lastIdx, err := getFundingsLatestIdx(model)

	for i := 0; i <= lastIdx; i++ {
		label := generateLabel(strconv.Itoa(i),fundingIdLabel)
		k, err := documents.AttrKeyFromLabel(label)
		if err != nil {
			return idx, err
		}

		attr, err := model.GetAttribute(k)
		if err != nil {
			return idx, err
		}

		if  attr.Value.Str == fundingId {
			return i, nil
		}

	}

	return idx, ErrFundingNotFound
}


func (s service) deriveFundingData(model documents.Model, idx string) (*clientfundingpb.FundingData, error) {
	//data := clientfundingpb.FundingData{}


	//reflect.ValueOf(&data).Elem().FieldByName("N").SetInt(7)

	return nil, nil

}

// DeriveFundingResponse returns create response from the added funding
func (s service) DeriveFundingResponse(model documents.Model,fundingId string) (*clientfundingpb.FundingResponse, error) {
	_, err := s.findFunding(model, fundingId)
	if err != nil {
		return nil, err
	}

	h, err := documents.DeriveResponseHeader(s.tokenRegFinder(), model)
	if err != nil {
		return nil, errors.New("failed to derive response: %v", err)
	}

	return &clientfundingpb.FundingResponse{
		Header: h,
		Data:   nil,
	}, nil

}

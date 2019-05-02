package funding

import (
	"context"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"reflect"
	"strconv"
	"strings"

	"github.com/centrifuge/go-centrifuge/centerrors"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	clientfundingpb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/funding"
)

// Service defines specific functions for extension funding
type Service interface {
	documents.Service
	DeriveFromPayload(ctx context.Context, req *clientfundingpb.FundingCreatePayload, identifier []byte) (documents.Model, error)
	DeriveFundingResponse(model documents.Model, payload *clientfundingpb.FundingCreatePayload) (*clientfundingpb.FundingResponse, error)
}

// service implements Service and handles all funding related persistence and validations
type service struct {
	documents.Service
	tokenRegFinder func() documents.TokenRegistry
}

const fundingLabel = "centrifuge_funding"
const fundingFieldKey = "centrifuge_funding[{IDX}]."
const fundingIdx = "{IDX}"

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

func generateKey(idx ,fieldName string) string {
	return strings.Replace(fundingFieldKey, fundingIdx, idx , -1) +fieldName
}

func keyFromJsonTag(idx, jsonTag string) (key string, err error) {
	correctJsonParts := 2
	jsonKeyIdx := 0

	// example `json:"days,omitempty"`
	jsonParts := strings.Split(jsonTag,",")
	if len(jsonParts) == correctJsonParts {
		return generateKey(idx,jsonParts[jsonKeyIdx]), nil

	}
	return key, ErrNoFundingField

}

func defineFundingIdx(model documents.Model) (attr documents.Attribute,err error) {
	key, err := documents.AttrKeyFromLabel(fundingLabel)
	if err != nil {
		return attr, err
	}

	if !model.AttributeExists(key) {
		return documents.NewAttribute(fundingLabel,documents.AttrString,"0")

	}

	attr ,err = model.GetAttribute(key)
	if err != nil {
		return attr ,err
	}

	idxInt, err := strconv.Atoi(attr.Value.Str)
	if err != nil {
		return attr ,err
	}

	if idxInt < 0 {
		return attr, ErrFundingIndex
	}

	newIdx, err := documents.AttrValFromString(documents.AttrString,strconv.Itoa(idxInt+1))
	if err != nil {
		return attr ,err
	}

	attr.Value = newIdx
	return attr, nil

}

func createAttributesList(current documents.Model, req *clientfundingpb.FundingCreatePayload) ([]documents.Attribute, error) {
	var attributes []documents.Attribute

	idx, err := defineFundingIdx(current)
	if err != nil {
		return nil,err
	}

	// define id
	req.Data.AgreementId = hexutil.Encode(utils.RandomSlice(32))

	// add updated idx
	attributes = append(attributes,idx)

	types := reflect.TypeOf(*req.Data)
	values := reflect.ValueOf(*req.Data)
	for i := 0; i < types.NumField(); i++ {

		jsonKey := types.Field(i).Tag.Get("json")
		key, err := keyFromJsonTag(idx.Value.Str, jsonKey)
		if err != nil {
			continue
		}

		value := values.Field(i).Interface().(string)
		attrType := types.Field(i).Type.String()

		attr, err := documents.NewAttribute(key, documents.AttributeType(attrType), value)
		if err != nil {
			return nil, err
		}

		attributes = append(attributes,attr)

	}

	return attributes, nil
}

func (s service) DeriveFromPayload(ctx context.Context, req *clientfundingpb.FundingCreatePayload, identifier []byte) (documents.Model, error) {
	model, err := s.GetCurrentVersion(ctx, identifier)
	if err != nil {
		apiLog.Error(err)
		return nil, centerrors.Wrap(err, "document not found")
	}

	attributes, err := createAttributesList(model,req)
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

// DeriveFundingResponse returns create response from the added funding
func (s service) DeriveFundingResponse(model documents.Model, payload *clientfundingpb.FundingCreatePayload) (*clientfundingpb.FundingResponse, error) {
	h, err := documents.DeriveResponseHeader(s.tokenRegFinder(), model)
	if err != nil {
		return nil, errors.New("failed to derive response: %v", err)
	}

	return &clientfundingpb.FundingResponse{
		Header: h,
		Data:   payload.Data,
	}, nil

}

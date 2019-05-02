package funding

import (
	"context"
	"fmt"
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

const fundingKey = "centrifuge_funding"
const fundingFieldKey = "centrifuge_funding[{IDX}]."
const fundingIdx = "{IDX}"

func generateKey(idx int, fieldName string) string {
	return strings.Replace(fundingFieldKey, fundingIdx, strconv.Itoa(idx) , -1) +fieldName
}

func keyFromJsonTag(idx int, jsonTag string) (key string, err error) {
	correctJsonParts := 2
	jsonKeyIdx := 0

	// example `json:"days,omitempty"`
	jsonParts := strings.Split(jsonTag,",")
	if len(jsonParts) == correctJsonParts {
		return generateKey(idx,jsonParts[jsonKeyIdx]), nil

	}
	return key, ErrNoFundingField

}

func defineFundingIdx(model documents.Model) (int, error) {
	_,_,_,idx,err := model.GetAttribute(fundingKey)
	if err != nil { // todo replace with Exists method
		return 0,nil
	}

	idxInt, err := strconv.Atoi(idx)
	if err != nil {
		return -1,err
	}

	if idxInt < 0 {
		return -1, ErrFundingIndex
	}

	return idxInt+1,nil

}

func createAttributeMap(current documents.Model, req *clientfundingpb.FundingCreatePayload) (map[string]int, error) {

	var attributes map[string]int

	idx, err := defineFundingIdx(current)
	if err != nil {
		return nil,err
	}

	req.Data.AgreementId = hexutil.Encode(utils.RandomSlice(32))
	
	types := reflect.TypeOf(*req.Data)
	values := reflect.ValueOf(*req.Data)
	for i := 0; i < types.NumField(); i++ {

		jsonKey := types.Field(i).Tag.Get("json")
		key, err := keyFromJsonTag(idx, jsonKey)
		if err != nil {
			continue
		}

		fmt.Println(key)
		fmt.Println(types.Field(i).Type.String())
		fmt.Println(values.Field(i).Interface())

	}

	// todo return attributeMap
	return attributes, nil
}

func (s service) DeriveFromPayload(ctx context.Context, req *clientfundingpb.FundingCreatePayload, identifier []byte) (documents.Model, error) {
	current, err := s.GetCurrentVersion(ctx, identifier)
	if err != nil {
		apiLog.Error(err)
		return nil, centerrors.Wrap(err, "document not found")
	}

	model, err := current.PrepareNewVersionWithExistingData()
	if err != nil {
		return nil, err
	}

	_, err = createAttributeMap(model,req)
	if err != nil {
		return nil, err
	}

	// todo validate funding payload
	// validate funding payload
	/*validator := CreateValidator()

	err = validator.Validate(current, model)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentInvalid, err)
	}
	*/

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

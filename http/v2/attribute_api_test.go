//go:build unit

package v2

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/http/coreapi"
	"github.com/centrifuge/go-centrifuge/pending"
	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/common"
	genericUtils "github.com/centrifuge/go-centrifuge/testingutils/generic"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types/codec"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandler_AddSignedAttribute(t *testing.T) {
	service, mocks := getServiceWithMocks(t)
	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	documentID := utils.RandomSlice(32)
	previousVersion := utils.RandomSlice(32)
	currentVersion := utils.RandomSlice(32)
	nextVersion := utils.RandomSlice(32)
	documentIDParam := hexutil.Encode(documentID)

	testURL := fmt.Sprintf("%s/documents/%s/signed_attribute", testServer.URL, documentIDParam)

	payloadLabel := "label"
	payloadType := "string"
	payloadValue := "value"

	payload := SignedAttributeRequest{
		Label:   payloadLabel,
		Type:    payloadType,
		Payload: payloadValue,
	}

	b, err := json.Marshal(payload)
	assert.NoError(t, err)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL, bytes.NewReader(b))
	assert.NoError(t, err)

	attributeType := documents.AttributeType(payloadType)
	attributeVal, err := documents.AttrValFromString(attributeType, payloadValue)
	assert.NoError(t, err)

	valBytes, err := attributeVal.ToBytes()
	assert.NoError(t, err)

	documentMock := documents.NewDocumentMock(t)

	genericUtils.GetMock[*pending.ServiceMock](mocks).On(
		"AddSignedAttribute",
		mock.Anything,
		documentID,
		payloadLabel,
		valBytes,
		attributeType,
	).Return(documentMock, nil).Once()

	mockDocumentResponseCalls(
		t,
		documentMock,
		payloadLabel,
		attributeVal,
		documentID,
		previousVersion,
		currentVersion,
		nextVersion,
	)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	resBody, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)

	var documentRes coreapi.DocumentResponse

	err = json.Unmarshal(resBody, &documentRes)
	assert.NoError(t, err)

	assertDocumentResponse(t, documentMock, documentRes)
}

func TestHandler_AddSignedAttribute_InvalidDocumentIDParam(t *testing.T) {
	service, _ := getServiceWithMocks(t)
	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	documentIDParam := "invalid-doc-id-param"

	testURL := fmt.Sprintf("%s/documents/%s/signed_attribute", testServer.URL, documentIDParam)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL, nil)
	assert.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_AddSignedAttribute_InvalidPayload(t *testing.T) {
	service, _ := getServiceWithMocks(t)
	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	documentID := utils.RandomSlice(32)
	documentIDParam := hexutil.Encode(documentID)

	testURL := fmt.Sprintf("%s/documents/%s/signed_attribute", testServer.URL, documentIDParam)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL, nil)
	assert.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_AddSignedAttribute_InvalidAttribute(t *testing.T) {
	service, _ := getServiceWithMocks(t)
	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	documentID := utils.RandomSlice(32)
	documentIDParam := hexutil.Encode(documentID)

	testURL := fmt.Sprintf("%s/documents/%s/signed_attribute", testServer.URL, documentIDParam)

	payloadLabel := "label"
	payloadType := "invalid-attr-type"
	payloadValue := "value"

	payload := SignedAttributeRequest{
		Label:   payloadLabel,
		Type:    payloadType,
		Payload: payloadValue,
	}

	b, err := json.Marshal(payload)
	assert.NoError(t, err)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL, bytes.NewReader(b))
	assert.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_AddSignedAttribute_PendingDocServerError(t *testing.T) {
	service, mocks := getServiceWithMocks(t)
	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	documentID := utils.RandomSlice(32)
	documentIDParam := hexutil.Encode(documentID)

	testURL := fmt.Sprintf("%s/documents/%s/signed_attribute", testServer.URL, documentIDParam)

	payloadLabel := "label"
	payloadType := "string"
	payloadValue := "value"

	payload := SignedAttributeRequest{
		Label:   payloadLabel,
		Type:    payloadType,
		Payload: payloadValue,
	}

	b, err := json.Marshal(payload)
	assert.NoError(t, err)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL, bytes.NewReader(b))
	assert.NoError(t, err)

	attributeType := documents.AttributeType(payloadType)
	attributeVal, err := documents.AttrValFromString(attributeType, payloadValue)
	assert.NoError(t, err)

	valBytes, err := attributeVal.ToBytes()
	assert.NoError(t, err)

	genericUtils.GetMock[*pending.ServiceMock](mocks).On(
		"AddSignedAttribute",
		mock.Anything,
		documentID,
		payloadLabel,
		valBytes,
		attributeType,
	).Return(nil, errors.New("error")).Once()

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_AddSignedAttribute_ResponseMappingError(t *testing.T) {
	service, mocks := getServiceWithMocks(t)
	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	documentID := utils.RandomSlice(32)
	documentIDParam := hexutil.Encode(documentID)

	testURL := fmt.Sprintf("%s/documents/%s/signed_attribute", testServer.URL, documentIDParam)

	payloadLabel := "label"
	payloadType := "string"
	payloadValue := "value"

	payload := SignedAttributeRequest{
		Label:   payloadLabel,
		Type:    payloadType,
		Payload: payloadValue,
	}

	b, err := json.Marshal(payload)
	assert.NoError(t, err)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL, bytes.NewReader(b))
	assert.NoError(t, err)

	attributeType := documents.AttributeType(payloadType)
	attributeVal, err := documents.AttrValFromString(attributeType, payloadValue)
	assert.NoError(t, err)

	valBytes, err := attributeVal.ToBytes()
	assert.NoError(t, err)

	documentMock := documents.NewDocumentMock(t)

	genericUtils.GetMock[*pending.ServiceMock](mocks).On(
		"AddSignedAttribute",
		mock.Anything,
		documentID,
		payloadLabel,
		valBytes,
		attributeType,
	).Return(documentMock, nil).Once()

	documentData := "document-data"
	documentMock.On("GetData").
		Return(documentData)

	documentScheme := "scheme"
	documentMock.On("Scheme").
		Return(documentScheme)

	attrs := []documents.Attribute{
		{
			KeyLabel: payloadLabel,
			Key:      documents.AttrKey(utils.RandomByte32()),
			Value:    attributeVal,
		},
	}
	documentMock.On("GetAttributes").
		Return(attrs)

	// Return error to ensure that response mapping fails.
	documentMock.On("GetCollaborators").
		Return(documents.CollaboratorsAccess{}, errors.New("error"))

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
}

func TestHandler_AddAttributes(t *testing.T) {
	service, mocks := getServiceWithMocks(t)
	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	documentID := utils.RandomSlice(32)
	previousVersion := utils.RandomSlice(32)
	currentVersion := utils.RandomSlice(32)
	nextVersion := utils.RandomSlice(32)
	documentIDParam := hexutil.Encode(documentID)

	testURL := fmt.Sprintf("%s/documents/%s/attributes", testServer.URL, documentIDParam)

	payloadLabel := "label"
	payloadType := "string"
	payloadValue := "value"

	payload := make(coreapi.AttributeMapRequest)
	payload[payloadLabel] = coreapi.AttributeRequest{
		Type:  payloadType,
		Value: payloadValue,
	}

	b, err := json.Marshal(payload)
	assert.NoError(t, err)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL, bytes.NewReader(b))
	assert.NoError(t, err)

	attrs, err := toDocumentAttributes(payload)
	assert.NoError(t, err)

	documentMock := documents.NewDocumentMock(t)

	genericUtils.GetMock[*pending.ServiceMock](mocks).On(
		"AddAttributes",
		mock.Anything,
		documentID,
		attrs,
	).Return(documentMock, nil).Once()

	attributeType := documents.AttributeType(payloadType)
	attributeVal, err := documents.AttrValFromString(attributeType, payloadValue)
	assert.NoError(t, err)

	mockDocumentResponseCalls(
		t,
		documentMock,
		payloadLabel,
		attributeVal,
		documentID,
		previousVersion,
		currentVersion,
		nextVersion,
	)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	resBody, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)

	var documentRes coreapi.DocumentResponse

	err = json.Unmarshal(resBody, &documentRes)
	assert.NoError(t, err)

	assertDocumentResponse(t, documentMock, documentRes)
}

func TestHandler_AddAttributes_InvalidDocumentIDParam(t *testing.T) {
	service, _ := getServiceWithMocks(t)
	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()
	testURL := fmt.Sprintf("%s/documents/invalid_document-id/attributes", testServer.URL)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL, nil)
	assert.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_AddAttributes_InvalidPayload(t *testing.T) {
	service, _ := getServiceWithMocks(t)
	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	documentID := utils.RandomSlice(32)
	documentIDParam := hexutil.Encode(documentID)

	testURL := fmt.Sprintf("%s/documents/%s/attributes", testServer.URL, documentIDParam)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL, nil)
	assert.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_AddAttributes_InvalidAttribute(t *testing.T) {
	service, _ := getServiceWithMocks(t)
	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	documentID := utils.RandomSlice(32)
	documentIDParam := hexutil.Encode(documentID)

	testURL := fmt.Sprintf("%s/documents/%s/attributes", testServer.URL, documentIDParam)

	payloadLabel := "label"
	payloadType := "invalidType"
	payloadValue := "value"

	payload := make(coreapi.AttributeMapRequest)
	payload[payloadLabel] = coreapi.AttributeRequest{
		Type:  payloadType,
		Value: payloadValue,
	}

	b, err := json.Marshal(payload)
	assert.NoError(t, err)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL, bytes.NewReader(b))
	assert.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_AddAttributes_PendingDocSrvError(t *testing.T) {
	service, mocks := getServiceWithMocks(t)
	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	documentID := utils.RandomSlice(32)
	documentIDParam := hexutil.Encode(documentID)

	testURL := fmt.Sprintf("%s/documents/%s/attributes", testServer.URL, documentIDParam)

	payloadLabel := "label"
	payloadType := "string"
	payloadValue := "value"

	payload := make(coreapi.AttributeMapRequest)
	payload[payloadLabel] = coreapi.AttributeRequest{
		Type:  payloadType,
		Value: payloadValue,
	}

	b, err := json.Marshal(payload)
	assert.NoError(t, err)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL, bytes.NewReader(b))
	assert.NoError(t, err)

	attrs, err := toDocumentAttributes(payload)
	assert.NoError(t, err)

	genericUtils.GetMock[*pending.ServiceMock](mocks).On(
		"AddAttributes",
		mock.Anything,
		documentID,
		attrs,
	).Return(nil, errors.New("error")).Once()

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_AddAttributes_ResponseMappingError(t *testing.T) {
	service, mocks := getServiceWithMocks(t)
	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	documentID := utils.RandomSlice(32)
	documentIDParam := hexutil.Encode(documentID)

	testURL := fmt.Sprintf("%s/documents/%s/attributes", testServer.URL, documentIDParam)

	payloadLabel := "label"
	payloadType := "string"
	payloadValue := "value"

	payload := make(coreapi.AttributeMapRequest)
	payload[payloadLabel] = coreapi.AttributeRequest{
		Type:  payloadType,
		Value: payloadValue,
	}

	b, err := json.Marshal(payload)
	assert.NoError(t, err)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL, bytes.NewReader(b))
	assert.NoError(t, err)

	attrs, err := toDocumentAttributes(payload)
	assert.NoError(t, err)

	documentMock := documents.NewDocumentMock(t)

	genericUtils.GetMock[*pending.ServiceMock](mocks).On(
		"AddAttributes",
		mock.Anything,
		documentID,
		attrs,
	).Return(documentMock, nil).Once()

	attributeType := documents.AttributeType(payloadType)
	attributeVal, err := documents.AttrValFromString(attributeType, payloadValue)
	assert.NoError(t, err)

	documentData := "document-data"
	documentMock.On("GetData").
		Return(documentData)

	documentScheme := "scheme"
	documentMock.On("Scheme").
		Return(documentScheme)

	attrs = []documents.Attribute{
		{
			KeyLabel: payloadLabel,
			Key:      utils.RandomByte32(),
			Value:    attributeVal,
		},
	}
	documentMock.On("GetAttributes").
		Return(attrs)

	// Return error to ensure that response mapping fails.
	documentMock.On("GetCollaborators").
		Return(documents.CollaboratorsAccess{}, errors.New("error"))

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
}

func TestHandler_DeleteAttribute(t *testing.T) {
	service, mocks := getServiceWithMocks(t)
	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	documentID := utils.RandomSlice(32)
	documentIDParam := hexutil.Encode(documentID)

	attributeKey := utils.RandomSlice(32)

	testURL := fmt.Sprintf("%s/documents/%s/attributes/%s", testServer.URL, documentIDParam, hexutil.Encode(attributeKey))

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, testURL, nil)
	assert.NoError(t, err)

	attrKey, err := documents.AttrKeyFromBytes(attributeKey)
	assert.NoError(t, err)

	documentMock := documents.NewDocumentMock(t)

	genericUtils.GetMock[*pending.ServiceMock](mocks).On("DeleteAttribute", mock.Anything, documentID, attrKey).
		Return(documentMock, nil).
		Once()

	mockDocumentResponseCalls(
		t,
		documentMock,
		"attr-key-label",
		documents.AttrVal{
			Type: "string",
			Str:  "value",
		},
		utils.RandomSlice(32),
		utils.RandomSlice(32),
		utils.RandomSlice(32),
		utils.RandomSlice(32),
	)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	resBody, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)

	var documentRes coreapi.DocumentResponse

	err = json.Unmarshal(resBody, &documentRes)
	assert.NoError(t, err)

	assertDocumentResponse(t, documentMock, documentRes)
}

func TestHandler_DeleteAttribute_InvalidDocumentIDParam(t *testing.T) {
	service, _ := getServiceWithMocks(t)
	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	documentIDParam := "invalid-doc-id"

	attributeKey := utils.RandomSlice(32)

	testURL := fmt.Sprintf("%s/documents/%s/attributes/%s", testServer.URL, documentIDParam, hexutil.Encode(attributeKey))

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, testURL, nil)
	assert.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_DeleteAttribute_InvalidAttributeKeyParam(t *testing.T) {
	service, _ := getServiceWithMocks(t)
	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	documentID := utils.RandomSlice(32)
	documentIDParam := hexutil.Encode(documentID)

	attributeKey := "invalid-attribute-key"

	testURL := fmt.Sprintf("%s/documents/%s/attributes/%s", testServer.URL, documentIDParam, attributeKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, testURL, nil)
	assert.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_DeleteAttribute_InvalidAttributeKey(t *testing.T) {
	service, _ := getServiceWithMocks(t)
	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	documentID := utils.RandomSlice(32)
	documentIDParam := hexutil.Encode(documentID)

	// Invalid byte slice length for an AttributeKey
	attributeKey := utils.RandomSlice(33)

	testURL := fmt.Sprintf("%s/documents/%s/attributes/%s", testServer.URL, documentIDParam, hexutil.Encode(attributeKey))

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, testURL, nil)
	assert.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_DeleteAttribute_PendingDocSrvError(t *testing.T) {
	service, mocks := getServiceWithMocks(t)
	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	documentID := utils.RandomSlice(32)
	documentIDParam := hexutil.Encode(documentID)

	attributeKey := utils.RandomSlice(32)

	testURL := fmt.Sprintf("%s/documents/%s/attributes/%s", testServer.URL, documentIDParam, hexutil.Encode(attributeKey))

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, testURL, nil)
	assert.NoError(t, err)

	attrKey, err := documents.AttrKeyFromBytes(attributeKey)
	assert.NoError(t, err)

	genericUtils.GetMock[*pending.ServiceMock](mocks).On("DeleteAttribute", mock.Anything, documentID, attrKey).
		Return(nil, errors.New("error")).
		Once()

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_DeleteAttribute_ResponseMappingError(t *testing.T) {
	service, mocks := getServiceWithMocks(t)
	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	documentID := utils.RandomSlice(32)
	documentIDParam := hexutil.Encode(documentID)

	attributeKey := utils.RandomSlice(32)

	testURL := fmt.Sprintf("%s/documents/%s/attributes/%s", testServer.URL, documentIDParam, hexutil.Encode(attributeKey))

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, testURL, nil)
	assert.NoError(t, err)

	attrKey, err := documents.AttrKeyFromBytes(attributeKey)
	assert.NoError(t, err)

	documentMock := documents.NewDocumentMock(t)

	genericUtils.GetMock[*pending.ServiceMock](mocks).On("DeleteAttribute", mock.Anything, documentID, attrKey).
		Return(documentMock, nil).
		Once()

	documentData := "document-data"
	documentMock.On("GetData").
		Return(documentData)

	documentScheme := "scheme"
	documentMock.On("Scheme").
		Return(documentScheme)

	documentMock.On("GetAttributes").
		Return(nil)

	// Return error to ensure that response mapping fails.
	documentMock.On("GetCollaborators").
		Return(documents.CollaboratorsAccess{}, errors.New("error"))

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
}

func assertDocumentResponse(t *testing.T, documentMock *documents.DocumentMock, documentRes coreapi.DocumentResponse) {
	assert.Equal(t, documentMock.Scheme(), documentRes.Scheme)
	assert.Equal(t, documentMock.GetData(), documentRes.Data)

	assertDocumentResponseAttributeMap(t, documentMock, documentRes)
	assertDocumentResponseHeader(t, documentMock, documentRes)

	assert.Equal(t, string(documentMock.GetStatus()), documentRes.Header.Status)
}

func assertDocumentResponseAttributeMap(t *testing.T, documentMock *documents.DocumentMock, documentRes coreapi.DocumentResponse) {
	for _, attribute := range documentMock.GetAttributes() {
		resAttr, ok := documentRes.Attributes[attribute.KeyLabel]
		assert.True(t, ok)

		assert.Equal(t, attribute.Key[:], resAttr.Key.Bytes())
		assert.Equal(t, attribute.Value.Type.String(), resAttr.Type)

		attrValStr, err := attribute.Value.String()
		assert.NoError(t, err)

		assert.Equal(t, attrValStr, resAttr.Value)
	}
}

func assertDocumentResponseHeader(t *testing.T, documentMock *documents.DocumentMock, documentRes coreapi.DocumentResponse) {
	assert.Equal(t, hexutil.Encode(documentMock.ID()), documentRes.Header.DocumentID)
	assert.Equal(t, hexutil.Encode(documentMock.PreviousVersion()), documentRes.Header.PreviousVersionID)
	assert.Equal(t, hexutil.Encode(documentMock.CurrentVersion()), documentRes.Header.VersionID)
	assert.Equal(t, hexutil.Encode(documentMock.NextVersion()), documentRes.Header.NextVersionID)

	author, err := documentMock.Author()
	assert.NoError(t, err)
	assert.Equal(t, author.ToHexString(), documentRes.Header.Author)

	timestamp, err := documentMock.Timestamp()
	assert.NoError(t, err)
	assert.Equal(t, timestamp.UTC().Format(time.RFC3339), documentRes.Header.CreatedAt)

	collabs, err := documentMock.GetCollaborators()
	assert.NoError(t, err)
	assert.Equal(t, collabs.ReadCollaborators, documentRes.Header.ReadAccess)
	assert.Equal(t, collabs.ReadWriteCollaborators, documentRes.Header.WriteAccess)

	assert.Equal(t, convertNFTs(t, documentMock), documentRes.Header.NFTs)

	fp, err := documentMock.CalculateTransitionRulesFingerprint()
	assert.NoError(t, err)
	assert.Equal(t, fp, documentRes.Header.Fingerprint.Bytes())
}

func convertNFTs(t *testing.T, documentMock *documents.DocumentMock) []*coreapi.NFT {
	var res []*coreapi.NFT

	for _, docNFT := range documentMock.NFTs() {
		var collectionID types.U64

		err := codec.Decode(docNFT.CollectionId, &collectionID)
		assert.NoError(t, err)

		var itemID types.U128

		err = codec.Decode(docNFT.ItemId, &itemID)
		assert.NoError(t, err)

		res = append(res, &coreapi.NFT{
			CollectionID: collectionID,
			ItemID:       itemID.String(),
		})
	}

	return res
}

func mockDocumentResponseCalls(
	t *testing.T,
	documentMock *documents.DocumentMock,
	attributeKeyLabel string,
	attributeValue documents.AttrVal,
	documentID []byte,
	previousVersion []byte,
	currentVersion []byte,
	nextVersion []byte,
) {
	documentData := "document-data"
	documentMock.On("GetData").
		Return(documentData)

	documentScheme := "scheme"
	documentMock.On("Scheme").
		Return(documentScheme)

	attrs := []documents.Attribute{
		{
			KeyLabel: attributeKeyLabel,
			Key:      documents.AttrKey(utils.RandomByte32()),
			Value:    attributeValue,
		},
	}
	documentMock.On("GetAttributes").
		Return(attrs)

	collab1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)
	collabs := documents.CollaboratorsAccess{
		ReadWriteCollaborators: []*types.AccountID{collab1},
	}
	documentMock.On("GetCollaborators").
		Return(collabs, nil)

	author, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)
	documentMock.On("Author").
		Return(author, nil)

	timestamp := time.Now()
	documentMock.On("Timestamp").
		Return(timestamp, nil)

	transitionRulesFingerprint := utils.RandomSlice(32)
	documentMock.On("CalculateTransitionRulesFingerprint").
		Return(transitionRulesFingerprint, nil)

	collectionID1 := types.U64(1111)
	encodedCollectionID1, err := codec.Encode(collectionID1)
	assert.NoError(t, err)

	itemID1 := types.NewU128(*big.NewInt(2222))
	encodedItemID1, err := codec.Encode(itemID1)
	assert.NoError(t, err)

	nfts := []*coredocumentpb.NFT{
		{
			CollectionId: encodedCollectionID1,
			ItemId:       encodedItemID1,
		},
	}
	documentMock.On("NFTs").
		Return(nfts)

	documentMock.On("ID").
		Return(documentID)
	documentMock.On("PreviousVersion").
		Return(previousVersion)
	documentMock.On("CurrentVersion").
		Return(currentVersion)
	documentMock.On("NextVersion").
		Return(nextVersion)
	documentMock.On("GetStatus").
		Return(documents.Committing)
}

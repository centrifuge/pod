//go:build unit

package v2

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	proofspb "github.com/centrifuge/precise-proofs/proofs/proto"

	"github.com/centrifuge/gocelery/v2"

	"github.com/centrifuge/go-centrifuge/errors"

	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/http/coreapi"
	"github.com/centrifuge/go-centrifuge/pending"
	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/common"
	mockUtils "github.com/centrifuge/go-centrifuge/testingutils/mocks"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/go-centrifuge/utils/byteutils"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandler_CreateDocument(t *testing.T) {
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
	documentScheme := "test-scheme"

	readCollab1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)
	writeCollab1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	readAccess := []*types.AccountID{readCollab1}
	writeAccess := []*types.AccountID{writeCollab1}

	documentData := utils.RandomSlice(32)

	documentAttrs := coreapi.AttributeMapRequest{
		"label1": {
			Type:          "string",
			Value:         "value",
			MonetaryValue: nil,
		},
	}

	payload := CreateDocumentRequest{
		CreateDocumentRequest: coreapi.CreateDocumentRequest{
			Scheme:      documentScheme,
			ReadAccess:  readAccess,
			WriteAccess: writeAccess,
			Data:        documentData,
			Attributes:  documentAttrs,
		},
		DocumentID: byteutils.OptionalHex{
			HexBytes: documentID,
		},
	}

	b, err := json.Marshal(payload)
	assert.NoError(t, err)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, testServer.URL+"/documents", bytes.NewReader(b))
	assert.NoError(t, err)

	docPayload, err := toDocumentsPayload(payload.CreateDocumentRequest, documentID)
	assert.NoError(t, err)

	documentMock := documents.NewDocumentMock(t)

	mockUtils.GetMock[*pending.ServiceMock](mocks).On("Create", mock.Anything, docPayload).
		Return(documentMock, nil).
		Once()

	mockDocumentResponseCalls(
		t,
		documentMock,
		"label1",
		documents.AttrVal{
			Type: "string",
			Str:  "value",
		},
		documentID,
		utils.RandomSlice(32),
		utils.RandomSlice(32),
		utils.RandomSlice(32),
	)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, res.StatusCode)

	resBody, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)

	var documentRes coreapi.DocumentResponse

	err = json.Unmarshal(resBody, &documentRes)
	assert.NoError(t, err)

	assertDocumentResponse(t, documentMock, documentRes)
}

func TestHandler_CreateDocument_InvalidPayload(t *testing.T) {
	service, _ := getServiceWithMocks(t)
	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, testServer.URL+"/documents", nil)
	assert.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_CreateDocument_InvalidPayloadAttrs(t *testing.T) {
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
	documentScheme := "test-scheme"

	readCollab1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)
	writeCollab1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	readAccess := []*types.AccountID{readCollab1}
	writeAccess := []*types.AccountID{writeCollab1}

	documentData := utils.RandomSlice(32)

	documentAttrs := coreapi.AttributeMapRequest{
		"label1": {
			Type:          "invalidType",
			Value:         "value",
			MonetaryValue: nil,
		},
	}

	payload := CreateDocumentRequest{
		CreateDocumentRequest: coreapi.CreateDocumentRequest{
			Scheme:      documentScheme,
			ReadAccess:  readAccess,
			WriteAccess: writeAccess,
			Data:        documentData,
			Attributes:  documentAttrs,
		},
		DocumentID: byteutils.OptionalHex{
			HexBytes: documentID,
		},
	}

	b, err := json.Marshal(payload)
	assert.NoError(t, err)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, testServer.URL+"/documents", bytes.NewReader(b))
	assert.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_CreateDocument_PendingDocSrvError(t *testing.T) {
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
	documentScheme := "test-scheme"

	readCollab1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)
	writeCollab1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	readAccess := []*types.AccountID{readCollab1}
	writeAccess := []*types.AccountID{writeCollab1}

	documentData := utils.RandomSlice(32)

	documentAttrs := coreapi.AttributeMapRequest{
		"label1": {
			Type:          "string",
			Value:         "value",
			MonetaryValue: nil,
		},
	}

	payload := CreateDocumentRequest{
		CreateDocumentRequest: coreapi.CreateDocumentRequest{
			Scheme:      documentScheme,
			ReadAccess:  readAccess,
			WriteAccess: writeAccess,
			Data:        documentData,
			Attributes:  documentAttrs,
		},
		DocumentID: byteutils.OptionalHex{
			HexBytes: documentID,
		},
	}

	b, err := json.Marshal(payload)
	assert.NoError(t, err)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, testServer.URL+"/documents", bytes.NewReader(b))
	assert.NoError(t, err)

	docPayload, err := toDocumentsPayload(payload.CreateDocumentRequest, documentID)
	assert.NoError(t, err)

	mockUtils.GetMock[*pending.ServiceMock](mocks).On("Create", mock.Anything, docPayload).
		Return(nil, errors.New("error")).
		Once()

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_CreateDocument_ResponseMappingError(t *testing.T) {
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
	documentScheme := "test-scheme"

	readCollab1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)
	writeCollab1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	readAccess := []*types.AccountID{readCollab1}
	writeAccess := []*types.AccountID{writeCollab1}

	documentData := utils.RandomSlice(32)

	documentAttrs := coreapi.AttributeMapRequest{
		"label1": {
			Type:          "string",
			Value:         "value",
			MonetaryValue: nil,
		},
	}

	payload := CreateDocumentRequest{
		CreateDocumentRequest: coreapi.CreateDocumentRequest{
			Scheme:      documentScheme,
			ReadAccess:  readAccess,
			WriteAccess: writeAccess,
			Data:        documentData,
			Attributes:  documentAttrs,
		},
		DocumentID: byteutils.OptionalHex{
			HexBytes: documentID,
		},
	}

	b, err := json.Marshal(payload)
	assert.NoError(t, err)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, testServer.URL+"/documents", bytes.NewReader(b))
	assert.NoError(t, err)

	docPayload, err := toDocumentsPayload(payload.CreateDocumentRequest, documentID)
	assert.NoError(t, err)

	documentMock := documents.NewDocumentMock(t)

	mockUtils.GetMock[*pending.ServiceMock](mocks).On("Create", mock.Anything, docPayload).
		Return(documentMock, nil).
		Once()

	documentMock.On("GetData").
		Return(documentData)

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

func TestHandler_CloneDocument(t *testing.T) {
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
	documentScheme := "test-scheme"

	payload := CloneDocumentRequest{
		Scheme: documentScheme,
	}

	b, err := json.Marshal(payload)
	assert.NoError(t, err)

	testURL := fmt.Sprintf("%s/documents/%s/clone", testServer.URL, hexutil.Encode(documentID))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL, bytes.NewReader(b))
	assert.NoError(t, err)

	clonePayload := documents.ClonePayload{
		Scheme:     documentScheme,
		TemplateID: documentID,
	}

	documentMock := documents.NewDocumentMock(t)

	mockUtils.GetMock[*pending.ServiceMock](mocks).On("Clone", mock.Anything, clonePayload).
		Return(documentMock, nil).
		Once()

	mockDocumentResponseCalls(
		t,
		documentMock,
		"label1",
		documents.AttrVal{
			Type: "string",
			Str:  "value",
		},
		documentID,
		utils.RandomSlice(32),
		utils.RandomSlice(32),
		utils.RandomSlice(32),
	)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, res.StatusCode)

	resBody, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)

	var documentRes coreapi.DocumentResponse

	err = json.Unmarshal(resBody, &documentRes)
	assert.NoError(t, err)

	assertDocumentResponse(t, documentMock, documentRes)
}

func TestHandler_CloneDocument_InvalidDocIDParam(t *testing.T) {
	service, _ := getServiceWithMocks(t)
	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	testURL := fmt.Sprintf("%s/documents/%s/clone", testServer.URL, "invalid-doc-id")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL, nil)
	assert.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_CloneDocument_InvalidPayload(t *testing.T) {
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

	testURL := fmt.Sprintf("%s/documents/%s/clone", testServer.URL, hexutil.Encode(documentID))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL, nil)
	assert.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_CloneDocument_PendingDocSrvError(t *testing.T) {
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
	documentScheme := "test-scheme"

	payload := CloneDocumentRequest{
		Scheme: documentScheme,
	}

	b, err := json.Marshal(payload)
	assert.NoError(t, err)

	testURL := fmt.Sprintf("%s/documents/%s/clone", testServer.URL, hexutil.Encode(documentID))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL, bytes.NewReader(b))
	assert.NoError(t, err)

	clonePayload := documents.ClonePayload{
		Scheme:     documentScheme,
		TemplateID: documentID,
	}

	mockUtils.GetMock[*pending.ServiceMock](mocks).On("Clone", mock.Anything, clonePayload).
		Return(nil, errors.New("error")).
		Once()

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_CloneDocument_ResponseMappingError(t *testing.T) {
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
	documentScheme := "test-scheme"

	payload := CloneDocumentRequest{
		Scheme: documentScheme,
	}

	b, err := json.Marshal(payload)
	assert.NoError(t, err)

	testURL := fmt.Sprintf("%s/documents/%s/clone", testServer.URL, hexutil.Encode(documentID))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL, bytes.NewReader(b))
	assert.NoError(t, err)

	clonePayload := documents.ClonePayload{
		Scheme:     documentScheme,
		TemplateID: documentID,
	}

	documentMock := documents.NewDocumentMock(t)

	mockUtils.GetMock[*pending.ServiceMock](mocks).On("Clone", mock.Anything, clonePayload).
		Return(documentMock, nil).
		Once()

	documentMock.On("GetData").
		Return(documentData)

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

func TestHandler_UpdateDocument(t *testing.T) {
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
	documentScheme := "test-scheme"

	readCollab1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)
	writeCollab1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	readAccess := []*types.AccountID{readCollab1}
	writeAccess := []*types.AccountID{writeCollab1}

	documentData := utils.RandomSlice(32)

	documentAttrs := coreapi.AttributeMapRequest{
		"label1": {
			Type:          "string",
			Value:         "value",
			MonetaryValue: nil,
		},
	}

	payload := UpdateDocumentRequest{
		CreateDocumentRequest: coreapi.CreateDocumentRequest{
			Scheme:      documentScheme,
			ReadAccess:  readAccess,
			WriteAccess: writeAccess,
			Data:        documentData,
			Attributes:  documentAttrs,
		},
	}

	b, err := json.Marshal(payload)
	assert.NoError(t, err)

	testURL := fmt.Sprintf("%s/documents/%s", testServer.URL, hexutil.Encode(documentID))

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, testURL, bytes.NewReader(b))
	assert.NoError(t, err)

	docPayload, err := toDocumentsPayload(payload.CreateDocumentRequest, documentID)
	assert.NoError(t, err)

	documentMock := documents.NewDocumentMock(t)

	mockUtils.GetMock[*pending.ServiceMock](mocks).On(
		"Update",
		mock.Anything,
		docPayload,
	).Return(documentMock, nil).Once()

	mockDocumentResponseCalls(
		t,
		documentMock,
		"label1",
		documents.AttrVal{
			Type: "string",
			Str:  "value",
		},
		documentID,
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

func TestHandler_UpdateDocument_InvalidDocIDParam(t *testing.T) {
	service, _ := getServiceWithMocks(t)
	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	documentID := "invalid-doc-id-param"

	testURL := fmt.Sprintf("%s/documents/%s", testServer.URL, documentID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, testURL, nil)
	assert.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_UpdateDocument_InvalidPayload(t *testing.T) {
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

	testURL := fmt.Sprintf("%s/documents/%s", testServer.URL, hexutil.Encode(documentID))

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, testURL, nil)
	assert.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_UpdateDocument_InvalidPayloadAttrs(t *testing.T) {
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
	documentScheme := "test-scheme"

	readCollab1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)
	writeCollab1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	readAccess := []*types.AccountID{readCollab1}
	writeAccess := []*types.AccountID{writeCollab1}

	documentData := utils.RandomSlice(32)

	documentAttrs := coreapi.AttributeMapRequest{
		"label1": {
			Type: "invalid-type",
		},
	}

	payload := UpdateDocumentRequest{
		CreateDocumentRequest: coreapi.CreateDocumentRequest{
			Scheme:      documentScheme,
			ReadAccess:  readAccess,
			WriteAccess: writeAccess,
			Data:        documentData,
			Attributes:  documentAttrs,
		},
	}

	b, err := json.Marshal(payload)
	assert.NoError(t, err)

	testURL := fmt.Sprintf("%s/documents/%s", testServer.URL, hexutil.Encode(documentID))

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, testURL, bytes.NewReader(b))
	assert.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_UpdateDocument_PendingDocSrvError(t *testing.T) {
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
	documentScheme := "test-scheme"

	readCollab1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)
	writeCollab1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	readAccess := []*types.AccountID{readCollab1}
	writeAccess := []*types.AccountID{writeCollab1}

	documentData := utils.RandomSlice(32)

	documentAttrs := coreapi.AttributeMapRequest{
		"label1": {
			Type:          "string",
			Value:         "value",
			MonetaryValue: nil,
		},
	}

	payload := UpdateDocumentRequest{
		CreateDocumentRequest: coreapi.CreateDocumentRequest{
			Scheme:      documentScheme,
			ReadAccess:  readAccess,
			WriteAccess: writeAccess,
			Data:        documentData,
			Attributes:  documentAttrs,
		},
	}

	b, err := json.Marshal(payload)
	assert.NoError(t, err)

	testURL := fmt.Sprintf("%s/documents/%s", testServer.URL, hexutil.Encode(documentID))

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, testURL, bytes.NewReader(b))
	assert.NoError(t, err)

	docPayload, err := toDocumentsPayload(payload.CreateDocumentRequest, documentID)
	assert.NoError(t, err)

	mockUtils.GetMock[*pending.ServiceMock](mocks).On(
		"Update",
		mock.Anything,
		docPayload,
	).Return(nil, errors.New("error")).Once()

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_UpdateDocument_ResponseMappingError(t *testing.T) {
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
	documentScheme := "test-scheme"

	readCollab1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)
	writeCollab1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	readAccess := []*types.AccountID{readCollab1}
	writeAccess := []*types.AccountID{writeCollab1}

	documentData := utils.RandomSlice(32)

	documentAttrs := coreapi.AttributeMapRequest{
		"label1": {
			Type:          "string",
			Value:         "value",
			MonetaryValue: nil,
		},
	}

	payload := UpdateDocumentRequest{
		CreateDocumentRequest: coreapi.CreateDocumentRequest{
			Scheme:      documentScheme,
			ReadAccess:  readAccess,
			WriteAccess: writeAccess,
			Data:        documentData,
			Attributes:  documentAttrs,
		},
	}

	b, err := json.Marshal(payload)
	assert.NoError(t, err)

	testURL := fmt.Sprintf("%s/documents/%s", testServer.URL, hexutil.Encode(documentID))

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, testURL, bytes.NewReader(b))
	assert.NoError(t, err)

	docPayload, err := toDocumentsPayload(payload.CreateDocumentRequest, documentID)
	assert.NoError(t, err)

	documentMock := documents.NewDocumentMock(t)

	mockUtils.GetMock[*pending.ServiceMock](mocks).On(
		"Update",
		mock.Anything,
		docPayload,
	).Return(documentMock, nil).Once()

	documentMock.On("GetData").
		Return(documentData)

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

func TestHandler_Commit(t *testing.T) {
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

	testURL := fmt.Sprintf("%s/documents/%s/commit", testServer.URL, hexutil.Encode(documentID))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL, nil)
	assert.NoError(t, err)

	documentMock := documents.NewDocumentMock(t)
	jobID := gocelery.JobID{}

	mockUtils.GetMock[*pending.ServiceMock](mocks).On(
		"Commit",
		mock.Anything,
		documentID,
	).Return(documentMock, jobID, nil).Once()

	mockDocumentResponseCalls(
		t,
		documentMock,
		"label1",
		documents.AttrVal{
			Type: "string",
			Str:  "value",
		},
		documentID,
		utils.RandomSlice(32),
		utils.RandomSlice(32),
		utils.RandomSlice(32),
	)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, res.StatusCode)

	resBody, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)

	var documentRes coreapi.DocumentResponse

	err = json.Unmarshal(resBody, &documentRes)
	assert.NoError(t, err)

	assertDocumentResponse(t, documentMock, documentRes)
}

func TestHandler_Commit_InvalidDocIDParam(t *testing.T) {
	service, _ := getServiceWithMocks(t)
	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	documentID := "invalid-doc-id-param"

	testURL := fmt.Sprintf("%s/documents/%s/commit", testServer.URL, documentID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL, nil)
	assert.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_Commit_PendingDocSrvError(t *testing.T) {
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

	testURL := fmt.Sprintf("%s/documents/%s/commit", testServer.URL, hexutil.Encode(documentID))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL, nil)
	assert.NoError(t, err)

	jobID := gocelery.JobID{}

	mockUtils.GetMock[*pending.ServiceMock](mocks).On(
		"Commit",
		mock.Anything,
		documentID,
	).Return(nil, jobID, errors.New("error")).Once()

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_Commit_ResponseMappingError(t *testing.T) {
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

	testURL := fmt.Sprintf("%s/documents/%s/commit", testServer.URL, hexutil.Encode(documentID))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL, nil)
	assert.NoError(t, err)

	documentMock := documents.NewDocumentMock(t)
	jobID := gocelery.JobID{}

	mockUtils.GetMock[*pending.ServiceMock](mocks).On(
		"Commit",
		mock.Anything,
		documentID,
	).Return(documentMock, jobID, nil).Once()

	documentMock.On("GetData").
		Return(documentData)

	documentMock.On("Scheme").
		Return("scheme")

	documentMock.On("GetAttributes").
		Return(nil)

	// Return error to ensure that response mapping fails.
	documentMock.On("GetCollaborators").
		Return(documents.CollaboratorsAccess{}, errors.New("error"))

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
}

func TestHandler_GetPendingDocument(t *testing.T) {
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

	testURL := fmt.Sprintf("%s/documents/%s/pending", testServer.URL, hexutil.Encode(documentID))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
	assert.NoError(t, err)

	documentMock := documents.NewDocumentMock(t)

	mockUtils.GetMock[*pending.ServiceMock](mocks).On(
		"Get",
		mock.Anything,
		documentID,
		documents.Pending,
	).Return(documentMock, nil).Once()

	mockDocumentResponseCalls(
		t,
		documentMock,
		"label1",
		documents.AttrVal{
			Type: "string",
			Str:  "value",
		},
		documentID,
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

func TestHandler_GetPendingDocument_InvalidDocIDParam(t *testing.T) {
	service, _ := getServiceWithMocks(t)
	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	documentID := "invalid-doc-id-param"

	testURL := fmt.Sprintf("%s/documents/%s/pending", testServer.URL, documentID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
	assert.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_GetPendingDocument_PendingDocSrvError(t *testing.T) {
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

	testURL := fmt.Sprintf("%s/documents/%s/pending", testServer.URL, hexutil.Encode(documentID))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
	assert.NoError(t, err)

	mockUtils.GetMock[*pending.ServiceMock](mocks).On(
		"Get",
		mock.Anything,
		documentID,
		documents.Pending,
	).Return(nil, errors.New("error")).Once()

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, res.StatusCode)
}

func TestHandler_GetPendingDocument_ResponseMappingError(t *testing.T) {
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

	testURL := fmt.Sprintf("%s/documents/%s/pending", testServer.URL, hexutil.Encode(documentID))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
	assert.NoError(t, err)

	documentMock := documents.NewDocumentMock(t)

	mockUtils.GetMock[*pending.ServiceMock](mocks).On(
		"Get",
		mock.Anything,
		documentID,
		documents.Pending,
	).Return(documentMock, nil).Once()

	documentMock.On("GetData").
		Return(documentData)

	documentMock.On("Scheme").
		Return("scheme")

	documentMock.On("GetAttributes").
		Return(nil)

	// Return error to ensure that response mapping fails.
	documentMock.On("GetCollaborators").
		Return(documents.CollaboratorsAccess{}, errors.New("error"))

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
}

func TestHandler_GetCommittedDocument(t *testing.T) {
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

	testURL := fmt.Sprintf("%s/documents/%s/committed", testServer.URL, hexutil.Encode(documentID))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
	assert.NoError(t, err)

	documentMock := documents.NewDocumentMock(t)

	mockUtils.GetMock[*pending.ServiceMock](mocks).On(
		"Get",
		mock.Anything,
		documentID,
		documents.Committed,
	).Return(documentMock, nil).Once()

	mockDocumentResponseCalls(
		t,
		documentMock,
		"label1",
		documents.AttrVal{
			Type: "string",
			Str:  "value",
		},
		documentID,
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

func TestHandler_GetCommittedDocument_InvalidDocIDParam(t *testing.T) {
	service, _ := getServiceWithMocks(t)
	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	documentID := "invalid-doc-id-param"

	testURL := fmt.Sprintf("%s/documents/%s/committed", testServer.URL, documentID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
	assert.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_GetCommittedDocument_PendingDocSrvError(t *testing.T) {
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

	testURL := fmt.Sprintf("%s/documents/%s/committed", testServer.URL, hexutil.Encode(documentID))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
	assert.NoError(t, err)

	mockUtils.GetMock[*pending.ServiceMock](mocks).On(
		"Get",
		mock.Anything,
		documentID,
		documents.Committed,
	).Return(nil, errors.New("error")).Once()

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, res.StatusCode)
}

func TestHandler_GetCommittedDocument_ResponseMappingError(t *testing.T) {
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

	testURL := fmt.Sprintf("%s/documents/%s/committed", testServer.URL, hexutil.Encode(documentID))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
	assert.NoError(t, err)

	documentMock := documents.NewDocumentMock(t)

	mockUtils.GetMock[*pending.ServiceMock](mocks).On(
		"Get",
		mock.Anything,
		documentID,
		documents.Committed,
	).Return(documentMock, nil).Once()

	documentMock.On("GetData").
		Return(documentData)

	documentMock.On("Scheme").
		Return("scheme")

	documentMock.On("GetAttributes").
		Return(nil)

	// Return error to ensure that response mapping fails.
	documentMock.On("GetCollaborators").
		Return(documents.CollaboratorsAccess{}, errors.New("error"))

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
}

func TestHandler_GetDocumentVersion(t *testing.T) {
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
	documentVersion := utils.RandomSlice(32)

	testURL := fmt.Sprintf(
		"%s/documents/%s/versions/%s",
		testServer.URL,
		hexutil.Encode(documentID),
		hexutil.Encode(documentVersion),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
	assert.NoError(t, err)

	documentMock := documents.NewDocumentMock(t)

	mockUtils.GetMock[*pending.ServiceMock](mocks).On(
		"GetVersion",
		mock.Anything,
		documentID,
		documentVersion,
	).Return(documentMock, nil).Once()

	mockDocumentResponseCalls(
		t,
		documentMock,
		"label1",
		documents.AttrVal{
			Type: "string",
			Str:  "value",
		},
		documentID,
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

func TestHandler_GetDocumentVersion_InvalidDocIDParam(t *testing.T) {
	service, _ := getServiceWithMocks(t)
	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	documentID := "invalid-doc-id-param"
	documentVersion := utils.RandomSlice(32)

	testURL := fmt.Sprintf(
		"%s/documents/%s/versions/%s",
		testServer.URL,
		documentID,
		hexutil.Encode(documentVersion),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
	assert.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_GetDocumentVersion_InvalidDocVersionParam(t *testing.T) {
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
	documentVersion := "invalid-doc-id-param"

	testURL := fmt.Sprintf(
		"%s/documents/%s/versions/%s",
		testServer.URL,
		hexutil.Encode(documentID),
		documentVersion,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
	assert.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_GetDocumentVersion_PendingDocSrvError(t *testing.T) {
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
	documentVersion := utils.RandomSlice(32)

	testURL := fmt.Sprintf(
		"%s/documents/%s/versions/%s",
		testServer.URL,
		hexutil.Encode(documentID),
		hexutil.Encode(documentVersion),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
	assert.NoError(t, err)

	mockUtils.GetMock[*pending.ServiceMock](mocks).On(
		"GetVersion",
		mock.Anything,
		documentID,
		documentVersion,
	).Return(nil, errors.New("error")).Once()

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, res.StatusCode)
}

func TestHandler_GetDocumentVersion_ResponseMappingError(t *testing.T) {
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
	documentVersion := utils.RandomSlice(32)

	testURL := fmt.Sprintf(
		"%s/documents/%s/versions/%s",
		testServer.URL,
		hexutil.Encode(documentID),
		hexutil.Encode(documentVersion),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
	assert.NoError(t, err)

	documentMock := documents.NewDocumentMock(t)

	mockUtils.GetMock[*pending.ServiceMock](mocks).On(
		"GetVersion",
		mock.Anything,
		documentID,
		documentVersion,
	).Return(documentMock, nil).Once()

	documentMock.On("GetData").
		Return(documentData)

	documentMock.On("Scheme").
		Return("scheme")

	documentMock.On("GetAttributes").
		Return(nil)

	// Return error to ensure that response mapping fails.
	documentMock.On("GetCollaborators").
		Return(documents.CollaboratorsAccess{}, errors.New("error"))

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
}

func TestHandler_RemoveCollaborators(t *testing.T) {
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

	collab1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collborators := []*types.AccountID{collab1}

	payload := RemoveCollaboratorsRequest{
		Collaborators: collborators,
	}

	b, err := json.Marshal(payload)
	assert.NoError(t, err)

	testURL := fmt.Sprintf("%s/documents/%s/collaborators", testServer.URL, hexutil.Encode(documentID))

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, testURL, bytes.NewReader(b))
	assert.NoError(t, err)

	documentMock := documents.NewDocumentMock(t)

	mockUtils.GetMock[*pending.ServiceMock](mocks).On(
		"RemoveCollaborators",
		mock.Anything,
		documentID,
		collborators,
	).Return(documentMock, nil).Once()

	mockDocumentResponseCalls(
		t,
		documentMock,
		"label1",
		documents.AttrVal{
			Type: "string",
			Str:  "value",
		},
		documentID,
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

func TestHandler_RemoveCollaborators_InvalidDocIDParam(t *testing.T) {
	service, _ := getServiceWithMocks(t)
	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	documentID := "invalid-doc-id-param"

	testURL := fmt.Sprintf("%s/documents/%s/collaborators", testServer.URL, documentID)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, testURL, nil)
	assert.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_RemoveCollaborators_InvalidPayload(t *testing.T) {
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

	testURL := fmt.Sprintf("%s/documents/%s/collaborators", testServer.URL, hexutil.Encode(documentID))

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, testURL, nil)
	assert.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_RemoveCollaborators_PendingDocSrvError(t *testing.T) {
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

	collab1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collborators := []*types.AccountID{collab1}

	payload := RemoveCollaboratorsRequest{
		Collaborators: collborators,
	}

	b, err := json.Marshal(payload)
	assert.NoError(t, err)

	testURL := fmt.Sprintf("%s/documents/%s/collaborators", testServer.URL, hexutil.Encode(documentID))

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, testURL, bytes.NewReader(b))
	assert.NoError(t, err)

	mockUtils.GetMock[*pending.ServiceMock](mocks).On(
		"RemoveCollaborators",
		mock.Anything,
		documentID,
		collborators,
	).Return(nil, errors.New("error")).Once()

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_RemoveCollaborators_ResponseMappingError(t *testing.T) {
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

	collab1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collborators := []*types.AccountID{collab1}

	payload := RemoveCollaboratorsRequest{
		Collaborators: collborators,
	}

	b, err := json.Marshal(payload)
	assert.NoError(t, err)

	testURL := fmt.Sprintf("%s/documents/%s/collaborators", testServer.URL, hexutil.Encode(documentID))

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, testURL, bytes.NewReader(b))
	assert.NoError(t, err)

	documentMock := documents.NewDocumentMock(t)

	mockUtils.GetMock[*pending.ServiceMock](mocks).On(
		"RemoveCollaborators",
		mock.Anything,
		documentID,
		collborators,
	).Return(documentMock, nil).Once()

	documentMock.On("GetData").
		Return(documentData)

	documentMock.On("Scheme").
		Return("scheme")

	documentMock.On("GetAttributes").
		Return(nil)

	// Return error to ensure that response mapping fails.
	documentMock.On("GetCollaborators").
		Return(documents.CollaboratorsAccess{}, errors.New("error"))

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
}

func TestHandler_GenerateProofs(t *testing.T) {
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

	fields := []string{
		"field1",
		"field2",
	}

	payload := coreapi.ProofsRequest{
		Fields: fields,
	}

	b, err := json.Marshal(payload)
	assert.NoError(t, err)

	testURL := fmt.Sprintf("%s/documents/%s/proofs", testServer.URL, hexutil.Encode(documentID))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL, bytes.NewReader(b))
	assert.NoError(t, err)

	documentProofs := &documents.DocumentProof{
		DocumentID: utils.RandomSlice(32),
		VersionID:  utils.RandomSlice(32),
		State:      "state",
		FieldProofs: []*proofspb.Proof{
			{
				Property: &proofspb.Proof_CompactName{
					CompactName: utils.RandomSlice(32),
				},
				Value: utils.RandomSlice(32),
				Salt:  utils.RandomSlice(32),
				Hash:  utils.RandomSlice(32),
				Hashes: []*proofspb.MerkleHash{
					{
						Left:  utils.RandomSlice(32),
						Right: utils.RandomSlice(32),
					},
				},
				SortedHashes: [][]byte{
					utils.RandomSlice(32),
				},
			},
		},
		SigningRoot:    utils.RandomSlice(32),
		SignaturesRoot: utils.RandomSlice(32),
	}

	mockUtils.GetMock[*documents.ServiceMock](mocks).On(
		"CreateProofs",
		mock.Anything,
		documentID,
		fields,
	).Return(documentProofs, nil).Once()

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	resBody, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)

	var proofRes coreapi.ProofsResponse

	err = json.Unmarshal(resBody, &proofRes)
	assert.NoError(t, err)

	expectedProofs := coreapi.ConvertProofs(documentProofs)

	assert.Equal(t, expectedProofs.Header, proofRes.Header)
	assert.Equal(t, expectedProofs.FieldProofs, proofRes.FieldProofs)
}

func TestHandler_GenerateProofs_InvalidDocIDParam(t *testing.T) {
	service, _ := getServiceWithMocks(t)
	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	documentID := "invalid-doc-id-param"

	testURL := fmt.Sprintf("%s/documents/%s/proofs", testServer.URL, documentID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL, nil)
	assert.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_GenerateProofs_InvalidPayload(t *testing.T) {
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

	testURL := fmt.Sprintf("%s/documents/%s/proofs", testServer.URL, hexutil.Encode(documentID))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL, nil)
	assert.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_GenerateProofs_DocumentSrvError(t *testing.T) {
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

	fields := []string{
		"field1",
		"field2",
	}

	payload := coreapi.ProofsRequest{
		Fields: fields,
	}

	b, err := json.Marshal(payload)
	assert.NoError(t, err)

	testURL := fmt.Sprintf("%s/documents/%s/proofs", testServer.URL, hexutil.Encode(documentID))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL, bytes.NewReader(b))
	assert.NoError(t, err)

	mockUtils.GetMock[*documents.ServiceMock](mocks).On(
		"CreateProofs",
		mock.Anything,
		documentID,
		fields,
	).Return(nil, errors.New("error")).Once()

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_GenerateProofsForVersion(t *testing.T) {
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
	documentVersion := utils.RandomSlice(32)

	fields := []string{
		"field1",
		"field2",
	}

	payload := coreapi.ProofsRequest{
		Fields: fields,
	}

	b, err := json.Marshal(payload)
	assert.NoError(t, err)

	testURL := fmt.Sprintf(
		"%s/documents/%s/versions/%s/proofs",
		testServer.URL,
		hexutil.Encode(documentID),
		hexutil.Encode(documentVersion),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL, bytes.NewReader(b))
	assert.NoError(t, err)

	documentProofs := &documents.DocumentProof{
		DocumentID: utils.RandomSlice(32),
		VersionID:  utils.RandomSlice(32),
		State:      "state",
		FieldProofs: []*proofspb.Proof{
			{
				Property: &proofspb.Proof_CompactName{
					CompactName: utils.RandomSlice(32),
				},
				Value: utils.RandomSlice(32),
				Salt:  utils.RandomSlice(32),
				Hash:  utils.RandomSlice(32),
				Hashes: []*proofspb.MerkleHash{
					{
						Left:  utils.RandomSlice(32),
						Right: utils.RandomSlice(32),
					},
				},
				SortedHashes: [][]byte{
					utils.RandomSlice(32),
				},
			},
		},
		SigningRoot:    utils.RandomSlice(32),
		SignaturesRoot: utils.RandomSlice(32),
	}

	mockUtils.GetMock[*documents.ServiceMock](mocks).On(
		"CreateProofsForVersion",
		mock.Anything,
		documentID,
		documentVersion,
		fields,
	).Return(documentProofs, nil).Once()

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	resBody, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)

	var proofRes coreapi.ProofsResponse

	err = json.Unmarshal(resBody, &proofRes)
	assert.NoError(t, err)

	expectedProofs := coreapi.ConvertProofs(documentProofs)

	assert.Equal(t, expectedProofs.Header, proofRes.Header)
	assert.Equal(t, expectedProofs.FieldProofs, proofRes.FieldProofs)
}

func TestHandler_GenerateProofsForVersion_InvalidDocIDParam(t *testing.T) {
	service, _ := getServiceWithMocks(t)
	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	documentID := "invalid-doc-id-param"
	documentVersion := utils.RandomSlice(32)

	testURL := fmt.Sprintf(
		"%s/documents/%s/versions/%s/proofs",
		testServer.URL,
		documentID,
		hexutil.Encode(documentVersion),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL, nil)
	assert.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_GenerateProofsForVersion_InvalidDocVersionParam(t *testing.T) {
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
	documentVersion := "invalid-doc-version-param"

	testURL := fmt.Sprintf(
		"%s/documents/%s/versions/%s/proofs",
		testServer.URL,
		hexutil.Encode(documentID),
		documentVersion,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL, nil)
	assert.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_GenerateProofsForVersion_InvalidPayload(t *testing.T) {
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
	documentVersion := utils.RandomSlice(32)

	testURL := fmt.Sprintf(
		"%s/documents/%s/versions/%s/proofs",
		testServer.URL,
		hexutil.Encode(documentID),
		hexutil.Encode(documentVersion),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL, nil)
	assert.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_GenerateProofsForVersion_DocumentSrvError(t *testing.T) {
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
	documentVersion := utils.RandomSlice(32)

	fields := []string{
		"field1",
		"field2",
	}

	payload := coreapi.ProofsRequest{
		Fields: fields,
	}

	b, err := json.Marshal(payload)
	assert.NoError(t, err)

	testURL := fmt.Sprintf(
		"%s/documents/%s/versions/%s/proofs",
		testServer.URL,
		hexutil.Encode(documentID),
		hexutil.Encode(documentVersion),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL, bytes.NewReader(b))
	assert.NoError(t, err)

	mockUtils.GetMock[*documents.ServiceMock](mocks).On(
		"CreateProofsForVersion",
		mock.Anything,
		documentID,
		documentVersion,
		fields,
	).Return(nil, errors.New("error")).Once()

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func documentData() map[string]interface{} {
	return map[string]interface{}{}
}

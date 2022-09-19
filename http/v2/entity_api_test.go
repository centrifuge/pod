//go:build unit

package v2

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/centrifuge/go-centrifuge/documents/entityrelationship"

	"github.com/centrifuge/go-centrifuge/errors"

	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/http/coreapi"

	"github.com/centrifuge/go-centrifuge/documents/entity"
	mockUtils "github.com/centrifuge/go-centrifuge/testingutils/mocks"
	"github.com/stretchr/testify/mock"

	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
)

func TestHandler_GetEntityThroughRelationship(t *testing.T) {
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

	testURL := fmt.Sprintf("%s/relationships/%s/entity", testServer.URL, hexutil.Encode(documentID))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
	assert.NoError(t, err)

	documentMock := documents.NewDocumentMock(t)

	mockUtils.GetMock[*entity.ServiceMock](mocks).On(
		"GetEntityByRelationship",
		mock.Anything,
		documentID,
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

func TestHandler_GetEntityThroughRelationship_InvalidDocIDParam(t *testing.T) {
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

	testURL := fmt.Sprintf("%s/relationships/%s/entity", testServer.URL, documentID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
	assert.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_GetEntityThroughRelationship_EntityServiceError(t *testing.T) {
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

	testURL := fmt.Sprintf("%s/relationships/%s/entity", testServer.URL, hexutil.Encode(documentID))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
	assert.NoError(t, err)

	mockUtils.GetMock[*entity.ServiceMock](mocks).On(
		"GetEntityByRelationship",
		mock.Anything,
		documentID,
	).Return(nil, errors.New("error")).Once()

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, res.StatusCode)
}

func TestHandler_GetEntityRelationships(t *testing.T) {
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

	testURL := fmt.Sprintf("%s/entities/%s/relationships", testServer.URL, hexutil.Encode(documentID))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
	assert.NoError(t, err)

	documentMock1 := documents.NewDocumentMock(t)
	documentMock2 := documents.NewDocumentMock(t)

	entityRelationships := []documents.Document{
		documentMock1,
		documentMock2,
	}

	mockUtils.GetMock[*entityrelationship.ServiceMock](mocks).On(
		"GetEntityRelationships",
		mock.Anything,
		documentID,
	).Return(entityRelationships, nil).Once()

	mockDocumentResponseCalls(
		t,
		documentMock1,
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

	mockDocumentResponseCalls(
		t,
		documentMock2,
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

	var documentRes []coreapi.DocumentResponse

	err = json.Unmarshal(resBody, &documentRes)
	assert.NoError(t, err)

	assertDocumentResponse(t, documentMock1, documentRes[0])
	assertDocumentResponse(t, documentMock2, documentRes[1])
}

func TestHandler_GetEntityRelationships_InvalidDocIDParam(t *testing.T) {
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

	testURL := fmt.Sprintf("%s/entities/%s/relationships", testServer.URL, documentID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
	assert.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_GetEntityRelationships_EntityRelationshipSrvError(t *testing.T) {
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

	testURL := fmt.Sprintf("%s/entities/%s/relationships", testServer.URL, hexutil.Encode(documentID))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
	assert.NoError(t, err)

	mockUtils.GetMock[*entityrelationship.ServiceMock](mocks).On(
		"GetEntityRelationships",
		mock.Anything,
		documentID,
	).Return(nil, errors.New("error")).Once()

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, res.StatusCode)
}

func TestHandler_GetEntityRelationships_ResponseMappingError(t *testing.T) {
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

	testURL := fmt.Sprintf("%s/entities/%s/relationships", testServer.URL, hexutil.Encode(documentID))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
	assert.NoError(t, err)

	documentMock1 := documents.NewDocumentMock(t)
	documentMock2 := documents.NewDocumentMock(t)

	entityRelationships := []documents.Document{
		documentMock1,
		documentMock2,
	}

	mockUtils.GetMock[*entityrelationship.ServiceMock](mocks).On(
		"GetEntityRelationships",
		mock.Anything,
		documentID,
	).Return(entityRelationships, nil).Once()

	documentMock1.On("GetData").
		Return(documentData)

	documentMock1.On("Scheme").
		Return("scheme")

	documentMock1.On("GetAttributes").
		Return(nil)

	// Return error to ensure that response mapping fails.
	documentMock1.On("GetCollaborators").
		Return(documents.CollaboratorsAccess{}, errors.New("error"))

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
}

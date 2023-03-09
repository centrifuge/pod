//go:build unit

package v2

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	"github.com/centrifuge/pod/pending"
	testingcommons "github.com/centrifuge/pod/testingutils/common"
	genericUtils "github.com/centrifuge/pod/testingutils/generic"
	"github.com/centrifuge/pod/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandler_GetRole(t *testing.T) {
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
	roleID := utils.RandomSlice(32)

	testURL := fmt.Sprintf(
		"%s/documents/%s/roles/%s",
		testServer.URL,
		hexutil.Encode(documentID),
		hexutil.Encode(roleID),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
	assert.NoError(t, err)

	role := &coredocumentpb.Role{
		RoleKey: utils.RandomSlice(32),
		Collaborators: [][]byte{
			utils.RandomSlice(32),
		},
		Nfts: [][]byte{
			utils.RandomSlice(32),
		},
	}

	genericUtils.GetMock[*pending.ServiceMock](mocks).On(
		"GetRole",
		mock.Anything,
		documentID,
		roleID,
	).Return(role, nil).Once()

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	resBody, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)

	expectedRoleRes := toClientRole(role)

	var roleRes Role

	err = json.Unmarshal(resBody, &roleRes)
	assert.NoError(t, err)

	assert.Equal(t, expectedRoleRes, roleRes)
}

func TestHandler_GetRole_InvalidDocIDParam(t *testing.T) {
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
	roleID := utils.RandomSlice(32)

	testURL := fmt.Sprintf(
		"%s/documents/%s/roles/%s",
		testServer.URL,
		documentID,
		hexutil.Encode(roleID),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
	assert.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_GetRole_InvalidRoleIDParam(t *testing.T) {
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
	roleID := "invalid-role-id-param"

	testURL := fmt.Sprintf(
		"%s/documents/%s/roles/%s",
		testServer.URL,
		hexutil.Encode(documentID),
		roleID,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
	assert.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_GetRole_PendingDocSrvError(t *testing.T) {
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
	roleID := utils.RandomSlice(32)

	testURL := fmt.Sprintf(
		"%s/documents/%s/roles/%s",
		testServer.URL,
		hexutil.Encode(documentID),
		hexutil.Encode(roleID),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
	assert.NoError(t, err)

	genericUtils.GetMock[*pending.ServiceMock](mocks).On(
		"GetRole",
		mock.Anything,
		documentID,
		roleID,
	).Return(nil, errors.New("error")).Once()

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, res.StatusCode)
}

func TestHandler_AddRole(t *testing.T) {
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

	roleKey := hexutil.Encode(utils.RandomSlice(32))
	collaborators := []*types.AccountID{
		collab1,
	}

	payload := AddRole{
		Key:           roleKey,
		Collaborators: collaborators,
	}

	b, err := json.Marshal(payload)
	assert.NoError(t, err)

	testURL := fmt.Sprintf(
		"%s/documents/%s/roles",
		testServer.URL,
		hexutil.Encode(documentID),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL, bytes.NewReader(b))
	assert.NoError(t, err)

	role := &coredocumentpb.Role{
		RoleKey: utils.RandomSlice(32),
		Collaborators: [][]byte{
			utils.RandomSlice(32),
		},
		Nfts: [][]byte{
			utils.RandomSlice(32),
		},
	}

	genericUtils.GetMock[*pending.ServiceMock](mocks).On(
		"AddRole",
		mock.Anything,
		documentID,
		roleKey,
		collaborators,
	).Return(role, nil).Once()

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	resBody, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)

	expectedRoleRes := toClientRole(role)

	var roleRes Role

	err = json.Unmarshal(resBody, &roleRes)
	assert.NoError(t, err)

	assert.Equal(t, expectedRoleRes, roleRes)
}

func TestHandler_AddRole_InvalidDocumentIDParam(t *testing.T) {
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

	testURL := fmt.Sprintf(
		"%s/documents/%s/roles",
		testServer.URL,
		documentID,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL, nil)
	assert.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_AddRole_InvalidPayload(t *testing.T) {
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

	testURL := fmt.Sprintf(
		"%s/documents/%s/roles",
		testServer.URL,
		hexutil.Encode(documentID),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL, nil)
	assert.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_AddRole_PendingDocSrvError(t *testing.T) {
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

	roleKey := hexutil.Encode(utils.RandomSlice(32))
	collaborators := []*types.AccountID{
		collab1,
	}

	payload := AddRole{
		Key:           roleKey,
		Collaborators: collaborators,
	}

	b, err := json.Marshal(payload)
	assert.NoError(t, err)

	testURL := fmt.Sprintf(
		"%s/documents/%s/roles",
		testServer.URL,
		hexutil.Encode(documentID),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL, bytes.NewReader(b))
	assert.NoError(t, err)

	genericUtils.GetMock[*pending.ServiceMock](mocks).On(
		"AddRole",
		mock.Anything,
		documentID,
		roleKey,
		collaborators,
	).Return(nil, errors.New("error")).Once()

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_UpdateRole(t *testing.T) {
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
	roleID := utils.RandomSlice(32)

	collab1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collaborators := []*types.AccountID{collab1}

	payload := UpdateRole{
		Collaborators: collaborators,
	}

	b, err := json.Marshal(payload)
	assert.NoError(t, err)

	testURL := fmt.Sprintf(
		"%s/documents/%s/roles/%s",
		testServer.URL,
		hexutil.Encode(documentID),
		hexutil.Encode(roleID),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, testURL, bytes.NewReader(b))
	assert.NoError(t, err)

	role := &coredocumentpb.Role{
		RoleKey: utils.RandomSlice(32),
		Collaborators: [][]byte{
			utils.RandomSlice(32),
		},
		Nfts: [][]byte{
			utils.RandomSlice(32),
		},
	}

	genericUtils.GetMock[*pending.ServiceMock](mocks).On(
		"UpdateRole",
		mock.Anything,
		documentID,
		roleID,
		collaborators,
	).Return(role, nil).Once()

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	resBody, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)

	expectedRoleRes := toClientRole(role)

	var roleRes Role

	err = json.Unmarshal(resBody, &roleRes)
	assert.NoError(t, err)

	assert.Equal(t, expectedRoleRes, roleRes)
}

func TestHandler_UpdateRole_InvalidDocIDParam(t *testing.T) {
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
	roleID := utils.RandomSlice(32)

	testURL := fmt.Sprintf(
		"%s/documents/%s/roles/%s",
		testServer.URL,
		documentID,
		hexutil.Encode(roleID),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, testURL, nil)
	assert.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_UpdateRole_InvalidRoleIDParam(t *testing.T) {
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
	roleID := "invalid-role-id-param"

	testURL := fmt.Sprintf(
		"%s/documents/%s/roles/%s",
		testServer.URL,
		hexutil.Encode(documentID),
		roleID,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, testURL, nil)
	assert.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_UpdateRole_InvalidPayload(t *testing.T) {
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
	roleID := utils.RandomSlice(32)

	testURL := fmt.Sprintf(
		"%s/documents/%s/roles/%s",
		testServer.URL,
		hexutil.Encode(documentID),
		hexutil.Encode(roleID),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, testURL, nil)
	assert.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_UpdateRole_PendingDocSrvError(t *testing.T) {
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
	roleID := utils.RandomSlice(32)

	collab1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	collaborators := []*types.AccountID{collab1}

	payload := UpdateRole{
		Collaborators: collaborators,
	}

	b, err := json.Marshal(payload)
	assert.NoError(t, err)

	testURL := fmt.Sprintf(
		"%s/documents/%s/roles/%s",
		testServer.URL,
		hexutil.Encode(documentID),
		hexutil.Encode(roleID),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, testURL, bytes.NewReader(b))
	assert.NoError(t, err)

	genericUtils.GetMock[*pending.ServiceMock](mocks).On(
		"UpdateRole",
		mock.Anything,
		documentID,
		roleID,
		collaborators,
	).Return(nil, errors.New("error")).Once()

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, res.StatusCode)
}

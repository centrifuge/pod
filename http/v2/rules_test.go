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

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/pod/errors"
	"github.com/centrifuge/pod/pending"
	genericUtils "github.com/centrifuge/pod/testingutils/generic"
	"github.com/centrifuge/pod/utils"
	"github.com/centrifuge/pod/utils/byteutils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandler_AddTransitionRules(t *testing.T) {
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

	payload := pending.AddTransitionRules{
		AttributeRules: []pending.AttributeRule{
			{
				KeyLabel: "label1",
				RoleID:   byteutils.HexBytes(utils.RandomSlice(32)),
			},
		},
		ComputeFieldsRules: []pending.ComputeFieldsRule{
			{
				WASM: byteutils.HexBytes(utils.RandomSlice(32)),
				AttributeLabels: []string{
					"label1",
				},
				TargetAttributeLabel: "target_label1",
			},
		},
	}

	b, err := json.Marshal(payload)
	assert.NoError(t, err)

	testURL := fmt.Sprintf(
		"%s/documents/%s/transition_rules",
		testServer.URL,
		hexutil.Encode(documentID),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL, bytes.NewReader(b))
	assert.NoError(t, err)

	transitionRules := []*coredocumentpb.TransitionRule{
		{
			RuleKey: utils.RandomSlice(32),
			Roles: [][]byte{
				utils.RandomSlice(32),
			},
			MatchType: 0,
			Field:     utils.RandomSlice(32),
			Action:    0,
			ComputeFields: [][]byte{
				utils.RandomSlice(32),
			},
			ComputeTargetField: utils.RandomSlice(32),
			ComputeCode:        utils.RandomSlice(32),
		},
	}

	genericUtils.GetMock[*pending.ServiceMock](mocks).On(
		"AddTransitionRules",
		mock.Anything,
		documentID,
		payload,
	).Return(transitionRules, nil).Once()

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	resBody, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)

	expectedRuleRes := toClientRules(transitionRules)

	var ruleRes TransitionRules

	err = json.Unmarshal(resBody, &ruleRes)
	assert.NoError(t, err)

	assert.Equal(t, expectedRuleRes, ruleRes)
}

func TestHandler_AddTransitionRules_InvalidDocIDParam(t *testing.T) {
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
		"%s/documents/%s/transition_rules",
		testServer.URL,
		documentID,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL, nil)
	assert.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_AddTransitionRules_InvalidPayload(t *testing.T) {
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
		"%s/documents/%s/transition_rules",
		testServer.URL,
		hexutil.Encode(documentID),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL, nil)
	assert.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_AddTransitionRules_PendingDocSrvError(t *testing.T) {
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

	payload := pending.AddTransitionRules{
		AttributeRules: []pending.AttributeRule{
			{
				KeyLabel: "label1",
				RoleID:   byteutils.HexBytes(utils.RandomSlice(32)),
			},
		},
		ComputeFieldsRules: []pending.ComputeFieldsRule{
			{
				WASM: byteutils.HexBytes(utils.RandomSlice(32)),
				AttributeLabels: []string{
					"label1",
				},
				TargetAttributeLabel: "target_label1",
			},
		},
	}

	b, err := json.Marshal(payload)
	assert.NoError(t, err)

	testURL := fmt.Sprintf(
		"%s/documents/%s/transition_rules",
		testServer.URL,
		hexutil.Encode(documentID),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL, bytes.NewReader(b))
	assert.NoError(t, err)

	genericUtils.GetMock[*pending.ServiceMock](mocks).On(
		"AddTransitionRules",
		mock.Anything,
		documentID,
		payload,
	).Return(nil, errors.New("error")).Once()

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_GetTransitionRule(t *testing.T) {
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
	ruleID := utils.RandomSlice(32)

	testURL := fmt.Sprintf(
		"%s/documents/%s/transition_rules/%s",
		testServer.URL,
		hexutil.Encode(documentID),
		hexutil.Encode(ruleID),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
	assert.NoError(t, err)

	transitionRule := &coredocumentpb.TransitionRule{
		RuleKey: utils.RandomSlice(32),
		Roles: [][]byte{
			utils.RandomSlice(32),
		},
		MatchType: 0,
		Field:     utils.RandomSlice(32),
		Action:    0,
		ComputeFields: [][]byte{
			utils.RandomSlice(32),
		},
		ComputeTargetField: utils.RandomSlice(32),
		ComputeCode:        utils.RandomSlice(32),
	}
	genericUtils.GetMock[*pending.ServiceMock](mocks).On(
		"GetTransitionRule",
		mock.Anything,
		documentID,
		ruleID,
	).Return(transitionRule, nil).Once()

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	resBody, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)

	expectedRuleRes := toClientRule(transitionRule)

	var ruleRes TransitionRule

	err = json.Unmarshal(resBody, &ruleRes)
	assert.NoError(t, err)

	assert.Equal(t, expectedRuleRes, ruleRes)
}

func TestHandler_GetTransitionRule_InvalidDocIDParam(t *testing.T) {
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
	ruleID := utils.RandomSlice(32)

	testURL := fmt.Sprintf(
		"%s/documents/%s/transition_rules/%s",
		testServer.URL,
		documentID,
		hexutil.Encode(ruleID),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
	assert.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_GetTransitionRule_InvalidRuleIDParam(t *testing.T) {
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
	ruleID := "invalid-rule-id-param"

	testURL := fmt.Sprintf(
		"%s/documents/%s/transition_rules/%s",
		testServer.URL,
		hexutil.Encode(documentID),
		ruleID,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
	assert.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_GetTransitionRule_PendingDocSrvError(t *testing.T) {
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
	ruleID := utils.RandomSlice(32)

	testURL := fmt.Sprintf(
		"%s/documents/%s/transition_rules/%s",
		testServer.URL,
		hexutil.Encode(documentID),
		hexutil.Encode(ruleID),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
	assert.NoError(t, err)

	genericUtils.GetMock[*pending.ServiceMock](mocks).On(
		"GetTransitionRule",
		mock.Anything,
		documentID,
		ruleID,
	).Return(nil, errors.New("error")).Once()

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, res.StatusCode)
}

func TestHandler_DeleteTransitionRule(t *testing.T) {
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
	ruleID := utils.RandomSlice(32)

	testURL := fmt.Sprintf(
		"%s/documents/%s/transition_rules/%s",
		testServer.URL,
		hexutil.Encode(documentID),
		hexutil.Encode(ruleID),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, testURL, nil)
	assert.NoError(t, err)

	genericUtils.GetMock[*pending.ServiceMock](mocks).On(
		"DeleteTransitionRule",
		mock.Anything,
		documentID,
		ruleID,
	).Return(nil).Once()

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, res.StatusCode)
}

func TestHandler_DeleteTransitionRule_InvalidDocIDParam(t *testing.T) {
	service, _ := getServiceWithMocks(t)
	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	ruleID := utils.RandomSlice(32)

	testURL := fmt.Sprintf(
		"%s/documents/%s/transition_rules/%s",
		testServer.URL,
		"invalid-doc-id-param",
		hexutil.Encode(ruleID),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, testURL, nil)
	assert.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_DeleteTransitionRule_InvalidRuleIDParam(t *testing.T) {
	service, _ := getServiceWithMocks(t)
	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	testURL := fmt.Sprintf(
		"%s/documents/%s/transition_rules/%s",
		testServer.URL,
		hexutil.Encode(utils.RandomSlice(32)),
		"invalid-rule-id-param",
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, testURL, nil)
	assert.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_DeleteTransitionRule_PendingDocSrvError(t *testing.T) {
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
	ruleID := utils.RandomSlice(32)

	testURL := fmt.Sprintf(
		"%s/documents/%s/transition_rules/%s",
		testServer.URL,
		hexutil.Encode(documentID),
		hexutil.Encode(ruleID),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, testURL, nil)
	assert.NoError(t, err)

	genericUtils.GetMock[*pending.ServiceMock](mocks).On(
		"DeleteTransitionRule",
		mock.Anything,
		documentID,
		ruleID,
	).Return(errors.New("error")).Once()

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, res.StatusCode)
}

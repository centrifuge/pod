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
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/entity"
	"github.com/centrifuge/go-centrifuge/documents/entityrelationship"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/http/coreapi"
	v2 "github.com/centrifuge/go-centrifuge/identity/v2"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/pending"
	testingcommons "github.com/centrifuge/go-centrifuge/testingutils/common"
	genericUtils "github.com/centrifuge/go-centrifuge/testingutils/generic"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandler_GenerateAccount(t *testing.T) {
	service, mocks := getServiceWithMocks(t)
	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	randomAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	webhookURL := "https://centrifuge.io"
	precommitEnabled := true
	documentSigningPublicKey := utils.RandomSlice(32)

	payload := coreapi.GenerateAccountPayload{
		Account: coreapi.Account{
			Identity:         randomAccountID,
			WebhookURL:       webhookURL,
			PrecommitEnabled: precommitEnabled,
		},
	}

	b, err := json.Marshal(payload)
	assert.NoError(t, err)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, testServer.URL+"/accounts/generate", bytes.NewReader(b))
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)

	genericUtils.GetMock[*v2.ServiceMock](mocks).On("CreateIdentity", mock.Anything, payload.ToCreateIdentityRequest()).
		Return(accountMock, nil).
		Once()

	accountMock.On("GetIdentity").
		Return(randomAccountID).
		Once()
	accountMock.On("GetWebhookURL").
		Return(webhookURL).
		Once()
	accountMock.On("GetPrecommitEnabled").
		Return(precommitEnabled).
		Once()
	accountMock.On("GetSigningPublicKey").
		Return(documentSigningPublicKey).
		Once()

	pubKey, _, err := testingcommons.GetTestSigningKeys()
	assert.NoError(t, err)

	rawPubKey, err := pubKey.Raw()
	assert.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, res.StatusCode)

	resBody, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)

	var resAccount coreapi.Account

	err = json.Unmarshal(resBody, &resAccount)
	assert.NoError(t, err)

	podOperator, err := genericUtils.GetMock[*config.ServiceMock](mocks).GetPodOperator()
	assert.NoError(t, err)

	assert.Equal(t, randomAccountID, resAccount.Identity)
	assert.Equal(t, webhookURL, resAccount.WebhookURL)
	assert.Equal(t, precommitEnabled, resAccount.PrecommitEnabled)
	assert.Equal(t, documentSigningPublicKey, resAccount.DocumentSigningPublicKey.Bytes())
	assert.Equal(t, rawPubKey, resAccount.P2PPublicSigningKey.Bytes())
	assert.Equal(t, podOperator.GetAccountID(), resAccount.PodOperatorAccountID)
}

func TestHandler_GenerateAccount_RequestBodyErrors(t *testing.T) {
	service, _ := getServiceWithMocks(t)

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	// Nil body

	req, err := http.NewRequest(http.MethodPost, testServer.URL+"/accounts/generate", nil)
	assert.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	// Invalid payload

	body := utils.RandomSlice(32)

	req, err = http.NewRequest(http.MethodPost, testServer.URL+"/accounts/generate", bytes.NewReader(body))
	assert.NoError(t, err)

	res, err = http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, res.StatusCode, http.StatusBadRequest)
}

func TestHandler_GenerateAccount_IdentityServiceError(t *testing.T) {
	service, mocks := getServiceWithMocks(t)
	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	randomAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	webhookURL := "https://centrifuge.io"
	precommitEnabled := true

	payload := coreapi.GenerateAccountPayload{
		Account: coreapi.Account{
			Identity:         randomAccountID,
			WebhookURL:       webhookURL,
			PrecommitEnabled: precommitEnabled,
		},
	}

	b, err := json.Marshal(payload)
	assert.NoError(t, err)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, testServer.URL+"/accounts/generate", bytes.NewReader(b))
	assert.NoError(t, err)

	genericUtils.GetMock[*v2.ServiceMock](mocks).On("CreateIdentity", mock.Anything, payload.ToCreateIdentityRequest()).
		Return(nil, errors.New("error")).
		Once()

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_SignPayload(t *testing.T) {
	service, mocks := getServiceWithMocks(t)
	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	signPayload := utils.RandomSlice(32)

	payload := coreapi.SignRequest{
		Payload: signPayload,
	}

	b, err := json.Marshal(payload)
	assert.NoError(t, err)

	randomAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	testURL := fmt.Sprintf("%s/accounts/%s/sign", testServer.URL, randomAccountID.ToHexString())

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL, bytes.NewReader(b))
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)

	genericUtils.GetMock[*config.ServiceMock](mocks).On("GetAccount", randomAccountID.ToBytes()).
		Return(accountMock, nil).
		Once()

	signatureBytes := utils.RandomSlice(32)
	signaturePubKey := utils.RandomSlice(32)

	signature := &coredocumentpb.Signature{
		SignerId:  randomAccountID.ToBytes(),
		PublicKey: signaturePubKey,
		Signature: signatureBytes,
	}

	accountMock.On("SignMsg", signPayload).
		Return(signature, nil).
		Once()

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	resBody, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)

	var signResponse coreapi.SignResponse

	err = json.Unmarshal(resBody, &signResponse)
	assert.NoError(t, err)

	assert.Equal(t, signPayload, signResponse.Payload.Bytes())
	assert.Equal(t, signaturePubKey, signResponse.PublicKey.Bytes())
	assert.Equal(t, signatureBytes, signResponse.Signature.Bytes())
	assert.Equal(t, randomAccountID.ToBytes(), signResponse.SignerID.Bytes())
}

func TestHandler_SignPayload_RequestBodyErrors(t *testing.T) {
	service, _ := getServiceWithMocks(t)
	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	randomAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	// Invalid accountID
	testURL := fmt.Sprintf("%s/accounts/invalidAccountID/sign", testServer.URL)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL, nil)
	assert.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, res.StatusCode, http.StatusBadRequest)

	// Nil request body
	testURL = fmt.Sprintf("%s/accounts/%s/sign", testServer.URL, randomAccountID.ToHexString())

	req, err = http.NewRequestWithContext(ctx, http.MethodPost, testURL, nil)
	assert.NoError(t, err)

	res, err = http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)

	// Invalid payload
	testURL = fmt.Sprintf("%s/accounts/%s/sign", testServer.URL, randomAccountID.ToHexString())

	req, err = http.NewRequestWithContext(ctx, http.MethodPost, testURL, bytes.NewReader(utils.RandomSlice(32)))
	assert.NoError(t, err)

	res, err = http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_SignPayload_SignError(t *testing.T) {
	service, mocks := getServiceWithMocks(t)
	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	signPayload := utils.RandomSlice(32)

	payload := coreapi.SignRequest{
		Payload: signPayload,
	}

	b, err := json.Marshal(payload)
	assert.NoError(t, err)

	randomAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	testURL := fmt.Sprintf("%s/accounts/%s/sign", testServer.URL, randomAccountID.ToHexString())

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, testURL, bytes.NewReader(b))
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)

	genericUtils.GetMock[*config.ServiceMock](mocks).On("GetAccount", randomAccountID.ToBytes()).
		Return(accountMock, nil).
		Once()

	accountMock.On("SignMsg", signPayload).
		Return(nil, errors.New("error")).
		Once()

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestHandler_GetSelf(t *testing.T) {
	service, mocks := getServiceWithMocks(t)
	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	accountMock := config.NewAccountMock(t)

	// Mimic the auth handler by adding the account to context.
	router.Use(func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			ctx = contextutil.WithAccount(request.Context(), accountMock)

			request = request.WithContext(ctx)

			h.ServeHTTP(writer, request)
		})
	})

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, testServer.URL+"/accounts/self", nil)
	assert.NoError(t, err)

	randomAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	webhookURL := "https://centrifuge.io"
	precommitEnabled := true
	documentSigningPublicKey := utils.RandomSlice(32)

	accountMock.On("GetIdentity").
		Return(randomAccountID).
		Once()
	accountMock.On("GetWebhookURL").
		Return(webhookURL).
		Once()
	accountMock.On("GetPrecommitEnabled").
		Return(precommitEnabled).
		Once()
	accountMock.On("GetSigningPublicKey").
		Return(documentSigningPublicKey).
		Once()

	pubKey, _, err := testingcommons.GetTestSigningKeys()
	assert.NoError(t, err)

	rawPubKey, err := pubKey.Raw()
	assert.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	resBody, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)

	var resAccount coreapi.Account

	err = json.Unmarshal(resBody, &resAccount)
	assert.NoError(t, err)

	podOperator, err := genericUtils.GetMock[*config.ServiceMock](mocks).GetPodOperator()
	assert.NoError(t, err)

	assert.Equal(t, randomAccountID, resAccount.Identity)
	assert.Equal(t, webhookURL, resAccount.WebhookURL)
	assert.Equal(t, precommitEnabled, resAccount.PrecommitEnabled)
	assert.Equal(t, documentSigningPublicKey, resAccount.DocumentSigningPublicKey.Bytes())
	assert.Equal(t, rawPubKey, resAccount.P2PPublicSigningKey.Bytes())
	assert.Equal(t, podOperator.GetAccountID(), resAccount.PodOperatorAccountID)
}

func TestHandler_GetSelf_AccountNotFound(t *testing.T) {
	service, _ := getServiceWithMocks(t)
	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, testServer.URL+"/accounts/self", nil)
	assert.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, res.StatusCode)
}

func TestHandler_GetAccount(t *testing.T) {
	service, mocks := getServiceWithMocks(t)
	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	randomAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	testURL := fmt.Sprintf("%s/accounts/%s", testServer.URL, randomAccountID.ToHexString())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
	assert.NoError(t, err)

	accountMock := config.NewAccountMock(t)

	genericUtils.GetMock[*config.ServiceMock](mocks).On("GetAccount", randomAccountID.ToBytes()).
		Return(accountMock, nil).
		Once()

	webhookURL := "https://centrifuge.io"
	precommitEnabled := true
	documentSigningPublicKey := utils.RandomSlice(32)

	accountMock.On("GetIdentity").
		Return(randomAccountID).
		Once()
	accountMock.On("GetWebhookURL").
		Return(webhookURL).
		Once()
	accountMock.On("GetPrecommitEnabled").
		Return(precommitEnabled).
		Once()
	accountMock.On("GetSigningPublicKey").
		Return(documentSigningPublicKey).
		Once()

	pubKey, _, err := testingcommons.GetTestSigningKeys()
	assert.NoError(t, err)

	rawPubKey, err := pubKey.Raw()
	assert.NoError(t, err)

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	resBody, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)

	var resAccount coreapi.Account

	err = json.Unmarshal(resBody, &resAccount)
	assert.NoError(t, err)

	podOperator, err := genericUtils.GetMock[*config.ServiceMock](mocks).GetPodOperator()
	assert.NoError(t, err)

	assert.Equal(t, randomAccountID, resAccount.Identity)
	assert.Equal(t, webhookURL, resAccount.WebhookURL)
	assert.Equal(t, precommitEnabled, resAccount.PrecommitEnabled)
	assert.Equal(t, documentSigningPublicKey, resAccount.DocumentSigningPublicKey.Bytes())
	assert.Equal(t, rawPubKey, resAccount.P2PPublicSigningKey.Bytes())
	assert.Equal(t, podOperator.GetAccountID(), resAccount.PodOperatorAccountID)
}

func TestHandler_GetAccount_ConfigServiceError(t *testing.T) {
	service, mocks := getServiceWithMocks(t)
	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	randomAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	testURL := fmt.Sprintf("%s/accounts/%s", testServer.URL, randomAccountID.ToHexString())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
	assert.NoError(t, err)

	genericUtils.GetMock[*config.ServiceMock](mocks).On("GetAccount", randomAccountID.ToBytes()).
		Return(nil, errors.New("error")).
		Once()

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, res.StatusCode)
}

func TestHandler_GetAccounts(t *testing.T) {
	service, mocks := getServiceWithMocks(t)
	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, testServer.URL+"/accounts", nil)
	assert.NoError(t, err)

	accountID1, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)
	signingPublicKey1 := utils.RandomSlice(32)
	webhookURL1 := "https://centrifuge.io/1"
	accountID2, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)
	signingPublicKey2 := utils.RandomSlice(32)
	webhookURL2 := "https://centrifuge.io/2"

	accountMock1 := config.NewAccountMock(t)
	accountMock1.On("GetIdentity").
		Return(accountID1)
	accountMock1.On("GetWebhookURL").
		Return(webhookURL1)
	accountMock1.On("GetPrecommitEnabled").
		Return(true)
	accountMock1.On("GetSigningPublicKey").
		Return(signingPublicKey1)
	accountMock2 := config.NewAccountMock(t)
	accountMock2.On("GetIdentity").
		Return(accountID2)
	accountMock2.On("GetWebhookURL").
		Return(webhookURL2)
	accountMock2.On("GetPrecommitEnabled").
		Return(false)
	accountMock2.On("GetSigningPublicKey").
		Return(signingPublicKey2)

	configAccounts := []config.Account{accountMock1, accountMock2}

	genericUtils.GetMock[*config.ServiceMock](mocks).On("GetAccounts").
		Return(configAccounts, nil).
		Once()

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	resBody, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)

	var resAccounts coreapi.Accounts

	err = json.Unmarshal(resBody, &resAccounts)
	assert.NoError(t, err)

	pubKey, _, err := testingcommons.GetTestSigningKeys()
	assert.NoError(t, err)

	rawPubKey, err := pubKey.Raw()
	assert.NoError(t, err)

	podOperator, err := genericUtils.GetMock[*config.ServiceMock](mocks).GetPodOperator()
	assert.NoError(t, err)

	assert.Equal(t, configAccounts[0].GetIdentity(), resAccounts.Data[0].Identity)
	assert.Equal(t, configAccounts[0].GetSigningPublicKey(), resAccounts.Data[0].DocumentSigningPublicKey.Bytes())
	assert.Equal(t, configAccounts[0].GetWebhookURL(), resAccounts.Data[0].WebhookURL)
	assert.Equal(t, configAccounts[0].GetPrecommitEnabled(), resAccounts.Data[0].PrecommitEnabled)
	assert.Equal(t, rawPubKey, resAccounts.Data[0].P2PPublicSigningKey.Bytes())
	assert.Equal(t, podOperator.GetAccountID(), resAccounts.Data[0].PodOperatorAccountID)

	assert.Equal(t, configAccounts[1].GetIdentity(), resAccounts.Data[1].Identity)
	assert.Equal(t, configAccounts[1].GetSigningPublicKey(), resAccounts.Data[1].DocumentSigningPublicKey.Bytes())
	assert.Equal(t, configAccounts[1].GetWebhookURL(), resAccounts.Data[1].WebhookURL)
	assert.Equal(t, configAccounts[1].GetPrecommitEnabled(), resAccounts.Data[1].PrecommitEnabled)
	assert.Equal(t, rawPubKey, resAccounts.Data[1].P2PPublicSigningKey.Bytes())
	assert.Equal(t, podOperator.GetAccountID(), resAccounts.Data[1].PodOperatorAccountID)
}

func TestHandler_GetAccounts_ConfigServiceError(t *testing.T) {
	service, mocks := getServiceWithMocks(t)
	ctx := context.Background()

	serviceContext := map[string]any{
		BootstrappedService: service,
	}

	router := chi.NewRouter()

	Register(serviceContext, router)

	testServer := httptest.NewServer(router)
	defer testServer.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, testServer.URL+"/accounts", nil)
	assert.NoError(t, err)

	genericUtils.GetMock[*config.ServiceMock](mocks).On("GetAccounts").
		Return(nil, errors.New("error")).
		Once()

	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
}

func getServiceWithMocks(t *testing.T) (*Service, []any) {
	pendingDocSrvMock := pending.NewServiceMock(t)
	dispatcherMock := jobs.NewDispatcherMock(t)
	cfgServiceMock := config.NewServiceMock(t)
	entityServiceMock := entity.NewServiceMock(t)
	identityServiceMock := v2.NewServiceMock(t)
	entityRelationshipServiceMock := entityrelationship.NewServiceMock(t)
	documentServiceMock := documents.NewServiceMock(t)

	configMock := config.NewConfigurationMock(t)

	cfgServiceMock.On("GetConfig").
		Return(configMock, nil)

	configMock.On("GetP2PKeyPair").
		Return(testingcommons.TestPublicSigningKeyPath, testingcommons.TestPrivateSigningKeyPath)

	podOperatorMock := config.NewPodOperatorMock(t)

	cfgServiceMock.On("GetPodOperator").
		Return(podOperatorMock, nil)

	podOperatorAccountID, err := testingcommons.GetRandomAccountID()
	assert.NoError(t, err)

	podOperatorMock.On("GetAccountID").
		Return(podOperatorAccountID)

	service, err := NewService(
		pendingDocSrvMock,
		dispatcherMock,
		cfgServiceMock,
		entityServiceMock,
		identityServiceMock,
		entityRelationshipServiceMock,
		documentServiceMock,
	)
	assert.NoError(t, err)

	return service, []any{pendingDocSrvMock,
		dispatcherMock,
		cfgServiceMock,
		entityServiceMock,
		identityServiceMock,
		entityRelationshipServiceMock,
		documentServiceMock,
	}
}

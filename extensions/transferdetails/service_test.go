// +build unit

package transferdetails

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testlogging"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/invoice"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/extensions"
	"github.com/centrifuge/go-centrifuge/httpapi/coreapi"
	"github.com/centrifuge/go-centrifuge/identity/ideth"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/p2p"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	"github.com/centrifuge/go-centrifuge/testingutils/config"
	"github.com/centrifuge/go-centrifuge/testingutils/documents"
	"github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/centrifuge/go-centrifuge/testingutils/testingjobs"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var ctx = map[string]interface{}{}
var cfg config.Configuration
var did = testingidentity.GenerateRandomDID()

func newCoreAPIService(docSrv documents.Service) coreapi.Service {
	return coreapi.NewService(docSrv, nil, nil, nil)
}

func TestMain(m *testing.M) {
	ethClient := new(ethereum.MockEthClient)
	ethClient.On("GetEthClient").Return(nil)
	ctx[ethereum.BootstrappedEthereumClient] = ethClient
	jobMan := &testingjobs.MockJobManager{}
	ctx[jobs.BootstrappedService] = jobMan
	done := make(chan error)
	jobMan.On("ExecuteWithinJob", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(jobs.NilJobID(), done, nil)
	ctx[bootstrap.BootstrappedInvoiceUnpaid] = new(testingdocuments.MockRegistry)
	ibootstrappers := []bootstrap.TestBootstrapper{
		&testlogging.TestLoggingBootstrapper{},
		&config.Bootstrapper{},
		&leveldb.Bootstrapper{},
		&queue.Bootstrapper{},
		&ideth.Bootstrapper{},
		&configstore.Bootstrapper{},
		anchors.Bootstrapper{},
		documents.Bootstrapper{},
		p2p.Bootstrapper{},
		documents.PostBootstrapper{},
		// &Bootstrapper{}, // todo add own bootstrapper
		&queue.Starter{},
	}
	bootstrap.RunTestBootstrappers(ibootstrappers, ctx)
	cfg = ctx[bootstrap.BootstrappedConfig].(config.Configuration)
	cfg.Set("identityId", did.String())
	result := m.Run()
	bootstrap.RunTestTeardown(ibootstrappers)
	os.Exit(result)
}

func TestDeriveFromPayload(t *testing.T) {
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	inv, _ := invoice.CreateInvoiceWithEmbedCD(t, nil, testingidentity.GenerateRandomDID(), nil)

	docSrv := new(testingdocuments.MockService)
	docSrv.On("GetCurrentVersion", mock.Anything, mock.Anything).Return(inv, nil)
	docSrv.On("UpdateModel", mock.Anything, mock.Anything).Return(inv, nil, nil)
	srv := DefaultService(newCoreAPIService(docSrv), nil)
	payload := createTestPayload()
	payload.DocumentID = hexutil.Encode(inv.Document.DocumentIdentifier)

	for i := 0; i < 10; i++ {
		model, _, err := srv.CreateTransferDetail(ctxh, payload)
		assert.NoError(t, err)
		label := fmt.Sprintf("transfer_details[%d].status", i)
		key, err := documents.AttrKeyFromLabel(label)
		assert.NoError(t, err)

		attr, err := model.GetAttribute(key)
		assert.NoError(t, err)
		assert.Equal(t, "open", attr.Value.Str)
	}
}

func TestDeriveTransferResponse(t *testing.T) {
	inv, _ := invoice.CreateInvoiceWithEmbedCD(t, nil, testingidentity.GenerateRandomDID(), nil)

	docSrv := new(testingdocuments.MockService)
	docSrv.On("GetCurrentVersion", mock.Anything, mock.Anything).Return(inv, nil)
	docSrv.On("UpdateModel", mock.Anything, mock.Anything).Return(inv, nil, nil)
	srv := DefaultService(newCoreAPIService(docSrv), nil)
	ctxh := testingconfig.CreateAccountContext(t, cfg)
	for i := 0; i < 10; i++ {
		payload := createTestPayload()
		payload.DocumentID = hexutil.Encode(inv.Document.DocumentIdentifier)
		model, _, err := srv.CreateTransferDetail(context.Background(), payload)
		assert.NoError(t, err)

		tID, err := hexutil.Decode(payload.Data.TransferID)
		assert.NoError(t, err)
		response, _, err := srv.DeriveTransferDetail(ctxh, model, tID)
		assert.NoError(t, err)
		checkResponse(t, payload, &response.Data)
	}

}

func TestService_DeriveTransferListWithNoAttributes(t *testing.T) {
	inv, _ := invoice.CreateInvoiceWithEmbedCD(t, nil, testingidentity.GenerateRandomDID(), nil)
	docSrv := new(testingdocuments.MockService)
	srv := DefaultService(newCoreAPIService(docSrv), nil)
	response, m, err := srv.DeriveTransferList(context.Background(), inv)
	assert.NotNil(t, response)
	assert.NotNil(t, m)
	assert.NoError(t, err)
}

func TestDeriveTransferListResponse(t *testing.T) {
	inv, _ := invoice.CreateInvoiceWithEmbedCD(t, nil, testingidentity.GenerateRandomDID(), nil)

	docSrv := new(testingdocuments.MockService)
	docSrv.On("GetCurrentVersion", mock.Anything, mock.Anything).Return(inv, nil)
	docSrv.On("UpdateModel", mock.Anything, mock.Anything).Return(inv, nil, nil)
	srv := DefaultService(newCoreAPIService(docSrv), nil)

	var model documents.Model
	var err error
	var payloads []CreateTransferDetailRequest
	for i := 0; i < 10; i++ {
		p := createTestPayload()
		p.DocumentID = hexutil.Encode(inv.Document.DocumentIdentifier)
		payloads = append(payloads, p)
		model, _, err = srv.CreateTransferDetail(context.Background(), p)
		assert.NoError(t, err)
	}

	response, _, err := srv.DeriveTransferList(context.Background(), model)
	assert.NoError(t, err)
	assert.Equal(t, 10, len(response.Data))

	for i := 0; i < 10; i++ {
		checkResponse(t, payloads[i], &response.Data[i])
	}

}

func TestService_DeriveFromUpdatePayload(t *testing.T) {
	inv, _ := invoice.CreateInvoiceWithEmbedCD(t, nil, testingidentity.GenerateRandomDID(), nil)

	docSrv := new(testingdocuments.MockService)
	docSrv.On("GetCurrentVersion", mock.Anything, mock.Anything).Return(inv, nil)
	docSrv.On("UpdateModel", mock.Anything, mock.Anything).Return(inv, nil, nil)
	srv := DefaultService(newCoreAPIService(docSrv), nil)
	var model documents.Model

	p := createTestPayload()
	p.DocumentID = hexutil.Encode(inv.Document.DocumentIdentifier)
	model, _, err := srv.CreateTransferDetail(context.Background(), p)
	assert.NoError(t, err)

	// update
	docSrv.On("GetCurrentVersion", mock.Anything, mock.Anything).Return(model, nil)
	docSrv.On("UpdateModel", mock.Anything, mock.Anything).Return(inv, nil, nil)
	p2 := &UpdateTransferDetailRequest{Data: createTestData(), DocumentID: p.DocumentID, TransferID: p.Data.TransferID}
	p2.Data.Currency = "USD"
	p2.Data.Amount = "1200"

	model, _, err = srv.UpdateTransferDetail(context.Background(), *p2)
	assert.NoError(t, err)

	response, _, err := srv.DeriveTransferList(context.Background(), model)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(response.Data))
	assert.Equal(t, p2.Data.Status, response.Data[0].Status)

	// Currency should have been updated
	assert.NotEqual(t, p.Data.Currency, response.Data[0].Currency)

	// attempted update of non-existent transfer details
	p3 := &UpdateTransferDetailRequest{Data: createTestData(), DocumentID: p.DocumentID, TransferID: hexutil.Encode(utils.RandomSlice(32))}
	model, _, err = srv.UpdateTransferDetail(context.Background(), *p3)
	assert.Error(t, err)
	assert.Contains(t, err, extensions.ErrAttributeSetNotFound)
}

func createTestData() Data {
	transferID := extensions.NewAttributeSetID()
	return Data{
		TransferID:          transferID,
		SenderID:            testingidentity.GenerateRandomDID().String(),
		RecipientID:         testingidentity.GenerateRandomDID().String(),
		ScheduledDate:       time.Now().UTC().Format(time.RFC3339),
		SettlementDate:      time.Now().UTC().Format(time.RFC3339),
		SettlementReference: hexutil.Encode(utils.RandomSlice(32)),
		Amount:              "1000",
		// the currency and amount will be combined once we have standardised multiformats
		Currency:     "EUR",
		Status:       "open",
		TransferType: "nft_transfer",
		Data:         hexutil.Encode(utils.RandomSlice(32)),
	}
}

func createTestPayload() CreateTransferDetailRequest {
	return CreateTransferDetailRequest{Data: createTestData()}
}

func checkResponse(t *testing.T, payload CreateTransferDetailRequest, response *Data) {
	assert.Equal(t, payload.Data.TransferID, response.TransferID)
	assert.Equal(t, payload.Data.Currency, response.Currency)
	assert.Equal(t, payload.Data.Status, response.Status)
	assert.Equal(t, payload.Data.TransferType, response.TransferType)
	assert.Equal(t, payload.Data.ScheduledDate, response.ScheduledDate)
}

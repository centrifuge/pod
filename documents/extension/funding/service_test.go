// +build unit

package funding

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/bootstrap"
	"github.com/centrifuge/go-centrifuge/bootstrap/bootstrappers/testlogging"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/config/configstore"
	"github.com/centrifuge/go-centrifuge/ethereum"
	"github.com/centrifuge/go-centrifuge/identity/ideth"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/p2p"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	"github.com/centrifuge/go-centrifuge/testingutils/commons"
	"github.com/centrifuge/go-centrifuge/testingutils/testingjobs"

	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/invoice"
	clientfundingpb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/funding"
	"github.com/centrifuge/go-centrifuge/testingutils/documents"
	"github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var ctx = map[string]interface{}{}
var cfg config.Configuration

var (
	did       = testingidentity.GenerateRandomDID()
	didBytes  = did[:]
	accountID = did[:]
)

func TestMain(m *testing.M) {
	ethClient := &testingcommons.MockEthClient{}
	ethClient.On("GetEthClient").Return(nil)
	ctx[ethereum.BootstrappedEthereumClient] = ethClient
	jobMan := &testingjobs.MockJobManager{}
	ctx[jobs.BootstrappedService] = jobMan
	done := make(chan bool)
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
	configService = ctx[config.BootstrappedConfigStorage].(config.Service)
	result := m.Run()
	bootstrap.RunTestTeardown(ibootstrappers)
	os.Exit(result)
}

func TestGenerateKey(t *testing.T) {
	assert.Equal(t, "centrifuge_funding[1].days", generateLabel("1", "days"))
	assert.Equal(t, "centrifuge_funding[0].", generateLabel("0", ""))

}

func TestCreateAttributesList(t *testing.T) {
	testingdocuments.CreateInvoicePayload()
	inv := &invoice.Invoice{}
	inv.InitInvoiceInput(testingdocuments.CreateInvoicePayload(), testingidentity.GenerateRandomDID())

	data := FundingData{Currency:"eur"}

	attributes, err := createAttributesList(inv, data)
	assert.NoError(t, err)
	assert.Equal(t, 12, len(attributes))

	for _, attribute := range attributes {
		if attribute.KeyLabel == "centrifuge_funding[0].currency" {
			assert.Equal(t, "eur", attribute.Value.Str)
			break
		}
	}
}

func TestDeriveFromPayload(t *testing.T) {
	testingdocuments.CreateInvoicePayload()
	inv := &invoice.Invoice{}
	inv.InitInvoiceInput(testingdocuments.CreateInvoicePayload(), testingidentity.GenerateRandomDID())

	docSrv := &testingdocuments.MockService{}
	docSrv.On("GetCurrentVersion", mock.Anything, mock.Anything).Return(inv, nil)
	srv := DefaultService(docSrv, nil)
	payload := &clientfundingpb.FundingCreatePayload{Data: &clientfundingpb.FundingData{Currency: "eur"}}

	for i := 0; i <10; i++ {
		model, err := srv.DeriveFromPayload(context.Background(), payload, utils.RandomSlice(32))
		assert.NoError(t, err)
		label := fmt.Sprintf("centrifuge_funding[%d].currency", i)
		key, err := documents.AttrKeyFromLabel(label)
		assert.NoError(t, err)

		attr, err := model.GetAttribute(key)
		assert.NoError(t, err)
		assert.Equal(t, "eur", attr.Value.Str)

	}

}

func TestDeriveFundingResponse(t *testing.T) {
	testingdocuments.CreateInvoicePayload()
	inv := &invoice.Invoice{}
	inv.InitInvoiceInput(testingdocuments.CreateInvoicePayload(), testingidentity.GenerateRandomDID())

	docSrv := &testingdocuments.MockService{}
	docSrv.On("GetCurrentVersion", mock.Anything, mock.Anything).Return(inv, nil)
	srv := DefaultService(docSrv, func() documents.TokenRegistry {
		return nil
	})

	for i := 0; i <10; i++ {
		fundingId := newFundingId()
		payload := &clientfundingpb.FundingCreatePayload{Data: &clientfundingpb.FundingData{FundingId: fundingId, Currency: "eur"}}
		model, err := srv.DeriveFromPayload(context.Background(), payload, utils.RandomSlice(32))
		assert.NoError(t, err)

		response, err := srv.DeriveFundingResponse(model, fundingId)
		assert.Equal(t, fundingId,response.Data.FundingId)
		assert.Equal(t, "eur", response.Data.Currency)

	}

}


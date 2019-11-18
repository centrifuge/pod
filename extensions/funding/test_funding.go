// +build unit integration testworld

package funding

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/invoice"
	"github.com/centrifuge/go-centrifuge/extensions"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	testingidentity "github.com/centrifuge/go-centrifuge/testingutils/identity"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func (b Bootstrapper) TestBootstrap(context map[string]interface{}) error {
	return b.Bootstrap(context)
}

func (Bootstrapper) TestTearDown() error {
	return nil
}

type MockService struct {
	Service
	mock.Mock
}

func (m *MockService) Sign(ctx context.Context, fundingID string, identifier []byte) (documents.Model, error) {
	args := m.Called(ctx, fundingID, identifier)
	model, _ := args.Get(0).(documents.Model)
	return model, args.Error(1)
}

func (m *MockService) CreateFundingAgreement(ctx context.Context, docID []byte, data *Data) (documents.Model, jobs.JobID, error) {
	args := m.Called(ctx, docID, data)
	model, _ := args.Get(0).(documents.Model)
	jobID, _ := args.Get(1).(jobs.JobID)
	return model, jobID, args.Error(2)
}

func (m *MockService) UpdateFundingAgreement(ctx context.Context, docID, fundingID []byte, data *Data) (documents.Model, jobs.JobID, error) {
	args := m.Called(ctx, docID, fundingID, data)
	model, _ := args.Get(0).(documents.Model)
	jobID, _ := args.Get(1).(jobs.JobID)
	return model, jobID, args.Error(2)
}

func (m *MockService) SignFundingAgreement(ctx context.Context, docID, fundingID []byte) (documents.Model, jobs.JobID, error) {
	args := m.Called(ctx, docID, fundingID)
	model, _ := args.Get(0).(documents.Model)
	jobID, _ := args.Get(1).(jobs.JobID)
	return model, jobID, args.Error(2)
}

func (m *MockService) GetDataAndSignatures(ctx context.Context, model documents.Model, fundingID, idx string) (Data, []Signature, error) {
	args := m.Called(ctx, model, fundingID)
	d, _ := args.Get(0).(Data)
	sigs, _ := args.Get(1).([]Signature)
	return d, sigs, args.Error(2)
}

func CreateData() Data {
	fundingId := extensions.NewAttributeSetID()
	return Data{
		AgreementID:           fundingId,
		Currency:              "eur",
		Days:                  "90",
		Amount:                "1000",
		RepaymentAmount:       "1200.12",
		Fee:                   "10",
		BorrowerID:            strings.ToLower(testingidentity.GenerateRandomDID().String()),
		FunderID:              strings.ToLower(testingidentity.GenerateRandomDID().String()),
		NFTAddress:            hexutil.Encode(utils.RandomSlice(32)),
		RepaymentDueDate:      time.Now().UTC().Format(time.RFC3339),
		RepaymentOccurredDate: time.Now().UTC().Format(time.RFC3339),
		PaymentDetailsID:      hexutil.Encode(utils.RandomSlice(32)),
		Status:                "submitted",
	}
}

func CreateInvoiceWithFunding(t *testing.T, ctx context.Context, did identity.DID) (*invoice.Invoice, string) {
	data := CreateData()
	inv, _ := invoice.CreateInvoiceWithEmbedCD(t, ctx, did, nil)
	attrs, err := extensions.CreateAttributesList(inv, data, fundingFieldKey, AttrFundingLabel)
	assert.NoError(t, err)
	err = inv.AddAttributes(documents.CollaboratorsAccess{}, false, attrs...)
	assert.NoError(t, err)
	return inv, data.AgreementID
}

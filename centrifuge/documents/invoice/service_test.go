// +build unit

package invoice

import (
	"context"
	"fmt"
	"testing"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/centrifuge/coredocument"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents"
	clientinvoicepb "github.com/centrifuge/go-centrifuge/centrifuge/protobufs/gen/go/invoice"
	"github.com/centrifuge/go-centrifuge/centrifuge/testingutils"
	"github.com/centrifuge/go-centrifuge/centrifuge/testingutils/documents"
	"github.com/centrifuge/go-centrifuge/centrifuge/tools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var invService Service

func createPayload() *clientinvoicepb.InvoiceCreatePayload {
	return &clientinvoicepb.InvoiceCreatePayload{
		Data: &clientinvoicepb.InvoiceData{
			Sender:      "0x010101010101",
			Recipient:   "0x010203040506",
			Payee:       "0x010203020406",
			GrossAmount: 42,
			ExtraData:   "0x",
			Currency:    "EUR",
		},
		Collaborators: []string{"0x010101010101"},
	}
}

func TestDefaultService(t *testing.T) {
	srv := DefaultService(GetRepository(), &testingutils.MockCoreDocumentProcessor{})
	assert.NotNil(t, srv, "must be non-nil")
}

func TestService_DeriveFromCoreDocument(t *testing.T) {
	// nil doc
	_, err := invService.DeriveFromCoreDocument(nil)
	assert.Error(t, err, "must fail to derive")

	// successful
	data := testinginvoice.CreateInvoiceData()
	coreDoc := testinginvoice.CreateCDWithEmbeddedInvoice(t, data)
	model, err := invService.DeriveFromCoreDocument(coreDoc)
	assert.Nil(t, err, "must return model")
	assert.NotNil(t, model, "model must be non-nil")
	inv, ok := model.(*InvoiceModel)
	assert.True(t, ok, "must be true")
	assert.Equal(t, inv.Payee[:], data.Payee)
	assert.Equal(t, inv.Sender[:], data.Sender)
	assert.Equal(t, inv.Recipient[:], data.Recipient)
	assert.Equal(t, inv.GrossAmount, data.GrossAmount)
}

func TestService_DeriveFromPayload(t *testing.T) {
	payload := createPayload()
	var model documents.Model
	var err error

	// fail due to nil payload
	_, err = invService.DeriveFromCreatePayload(nil)
	assert.Error(t, err, "DeriveWithInvoiceInput should produce an error if invoiceInput equals nil")

	model, err = invService.DeriveFromCreatePayload(payload)
	assert.Nil(t, err, "valid invoiceData shouldn't produce an error")

	receivedCoreDocument, err := model.PackCoreDocument()
	assert.Nil(t, err, "model should be able to return the core document with embedded invoice")
	assert.NotNil(t, receivedCoreDocument.EmbeddedData, "embeddedData should be field")
}

func TestService_GetLastVersion(t *testing.T) {
	documentIdentifier := tools.RandomSlice(32)
	nextIdentifier := tools.RandomSlice(32)
	thirdIdentifier := tools.RandomSlice(32)
	inv1 := &InvoiceModel{
		GrossAmount: 60,
		CoreDocument: &coredocumentpb.CoreDocument{
			DocumentIdentifier: documentIdentifier,
			CurrentVersion:  documentIdentifier,
			NextVersion:     nextIdentifier,
		},
	}
	err := GetRepository().Create(documentIdentifier, inv1)
	assert.Nil(t, err)

	mod1, err := invService.GetLastVersion(documentIdentifier)
	assert.Nil(t, err)

	invLoad1, _ := mod1.(*InvoiceModel)
	assert.Equal(t, invLoad1.CoreDocument.CurrentVersion, documentIdentifier)

	inv2 := &InvoiceModel{
		GrossAmount: 60,
		CoreDocument: &coredocumentpb.CoreDocument{
			DocumentIdentifier: documentIdentifier,
			CurrentVersion:  nextIdentifier,
			NextVersion:     thirdIdentifier,
		},
	}

	err = GetRepository().Create(nextIdentifier, inv2)
	assert.Nil(t, err)

	mod2, err := invService.GetLastVersion(documentIdentifier)
	assert.Nil(t, err)

	invLoad2, _ := mod2.(*InvoiceModel)
	assert.Equal(t, invLoad2.CoreDocument.CurrentVersion, nextIdentifier)
	assert.Equal(t, invLoad2.CoreDocument.NextVersion, thirdIdentifier)
}

func TestService_GetVersion_invalid_version(t *testing.T) {
	currentVersion := tools.RandomSlice(32)

	inv := &InvoiceModel{
		GrossAmount: 60,
		CoreDocument: &coredocumentpb.CoreDocument{
			DocumentIdentifier: tools.RandomSlice(32),
			CurrentVersion:  currentVersion,
		},
	}
	err := GetRepository().Create(currentVersion, inv)
	assert.Nil(t, err)

	mod, err := invService.GetVersion(tools.RandomSlice(32), currentVersion)
	assert.EqualError(t, err, "[4]version is not valid for this identifier")
	assert.Nil(t, mod)
}

func TestService_GetVersion(t *testing.T) {
	documentIdentifier := tools.RandomSlice(32)
	currentVersion := tools.RandomSlice(32)

	inv := &InvoiceModel{
		GrossAmount: 60,
		CoreDocument: &coredocumentpb.CoreDocument{
			DocumentIdentifier: documentIdentifier,
			CurrentVersion:  currentVersion,
		},
	}
	err := GetRepository().Create(currentVersion, inv)
	assert.Nil(t, err)

	mod, err := invService.GetVersion(documentIdentifier, currentVersion)
	assert.Nil(t, err)
	loadInv, _ := mod.(*InvoiceModel)
	assert.Equal(t, loadInv.CoreDocument.CurrentVersion, currentVersion)
	assert.Equal(t, loadInv.CoreDocument.DocumentIdentifier, documentIdentifier)

	mod, err = invService.GetVersion(documentIdentifier, []byte{})
	assert.Error(t, err)
}

func TestService_Create_validation_fail(t *testing.T) {
	// fail Validations
	ctx := context.Background()
	_, err := invService.Create(ctx, &InvoiceModel{})
	assert.Error(t, err, "must be non nil")
	assert.Contains(t, err.Error(), "currency is invalid")
}

func TestService_Create_db_fail(t *testing.T) {
	ctx := context.Background()
	model := &mockModel{}
	cd := coredocument.New()
	model.On("JSON").Return([]byte{1, 2, 3}, nil).Once()
	err := GetRepository().Create(cd.CurrentVersion, model)
	model.AssertExpectations(t)

	payload := createPayload()
	inv, err := invService.DeriveFromCreatePayload(payload)
	assert.Nil(t, err, "must be non nil")
	assert.NotNil(t, inv)
	inv.(*InvoiceModel).CoreDocument = cd

	_, err = invService.Create(ctx, inv)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "document already exists")
}

func TestService_Create_anchor_fail(t *testing.T) {
	srv := invService.(*service)
	proc := &testingutils.MockCoreDocumentProcessor{}
	proc.On("Anchor", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("failed to anchor document"))
	srv.coreDocProcessor = proc
	payload := createPayload()
	inv, err := invService.DeriveFromCreatePayload(payload)
	_, err = srv.Create(context.Background(), inv)
	proc.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to anchor document")
}

func TestService_Create_send_fail(t *testing.T) {
	srv := invService.(*service)
	proc := &testingutils.MockCoreDocumentProcessor{}
	proc.On("Anchor", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	proc.On("Send", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("failed to send"))
	srv.coreDocProcessor = proc
	payload := createPayload()
	payload.Collaborators = []string{"0x010203040506"}
	inv, err := invService.DeriveFromCreatePayload(payload)
	_, err = srv.Create(context.Background(), inv)
	proc.AssertExpectations(t)
	assert.Nil(t, err)
}

func TestService_Create_saveState_fail(t *testing.T) {
	srv := invService.(*service)
	proc := &testingutils.MockCoreDocumentProcessor{}
	proc.On("Anchor", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	proc.On("Send", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("failed to send"))
	srv.coreDocProcessor = proc
	payload := createPayload()
	payload.Collaborators = []string{"0x010203040506"}
	inv, err := invService.DeriveFromCreatePayload(payload)
	_, err = srv.Create(context.Background(), inv)
	proc.AssertExpectations(t)
	assert.Nil(t, err)
}

func TestService_Create(t *testing.T) {
	srv := invService.(*service)
	proc := &testingutils.MockCoreDocumentProcessor{}
	proc.On("Anchor", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	proc.On("Send", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	srv.coreDocProcessor = proc
	payload := createPayload()
	payload.Collaborators = []string{"0x010203040506"}
	inv, err := invService.DeriveFromCreatePayload(payload)
	_, err = srv.Create(context.Background(), inv)
	proc.AssertExpectations(t)
	assert.Nil(t, err)
}

func TestService_DeriveInvoiceData(t *testing.T) {
	// some random model
	_, err := invService.DeriveInvoiceData(&mockModel{})
	assert.Error(t, err, "Derive must fail")

	// success
	payload := createPayload()
	inv, err := invService.DeriveFromCreatePayload(payload)
	assert.Nil(t, err, "must be non nil")
	data, err := invService.DeriveInvoiceData(inv)
	assert.Nil(t, err, "Derive must succeed")
	assert.NotNil(t, data, "data must be non nil")
}

func TestService_DeriveInvoiceResponse(t *testing.T) {
	model := &mockModel{
		CoreDocument: &coredocumentpb.CoreDocument{
			DocumentIdentifier: []byte{},
		},
	}
	// some random model
	_, err := invService.DeriveInvoiceResponse(model)
	assert.Error(t, err, "Derive must fail")

	// success
	payload := createPayload()
	inv1, err := invService.DeriveFromCreatePayload(payload)
	assert.Nil(t, err, "must be non nil")
	inv, ok := inv1.(*InvoiceModel)
	assert.True(t, ok)
	inv.CoreDocument = &coredocumentpb.CoreDocument{
		DocumentIdentifier: []byte{},
	}
	resp, err := invService.DeriveInvoiceResponse(inv)
	assert.Nil(t, err, "Derive must succeed")
	assert.NotNil(t, resp, "data must be non nil")
	assert.Equal(t, resp.Data, payload.Data, "data mismatch")
}

func TestService_SaveState(t *testing.T) {
	// unknown type
	err := invService.SaveState(&mockModel{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "document of invalid type")

	inv := new(InvoiceModel)
	err = inv.InitInvoiceInput(createPayload())
	assert.Nil(t, err)

	// save state must fail missing core document
	invEmpty := new(InvoiceModel)
	err = invService.SaveState(invEmpty)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "core document missing")

	// update fail
	coreDoc, err := inv.PackCoreDocument()
	assert.Nil(t, err)
	assert.NotNil(t, coreDoc)
	err = invService.SaveState(inv)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "document doesn't exists")

	// successful
	err = GetRepository().Create(coreDoc.CurrentVersion, inv)
	assert.Nil(t, err)
	assert.Equal(t, inv.Currency, "EUR")
	assert.Nil(t, inv.CoreDocument.DataRoot)

	inv.Currency = "INR"
	inv.CoreDocument.DataRoot = tools.RandomSlice(32)
	err = invService.SaveState(inv)
	assert.Nil(t, err)

	loadInv := new(InvoiceModel)
	err = GetRepository().LoadByID(coreDoc.CurrentVersion, loadInv)
	assert.Nil(t, err)
	assert.Equal(t, loadInv.Currency, inv.Currency)
	assert.Equal(t, loadInv.CoreDocument.DataRoot, inv.CoreDocument.DataRoot)
}

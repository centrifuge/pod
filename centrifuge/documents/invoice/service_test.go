// +build unit

package invoice

import (
	"context"
	"fmt"
	"testing"

	"strconv"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/centrifuge/code"
	"github.com/centrifuge/go-centrifuge/centrifuge/coredocument"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents"
	clientinvoicepb "github.com/centrifuge/go-centrifuge/centrifuge/protobufs/gen/go/invoice"
	"github.com/centrifuge/go-centrifuge/centrifuge/testingutils"
	"github.com/centrifuge/go-centrifuge/centrifuge/testingutils/documents"
	"github.com/centrifuge/go-centrifuge/centrifuge/tools"
	"github.com/ethereum/go-ethereum/common/hexutil"
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
			ExtraData:   "0x01020302010203",
			Currency:    "EUR",
		},
		Collaborators: []string{"0x010101010101"},
	}
}

func TestDefaultService(t *testing.T) {
	srv := DefaultService(getRepository(), &testingutils.MockCoreDocumentProcessor{})
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
	thirdIdentifier := tools.RandomSlice(32)
	doc, err := createMockDocument()
	assert.Nil(t, err)

	mod1, err := invService.GetLastVersion(doc.CoreDocument.DocumentIdentifier)
	assert.Nil(t, err)

	invLoad1, _ := mod1.(*InvoiceModel)
	assert.Equal(t, invLoad1.CoreDocument.CurrentVersion, doc.CoreDocument.DocumentIdentifier)

	inv2 := &InvoiceModel{
		GrossAmount: 60,
		CoreDocument: &coredocumentpb.CoreDocument{
			DocumentIdentifier: doc.CoreDocument.DocumentIdentifier,
			CurrentVersion:     doc.CoreDocument.NextVersion,
			NextVersion:        thirdIdentifier,
		},
	}

	err = getRepository().Create(doc.CoreDocument.NextVersion, inv2)
	assert.Nil(t, err)

	mod2, err := invService.GetLastVersion(doc.CoreDocument.DocumentIdentifier)
	assert.Nil(t, err)

	invLoad2, _ := mod2.(*InvoiceModel)
	assert.Equal(t, invLoad2.CoreDocument.CurrentVersion, doc.CoreDocument.NextVersion)
	assert.Equal(t, invLoad2.CoreDocument.NextVersion, thirdIdentifier)
}

func TestService_GetVersion_invalid_version(t *testing.T) {
	currentVersion := tools.RandomSlice(32)

	inv := &InvoiceModel{
		GrossAmount: 60,
		CoreDocument: &coredocumentpb.CoreDocument{
			DocumentIdentifier: tools.RandomSlice(32),
			CurrentVersion:     currentVersion,
		},
	}
	err := getRepository().Create(currentVersion, inv)
	assert.Nil(t, err)

	mod, err := invService.GetVersion(tools.RandomSlice(32), currentVersion)
	assert.EqualError(t, err, "[4]document not found for the given version: version is not valid for this identifier")
	assert.Nil(t, mod)
}

func TestService_GetVersion(t *testing.T) {
	documentIdentifier := tools.RandomSlice(32)
	currentVersion := tools.RandomSlice(32)

	inv := &InvoiceModel{
		GrossAmount: 60,
		CoreDocument: &coredocumentpb.CoreDocument{
			DocumentIdentifier: documentIdentifier,
			CurrentVersion:     currentVersion,
		},
	}
	err := getRepository().Create(currentVersion, inv)
	assert.Nil(t, err)

	mod, err := invService.GetVersion(documentIdentifier, currentVersion)
	assert.Nil(t, err)
	loadInv, _ := mod.(*InvoiceModel)
	assert.Equal(t, loadInv.CoreDocument.CurrentVersion, currentVersion)
	assert.Equal(t, loadInv.CoreDocument.DocumentIdentifier, documentIdentifier)

	mod, err = invService.GetVersion(documentIdentifier, []byte{})
	assert.Error(t, err)
}

func TestService_Create_unknown_type_fail(t *testing.T) {
	ctx := context.Background()
	_, err := invService.Create(ctx, &mockModel{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown document type")
}

func TestService_Create_validation_fail(t *testing.T) {
	// fail Validations
	ctx := context.Background()
	_, err := invService.Create(ctx, &InvoiceModel{CoreDocument: &coredocumentpb.CoreDocument{}})
	assert.Error(t, err, "must be non nil")
	assert.Contains(t, err.Error(), "currency is invalid")
}

func TestService_Create_db_fail(t *testing.T) {
	ctx := context.Background()
	model := &mockModel{}
	cd := coredocument.New()
	model.On("JSON").Return([]byte{1, 2, 3}, nil).Once()
	err := getRepository().Create(cd.CurrentVersion, model)
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
	err = getRepository().Create(coreDoc.CurrentVersion, inv)
	assert.Nil(t, err)
	assert.Equal(t, inv.Currency, "EUR")
	assert.Nil(t, inv.CoreDocument.DataRoot)

	inv.Currency = "INR"
	inv.CoreDocument.DataRoot = tools.RandomSlice(32)
	err = invService.SaveState(inv)
	assert.Nil(t, err)

	loadInv := new(InvoiceModel)
	err = getRepository().LoadByID(coreDoc.CurrentVersion, loadInv)
	assert.Nil(t, err)
	assert.Equal(t, loadInv.Currency, inv.Currency)
	assert.Equal(t, loadInv.CoreDocument.DataRoot, inv.CoreDocument.DataRoot)
}

func TestService_CreateProofs(t *testing.T) {
	i, err := createAnchoredMockDocument(t, false)
	assert.Nil(t, err)
	proof, err := invService.CreateProofs(i.CoreDocument.DocumentIdentifier, []string{"invoice_number"})
	assert.Nil(t, err)
	assert.Equal(t, proof.Header.DocumentId, hexutil.Encode(i.CoreDocument.DocumentIdentifier))
	assert.Equal(t, proof.Header.VersionId, hexutil.Encode(i.CoreDocument.DocumentIdentifier))
	assert.Equal(t, len(proof.FieldProofs), 1)
	assert.Equal(t, proof.FieldProofs[0].GetProperty(), "invoice_number")
}

func TestService_CreateProofsInvalidField(t *testing.T) {
	i, err := createAnchoredMockDocument(t, false)
	assert.Nil(t, err)
	_, err = invService.CreateProofs(i.CoreDocument.DocumentIdentifier, []string{"invalid_field"})
	assert.Error(t, err)
	assert.Equal(t, "createProofs error No such field: invalid_field in obj", err.Error())
}

func TestService_CreateProofsDocumentDoesntExist(t *testing.T) {
	_, err := invService.CreateProofs(tools.RandomSlice(32), []string{"invoice_number"})
	assert.Error(t, err)
	assert.Equal(t, "document not found: leveldb: not found", err.Error())
}

func TestService_CreateProofsForVersion(t *testing.T) {
	i, err := createAnchoredMockDocument(t, false)
	assert.Nil(t, err)
	olderVersion := i.CoreDocument.CurrentVersion
	i, err = updatedAnchoredMockDocument(t, i)
	assert.Nil(t, err)
	proof, err := invService.CreateProofsForVersion(i.CoreDocument.DocumentIdentifier, olderVersion, []string{"invoice_number"})
	assert.Nil(t, err)
	assert.Equal(t, proof.Header.DocumentId, hexutil.Encode(i.CoreDocument.DocumentIdentifier))
	assert.Equal(t, proof.Header.VersionId, hexutil.Encode(olderVersion))
	assert.Equal(t, len(proof.FieldProofs), 1)
	assert.Equal(t, proof.FieldProofs[0].GetProperty(), "invoice_number")
}

func TestService_CreateProofsForVersionDocumentDoesntExist(t *testing.T) {
	i, err := createAnchoredMockDocument(t, false)
	assert.Nil(t, err)
	_, err = invService.CreateProofsForVersion(i.CoreDocument.DocumentIdentifier, tools.RandomSlice(32), []string{"invoice_number"})
	assert.Error(t, err)
	assert.Equal(t, "document not found for the given version: leveldb: not found", err.Error())
}

func TestService_RequestDocumentSignature_SigningRootNil(t *testing.T) {
	i, err := createAnchoredMockDocument(t, true)
	assert.Nil(t, err)
	i.CoreDocument.SigningRoot = nil
	signature, err := invService.RequestDocumentSignature(i)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), strconv.Itoa(int(code.DocumentInvalid)))
	assert.Contains(t, err.Error(), "signing root missing")
	assert.Nil(t, signature)
}

func createAnchoredMockDocument(t *testing.T, skipSave bool) (*InvoiceModel, error) {
	i := &InvoiceModel{
		InvoiceNumber: "test_invoice",
		GrossAmount:   60,
		CoreDocument:  coredocument.New(),
	}
	err := i.calculateDataRoot()
	if err != nil {
		return nil, err
	}
	// get the coreDoc for the invoice
	corDoc, err := i.PackCoreDocument()
	if err != nil {
		return nil, err
	}
	coredocument.FillSalts(corDoc)
	err = coredocument.CalculateSigningRoot(corDoc)
	if err != nil {
		return nil, err
	}
	err = coredocument.CalculateDocumentRoot(corDoc)
	if err != nil {
		return nil, err
	}
	err = i.UnpackCoreDocument(corDoc)
	if err != nil {
		return nil, err
	}

	if !skipSave {
		err = getRepository().Create(i.CoreDocument.CurrentVersion, i)
		if err != nil {
			return nil, err
		}
	}

	return i, nil
}

func updatedAnchoredMockDocument(t *testing.T, i *InvoiceModel) (*InvoiceModel, error) {
	i.GrossAmount = 50
	err := i.calculateDataRoot()
	if err != nil {
		return nil, err
	}
	// get the coreDoc for the invoice
	corDoc, err := i.PackCoreDocument()
	if err != nil {
		return nil, err
	}
	// hacky update to version
	corDoc.CurrentVersion = corDoc.NextVersion
	corDoc.NextVersion = tools.RandomSlice(32)
	if err != nil {
		return nil, err
	}
	err = coredocument.CalculateSigningRoot(corDoc)
	if err != nil {
		return nil, err
	}
	err = coredocument.CalculateDocumentRoot(corDoc)
	if err != nil {
		return nil, err
	}
	err = i.UnpackCoreDocument(corDoc)
	if err != nil {
		return nil, err
	}
	err = getRepository().Create(i.CoreDocument.CurrentVersion, i)
	if err != nil {
		return nil, err
	}
	return i, nil
}

func createMockDocument() (*InvoiceModel, error) {
	documentIdentifier := tools.RandomSlice(32)
	nextIdentifier := tools.RandomSlice(32)
	inv1 := &InvoiceModel{
		InvoiceNumber: "test_invoice",
		GrossAmount:   60,
		CoreDocument: &coredocumentpb.CoreDocument{
			DocumentIdentifier: documentIdentifier,
			CurrentVersion:     documentIdentifier,
			NextVersion:        nextIdentifier,
		},
	}
	err := getRepository().Create(documentIdentifier, inv1)
	return inv1, err
}

func TestService_DeriveFromUpdatePayload(t *testing.T) {
	// nil payload
	doc, err := invService.DeriveFromUpdatePayload(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid payload")
	assert.Nil(t, doc)

	// messed up identifier
	payload := &clientinvoicepb.InvoiceUpdatePayload{Identifier: "some identifier"}
	doc, err = invService.DeriveFromUpdatePayload(payload)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode identifier")
	assert.Nil(t, doc)

	// missing last version
	id := tools.RandomSlice(32)
	payload.Identifier = hexutil.Encode(id)
	doc, err = invService.DeriveFromUpdatePayload(payload)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch old version")
	assert.Nil(t, doc)

	// failed to load from data
	old := new(InvoiceModel)
	err = old.InitInvoiceInput(createPayload())
	assert.Nil(t, err)
	old.CoreDocument.DocumentIdentifier = id
	old.CoreDocument.CurrentVersion = id
	old.CoreDocument.DocumentRoot = tools.RandomSlice(32)
	err = getRepository().Create(id, old)
	assert.Nil(t, err)
	payload.Data = &clientinvoicepb.InvoiceData{
		Sender:      "0x010101010101",
		Recipient:   "0x010203040506",
		Payee:       "0x010203020406",
		GrossAmount: 42,
		ExtraData:   "some data",
		Currency:    "EUR",
	}
	doc, err = invService.DeriveFromUpdatePayload(payload)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load invoice from data")
	assert.Nil(t, doc)

	// failed core document new version
	payload.Data.ExtraData = hexutil.Encode(tools.RandomSlice(32))
	payload.Collaborators = []string{"some wrong ID"}
	doc, err = invService.DeriveFromUpdatePayload(payload)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to prepare new version")
	assert.Nil(t, doc)

	// success
	wantCollab := tools.RandomSlice(6)
	payload.Collaborators = []string{hexutil.Encode(wantCollab)}
	doc, err = invService.DeriveFromUpdatePayload(payload)
	assert.Nil(t, err)
	assert.NotNil(t, doc)
	cd, err := doc.PackCoreDocument()
	assert.Nil(t, err)
	assert.Equal(t, wantCollab, cd.Collaborators[0])
	assert.Len(t, cd.Collaborators, 1)
	oldCD, err := old.PackCoreDocument()
	assert.Nil(t, err)
	assert.Equal(t, oldCD.DocumentIdentifier, cd.DocumentIdentifier)
	assert.Equal(t, payload.Identifier, hexutil.Encode(cd.DocumentIdentifier))
	assert.Equal(t, oldCD.CurrentVersion, cd.PreviousVersion)
	assert.Equal(t, oldCD.NextVersion, cd.CurrentVersion)
	assert.NotNil(t, cd.NextVersion)
	assert.Equal(t, payload.Data, doc.(*InvoiceModel).getClientData())
}

func TestService_Update_packCoreDocument_fail(t *testing.T) {
	model := &mockModel{}
	model.On("PackCoreDocument").Return(nil, fmt.Errorf("pack error")).Once()
	_, err := invService.Update(context.Background(), model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "pack error")
}

func TestService_Update_Missing_last_version(t *testing.T) {
	model := &mockModel{}
	cd := coredocument.New()
	model.On("PackCoreDocument").Return(cd, nil).Once()
	_, err := invService.Update(context.Background(), model)
	model.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "document not found")
}

func TestService_Update_unknown_type(t *testing.T) {
	model, err := invService.DeriveFromCreatePayload(createPayload())
	assert.Nil(t, err)
	cd, err := model.PackCoreDocument()
	assert.Nil(t, err)
	getRepository().Create(cd.DocumentIdentifier, model)

	newModel := &mockModel{}
	newModel.On("PackCoreDocument").Return(cd, nil).Once()
	_, err = invService.Update(context.Background(), newModel)
	newModel.AssertExpectations(t)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown document type")
}

func TestService_Update(t *testing.T) {
	srv := invService.(*service)
	proc := &testingutils.MockCoreDocumentProcessor{}
	proc.On("Anchor", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	proc.On("Send", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	srv.coreDocProcessor = proc
	payload := createPayload()
	payload.Collaborators = []string{"0x010203040506"}
	inv, err := invService.DeriveFromCreatePayload(payload)
	assert.Nil(t, err)
	cd, err := inv.PackCoreDocument()
	assert.Nil(t, err)
	cd.DocumentRoot = tools.RandomSlice(32)
	err = inv.UnpackCoreDocument(cd)
	assert.Nil(t, err)
	getRepository().Create(cd.DocumentIdentifier, inv)

	data, err := invService.DeriveInvoiceData(inv)
	assert.Nil(t, err)
	data.GrossAmount = 100
	data.ExtraData = hexutil.Encode(tools.RandomSlice(32))
	collab := hexutil.Encode(tools.RandomSlice(6))

	newInv, err := invService.DeriveFromUpdatePayload(&clientinvoicepb.InvoiceUpdatePayload{
		Identifier:    hexutil.Encode(cd.DocumentIdentifier),
		Collaborators: []string{collab},
		Data:          data,
	})
	assert.Nil(t, err)
	newData, err := invService.DeriveInvoiceData(newInv)
	assert.Nil(t, err)
	assert.Equal(t, data, newData)

	model, err := invService.Update(context.Background(), newInv)
	proc.AssertExpectations(t)
	assert.Nil(t, err)
	assert.Equal(t, newInv, model.(*InvoiceModel))

	newCD, err := model.PackCoreDocument()
	assert.Nil(t, err)
	assert.True(t, getRepository().Exists(newCD.DocumentIdentifier))
	assert.True(t, getRepository().Exists(newCD.CurrentVersion))
	assert.True(t, getRepository().Exists(newCD.PreviousVersion))

	newData, err = invService.DeriveInvoiceData(model)
	assert.Nil(t, err)
	assert.Equal(t, data, newData)
}

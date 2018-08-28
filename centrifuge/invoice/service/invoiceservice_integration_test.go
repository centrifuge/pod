// +build ethereum

package invoiceservice_test

import (
	"context"
	"os"
	"testing"

	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/invoice"
	cc "github.com/CentrifugeInc/go-centrifuge/centrifuge/context/testing"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument/processor"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice/repository"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice/service"
	clientinvoicepb "github.com/CentrifugeInc/go-centrifuge/centrifuge/protobufs/gen/go/invoice"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/testingutils"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	cc.TestFunctionalEthereumBootstrap()
	invoicerepository.NewLevelDBInvoiceRepository(&invoicerepository.LevelDBInvoiceRepository{Leveldb: cc.GetLevelDBStorage()})

	result := m.Run()
	cc.TestIntegrationTearDown()
	os.Exit(result)
}

func generateEmptyInvoiceForProcessing() (doc *invoice.Invoice) {
	identifier := testingutils.Rand32Bytes()
	doc = invoice.Empty()
	salts := &coredocumentpb.CoreDocumentSalts{}
	proofs.FillSalts(salts)
	doc.Document.CoreDocument = &coredocumentpb.CoreDocument{
		DocumentIdentifier: identifier,
		CurrentIdentifier:  identifier,
		NextIdentifier:     testingutils.Rand32Bytes(),
		CoredocumentSalts:  salts,
	}
	return
}

func TestInvoiceDocumentService_HandleAnchorInvoiceDocument_Integration(t *testing.T) {
	s := invoiceservice.InvoiceDocumentService{
		InvoiceRepository:     invoicerepository.GetInvoiceRepository(),
		CoreDocumentProcessor: coredocumentprocessor.NewDefaultProcessor(),
	}
	doc := generateEmptyInvoiceForProcessing()
	doc.Document.Data = &invoicepb.InvoiceData{
		InvoiceNumber:    "inv1234",
		SenderName:       "Jack",
		SenderZipcode:    "921007",
		SenderCountry:    "AUS",
		RecipientName:    "John",
		RecipientZipcode: "12345",
		RecipientCountry: "Germany",
		Currency:         "EUR",
		GrossAmount:      800,
	}

	anchoredDoc, err := s.HandleAnchorInvoiceDocument(context.Background(), &clientinvoicepb.AnchorInvoiceEnvelope{Document: doc.Document})

	//Call overall worked well and receive roughly sensical data back
	assert.Nil(t, err)
	assert.Equal(t, anchoredDoc.CoreDocument.DocumentIdentifier, doc.Document.CoreDocument.DocumentIdentifier,
		"DocumentIdentifier doesn't match")

	//Invoice document got stored in the DB
	loadedInvoice, _ := invoicerepository.GetInvoiceRepository().FindById(doc.Document.CoreDocument.DocumentIdentifier)
	assert.Equal(t, "AUS", loadedInvoice.Data.SenderCountry,
		"Didn't save the invoice data correctly")

	//Invoice Service should error out if trying to anchor the same document ID again
	doc.Document.Data.SenderCountry = "ES"
	anchoredDoc2, err := s.HandleAnchorInvoiceDocument(context.Background(), &clientinvoicepb.AnchorInvoiceEnvelope{Document: doc.Document})
	assert.Nil(t, anchoredDoc2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Document already exists")

	loadedInvoice2, _ := invoicerepository.GetInvoiceRepository().FindById(doc.Document.CoreDocument.DocumentIdentifier)
	assert.Equal(t, "DE", loadedInvoice2.Data.SenderCountry,
		"Invoice document on DB should have not not gotten overwritten after rejected anchor call")
}

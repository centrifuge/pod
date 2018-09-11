// +build ethereum

package invoiceservice_test

import (
	"context"
	"os"
	"testing"

	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/invoice"
	cc "github.com/CentrifugeInc/go-centrifuge/centrifuge/context/testingbootstrap"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument/processor"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice/repository"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice/service"
	clientinvoicepb "github.com/CentrifugeInc/go-centrifuge/centrifuge/protobufs/gen/go/invoice"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/testingutils"
	"github.com/centrifuge/precise-proofs/proofs"
	"github.com/stretchr/testify/assert"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument/repository"
)

func TestMain(m *testing.M) {
	cc.TestFunctionalEthereumBootstrap()
	db := cc.GetLevelDBStorage()
	invoicerepository.InitLevelDBRepository(db)
	coredocumentrepository.InitLevelDBRepository(db)
	result := m.Run()
	cc.TestFunctionalEthereumTearDown()
	os.Exit(result)
}

func generateEmptyInvoiceForProcessing() (doc *invoice.Invoice) {
	identifier := testingutils.Rand32Bytes()
	doc = invoice.Empty()
	salts := &coredocumentpb.CoreDocumentSalts{}
	proofs.FillSalts(doc.Document.Data, salts)
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
		InvoiceRepository:     invoicerepository.GetRepository(),
		CoreDocumentProcessor: coredocumentprocessor.DefaultProcessor(),
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
	loadedInvoice := new(invoicepb.InvoiceDocument)
	err = invoicerepository.GetRepository().GetByID(doc.Document.CoreDocument.DocumentIdentifier, loadedInvoice)
	assert.Equal(t, "AUS", loadedInvoice.Data.SenderCountry,
		"Didn't save the invoice data correctly")

	// Invoice stored after anchoring has Salts populated
	assert.NotNil(t, loadedInvoice.Salts.SenderCountry)

	//Invoice Service should error out if trying to anchor the same document ID again
	doc.Document.Data.SenderCountry = "ES"
	anchoredDoc2, err := s.HandleAnchorInvoiceDocument(context.Background(), &clientinvoicepb.AnchorInvoiceEnvelope{Document: doc.Document})
	assert.Nil(t, anchoredDoc2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "document already exists")

	loadedInvoice2 := new(invoicepb.InvoiceDocument)
	err = invoicerepository.GetRepository().GetByID(doc.Document.CoreDocument.DocumentIdentifier, loadedInvoice2)
	assert.Equal(t, "AUS", loadedInvoice2.Data.SenderCountry,
		"Invoice document on DB should have not not gotten overwritten after rejected anchor call")
}

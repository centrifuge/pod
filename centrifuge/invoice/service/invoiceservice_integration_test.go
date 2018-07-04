// +build ethereum

package invoiceservice_test

import (
	"context"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/invoice"
	cc "github.com/CentrifugeInc/go-centrifuge/centrifuge/context"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice/service"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice/repository"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/testingutils"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	cc.TestBootstrap()
	result := m.Run()
	cc.TestTearDown()
	os.Exit(result)
}

func generateEmptyInvoiceForProcessing() (doc *invoice.Invoice) {
	identifier := testingutils.Rand32Bytes()
	doc = invoice.NewEmptyInvoice()
	doc.Document.CoreDocument = &coredocumentpb.CoreDocument{
		DocumentIdentifier: identifier,
		CurrentIdentifier:  identifier,
		NextIdentifier:     testingutils.Rand32Bytes(),
		DataMerkleRoot:     testingutils.Rand32Bytes(),
	}
	return
}

func TestInvoiceDocumentService_HandleAnchorInvoiceDocument_Integration(t *testing.T) {
	s := invoiceservice.InvoiceDocumentService{
		InvoiceRepository:     invoicerepository.GetInvoiceRepository(),
		CoreDocumentProcessor: coredocument.GetDefaultCoreDocumentProcessor(),
	}
	doc := generateEmptyInvoiceForProcessing()
	doc.Document.Data.SenderCountry = "DE"

	anchoredDoc, err := s.HandleAnchorInvoiceDocument(context.Background(), &invoicepb.AnchorInvoiceEnvelope{Document: doc.Document})

	//Call overall worked well and receive roughly sensical data back
	assert.Nil(t, err)
	assert.Equal(t, anchoredDoc.CoreDocument.DocumentIdentifier, doc.Document.CoreDocument.DocumentIdentifier,
		"DocumentIdentifier doesn't match")

	//Invoice document got stored in the DB
	loadedInvoice, _ := invoicerepository.GetInvoiceRepository().FindById(doc.Document.CoreDocument.DocumentIdentifier)
	assert.Equal(t, "DE", loadedInvoice.Data.SenderCountry,
		"Didn't save the invoice data correctly")
}

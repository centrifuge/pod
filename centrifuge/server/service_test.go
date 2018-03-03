package server

import (
	"testing"
	"context"
	"bytes"
	"fmt"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice/documentservice"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument"
	cc "github.com/CentrifugeInc/go-centrifuge/centrifuge/context"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice"
	"os"
)

func TestMain(m *testing.M) {
	mockBootstrap()
	os.Exit(m.Run())
}

func TestCoreDocumentService(t *testing.T) {
	identifier := []byte("identifier")
	identifierIncorrect := []byte("incorrectIdentifier")
	s := documentservice.InvoiceDocumentService{}
	doc := invoice.InvoiceDocument{
		CoreDocument: &coredocument.CoreDocument{DocumentIdentifier:identifier},
	}

	sentDoc, err := s.SendInvoiceDocument(context.Background(), &invoice.SendInvoiceEnvelope{[][]byte{}, &doc})
	if err != nil {
		t.Fatal("Error in RPC Call", err)
	}
	if !bytes.Equal(sentDoc.CoreDocument.DocumentIdentifier, identifier) {
		t.Fatal("DocumentIdentifier doesn't match")
	}

	receivedDoc, err := s.GetInvoiceDocument(context.Background(),
		&invoice.GetInvoiceDocumentEnvelope{DocumentIdentifier: identifier})
	if err != nil {
		t.Fatal("Error in RPC Call", err)
	}
	if !bytes.Equal(receivedDoc.CoreDocument.DocumentIdentifier, identifier) {
		t.Fatal("DocumentIdentifier doesn't match")
	}

	docIncorrect, err := s.GetInvoiceDocument(context.Background(),
		&invoice.GetInvoiceDocumentEnvelope{DocumentIdentifier: identifierIncorrect})
	fmt.Println(docIncorrect)
	if err == nil {
		t.Fatal("RPC call should have raised exception")
	}
}

func mockBootstrap() {
	(&cc.MockCentNode{}).BootstrapDependencies()
}
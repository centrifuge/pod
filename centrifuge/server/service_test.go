package server

import (
	"testing"
	"context"
	pb "github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument"
	"bytes"
)

func TestCoreDocumentService(t *testing.T) {

	s := newCentrifugeNodeService()
	doc := pb.InvoiceDocument{
		DocumentIdentifier:[]byte("1"),
	}
	sentDoc, err := s.SendInvoiceDocument(context.Background(), &pb.SendInvoiceEnvelope{[][]byte{}, &doc})
	if err != nil {
		t.Fatal("Error in RPC Call", err)
	}
	if !bytes.Equal(sentDoc.DocumentIdentifier, []byte("1")) {
		t.Fatal("DocumentIdentifier doesn't match")
	}

}
package invoicestorage

import (
	"testing"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument"
	"bytes"
)

func TestStorageService(t *testing.T) {
	service := StorageService{}

	invoice := coredocument.InvoiceDocument{DocumentIdentifier: []byte("1"), CoreDocument: &coredocument.CoreDocument{DocumentIdentifier:[]byte("1")}}
	service.PutDocument(&invoice)

	doc, err := service.GetDocument([]byte("1"))
	if err != nil {
		t.Fatal("Error getting document")
	}

	if !bytes.Equal(doc.DocumentIdentifier, []byte("1")) {
		t.Fatal("Id doesn't match")
	}
}


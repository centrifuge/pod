package invoicestorage

import (
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/storage"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/invoice"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/errors"
	"github.com/golang/protobuf/proto"
)

var ReceivedInvoiceDocumentsKey = []byte("received-invoice-documents")


type StorageService struct {
	storage storage.DataStore
}

func (srv *StorageService) SetStorageBackend (s storage.DataStore) {
	srv.storage = s
}

func (srv *StorageService) GetDocumentKey(id []byte) (key []byte) {
	key = append([]byte("invoice"), id...)
	return key
}

func (srv *StorageService) GetDocument(id []byte) (doc *invoice.InvoiceDocument, err error) {
	doc_bytes, err := srv.storage.Get(srv.GetDocumentKey(id))
	if err != nil {
		return nil, err
	}

	doc = &invoice.InvoiceDocument{}
	err = proto.Unmarshal(doc_bytes, doc)
	if err != nil {
		return nil, err
	}
	return
}

func (srv *StorageService) PutDocument(doc *invoice.InvoiceDocument) (err error) {
	if doc.CoreDocument == nil {
		err = &errors.GenericError{"Invalid Empty (NIL) Invoice Document"}
		return
	}
	key := srv.GetDocumentKey(doc.CoreDocument.DocumentIdentifier)
	data, err := proto.Marshal(doc)

	if err != nil {
		return
	}
	err = srv.storage.Put(key, data)
	return
}

func (srv *StorageService) ReceiveDocument (doc *invoice.InvoiceDocument) (err error) {
	invoices, err := srv.GetReceivedDocuments()
	if err != nil {
		return
	}

	invoices.Invoices = append(invoices.Invoices, doc)

	data, err := proto.Marshal(invoices)
	if err != nil {
		return
	}

	err = srv.storage.Put(ReceivedInvoiceDocumentsKey, data)
	return
}

func (srv *StorageService) GetReceivedDocuments () (docs *invoice.ReceivedInvoices, err error) {
	doc_bytes, err := srv.storage.Get(ReceivedInvoiceDocumentsKey)
	invoices := &invoice.ReceivedInvoices{}

	if err == nil {
		err = proto.Unmarshal(doc_bytes, invoices)
		if err != nil {
			return
		}
	}
	return
}
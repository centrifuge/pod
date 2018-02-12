package invoicestorage

import (
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/storage"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument"
	"github.com/golang/protobuf/proto"
)

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

func (srv *StorageService) GetDocument(id []byte) (doc *coredocument.InvoiceDocument, err error) {
	doc_bytes, err := srv.storage.Get(srv.GetDocumentKey(id))
	if err != nil {
		return nil, err
	}

	doc = &coredocument.InvoiceDocument{}
	err = proto.Unmarshal(doc_bytes, doc)
	if err != nil {
		return nil, err
	}
	return
}

func (srv *StorageService) PutDocument(doc *coredocument.InvoiceDocument) (err error) {
	key := srv.GetDocumentKey(doc.DocumentIdentifier)
	data, err := proto.Marshal(doc)

	if err != nil {
		return
	}
	err = srv.storage.Put(key, data)
	return
}
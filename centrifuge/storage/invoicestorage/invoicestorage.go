package invoicestorage

import (
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/storage"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument"
	"github.com/golang/protobuf/proto"
)

var storageInstance = *storage.GetStorage()

type StorageService struct {
}

func (srv *StorageService) GetDocumentKey(id []byte) (key []byte) {
	key = append([]byte("invoice"), id...)
	return key
}

func (srv *StorageService) GetDocument(id []byte) (doc *coredocument.InvoiceDocument, err error) {
	doc_bytes, err := storageInstance.Get(srv.GetDocumentKey(id))

	doc = &coredocument.InvoiceDocument{}
	err = proto.Unmarshal(doc_bytes, doc)

	return
}

func (srv *StorageService) PutDocument(doc *coredocument.InvoiceDocument) (err error) {
	key := srv.GetDocumentKey(doc.DocumentIdentifier)
	data, err := proto.Marshal(doc)

	if err != nil {
		return
	}
	err = storageInstance.Put(key, data)
	return
}

package coredocumentstorage

import (
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/storage"
	"github.com/CentrifugeInc/centrifuge-protobufs/coredocument"
)

type StorageService struct {
	storage storage.DataStore
}

func (srv *StorageService) SetStorageBackend (s storage.DataStore) {
	srv.storage = s
}

func (srv *StorageService) GetDocument(id []byte) (doc *coredocumentpb.CoreDocument, err error) {
	doc, err = srv.storage.GetDocument(id)
	return
}

func (srv *StorageService) PutDocument(doc *coredocumentpb.CoreDocument) (err error) {
	return srv.storage.PutDocument(doc)
}

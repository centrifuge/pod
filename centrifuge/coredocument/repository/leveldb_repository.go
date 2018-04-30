package repository

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/CentrifugeInc/centrifuge-protobufs/coredocument"
	"github.com/golang/protobuf/proto"
)

type levelDBCoreDocumentRepository struct {
	leveldb *leveldb.DB
}

func NewLevelDBCoreDocumentRepository(Conn *leveldb.DB) CoreDocumentRepository {
	return &levelDBCoreDocumentRepository{Conn}
}

func (repo *levelDBCoreDocumentRepository) GetKey(id []byte) ([]byte) {
	return append([]byte("coredoc"), id...)
}

func (repo *levelDBCoreDocumentRepository) FindById(id []byte) (doc *coredocumentpb.CoreDocument, err error) {
	doc_bytes, err := repo.leveldb.Get(repo.GetKey(id), nil)
	if err != nil {
		return nil, err
	}

	doc = &coredocumentpb.CoreDocument{}
	err = proto.Unmarshal(doc_bytes, doc)
	if err != nil {
		return nil, err
	}
	return
}

func (repo *levelDBCoreDocumentRepository) Store(doc *coredocumentpb.CoreDocument) (err error) {
	key := repo.GetKey(doc.DocumentIdentifier)
	data, err := proto.Marshal(doc)

	if err != nil {
		return
	}
	err = repo.leveldb.Put(key, data, nil)
	return
}

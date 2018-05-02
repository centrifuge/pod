package coredocumentrepository

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/CentrifugeInc/centrifuge-protobufs/coredocument"
	"github.com/golang/protobuf/proto"
	"sync"
)

var once sync.Once

type LevelDBCoreDocumentRepository struct {
	Leveldb *leveldb.DB
}

func NewLevelDBCoreDocumentRepository(cdr CoreDocumentRepository) {
	once.Do(func() {
		coreDocumentRepository = cdr
	})
	return
}

func (repo *LevelDBCoreDocumentRepository) GetKey(id []byte) ([]byte) {
	return append([]byte("coredoc"), id...)
}

func (repo *LevelDBCoreDocumentRepository) FindById(id []byte) (doc *coredocumentpb.CoreDocument, err error) {
	doc_bytes, err := repo.Leveldb.Get(repo.GetKey(id), nil)
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

func (repo *LevelDBCoreDocumentRepository) Store(doc *coredocumentpb.CoreDocument) (err error) {
	key := repo.GetKey(doc.DocumentIdentifier)
	data, err := proto.Marshal(doc)

	if err != nil {
		return
	}
	err = repo.Leveldb.Put(key, data, nil)
	return
}

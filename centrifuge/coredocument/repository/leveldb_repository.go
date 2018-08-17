package coredocumentrepository

import (
	"sync"

	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"
	"github.com/golang/protobuf/proto"
	"github.com/syndtr/goleveldb/leveldb"
)

var once sync.Once

// LevelDBRepository is an implementation of Core Document Repository
type LevelDBRepository struct {
	LevelDB *leveldb.DB
}

func NewLevelDBRepository(cdr Repository) {
	once.Do(func() {
		coreDocumentRepository = cdr
	})
	return
}

func (repo *LevelDBRepository) GetKey(id []byte) []byte {
	return append([]byte("coredoc"), id...)
}

func (repo *LevelDBRepository) FindById(id []byte) (doc *coredocumentpb.CoreDocument, err error) {
	docBytes, err := repo.LevelDB.Get(repo.GetKey(id), nil)
	if err != nil {
		return nil, err
	}

	doc = &coredocumentpb.CoreDocument{}
	err = proto.Unmarshal(docBytes, doc)
	if err != nil {
		return nil, err
	}
	return
}

func (repo *LevelDBRepository) CreateOrUpdate(doc *coredocumentpb.CoreDocument) (err error) {
	key := repo.GetKey(doc.DocumentIdentifier)
	data, err := proto.Marshal(doc)

	if err != nil {
		return
	}
	err = repo.LevelDB.Put(key, data, nil)
	return
}

package invoicerepository

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/CentrifugeInc/centrifuge-protobufs/invoice"
	"github.com/golang/protobuf/proto"
	"github.com/go-errors/errors"
	"sync"
)

var once sync.Once

type LevelDBInvoiceRepository struct {
	Leveldb *leveldb.DB
}

func NewLevelDBInvoiceRepository(ir InvoiceRepository) {
	once.Do(func() {
		invoiceRepository = ir
	})
	return
}

func (repo *LevelDBInvoiceRepository) GetKey(id []byte) ([]byte) {
	return append([]byte("invoice"), id...)
}

func (repo *LevelDBInvoiceRepository) FindById(id []byte) (inv *invoicepb.InvoiceDocument, err error) {
	doc_bytes, err := repo.Leveldb.Get(repo.GetKey(id), nil)
	if err != nil {
		return nil, err
	}

	inv = &invoicepb.InvoiceDocument{}
	err = proto.Unmarshal(doc_bytes, inv)
	if err != nil {
		return nil, err
	}
	return
}

func (repo *LevelDBInvoiceRepository) Store(inv *invoicepb.InvoiceDocument) (err error) {
	if inv.CoreDocument == nil {
		err = errors.Errorf("Invalid Empty (NIL) Invoice Document")
		return
	}
	key := repo.GetKey(inv.CoreDocument.DocumentIdentifier)
	data, err := proto.Marshal(inv)

	if err != nil {
		return
	}
	err = repo.Leveldb.Put(key, data, nil)
	return
}

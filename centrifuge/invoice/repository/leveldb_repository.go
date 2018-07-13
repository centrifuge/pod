package invoicerepository

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/invoice"
	"github.com/golang/protobuf/proto"
	"sync"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/errors"
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
	if inv == nil {
		err = errors.GenerateNilParameterError(inv)
	}
	if inv.CoreDocument == nil {
		err = errors.GenerateNilParameterError(inv.CoreDocument)
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

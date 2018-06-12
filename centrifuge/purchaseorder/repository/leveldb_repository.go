package purchaseorderrepository

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/purchaseorder"
	"github.com/golang/protobuf/proto"
	"github.com/go-errors/errors"
	"sync"
)

var once sync.Once

type LevelDBPurchaseOrderRepository struct {
	Leveldb *leveldb.DB
}

func NewLevelDBPurchaseOrderRepository(ir PurchaseOrderRepository) {
	once.Do(func() {
		purchaseorderRepository = ir
	})
	return
}

func (repo *LevelDBPurchaseOrderRepository) GetKey(id []byte) ([]byte) {
	return append([]byte("purchaseorder"), id...)
}

func (repo *LevelDBPurchaseOrderRepository) FindById(id []byte) (inv *purchaseorderpb.PurchaseOrderDocument, err error) {
	doc_bytes, err := repo.Leveldb.Get(repo.GetKey(id), nil)
	if err != nil {
		return nil, err
	}

	inv = &purchaseorderpb.PurchaseOrderDocument{}
	err = proto.Unmarshal(doc_bytes, inv)
	if err != nil {
		return nil, err
	}
	return
}

func (repo *LevelDBPurchaseOrderRepository) Store(inv *purchaseorderpb.PurchaseOrderDocument) (err error) {
	if inv.CoreDocument == nil {
		err = errors.Errorf("Invalid Empty (NIL) PurchaseOrder Document")
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

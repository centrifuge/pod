package purchaseorderrepository

import (
	"sync"

	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/purchaseorder"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/errors"
	gerrors "github.com/go-errors/errors"
	"github.com/golang/protobuf/proto"
	"github.com/syndtr/goleveldb/leveldb"
)

var once sync.Once

type LevelDBPurchaseOrderRepository struct {
	Leveldb *leveldb.DB
}

func checkIfCoreDocumentFilledCorrectly(doc *purchaseorderpb.PurchaseOrderDocument) error {
	if doc.CoreDocument == nil {
		return errors.NilError(doc.CoreDocument)
	}
	if doc.CoreDocument.DocumentIdentifier == nil {
		return errors.NilError(doc.CoreDocument.DocumentIdentifier)
	}
	return nil
}

func NewLevelDBPurchaseOrderRepository(ir PurchaseOrderRepository) {
	once.Do(func() {
		purchaseorderRepository = ir
	})
	return
}

func (repo *LevelDBPurchaseOrderRepository) GetKey(id []byte) []byte {
	return append([]byte("purchaseorder"), id...)
}

func (repo *LevelDBPurchaseOrderRepository) FindById(id []byte) (orderDocument *purchaseorderpb.PurchaseOrderDocument, err error) {
	doc_bytes, err := repo.Leveldb.Get(repo.GetKey(id), nil)
	if err != nil {
		return nil, err
	}

	orderDocument = &purchaseorderpb.PurchaseOrderDocument{}
	err = proto.Unmarshal(doc_bytes, orderDocument)
	if err != nil {
		return nil, err
	}
	return
}

func (repo *LevelDBPurchaseOrderRepository) CreateOrUpdate(orderDocument *purchaseorderpb.PurchaseOrderDocument) (err error) {
	if orderDocument == nil {
		return errors.NilError(orderDocument)
	}
	if orderDocument.CoreDocument == nil {
		return errors.NilError(orderDocument.CoreDocument)
	}

	key := repo.GetKey(orderDocument.CoreDocument.DocumentIdentifier)
	data, err := proto.Marshal(orderDocument)

	if err != nil {
		return
	}
	err = repo.Leveldb.Put(key, data, nil)
	return
}

func (repo *LevelDBPurchaseOrderRepository) Create(doc *purchaseorderpb.PurchaseOrderDocument) (err error) {
	err = checkIfCoreDocumentFilledCorrectly(doc)
	if err != nil {
		return err
	}
	loadDoc, readErr := repo.FindById(doc.CoreDocument.DocumentIdentifier)
	if loadDoc != nil {
		return gerrors.Errorf("Document already exists. Create will not overwrite.")
	} else if readErr != nil && !gerrors.Is(leveldb.ErrNotFound, readErr) {
		return readErr
	} else {
		return repo.CreateOrUpdate(doc)
	}
}

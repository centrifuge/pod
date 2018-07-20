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

func checkIfCoreDocumentFilledCorrectly(inv *invoicepb.InvoiceDocument) error {
	//TODO use the new error coming from Miguel's changes
	if inv.CoreDocument == nil {
		return errors.Errorf("Invalid Empty (NIL) Invoice Document")
	}
	if inv.CoreDocument.DocumentIdentifier == nil {
		return errors.Errorf("Invalid Empty (NIL) Invoice Document")
	}
	return nil
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
		return errors.GenerateNilParameterError(inv)
	}
	if inv.CoreDocument == nil {
		return errors.GenerateNilParameterError(inv.CoreDocument)
	}
	key := repo.GetKey(inv.CoreDocument.DocumentIdentifier)
	data, err := proto.Marshal(inv)

	if err != nil {
		return
	}
	err = repo.Leveldb.Put(key, data, nil)
	return
}

func (repo *LevelDBInvoiceRepository) StoreOnce(inv *invoicepb.InvoiceDocument) (err error) {
	err = checkIfCoreDocumentFilledCorrectly(inv)
	if err != nil {
		return err
	}
	loadDoc, readErr := repo.FindById(inv.CoreDocument.DocumentIdentifier)
	if loadDoc != nil {
		return errors.Errorf("Document already exists. StoreOnce will not overwrite.")
	} else if readErr != nil && !errors.Is(leveldb.ErrNotFound, readErr) {
		return readErr
	} else {
		return repo.Store(inv)
	}
}

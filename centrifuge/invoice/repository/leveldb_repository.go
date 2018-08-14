package invoicerepository

import (
	"fmt"
	"sync"

	"github.com/CentrifugeInc/centrifuge-protobufs/gen/go/invoice"
	"github.com/CentrifugeInc/go-centrifuge/centrifuge/errors"
	gerrors "github.com/go-errors/errors"
	"github.com/golang/protobuf/proto"
	"github.com/syndtr/goleveldb/leveldb"
)

var once sync.Once
var ErrDocumentExists = fmt.Errorf("syntax error in pattern")

type LevelDBInvoiceRepository struct {
	Leveldb *leveldb.DB
}

func checkIfCoreDocumentFilledCorrectly(doc *invoicepb.InvoiceDocument) error {
	if doc.CoreDocument == nil {
		return errors.NilError(doc.CoreDocument)
	}
	if doc.CoreDocument.DocumentIdentifier == nil {
		return errors.NilError(doc.CoreDocument.DocumentIdentifier)
	}
	return nil
}

func NewLevelDBInvoiceRepository(ir InvoiceRepository) {
	once.Do(func() {
		invoiceRepository = ir
	})
	return
}

func (repo *LevelDBInvoiceRepository) GetKey(id []byte) []byte {
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

func (repo *LevelDBInvoiceRepository) CreateOrUpdate(inv *invoicepb.InvoiceDocument) (err error) {
	if inv == nil {
		return errors.NilError(inv)
	}
	if inv.CoreDocument == nil {
		return errors.NilError(inv.CoreDocument)
	}
	key := repo.GetKey(inv.CoreDocument.DocumentIdentifier)
	data, err := proto.Marshal(inv)

	if err != nil {
		return
	}
	err = repo.Leveldb.Put(key, data, nil)
	return
}

func (repo *LevelDBInvoiceRepository) Create(inv *invoicepb.InvoiceDocument) (err error) {
	err = checkIfCoreDocumentFilledCorrectly(inv)
	if err != nil {
		return err
	}
	loadDoc, readErr := repo.FindById(inv.CoreDocument.DocumentIdentifier)
	if loadDoc != nil {
		return gerrors.Errorf("Document already exists. Create will not overwrite.")
	} else if readErr != nil && !gerrors.Is(leveldb.ErrNotFound, readErr) {
		return readErr
	} else {
		return repo.CreateOrUpdate(inv)
	}
}

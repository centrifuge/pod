package repository

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/CentrifugeInc/centrifuge-protobufs/invoice"
	"github.com/golang/protobuf/proto"
	"github.com/go-errors/errors"
)

type levelDBInvoiceRepository struct {
	leveldb *leveldb.DB
}

func NewLevelDBInvoiceRepository(Conn *leveldb.DB) InvoiceRepository {
	return &levelDBInvoiceRepository{Conn}
}

func (repo *levelDBInvoiceRepository) GetKey(id []byte) ([]byte) {
	return append([]byte("invoice"), id...)
}

func (repo *levelDBInvoiceRepository) FindById(id []byte) (inv *invoicepb.InvoiceDocument, err error) {
	doc_bytes, err := repo.leveldb.Get(repo.GetKey(id), nil)
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

func (repo *levelDBInvoiceRepository) Store(inv *invoicepb.InvoiceDocument) (err error) {
	if inv.CoreDocument == nil {
		err = errors.Errorf("Invalid Empty (NIL) Invoice Document")
		return
	}
	key := repo.GetKey(inv.CoreDocument.DocumentIdentifier)
	data, err := proto.Marshal(inv)

	if err != nil {
		return
	}
	err = repo.leveldb.Put(key, data, nil)
	return
}

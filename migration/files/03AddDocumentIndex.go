package migrationfiles

import (
	"strings"

	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/entity"
	"github.com/centrifuge/go-centrifuge/documents/entityrelationship"
	"github.com/centrifuge/go-centrifuge/documents/generic"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ldb "github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

// AddDocumentIndex03 adds index to the document for efficient fetching.
func AddDocumentIndex03(db *ldb.DB) error {
	strRepo := leveldb.NewLevelDBRepository(db)
	repo := documents.NewDBRepository(strRepo)
	repo.Register(new(entityrelationship.EntityRelationship))
	repo.Register(new(entity.Entity))
	repo.Register(new(generic.Generic))
	iter := db.NewIterator(util.BytesPrefix([]byte("document_")), nil)
	var c, e int
	for iter.Next() {
		c++
		key := iter.Key()
		acc, id, err := getAccountAndID(key)
		if err != nil {
			e++
			// must have been an older document. skipping
			continue
		}
		m, err := strRepo.Get(key)
		if err != nil {
			return err
		}

		mm, ok := m.(documents.Model)
		if !ok {
			return documents.ErrDocumentInvalidType
		}

		err = repo.Update(acc, id, mm)
		if err != nil {
			return err
		}

	}
	log.Infof("Updated index for %d documents\n", c-e)
	iter.Release()
	err := iter.Error()
	if err != nil {
		return err
	}

	log.Infof("AddDocumentIndex03 Migration Run successfully")
	return nil
}

func getAccountAndID(key []byte) (acc, id []byte, err error) {
	str := strings.TrimSpace(strings.TrimPrefix(string(key), "document_"))
	d, err := hexutil.Decode(str)
	if err != nil {
		return nil, nil, err
	}

	if len(d) != 52 {
		return nil, nil, errors.New("invalid key(%v) of length %d found", string(key), len(d))
	}

	// first 20 bytes are account and last 32 are id
	return d[:20], d[20:], nil
}

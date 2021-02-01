package migrationfiles

import (
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/documents/entity"
	"github.com/centrifuge/go-centrifuge/documents/entityrelationship"
	"github.com/centrifuge/go-centrifuge/documents/generic"
	"github.com/centrifuge/go-centrifuge/storage/leveldb"
	ldb "github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

// AddStatusToDocuments04 adds status to committed.
func AddStatusToDocuments04(db *ldb.DB) error {
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
		m, err := strRepo.Get(key)
		if err != nil {
			// model fetch failed, skip
			e++
			continue
		}

		mm, ok := m.(documents.Document)
		if !ok {
			return documents.ErrDocumentInvalidType
		}

		err = mm.SetStatus(documents.Committed)
		if err != nil {
			return err
		}

		err = strRepo.Update(key, mm)
		if err != nil {
			return err
		}

	}

	log.Infof("Updated status for %d documents\n", c-e)
	iter.Release()
	err := iter.Error()
	if err != nil {
		return err
	}

	log.Infof("AddStatusToDocuments04 Migration Run successfully")
	return nil
}

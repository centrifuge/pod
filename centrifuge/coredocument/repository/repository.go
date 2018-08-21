package coredocumentrepository

import "github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"

var coreDocumentRepository CoreDocumentRepository

func GetCoreDocumentRepository() CoreDocumentRepository {
	return coreDocumentRepository
}

type CoreDocumentRepository interface {
	GetKey(id []byte) []byte
	FindById(id []byte) (doc *coredocumentpb.CoreDocument, err error)

	// CreateOrUpdate functions similar to a REST HTTP PUT where the document is either created or updated regardless if it existed before
	CreateOrUpdate(doc *coredocumentpb.CoreDocument) (err error)
}

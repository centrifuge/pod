package coredocumentrepository

import "github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"

var coreDocumentRepository CoreDocumentRepository

func GetCoreDocumentRepository() CoreDocumentRepository {
	return coreDocumentRepository
}

type CoreDocumentRepository interface {
	GetKey(id []byte) ([]byte)
	FindById(id []byte) (doc *coredocumentpb.CoreDocument, err error)
	Store(doc *coredocumentpb.CoreDocument) (err error)
}

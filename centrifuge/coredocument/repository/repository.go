package coredocumentrepository

import "github.com/CentrifugeInc/centrifuge-protobufs/gen/go/coredocument"

var coreDocumentRepository Repository

// GetRepository returns CoreDocument repository implementation
func GetRepository() Repository {
	return coreDocumentRepository
}

// Repository defines functions for Repository
type Repository interface {
	GetKey(id []byte) []byte
	FindById(id []byte) (doc *coredocumentpb.CoreDocument, err error)

	// CreateOrUpdate functions similar to a REST HTTP PUT where the document is either created or updated regardless if it existed before
	CreateOrUpdate(doc *coredocumentpb.CoreDocument) (err error)
}

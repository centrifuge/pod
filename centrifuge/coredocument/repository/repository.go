package repository

import "github.com/CentrifugeInc/centrifuge-protobufs/coredocument"

type CoreDocumentRepository interface {
	GetKey(id []byte) ([]byte)
	FindById(id []byte) (doc *coredocumentpb.CoreDocument, err error)
	Store(doc *coredocumentpb.CoreDocument) (err error)
}

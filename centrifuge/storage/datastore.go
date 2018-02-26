package storage

import "github.com/CentrifugeInc/go-centrifuge/centrifuge/coredocument"

type DataStore interface {
	Open () error
	Close ()
	Get([]byte) ([]byte, error)
	Put([]byte, []byte) error
	GetDocumentKey ([]byte) []byte
	GetDocument([]byte) (*coredocument.CoreDocument, error)
	PutDocument(document *coredocument.CoreDocument) (error)
}
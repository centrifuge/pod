package storage

import coredocumentpb "github.com/CentrifugeInc/centrifuge-protobufs/coredocument"

type DataStore interface {
	Open () error
	Close ()
	Get([]byte) ([]byte, error)
	Put([]byte, []byte) error
	GetDocumentKey ([]byte) []byte
	GetDocument([]byte) (*coredocumentpb.CoreDocument, error)
	PutDocument(document *coredocumentpb.CoreDocument) (error)
}
package storage

type DataStore interface {
	Close()
	GetDocumentKey()
}

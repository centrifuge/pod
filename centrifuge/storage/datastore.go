package storage

type DataStore interface {
	Open () error
	Close ()
	Get([]byte) ([]byte, error)
	Put([]byte, []byte) error
}
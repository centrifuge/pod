package storage

import (
	"github.com/golang/protobuf/proto"
)

// Getter interface can be implemented by any repository that handles document retrieval
type Getter interface {
	// GetKey will prepare the the identifier key from ID
	GetKey(id []byte) (key []byte)

	// GetByID finds the doc with identifier and marshalls it into message
	GetByID(id []byte, msg proto.Message) error
}

// Checker interface can be implemented by any repository that handles document retrieval
type Checker interface {
	// Exists checks for document existence
	// True if exists else false
	Exists(id []byte) bool
}

// Creator interface can be implemented by any repository that handles document storage
type Creator interface {
	// Create stores the initial document
	// If document exist, it errors out
	Create(id []byte, msg proto.Message) error
}

// Updater interface can be implemented by any repository that handles document storage
type Updater interface {
	// Update updates the already stored document
	// errors out when document is missing
	Update(id []byte, msg proto.Message) error
}

// LegacyRepository interface combines above interfaces
type LegacyRepository interface {
	Checker
	Getter
	Creator
	Updater
}

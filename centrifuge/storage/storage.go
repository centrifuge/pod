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

// Repository interface combines above interfaces
type Repository interface {
	Checker
	Getter
	Creator
	Updater
}

// Marshaler is an interface that is implemented by types that can marshall themselves into JSON
type Marshaler interface {
	JSON() ([]byte, error)
}

// Unmarshaler is an interface that is implemented by types that can unmarshal themselves from JSON data
type Unmarshaler interface {
	FromJSON([]byte) error
}

// CreatorModel interface can be implemented by any repository that handles model retrieval
// TODO(ved): rename the interfaces once model storage is implemented across documents
type GetterModel interface {
	// GetKey will prepare the the identifier key from ID
	GetKey(id []byte) (key []byte)

	// GetByID finds the doc with identifier and marshals it into message
	GetModelByID(id []byte, msg Unmarshaler) error
}

// CreatorModel interface can be implemented by any repository that handles model storage
type CreatorModel interface {
	// Create stores the initial document
	// If document exist, it errors out
	CreateModel(id []byte, msg Marshaler) error
}

// UpdaterModel interface can be implemented by any repository that handles model storage
type UpdaterModel interface {
	// Update updates the already stored document
	// errors out when document is missing
	UpdateModel(id []byte, msg Marshaler) error
}

// ModelRepository interface combines above interfaces
type ModelRepository interface {
	Checker
	GetterModel
	CreatorModel
	UpdaterModel
}

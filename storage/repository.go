package storage

import (
	"reflect"
)

const (
	// BootstrappedDB is a key mapped to DB at boot
	BootstrappedDB string = "BootstrappedDB"
	// BootstrappedConfigDB is a key mapped to DB for configs at boot
	BootstrappedConfigDB string = "BootstrappedConfigDB"
)

//go:generate mockery --name Model --structname ModelMock --filename model_mock.go --inpackage

// Model is an interface to abstract away storage model specificness
type Model interface {
	// Type Returns the underlying type of the Model
	Type() reflect.Type

	// JSON return the json representation of the model
	JSON() ([]byte, error)

	// FromJSON initialize the model with a json
	FromJSON(json []byte) error
}

//go:generate mockery --name Repository --structname RepositoryMock --filename repository_mock.go --inpackage

// Repository defines the required methods for standard storage repository.
type Repository interface {
	Register(model Model)
	Exists(key []byte) bool
	Get(key []byte) (Model, error)
	GetAllByPrefix(prefix string) ([]Model, error)
	Create(key []byte, model Model) error
	Update(key []byte, model Model) error
	Delete(key []byte) error
	Close() error
}

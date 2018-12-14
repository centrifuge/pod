package storage

import (
	"reflect"
)

type Model interface {
	// Get the ID of the document represented by this model
	ID() ([]byte, error)

	//Returns the underlying type of the Model
	Type() reflect.Type

	// JSON return the json representation of the model
	JSON() ([]byte, error)

	// FromJSON initialize the model with a json
	FromJSON(json []byte) error
}

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

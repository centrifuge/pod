package storage

import "github.com/centrifuge/go-centrifuge/errors"

const (
	// ErrModelRepositorySerialisation must be used when document repository encounters a marshalling error
	ErrModelRepositorySerialisation = errors.Error("model repository encountered a marshalling error")

	// ErrModelRepositoryNotFound must be used when model is not found in db
	ErrModelRepositoryNotFound = errors.Error("model not found in db")

	// ErrRepositoryModelSave must be used when db repository can not save the given model
	ErrRepositoryModelSave = errors.Error("db repository could not save the given model")

	// ErrRepositoryModelUpdateKeyNotFound must be used when db repository can not update the given model
	ErrRepositoryModelUpdateKeyNotFound = errors.Error("db repository could not update the given model, key not found")

	// ErrRepositoryModelCreateKeyExists must be used when db repository can not create the given model
	ErrRepositoryModelCreateKeyExists = errors.Error("db repository could not create the given model, key already exists")

	// ErrModelTypeNotRegistered must be used when model hasn't been registered in db
	ErrModelTypeNotRegistered = errors.Error("type not registered")
)

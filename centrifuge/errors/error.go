package errors

import (
	"github.com/go-errors/errors"
	"reflect"
)

func GenerateNilParameterError(param interface{}) (error) {
	return errors.Errorf("NIL %v provided", reflect.TypeOf(param))
}

func New(message string) (error) {
	return errors.New(message)
}
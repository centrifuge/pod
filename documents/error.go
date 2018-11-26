package documents

import (
	"bytes"
	"fmt"

	"github.com/hashicorp/go-multierror"
)

// Error wraps an error with specific key
type Error struct {
	key string
	err error
}

// Error returns the underlying error message
func (e Error) Error() string {
	return e.err.Error()
}

// NewError creates a new error from a key and a msg.
func NewError(key, msg string) error {
	err := fmt.Errorf(msg)
	return Error{key: key, err: err}
}

// AppendError function is used to create a list of errors.
// First argument can be nil, a multierror.Error, or any other error
func AppendError(dstErr, srcErr error) error {
	_, ok := dstErr.(*multierror.Error)
	result := multierror.Append(dstErr, srcErr)

	// if dstErr is not a multierror.Error newly created multierror.Error result needs formatting
	if !ok {
		return format(result)
	}
	return result
}

func format(err *multierror.Error) error {
	err.ErrorFormat = func(errorList []error) string {
		var buffer bytes.Buffer
		for i, err := range errorList {
			if errt, ok := err.(Error); ok {
				buffer.WriteString(fmt.Sprintf("%s : %s\n", errt.key, errt.err.Error()))
				continue
			}

			buffer.WriteString(fmt.Sprintf("Error %v : %s\n", i+1, err.Error()))
		}

		buffer.WriteString(fmt.Sprintf("Total Errors: %v\n", len(errorList)))
		return buffer.String()
	}

	return err
}

// Errors returns an array of errors
func Errors(err error) []error {
	if err == nil {
		return nil
	}

	if multiErr, ok := err.(*multierror.Error); ok {
		return multiErr.Errors
	}

	return []error{err}
}

// LenError returns the amount of embedded errors.
func LenError(err error) int {

	if err == nil {
		return 0
	}

	if multiErr, ok := err.(*multierror.Error); ok {

		return multiErr.Len()
	}
	return 1

}

func addToMap(errorMap map[string]string, key, msg string) map[string]string {
	if errorMap[key] != "" {
		errorMap[key] = fmt.Sprintf("%s\n%s", errorMap[key], msg)

	} else {
		errorMap[key] = msg

	}
	return errorMap
}

// ConvertToMap converts errors into a map.
func ConvertToMap(err error) map[string]string {

	errorMap := make(map[string]string)
	var key string
	var standardErrorCounter int

	errors := Errors(err)

	for _, err := range errors {
		if err, ok := err.(Error); ok {
			key = err.key

		} else {
			standardErrorCounter++
			key = fmt.Sprintf("error_%v", standardErrorCounter)

		}

		addToMap(errorMap, key, err.Error())

	}
	return errorMap

}

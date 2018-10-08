package documenterror

import (
	"bytes"
	"fmt"

	"github.com/hashicorp/go-multierror"
)

// New creates a new error from a string message
func New(msg string) error {
	return fmt.Errorf(msg)
}

// Append function is used to create a list of errors.
// First argument can be nil, a multierror.Error, or any other error
func Append(dstErr, srcErr error) error {

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

// Len returns the amount of embedded errors
func Len(err error) int {

	if err == nil {
		return 0
	}

	if multiErr, ok := err.(*multierror.Error); ok {

		return multiErr.Len()
	}
	return 1

}

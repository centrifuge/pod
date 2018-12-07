package errors

import (
	"fmt"
	"strings"
)

// ErrUnknown is an unknown error type
const ErrUnknown = Error("unknown error")

// Error is a string that implements error
// this will have interesting side effects of having constant errors
type Error string

// Error returns error message
func (e Error) Error() string {
	return string(e)
}

// New returns a new error with message passed.
// if args are passed, we will format the message with args
// Example:
// New("some error") returns error with exact string passed
// New("some error: %v", "some context") returns error with message "some error: some context"
func New(format string, args ...interface{}) error {
	return Error(fmt.Sprintf(format, args...))
}

// listError holds a list of errors
type listError struct {
	errs []error
}

// Error formats the underlying errs into string
func (l *listError) Error() string {
	if len(l.errs) == 0 {
		return ""
	}

	var errs []string
	for _, err := range l.errs {
		errs = append(errs, err.Error())
	}

	res := strings.Join(errs, "; ")
	return "[" + res + "]"
}

// append appends the err to the list of errs
func getErrs(err error) []error {
	if err == nil {
		return nil
	}

	if errl, ok := err.(*listError); ok {
		return errl.errs
	}

	return []error{err}
}

// AppendError returns a new listError
// if errn == nil, return err
// if err is of type listError and if errn is of type listerror,
// append errn errors to err and return err
func AppendError(err, errn error) error {
	var errs []error

	for _, e := range []error{err, errn} {
		if serrs := getErrs(e); len(serrs) > 0 {
			errs = append(errs, serrs...)
		}
	}

	if len(errs) < 1 {
		return nil
	}

	return &listError{errs}
}

// Len returns the total number of errors
// if err is listError, return len(listError.errs)
// if err == nil, return 0
// else return 1
func Len(err error) int {
	if err == nil {
		return 0
	}

	if lerr, ok := err.(*listError); ok {
		return len(lerr.errs)
	}

	return 1
}

// typeError holds a type of error and an context error
type typeError struct {
	terr   error
	ctxErr error
}

// Error returns the error in string
func (t *typeError) Error() string {
	return fmt.Sprintf("%v: %v", t.terr, t.ctxErr)
}

// NewTypeError returns a new error of type typeError
func NewTypeError(terr, err error) error {
	if terr == nil {
		terr = ErrUnknown
	}

	return &typeError{terr: terr, ctxErr: err}
}

// TypeError can be implemented by any type error
type TypeError interface {
	IsOfType(terr error) bool
}

// IsOfType returns if the err t is of type terr
func (t *typeError) IsOfType(terr error) bool {
	if t.terr.Error() == terr.Error() {
		return true
	}

	if cterr, ok := t.ctxErr.(TypeError); ok {
		return cterr.IsOfType(terr)
	}

	return t.ctxErr.Error() == terr.Error()
}

// IsOfType returns if the err is of type terr
func IsOfType(terr, err error) bool {
	if errt, ok := err.(TypeError); ok {
		return errt.IsOfType(terr)
	}

	return err.Error() == terr.Error()
}

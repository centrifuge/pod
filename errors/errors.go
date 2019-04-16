package errors

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ErrUnknown is an unknown error type
const ErrUnknown = Error("unknown error")

// MaskErrs this is a compile time flag to indicate whether to mask errors that can leak private or sensitive information.
// IMPORTANT!!! DO NOT CHANGE AT RUNTIME in production code.
var MaskErrs = true

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
	err := Error(fmt.Sprintf(format, args...))
	return &withStack{err: err, trace: callers()}
}

// listError holds a list of errors
type listError []error

// Error formats the underlying errs into string
func (l listError) Error() string {
	if len(l) == 0 {
		return ""
	}

	var errs []string
	for _, err := range l {
		errs = append(errs, err.Error())
	}

	res := strings.Join(errs, "; ")
	return "[" + res + "]"
}

// GetErrs gets the list of errors if its a list
func GetErrs(err error) []error {
	if err == nil {
		return nil
	}

	if errl, ok := err.(listError); ok {
		return errl
	}

	return []error{err}
}

// AppendError returns a new listError
// if errn == nil, return err
// if err is of type listError and if errn is of type listerror,
// append errn errors to err and return err
func AppendError(err, errn error) error {
	var errs listError

	for _, e := range []error{err, errn} {
		if serrs := GetErrs(e); len(serrs) > 0 {
			errs = append(errs, serrs...)
		}
	}

	if len(errs) < 1 {
		return nil
	}

	return errs
}

// Len returns the total number of errors
// if err is listError, return len(listError.errs)
// if err == nil, return 0
// else return 1
func Len(err error) int {
	if err == nil {
		return 0
	}

	if lerr, ok := err.(listError); ok {
		return len(lerr)
	}

	return 1
}

// typedError holds a type of error and an context error
type typedError struct {
	terr   error
	ctxErr error
	mask   string
}

// Error returns the error in string
func (t *typedError) Error() string {
	return fmt.Sprintf("%v: %v", t.terr, t.ctxErr)
}

// NewTypedError returns a new error of type typedError
func NewTypedError(terr, err error) error {
	if terr == nil {
		terr = ErrUnknown
	}

	return &typedError{terr: terr, ctxErr: err, mask: "error has been masked"}
}

// TypedError can be implemented by any type error
type TypedError interface {
	IsOfType(terr error) bool
	Mask() error
}

// IsOfType returns if the err t is of type terr
func (t *typedError) IsOfType(terr error) bool {
	if t.terr.Error() == terr.Error() {
		return true
	}

	if cterr, ok := t.ctxErr.(TypedError); ok {
		return cterr.IsOfType(terr)
	}

	return t.ctxErr.Error() == terr.Error()
}

// Mask returns a mask to hide the actual error to prevent guessing attacks using error messages on p2p
func (t *typedError) Mask() error {
	return New(t.mask)
}

// IsOfType returns if the err is of type terr
func IsOfType(terr, err error) bool {
	err = detachStackTrace(err)
	if errt, ok := err.(TypedError); ok {
		return errt.IsOfType(terr)
	}

	if serr, ok := status.FromError(err); ok {
		return serr.Message() == terr.Error()
	}

	return err.Error() == terr.Error()
}

// Mask returns the mask for the error
func Mask(err error) error {
	if !MaskErrs {
		return err
	}

	err = detachStackTrace(err)
	if errt, ok := err.(TypedError); ok {
		return errt.Mask()
	}

	return New("error has been masked")
}

// NewHTTPError returns an HTTPError.
func NewHTTPError(c int, err error) error {
	// there is a limitation with how err is handled by grpc library.
	// we will come to this once we have format for error types
	return status.Error(codes.Code(c), err.Error())
}

// GetHTTPDetails returns a http code and message
// default http code is 500.
func GetHTTPDetails(err error) (code int, msg string) {
	serr, ok := status.FromError(err)
	if !ok {
		return http.StatusInternalServerError, err.Error()
	}

	code = int(serr.Code())

	// if this is a grpc code, then convert it
	if code < http.StatusContinue {
		code = runtime.HTTPStatusFromCode(serr.Code())
	}

	return code, serr.Message()
}

// withStack holds the error and caller stack trace
type withStack struct {
	err   error
	trace *stack
}

// Error returns the error string.
func (ws *withStack) Error() string {
	return ws.err.Error()
}

// WithStackTrace attaches stack trace to error.
// Note: if the err already holds a stack trace, that trace will be replace with latest
func WithStackTrace(err error) error {
	if err == nil {
		return nil
	}

	err = detachStackTrace(err)
	return &withStack{
		err:   err,
		trace: callers(),
	}
}

// detachStackTrace is a helper function to detach the stack trace attached to error if any.
func detachStackTrace(err error) error {
	if err == nil {
		return nil
	}

	if wst, ok := err.(*withStack); ok {
		return wst.err
	}

	return err
}

// StackTrace returns the stack trace attached to the error if any
func StackTrace(err error) string {
	if err == nil {
		return ""
	}

	wst, ok := err.(*withStack)
	if !ok {
		return ""
	}

	st := wst.trace.StackTrace()
	return fmt.Sprintf("%+v", st)
}

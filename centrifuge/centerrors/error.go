package centerrors

import (
	"fmt"
	"reflect"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/errors"
	"github.com/centrifuge/go-centrifuge/centrifuge/code"
	"github.com/go-errors/errors"
)

const (
	// RequiredField error when required field is empty
	RequiredField = "Required field"

	// NilDocument error when document passed is Nil
	NilDocument = "Nil document"

	// IdentifierReUsed error when same identifier is re-used
	IdentifierReUsed = "Identifier re-used"

	// NilDocumentData error when document data is Nil
	NilDocumentData = "Nil document data"

	// RequirePositiveNumber error when amount or any such is zero or negative
	RequirePositiveNumber = "Require positive number"
)

// errpb is the type alias for errorspb.Error
type errpb errorspb.Error

// Error implements the error interface
// message format: [code]message: [sub errors if any]
func (err *errpb) Error() string {
	if err.Errors == nil || len(err.Errors) == 0 {
		return fmt.Sprintf("[%d]%s", err.Code, err.Message)
	}

	return fmt.Sprintf("[%d]%s: %v", err.Code, err.Message, err.Errors)
}

// New constructs a new error with code and error message
func New(code code.Code, message string) error {
	return NewWithErrors(code, message, nil)
}

// NewWithErrors constructs a new error with code, error message, and errors
func NewWithErrors(c code.Code, message string, errors map[string]string) error {
	if c == code.Ok {
		return nil
	}

	return &errpb{
		Code:    int32(c),
		Message: message,
		Errors:  errors,
	}
}

// P2PError represents p2p error type
type Error struct {
	err *errorspb.Error
}

// FromError constructs and returns errorspb.Error
// if bool true, conversion to Error successful
// else failed and returns unknown Error
func FromError(err error) (*Error, bool) {
	if err == nil {
		return &Error{err: &errorspb.Error{Code: int32(code.Ok)}}, true
	}

	errpb, ok := err.(*errpb)
	if !ok {
		errors := make(map[string]string)
		errors["unkownError"] = err.Error()
		return &Error{err: &errorspb.Error{Code: int32(code.Unknown), Message: err.Error(), Errors: errors}}, false
	}

	return &Error{err: (*errorspb.Error)(errpb)}, true
}

// Code returns the error code
func (p2pErr *Error) Code() code.Code {
	if p2pErr == nil || p2pErr.err == nil {
		return code.Ok
	}

	return code.To(p2pErr.err.Code)
}

// Message returns error message
func (p2pErr *Error) Message() string {
	if p2pErr == nil || p2pErr.err == nil {
		return ""
	}

	return p2pErr.err.Message
}

// Errors returns map errors passed
func (p2pErr *Error) Errors() map[string]string {
	if p2pErr == nil || p2pErr.err == nil {
		return nil
	}

	return p2pErr.err.Errors
}

// NilError returns error with Type added to message
func NilError(param interface{}) error {
	return errors.Errorf("NIL %v provided", reflect.TypeOf(param))
}

// Wrap appends msg to errpb.Message if it is of type *errpb
// else appends the msg to error through fmt.Errorf
func Wrap(err error, msg string) error {
	if err == nil {
		return fmt.Errorf(msg)
	}

	errpb, ok := err.(*errpb)
	if !ok {
		return fmt.Errorf("%s: %v", msg, err)
	}

	errpb.Message = fmt.Sprintf("%s: %v", msg, errpb.Message)
	return errpb
}

//getNextErrorId returns a new unique key for an error
//For example two error having the same key called 'errorX'
// the second key would use 'errorX_2' instead of 'errorX'
func getNextErrorId(errors map[string]string, key string) string {
	counter := 2
	isUnique := false
	for isUnique != true {

		uniqueKey := fmt.Sprintf("%s_%v", key, counter)
		if errors[uniqueKey] == "" {
			return uniqueKey
		}
		counter++

	}
	return ""
}

func WrapErrors(errDst error, errSrc error) error {

	if errDst == nil {
		return errSrc
	}

	if errSrc == nil {
		return errDst
	}

	errorSrc, okSrc := FromError(errSrc)

	errorDst, okDst := FromError(errDst)

	if !okDst && okSrc || errorDst.Code() == code.Unknown {
		// unknown error in dst prefers src error code and message
		errorDst.err.Code = errorSrc.err.Code
		errorDst.err.Message = errorSrc.err.Message
	}

	for errorKey, errorValue := range errorSrc.err.Errors {
		if errorDst.err.Errors[errorKey] != "" {

			uniqueKey := getNextErrorId(errorDst.err.Errors, errorKey)
			errorDst.err.Errors[uniqueKey] = errorValue

		} else {
			errorDst.err.Errors[errorKey] = errorValue
		}

	}

	return NewWithErrors(errorDst.Code(), errorDst.Message(), errorDst.Errors())

}

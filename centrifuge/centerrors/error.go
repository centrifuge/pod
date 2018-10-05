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

// Error represents cent error type
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
		return &Error{err: &errorspb.Error{Code: int32(code.Unknown), Message: err.Error()}}, false
	}

	return &Error{err: (*errorspb.Error)(errpb)}, true
}

// Code returns the error code
func (centErr *Error) Code() code.Code {
	if centErr == nil || centErr.err == nil {
		return code.Ok
	}

	return code.To(centErr.err.Code)
}

// Message returns error message
func (centErr *Error) Message() string {
	if centErr == nil || centErr.err == nil {
		return ""
	}

	return centErr.err.Message
}

// Errors returns map errors passed
func (centErr *Error) Errors() map[string]string {
	if centErr == nil || centErr.err == nil {
		return nil
	}

	return centErr.err.Errors
}

// NilError returns error with Type added to message
func NilError(param interface{}) error {
	return errors.Errorf("NIL %v provided", reflect.TypeOf(param))
}

// wrapMsg appends msg to errpb.Message if it is of type *errpb
// else appends the msg to error through fmt.Errorf
func wrapMsg(err error, msg string) error {
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

// getNextErrorId returns a new unique key for an error
// For example two error having the same key called 'errorX'
// the second key would use 'errorX_2' instead of 'errorX'
func getNextErrorId(errors map[string]string, key string) string {
	counter := 2
	isUnique := false
	uniqueKey := ""
	for isUnique != true {

		uniqueKey = fmt.Sprintf("%s_%v", key, counter)
		if errors[uniqueKey] == "" {
			isUnique = true
		}
		counter++

	}
	return uniqueKey
}

func wrapErrorMaps(errorDst, errorSrc map[string]string) map[string]string {

	if errorDst == nil {
		return errorSrc
	}

	if errorSrc == nil {
		return errorDst
	}

	for errorKey, errorValue := range errorSrc {
		if errorDst[errorKey] != "" {

			uniqueKey := getNextErrorId(errorDst, errorKey)
			errorDst[uniqueKey] = errorValue

		} else {
			errorDst[errorKey] = errorValue
		}
	}

	return errorDst

}

func wrapErrors(errorDst *errpb, errorSrc *errpb) error {

	if errorDst.Code == int32(code.Unknown) {

		errorDst.Code = errorSrc.Code
	}

	errorDst.Message = fmt.Sprintf("%v: %v", errorDst.Message, errorSrc.Message)

	errorDst.Errors = wrapErrorMaps(errorDst.Errors, errorSrc.Errors)

	return errorDst

}

func handleErrorWrapCases(errDst, errSrc error) error {

	errorDst, dstIsCentError := errDst.(*errpb)
	errorSrc, srcIsCentError := errSrc.(*errpb)

	// if no centerror is used a standard error will be returned
	if !dstIsCentError && !srcIsCentError {
		return fmt.Errorf("%v: %v", errDst, errSrc)
	}

	// if one of the two errors is a centerror the standard error will be appended to Error.Message
	if dstIsCentError && !srcIsCentError {
		errorDst.Message = fmt.Sprintf("%s: %v", errorDst.Message, errSrc)
		return errorDst

	}
	if !dstIsCentError && srcIsCentError {
		errorSrc.Message = fmt.Sprintf("%v: %s", errDst, errorSrc.Message)
		return errorSrc

	}

	// case: centError & centError
	if dstIsCentError && srcIsCentError {
		return wrapErrors(errorDst, errorSrc)

	}

	return fmt.Errorf("a error occured while wrapping other error messages")

}

// Wrap can wrap an error into an other.
// src can be a string message, a implementation of the error interface or from type centerrors.Error
func Wrap(errDst error, src interface{}) error {

	errMsg, isStringSrc := src.(string)
	if isStringSrc {

		return wrapMsg(errDst, errMsg)
	}

	errSrc, ok := src.(error)
	if !ok {
		return fmt.Errorf("wrap error needs a string or an error as source")
	}

	if errDst == nil {
		return errSrc
	}

	if errSrc == nil {
		return errDst
	}

	return handleErrorWrapCases(errDst, errSrc)

}

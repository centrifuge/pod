package coreapi

import "github.com/centrifuge/pod/errors"

const (
	ErrRequestBodyRead          = errors.Error("couldn't read request body")
	ErrRequestPayloadJSONDecode = errors.Error("couldn't JSON decode the request body")
	ErrInvalidDocumentID        = errors.Error("invalid document identifier")
	ErrDocumentNotFound         = errors.Error("document not found")
	ErrAccountIDInvalid         = errors.Error("account ID is invalid")
	ErrAccountNotFound          = errors.Error("account not found")
	ErrAccountGeneration        = errors.Error("couldn't generate account")
	ErrPayloadSigning           = errors.Error("couldn't sign payload")
	ErrAccountsRetrieval        = errors.Error("couldn't retrieve accounts")
)

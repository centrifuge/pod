package coreapi

import "github.com/centrifuge/go-centrifuge/errors"

const (
	// ErrInvalidDocumentID is a sentinel error for invalid document identifiers.
	ErrInvalidDocumentID = errors.Error("invalid document identifier")

	// ErrDocumentNotFound is a sentinel error for missing documents.
	ErrDocumentNotFound = errors.Error("document not found")

	// ErrAccountIDInvalid is a sentinel error for invalid account IDs.
	ErrAccountIDInvalid = errors.Error("account ID is invalid")

	// ErrAccountNotFound is a sentinel error for when account is missing.
	ErrAccountNotFound = errors.Error("account not found")
)

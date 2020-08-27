package coreapi

import "github.com/centrifuge/go-centrifuge/errors"

const (
	// ErrInvalidDocumentID is a sentinel error for invalid document identifiers.
	ErrInvalidDocumentID = errors.Error("invalid document identifier")

	// ErrDocumentNotFound is a sentinel error for missing documents.
	ErrDocumentNotFound = errors.Error("document not found")
)

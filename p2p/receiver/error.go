package receiver

import (
	"github.com/centrifuge/pod/errors"
)

const (
	// ErrAccessDenied must be used when the requester does not have access rights for the document requested
	ErrAccessDenied = errors.Error("requester does not have access")

	// ErrInvalidAccessType must be used when the access type found in the request is invalid
	ErrInvalidAccessType = errors.Error("invalid access type")
)

package identity

import (
	"github.com/centrifuge/go-centrifuge/errors"
)

const (
	// ErrInvalidDIDLength must be used with invalid bytelength when attempting to convert to a DID
	ErrInvalidDIDLength = errors.Error("invalid DID length")
)

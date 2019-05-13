package ideth

import (
	"github.com/centrifuge/go-centrifuge/errors"
)

const (
	// ErrSignature must be used if a signature is invalid
	ErrSignature = errors.Error("invalid signature")
)

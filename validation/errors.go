package validation

import "github.com/centrifuge/go-centrifuge/errors"

const (
	ErrMissingURL       = errors.Error("missing URL")
	ErrInvalidURL       = errors.Error("invalid URL")
	ErrInvalidAccountID = errors.Error("invalid account ID")
)

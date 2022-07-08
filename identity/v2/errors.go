package v2

import "github.com/centrifuge/go-centrifuge/errors"

const (
	ErrAccountRetrieval = errors.Error("couldn't retrieve account")
	ErrKeyRetrieval     = errors.Error("couldn't retrieve key")
	ErrKeyNotFound      = errors.Error("key not found")
	ErrKeyRevoked       = errors.Error("key is revoked")
)

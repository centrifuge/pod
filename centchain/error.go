package centchain

import "github.com/centrifuge/go-centrifuge/errors"

const (
	// ErrExtrinsic is a generic error type to be used for cent-chain errors
	ErrExtrinsic = errors.Error("error on cent-chain tx layer")

	// ErrBlockNotReady error when block is not ready yet
	ErrBlockNotReady = errors.Error("required result to be 32 bytes, but got 0")
)

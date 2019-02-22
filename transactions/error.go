package transactions

import "github.com/centrifuge/go-centrifuge/errors"

const (

	// ErrTransactionBootstrap error when bootstrap fails.
	ErrTransactionBootstrap = errors.Error("failed to bootstrap transactions")

	// ErrTransactionMissing error when transaction doesn't exist in Repository.
	ErrTransactionMissing = errors.Error("transaction doesn't exist")

	// ErrKeyConstructionFailed error when the key construction failed.
	ErrKeyConstructionFailed = errors.Error("failed to construct transaction key")
)

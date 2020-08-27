package ethereum

import "github.com/centrifuge/go-centrifuge/errors"

const (
	// ErrTransactionUnderpriced transaction is under priced
	ErrTransactionUnderpriced = errors.Error("replacement transaction underpriced")

	// ErrNonceTooLow nonce is too low
	ErrNonceTooLow = errors.Error("nonce too low")

	// ErrTransactionFailed error when transaction failed
	ErrTransactionFailed = errors.Error("Transaction failed")

	// ErrEthTransaction is a generic error type to be used for Ethereum errors
	ErrEthTransaction = errors.Error("error on ethereum tx layer")

	// ErrEthURL is used when failed to parse ethereum node URL
	ErrEthURL = errors.Error("failed to parse ethereum node URL")

	// ErrEthKeyNotProvided holds specific error when ethereum key is not provided
	ErrEthKeyNotProvided = errors.Error("Ethereum Key not provided")
)

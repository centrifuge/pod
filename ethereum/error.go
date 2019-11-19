package ethereum

import "github.com/centrifuge/go-centrifuge/errors"

const (
	// ErrTransactionUnderpriced transaction is under priced
	ErrTransactionUnderpriced = errors.Error("replacement transaction underpriced")

	// ErrUsrTransactionUnderpriced transaction is under priced
	ErrUsrTransactionUnderpriced = errors.Error("Transaction gas price supplied is too low")

	// ErrNonceTooLow nonce is too low
	ErrNonceTooLow = errors.Error("nonce too low")

	// ErrUsrNonceTooLow nonce is too low
	ErrUsrNonceTooLow = errors.Error("There is another transaction with same nonce in the queue")

	// ErrTransactionFailed error when transaction failed
	ErrTransactionFailed = errors.Error("Transaction failed")

	// ErrEthTransaction is a generic error type to be used for Ethereum errors
	ErrEthTransaction = errors.Error("error on ethereum tx layer")

	// ErrEthURL is used when failed to parse ethereum node URL
	ErrEthURL = errors.Error("failed to parse ethereum node URL")

	// ErrEthKeyNotProvided holds specific error when ethereum key is not provided
	ErrEthKeyNotProvided = errors.Error("Ethereum Key not provided")
)

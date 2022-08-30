package v2

import "github.com/centrifuge/go-centrifuge/errors"

const (
	ErrAccountRetrieval          = errors.Error("couldn't retrieve account")
	ErrKeyRetrieval              = errors.Error("couldn't retrieve key")
	ErrKeyNotFound               = errors.Error("key not found")
	ErrLatestBlockRetrieval      = errors.Error("couldn't retrieve last block")
	ErrInvalidKeyData            = errors.Error("invalid key data")
	ErrKeyRevoked                = errors.Error("key is revoked")
	ErrInvalidSignature          = errors.Error("invalid signature")
	ErrMetadataRetrieval         = errors.Error("couldn't retrieve latest metadata")
	ErrAccountStorageKeyCreation = errors.Error("couldn't create account storage key")
	ErrAccountStorageRetrieval   = errors.Error("couldn't retrieve account from storage")
	ErrInvalidAccount            = errors.Error("invalid account")
	ErrInvalidWebhookURL         = errors.Error("invalid webhook URL")
	ErrSigningKeyPairGeneration  = errors.Error("couldn't generate signing key pair")
	ErrAccountCreation           = errors.Error("couldn't create account")
	ErrAccountStorage            = errors.Error("couldn't store account")
	ErrProtocolIDDispatch        = errors.Error("couldn't dispatch protocol ID")
	ErrAccountIDCreation         = errors.Error("couldn't create account ID")
	ErrAccountNotFound           = errors.Error("account not found")
)

package access

import "github.com/centrifuge/pod/errors"

const (
	ErrSS58AddressDecode             = errors.Error("couldn't decode SS58 address")
	ErrInvalidSignature              = errors.Error("invalid signature")
	ErrInvalidIdentity               = errors.Error("invalid identity")
	ErrDelegatorNotFound             = errors.Error("delegator not found")
	ErrAccountProxiesRetrieval       = errors.Error("couldn't retrieve account proxies")
	ErrInvalidProxyType              = errors.Error("invalid proxy type")
	ErrInvalidDelegate               = errors.Error("invalid delegate")
	ErrPodAdminRetrieval             = errors.Error("couldn't retrieve pod admin")
	ErrNotAdminAccount               = errors.Error("provided account is not an admin")
	ErrInvalidPoolPermissions        = errors.Error("invalid pool permissions")
	ErrInvestorAccessParamsRetrieval = errors.Error("investor access params retrieval")
	ErrActiveLoansRetrieval          = errors.Error("active loans retrieval")
	ErrActiveLoanNotFound            = errors.Error("active loan not found")
	ErrDocumentIDRetrieval           = errors.Error("document ID retrieval")
	ErrDocumentIDMismatch            = errors.Error("document IDs do not match")
)

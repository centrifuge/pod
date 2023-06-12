package access

import "github.com/centrifuge/pod/errors"

const (
	ErrSS58AddressDecode             = errors.Error("SS58 address decoding")
	ErrAccountProxiesRetrieval       = errors.Error("account proxies retrieval")
	ErrInvalidDelegate               = errors.Error("invalid delegate")
	ErrPodAdminRetrieval             = errors.Error("pod admin retrieval")
	ErrNotAdminAccount               = errors.Error("provided account is not an admin")
	ErrPermissionRolesRetrievalError = errors.Error("permission roles retrieval")
	ErrInvalidPoolPermissions        = errors.Error("invalid pool permissions")
	ErrInvestorAccessParamsRetrieval = errors.Error("investor access params retrieval")
	ErrDocumentIDRetrieval           = errors.Error("document ID retrieval")
	ErrDocumentIDMismatch            = errors.Error("document IDs do not match")
	ErrCreatedLoanRetrieval          = errors.Error("created loan retrieval")
	ErrNoValidationServiceForPath    = errors.Error("no validator service for request path")
	ErrIdentityNotFound              = errors.Error("identity not found")
	ErrInvalidProxyType              = errors.Error("invalid proxy type")
	ErrInvalidAuthorizationHeader    = errors.Error("invalid authorization header")
)

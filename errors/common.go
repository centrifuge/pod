package errors

const (
	ErrContextAccountRetrieval  = Error("couldn't retrieve account from context")
	ErrContextIdentityRetrieval = Error("couldn't retrieve identity from context")
	ErrMetadataRetrieval        = Error("couldn't retrieve metadata")
	ErrCallCreation             = Error("couldn't create call")
	ErrProxyCall                = Error("couldn't execute proxy call")
	ErrValidation               = Error("validation failed")
	ErrRequestNil               = Error("nil request")
	ErrRequestInvalid           = Error("invalid request")
	ErrStorageKeyCreation       = Error("couldn't create storage key")
	ErrPodOperatorRetrieval     = Error("couldn't retrieve pod operator")
)

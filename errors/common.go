package errors

const (
	ErrContextAccountRetrieval  = Error("couldn't retrieve account from context")
	ErrContextIdentityRetrieval = Error("couldn't retrieve identity from context")
	ErrAccountProxyRetrieval    = Error("couldn't retrieve account proxy")
	ErrMetadataRetrieval        = Error("couldn't retrieve metadata")
	ErrCallCreation             = Error("couldn't create call")
	ErrProxyCall                = Error("couldn't execute proxy call")
)

package errors

const (
	ErrAccountRetrieval      = Error("couldn't retrieve account")
	ErrAccountProxyRetrieval = Error("couldn't retrieve account proxy")
	ErrMetadataRetrieval     = Error("couldn't retrieve metadata")
	ErrCallCreation          = Error("couldn't create call")
	ErrProxyCall             = Error("couldn't execute proxy call")
)

package errors

const (
	ErrContextAccountRetrieval  = Error("couldn't retrieve account from context")
	ErrContextIdentityRetrieval = Error("couldn't retrieve identity from context")
	ErrMetadataRetrieval        = Error("couldn't retrieve metadata")
	ErrCallCreation             = Error("couldn't create call")
	ErrExtrinsicSubmission      = Error("couldn't submit extrinsic")
	ErrExtrinsicSubmitAndWatch  = Error("couldn't submit and watch extrinsic")
	ErrProxyCall                = Error("couldn't execute proxy call")
	ErrRequestNil               = Error("nil request")
	ErrRequestInvalid           = Error("invalid request")
	ErrStorageKeyCreation       = Error("couldn't create storage key")
	ErrPodOperatorRetrieval     = Error("couldn't retrieve pod operator")
)

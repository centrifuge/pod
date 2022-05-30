package ipfs_pinning

import "github.com/centrifuge/go-centrifuge/errors"

const (
	ErrMissingRequest        = errors.Error("request missing")
	ErrInvalidCIDVersion     = errors.Error("invalid CID version")
	ErrMissingPinningData    = errors.Error("missing pinning data")
	ErrInvalidPinningRequest = errors.Error("invalid pinning request")
	ErrMissingIPFSHash       = errors.Error("IPFS hash missing")
	ErrMissingAPIJWT         = errors.Error("API JWT missing")
	ErrHTTPRequestCreation   = errors.Error("couldn't create HTTP request")
	ErrHTTPRequest           = errors.Error("couldn't perform HTTP request")
	ErrHTTPResponseBodyRead  = errors.Error("couldn't read HTTP response body")
	ErrHTTPResponse          = errors.Error("HTTP response error")
	ErrResponseJSONUnmarshal = errors.Error("couldn't unmarshal response")
	ErrRequestJSONMarshal    = errors.Error("couldn't marshal request to JSON")
)

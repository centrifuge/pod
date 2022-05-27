package ipfs_pinning

import "github.com/centrifuge/go-centrifuge/errors"

const (
	ErrInvalidURL            = errors.Error("invalid URL")
	ErrHTTPRequestCreation   = errors.Error("couldn't create HTTP request")
	ErrHTTPRequest           = errors.Error("couldn't perform HTTP request")
	ErrHTTPResponseBodyRead  = errors.Error("couldn't read HTTP response body")
	ErrHTTPResponse          = errors.Error("HTTP response error")
	ErrResponseJSONUnmarshal = errors.Error("couldn't unmarshal response")
	ErrRequestJSONMarshal    = errors.Error("couldn't marshal request to JSON")
)

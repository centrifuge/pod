package errors

import "net/http"

// Code represents an error code
// Alias for int32
type Code int32

const (
	// Ok no error code
	Ok Code = 0

	// Unknown operation cancelled due  unhandled error
	Unknown Code = 1

	// NetworkMismatch operation cancelled due to Node network mismatch
	NetworkMismatch Code = 2

	// VersionMismatch operation cancelled due to Node version mismatch
	VersionMismatch Code = 3

	// DocumentInvalid operation cancelled due to invalid document. Check for sub errors if any
	DocumentInvalid Code = 4

	// AuthenticationFailed operation called due to failed auth
	AuthenticationFailed Code = 5

	// AuthorizationFailed operation cancelled due to insufficient permissions
	AuthorizationFailed Code = 6

	// DocumentNotFound operation cancelled due to missing document
	DocumentNotFound Code = 7

	// maxCode for boundary limit. increment this to add new error code
	maxCode Code = 8
)

// httpMapping maps known error codes to HTTP codes
var httpMapping = map[Code]int{
	Ok:                   http.StatusOK,
	Unknown:              http.StatusInternalServerError,
	NetworkMismatch:      http.StatusBadRequest,
	VersionMismatch:      http.StatusBadRequest,
	DocumentInvalid:      http.StatusBadRequest,
	AuthenticationFailed: http.StatusUnauthorized,
	AuthorizationFailed:  http.StatusForbidden,
	DocumentNotFound:     http.StatusNotFound,
}

// GetHTTPCode returns mapped HTTP code for error code
// returns 500 for unknown error
func GetHTTPCode(code Code) int {
	if httpCode, ok := httpMapping[code]; ok {
		return httpCode
	}

	return http.StatusInternalServerError
}

// getCode converts int32 to Code
func getCode(code int32) Code {
	if code >= int32(maxCode) {
		return Unknown
	}

	return Code(code)
}

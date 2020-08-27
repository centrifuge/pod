package extensions

import "github.com/centrifuge/go-centrifuge/errors"

const (

	// ErrArrayIndex must be used for invalid funding array indexes
	ErrArrayIndex = errors.Error("invalid array index")

	// ErrAttributeSetNotFound must be used if the id of a custom attribute set is not found in a model
	ErrAttributeSetNotFound = errors.Error("id of custom attribute set not found in model")

	// ErrDeriveAttr must be used if it is not possible to derive an attribute set from a document
	ErrDeriveAttr = errors.Error("stored attribute set in document has an error")

	// ErrPayload must be used if it is not possible to derive an attribute set from a payload
	ErrPayload = errors.Error("could not derive attribute set from payload")

	// ErrJSON must be used if it is not possible to derive a json
	ErrJSON = errors.Error("could not derive JSON from the attributes")

	// ErrAttributeSetID must be used if the provided attribute set id has an error
	ErrAttributeSetID = errors.Error("attribute set ID needs to be hex or empty")

	// ErrAttrSetSignature must be used if a signature on an attribute set is invalid
	ErrAttrSetSignature = errors.Error("stored signature on attribute set in document has an error")
)

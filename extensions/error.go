package extension

import "github.com/centrifuge/go-centrifuge/errors"

const (

	// ErrFundingIndex must be used for invalid funding array indexes
	ErrArrayIndex = errors.Error("invalid array index")

	// ErrNoFundingField must be used if a field is not a funding field
	ErrNoFundingField = errors.Error("field not a funding related field")

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

	// ErrFundingSignature must be used if a funding signature is invalid
	ErrFundingSignature = errors.Error("stored funding signature in document has an error")
)

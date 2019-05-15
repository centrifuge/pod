package funding

import "github.com/centrifuge/go-centrifuge/errors"

const (

	// ErrFundingIndex must be used for invalid funding array indexes
	ErrFundingIndex = errors.Error("invalid funding array index")

	// ErrNoFundingField must be used if a field is not a funding field
	ErrNoFundingField = errors.Error("field not a funding related field")

	// ErrFundingNotFound must be used if a funding id is not found in a model
	ErrFundingNotFound = errors.Error("funding not found in model")

	// ErrFundingAttr must be used if it is not possible to derive a funding from a document
	ErrFundingAttr = errors.Error("stored funding in document has an error")

	// ErrPayload must be used if it is not possible to derive an a funding from a payload
	ErrPayload = errors.Error("could not derive funding from payload")

	// ErrJSON must be used if it is not possible to derive a json from a funding
	ErrJSON = errors.Error("could not create json for signing funding ")
)

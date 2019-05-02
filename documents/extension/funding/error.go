package funding

import "github.com/centrifuge/go-centrifuge/errors"

const (

	// ErrFundingIndex must be used for invalid funding array indexes
	ErrFundingIndex = errors.Error("invalid funding array index")

	// ErrNoFundingField must be used if a field is not a funding field
	ErrNoFundingField = errors.Error("field not a funding related field")
)
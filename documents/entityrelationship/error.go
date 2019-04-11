package entityrelationship

import "github.com/centrifuge/go-centrifuge/errors"

const (

	// ErrERNotFound must be used to indicate that entity relationship for provided id is not found in the system
	ErrERNotFound = errors.Error("entity relationship not found in the system database.")

	// ErrERInvalidIdentifier must be used to indicate different identifier
	ErrERInvalidIdentifier = errors.Error("entity relationship contains different entity identifier")

	// ErrERNoToken must be used to indicate missing tokens
	ErrERNoToken = errors.Error("entity relationship contains no access token")

	// ErrNotEntityRelationship must be used if an expected entityRelationship model is not a entityRelationship
	ErrNotEntityRelationship = errors.Error("model not entity relationship")
)

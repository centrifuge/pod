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

	// ErrERInvalidData is sent when the entity relationship data is invalid
	ErrERInvalidData = errors.Error("invalid entity relationship data")

	// ErrDocumentsStorageRetrieval is sent when documents cannot be retrieved from storage
	ErrDocumentsStorageRetrieval = errors.Error("couldn't retrieve documents from storage")

	// ErrRelationshipsStorageRetrieval is sent when relationships cannot be retrieved from storage
	ErrRelationshipsStorageRetrieval = errors.Error("couldn't retrieve relationships from storage")

	// ErrEntityIDNil is sent when the entity ID is nil
	ErrEntityIDNil = errors.Error("entity ID is nil")
)

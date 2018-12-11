package documents

import (
	"fmt"

	"github.com/centrifuge/go-centrifuge/errors"
)

const (

	// ErrDocumentConfigTenantID must be used for errors related to tenantID operations
	ErrDocumentConfigTenantID = errors.Error("error with tenantID operations")

	// ErrDocumentBootstrap must be used for errors related to documents package bootstrapping
	ErrDocumentBootstrap = errors.Error("error when bootstrapping document package")

	// ErrDocumentIdentifier must be used for errors caused by document identifier problems
	ErrDocumentIdentifier = errors.Error("document identifier error")

	// ErrDocumentInvalidType must be used when a provided document type is not valid to be processed by the service
	ErrDocumentInvalidType = errors.Error("document is of invalid type")

	// ErrDocumentProvidedIsNil must be used when the provided document through a function is nil
	ErrDocumentProvidedIsNil = errors.Error("no(nil) document provided")

	// ErrDocumentInvalid must only be used when the reason for invalidity is impossible to determine or the invalidity is caused by validation errors
	ErrDocumentInvalid = errors.Error("document is invalid")

	// ErrDocumentNotFound must be used to indicate that the document for provided id is not found in the system
	ErrDocumentNotFound = errors.Error("document not found in the system database")

	// ErrDocumentVersionNotFound must be used to indicate that the specified version of the document for provided id is not found in the system
	ErrDocumentVersionNotFound = errors.Error("specified version of the document not found in the system database")

	// ErrDocumentPersistence must be used when creating or updating a document in the system database failed
	ErrDocumentPersistence = errors.Error("error encountered when storing document in the system database")

	// ErrDocumentPackingCoreDocument must be used when packing of core document for the given document failed
	ErrDocumentPackingCoreDocument = errors.Error("core document packing failed")

	// ErrDocumentUnPackingCoreDocument must be used when unpacking of core document for the given document failed
	ErrDocumentUnPackingCoreDocument = errors.Error("core document unpacking failed")

	// ErrDocumentPrepareCoreDocument must be used when preparing a new core document fails for the given document
	ErrDocumentPrepareCoreDocument = errors.Error("core document preparation failed")

	// ErrDocumentSigning must be used when document signing related functionality fails
	ErrDocumentSigning = errors.Error("document signing failed")

	// ErrDocumentAnchoring must be used when document anchoring fails
	ErrDocumentAnchoring = errors.Error("document anchoring failed")

	// ErrDocumentCollaborator must be used when there is an error in processing collaborators
	ErrDocumentCollaborator = errors.Error("document collaborator issue")

	// ErrDocumentProof must be used when document proof creation fails
	ErrDocumentProof = errors.Error("document proof error")

	// Document repository errors

	// ErrDocumentRepositoryModelNotRegistered must be used when the model hasn't been registered in the database repository
	ErrDocumentRepositoryModelNotRegistered = errors.Error("document model hasn't been registered in the database repository")

	// ErrDocumentRepositorySerialisation must be used when document repository encounters a marshalling error
	ErrDocumentRepositorySerialisation = errors.Error("document repository encountered a marshalling error")

	// ErrDocumentRepositoryModelNotFound must be used when document repository can not locate the given model
	ErrDocumentRepositoryModelNotFound = errors.Error("document repository could not locate the given model")

	// ErrDocumentRepositoryModelSave must be used when document repository can not save the given model
	ErrDocumentRepositoryModelSave = errors.Error("document repository could not save the given model")

	// ErrDocumentRepositoryModelAllReadyExists must be used when document repository finds an already existing model when saving
	ErrDocumentRepositoryModelAllReadyExists = errors.Error("document repository found an already existing model when saving")

	// ErrDocumentRepositoryModelDoesntExist must be used when document repository does not find an existing model for an update
	ErrDocumentRepositoryModelDoesntExist = errors.Error("document repository did not find an existing model for an update")
)

// Error wraps an error with specific key
// Deprecated: in favour of Error type in `github.com/centrifuge/go-centrifuge/errors`
type Error struct {
	key string
	err error
}

// Error returns the underlying error message
func (e Error) Error() string {
	return fmt.Sprintf("%s : %s", e.key, e.err)
}

// NewError creates a new error from a key and a msg.
// Deprecated: in favour of Error type in `github.com/centrifuge/go-centrifuge/errors`
func NewError(key, msg string) error {
	err := fmt.Errorf(msg)
	return Error{key: key, err: err}
}

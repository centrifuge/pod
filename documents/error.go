package documents

import (
	"fmt"

	"github.com/centrifuge/go-centrifuge/errors"
)

const (

	// ErrDocumentConfigAccountID must be used for errors related to accountID operations
	ErrDocumentConfigAccountID = errors.Error("error with accountID operations")

	// ErrDocumentBootstrap must be used for errors related to documents package bootstrapping
	ErrDocumentBootstrap = errors.Error("error when bootstrapping documents package")

	// ErrDocumentIdentifier must be used for errors caused by Document identifier problems
	ErrDocumentIdentifier = errors.Error("Document identifier error")

	// ErrDocumentInvalidType must be used when a provided Document type is not valid to be processed by the service
	ErrDocumentInvalidType = errors.Error("Document is of invalid type")

	// ErrDocumentNil must be used when the provided Document through a function is nil
	ErrDocumentNil = errors.Error("no(nil) Document provided")

	// ErrDocumentInvalid must only be used when the reason for invalidity is impossible to determine or the invalidity is caused by validation errors
	ErrDocumentInvalid = errors.Error("Document is invalid")

	// ErrDocumentNotFound must be used to indicate that the Document for provided id is not found in the system
	ErrDocumentNotFound = errors.Error("Document not found in the system database")

	// ErrDocumentVersionNotFound must be used to indicate that the specified version of the Document for provided id is not found in the system
	ErrDocumentVersionNotFound = errors.Error("specified version of the Document not found in the system database")

	// ErrDocumentPersistence must be used when creating or updating a Document in the system database failed
	ErrDocumentPersistence = errors.Error("error encountered when storing Document in the system database")

	// ErrDocumentPackingCoreDocument must be used when packing of core Document for the given Document failed
	ErrDocumentPackingCoreDocument = errors.Error("core Document packing failed")

	// ErrDocumentUnPackingCoreDocument must be used when unpacking of core Document for the given Document failed
	ErrDocumentUnPackingCoreDocument = errors.Error("core Document unpacking failed")

	// ErrDocumentPrepareCoreDocument must be used when preparing a new core Document fails for the given Document
	ErrDocumentPrepareCoreDocument = errors.Error("core Document preparation failed")

	// ErrDocumentSigning must be used when Document signing related functionality fails
	ErrDocumentSigning = errors.Error("Document signing failed")

	// ErrDocumentAnchoring must be used when Document anchoring fails
	ErrDocumentAnchoring = errors.Error("Document anchoring failed")

	// ErrDocumentCollaborator must be used when there is an error in processing collaborators
	ErrDocumentCollaborator = errors.Error("Document collaborator issue")

	// ErrDocumentProof must be used when Document proof creation fails
	ErrDocumentProof = errors.Error("Document proof error")

	// Document repository errors

	// ErrDocumentRepositoryModelNotRegistered must be used when the model hasn't been registered in the database repository
	ErrDocumentRepositoryModelNotRegistered = errors.Error("Document model hasn't been registered in the database repository")

	// ErrDocumentRepositorySerialisation must be used when Document repository encounters a marshalling error
	ErrDocumentRepositorySerialisation = errors.Error("Document repository encountered a marshalling error")

	// ErrDocumentRepositoryModelNotFound must be used when Document repository can not locate the given model
	ErrDocumentRepositoryModelNotFound = errors.Error("Document repository could not locate the given model")

	// ErrDocumentRepositoryModelSave must be used when Document repository can not save the given model
	ErrDocumentRepositoryModelSave = errors.Error("Document repository could not save the given model")

	// ErrDocumentRepositoryModelAllReadyExists must be used when Document repository finds an already existing model when saving
	ErrDocumentRepositoryModelAllReadyExists = errors.Error("Document repository found an already existing model when saving")

	// ErrDocumentRepositoryModelDoesntExist must be used when Document repository does not find an existing model for an update
	ErrDocumentRepositoryModelDoesntExist = errors.Error("Document repository did not find an existing model for an update")
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
	err := errors.New(msg)
	return Error{key: key, err: err}
}

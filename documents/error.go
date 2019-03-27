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

	// ErrDocumentIdentifier must be used for errors caused by document identifier problems
	ErrDocumentIdentifier = errors.Error("document identifier error")

	// ErrDocumentInvalidType must be used when a provided document type is not valid to be processed by the service
	ErrDocumentInvalidType = errors.Error("document is of invalid type")

	// ErrDocumentNil must be used when the provided document through a function is nil
	ErrDocumentNil = errors.Error("no(nil) document provided")

	// ErrDocumentNotification must be used when a notification about a document could not be delivered
	ErrDocumentNotification = errors.Error("could not notify of the document")

	// ErrDocumentInvalid must only be used when the reason for invalidity is impossible to determine or the invalidity is caused by validation errors
	ErrDocumentInvalid = errors.Error("document is invalid")

	// ErrDocumentNotFound must be used to indicate that the document for provided id is not found in the system
	ErrDocumentNotFound = errors.Error("document not found in the system database")

	// ErrDocumentVersionNotFound must be used to indicate that the specified version of the document for provided id is not found in the system
	ErrDocumentVersionNotFound = errors.Error("specified version of the document not found in the system database")

	// ErrDocumentPersistence must be used when creating or updating a document in the system database failed
	ErrDocumentPersistence = errors.Error("error encountered when storing document in the system database")

	// ErrDocumentUnPackingCoreDocument must be used when unpacking of core document for the given document failed
	ErrDocumentUnPackingCoreDocument = errors.Error("core document unpacking failed")

	// ErrDocumentPrepareCoreDocument must be used when preparing a new core document fails for the given document
	ErrDocumentPrepareCoreDocument = errors.Error("core document preparation failed")

	// ErrDocumentAnchoring must be used when document anchoring fails
	ErrDocumentAnchoring = errors.Error("document anchoring failed")

	// ErrDocumentProof must be used when document proof creation fails
	ErrDocumentProof = errors.Error("document proof error")

	// Read ACL errors

	// ErrNftNotFound must be used when the NFT is not found in the document
	ErrNftNotFound = errors.Error("nft not found in the Document")

	// ErrNftByteLength must be used when there is a byte length mismatch
	ErrNftByteLength = errors.Error("byte length mismatch")

	// ErrAccessTokenInvalid must be used when the access token is invalid
	ErrAccessTokenInvalid = errors.Error("access token is invalid")

	// ErrAccessTokenNotFound must be used when the access token was not found
	ErrAccessTokenNotFound = errors.Error("access token not found")

	// ErrRequesterNotGrantee must be used when the document requester is not the grantee of the access token
	ErrRequesterNotGrantee = errors.Error("requester is not the same as the access token grantee")

	// ErrGranterNotCollab must be used when the granter of the access token is not a collaborator on the document
	ErrGranterNotCollab = errors.Error("access token granter is not a collaborator on this document")

	// ErrFailedCollaborators must be used when collaborators are not valid
	ErrFailedCollaborators = errors.Error("invalid collaborators")

	// ErrReqDocNotMatch must be used when the requested document does not match the access granted by the access token
	ErrReqDocNotMatch = errors.Error("the document requested does not match the document to which the access token grants access")

	// ErrNFTRoleMissing errors when role to generate proof doesn't exist
	ErrNFTRoleMissing = errors.Error("NFT Role doesn't exist")

	// ErrInvalidIDLength must be used when the identifier bytelength is not 32
	ErrInvalidIDLength = errors.Error("invalid identifier length")

	// ErrInvalidDecimal must be used when given decimal is invalid
	ErrInvalidDecimal = errors.Error("invalid decimal")

	// ErrIdentityNotOwner must be used when an identity which does not own the entity relationship attempts to update the document
	ErrIdentityNotOwner = errors.Error("identity attempting to update the document does not own this entity relationship")
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

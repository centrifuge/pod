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

	// ErrPayloadNil must be used when a required payload is nil
	ErrPayloadNil = errors.Error("no(nil) payload provided")

	// ErrDocumentSchemeUnknown is a sentinel error when the scheme provided is missing in the registry.
	ErrDocumentSchemeUnknown = errors.Error("unknown document scheme provided")

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

	// ErrDocumentAnchoring must be used when document anchoring fails
	ErrDocumentAnchoring = errors.Error("document anchoring failed")

	// ErrDocumentProof must be used when document proof creation fails
	ErrDocumentProof = errors.Error("document proof error")

	// ErrNotPatcher must be used if an expected patcher model does not support patching
	ErrNotPatcher = errors.Error("document doesn't support patching")

	// Coredoc errors

	// ErrCDCreate must be used for coredoc creation/generation errors
	ErrCDCreate = errors.Error("error creating core document")

	// ErrCDNewVersion must be used for coredoc creation/generation errors
	ErrCDNewVersion = errors.Error("error creating new version of core document")

	// ErrCDTree must be used when there are errors during precise-proof tree and root generation
	ErrCDTree = errors.Error("error when generating trees/roots")

	// ErrCDAttribute must be used when there are errors caused by custom model attributes
	ErrCDAttribute = errors.Error("model attribute error")

	// ErrCDStatus is a sentinel error used when status is being chnaged from Committed to anything else.
	ErrCDStatus = errors.Error("cannot change the status of a committed document")

	// ErrDocumentNotInAllowedState is a sentinel error used when a document is not in allowed state for certain op
	ErrDocumentNotInAllowedState = errors.Error("document is not in allowed state")

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

	// ErrReqDocNotMatch must be used when the requested document does not match the access granted by the access token
	ErrReqDocNotMatch = errors.Error("the document requested does not match the document to which the access token grants access")

	// ErrNFTRoleMissing errors when role to generate proof doesn't exist
	ErrNFTRoleMissing = errors.Error("NFT Role doesn't exist")

	// ErrInvalidIDLength must be used when the identifier bytelength is not 32
	ErrInvalidIDLength = errors.Error("invalid identifier length")

	// ErrDocumentNotLatest must be used if document is not the latest version
	ErrDocumentNotLatest = errors.Error("document is not the latest version")

	// others

	// ErrModelNil must be used if the model is nil
	ErrModelNil = errors.Error("model is empty")

	// ErrInvalidDecimal must be used when given decimal is invalid
	ErrInvalidDecimal = errors.Error("invalid decimal")

	// ErrInvalidInt256 must be used when given 256 bit signed integer is invalid
	ErrInvalidInt256 = errors.Error("invalid 256 bit signed integer")

	// ErrIdentityNotOwner must be used when an identity which does not own the entity relationship attempts to update the document
	ErrIdentityNotOwner = errors.Error("identity attempting to update the document does not own this entity relationship")

	// ErrNotImplemented must be used when an method has not been implemented
	ErrNotImplemented = errors.Error("Method not implemented")

	// ErrDocumentConfigNotInitialised is a sentinel error when document config is missing
	ErrDocumentConfigNotInitialised = errors.Error("document config not initialised")

	// ErrDifferentAnchoredAddress is a sentinel error when anchor address is different from the configured one.
	ErrDifferentAnchoredAddress = errors.Error("anchor address is not the node configured address")

	// ErrDocumentIDReused is a sentinel error when identifier is re-used
	ErrDocumentIDReused = errors.Error("document identifier is already used")

	// ErrNotValidAttrType is a sentinel error when an unknown attribute type is given
	ErrNotValidAttrType = errors.Error("not a valid attribute type")

	// ErrEmptyAttrLabel is a sentinel error when the attribute label is empty
	ErrEmptyAttrLabel = errors.Error("empty attribute label")

	// ErrWrongAttrFormat is a sentinel error when the attribute format is wrong
	ErrWrongAttrFormat = errors.Error("wrong attribute format")

	// ErrDocumentValidation must be used when document validation fails
	ErrDocumentValidation = errors.Error("document validation failure")
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

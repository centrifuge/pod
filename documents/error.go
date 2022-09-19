package documents

import (
	"github.com/centrifuge/go-centrifuge/errors"
)

const (

	// ErrAccountNotFoundInContext must be used when the account cannot be retrieved from context
	ErrAccountNotFoundInContext = errors.Error("account not found in context")

	// ErrDocumentBootstrap must be used for errors related to documents package bootstrapping
	ErrDocumentBootstrap = errors.Error("error when bootstrapping documents package")

	// ErrDocumentIdentifier must be used for errors caused by document identifier problems
	ErrDocumentIdentifier = errors.Error("document identifier error")

	// ErrDocumentInvalidType must be used when a provided document type is not valid to be processed by the service
	ErrDocumentInvalidType = errors.Error("document is of invalid type")

	// ErrDocumentNil must be used when the provided document through a function is nil
	ErrDocumentNil = errors.Error("no document provided")

	// ErrPayloadNil must be used when a required payload is nil
	ErrPayloadNil = errors.Error("no payload provided")

	// ErrDocumentSchemeUnknown is a sentinel error when the scheme provided is missing in the registry.
	ErrDocumentSchemeUnknown = errors.Error("unknown document scheme provided")

	// ErrDocumentConvertInvalidSchema is sent when attempting to convert a document with an invalid schema
	ErrDocumentConvertInvalidSchema = errors.Error("trying to convert document with invalid schema")

	// ErrDocumentInvalid must only be used when the reason for invalidity is impossible to determine or the invalidity is caused by validation errors
	ErrDocumentInvalid = errors.Error("document is invalid")

	// ErrDocumentTimestampInvalid is used when the document timestamp is invalid.
	ErrDocumentTimestampInvalid = errors.Error("document timestamp is invalid")

	// ErrDocumentNotFound must be used to indicate that the document for provided id is not found in the system
	ErrDocumentNotFound = errors.Error("document not found in the system database")

	// ErrDocumentVersionNotFound must be used to indicate that the specified version of the document for provided id is not found in the system
	ErrDocumentVersionNotFound = errors.Error("specified version of the document not found in the system database")

	// ErrDocumentPersistence must be used when creating or updating a document in the system database failed
	ErrDocumentPersistence = errors.Error("error encountered when storing document in the system database")

	// ErrDocumentUnPackingCoreDocument must be used when unpacking of core document for the given document failed
	ErrDocumentUnPackingCoreDocument = errors.Error("core document unpacking failed")

	// ErrDocumentPackingCoreDocument must be used when packing of core document for the given document failed
	ErrDocumentPackingCoreDocument = errors.Error("core document packing failed")

	// ErrDocumentAnchoring must be used when document anchoring fails
	ErrDocumentAnchoring = errors.Error("document anchoring failed")

	// ErrDocumentProof must be used when document proof creation fails
	ErrDocumentProof = errors.Error("document proof error")

	// ErrNotPatcher must be used if an expected patcher model does not support patching
	ErrNotPatcher = errors.Error("document doesn't support patching")

	// Coredoc errors

	// ErrCDCreate must be used for coredoc creation/generation errors
	ErrCDCreate = errors.Error("error creating core document")

	// ErrCDClone must be used for coredoc clone errors
	ErrCDClone = errors.Error("error cloning core document")

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

	// ErrDataTree must be used for data tree errors
	ErrDataTree = errors.Error("data tree error")

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

	// ErrRequesterInvalidAccountID must be used when the requester account ID is invalid
	ErrRequesterInvalidAccountID = errors.Error("invalid requester account ID")

	// ErrGranterInvalidAccountID must be used when the granter account ID is invalid
	ErrGranterInvalidAccountID = errors.Error("invalid granter account ID")

	// ErrGranteeInvalidAccountID must be used when the grantee account ID is invalid
	ErrGranteeInvalidAccountID = errors.Error("invalid grantee account ID")

	// ErrDocumentRetrieval must be used when a document cannot be retrieved from the document service
	ErrDocumentRetrieval = errors.Error("couldn't retrieve document")

	// ErrDocumentSigningKeyValidation must be used when the document signing key cannot be validated
	ErrDocumentSigningKeyValidation = errors.Error("couldn't validate document signing key")

	// others

	// ErrModelNil must be used if the model is nil
	ErrModelNil = errors.Error("model is empty")

	// ErrInvalidDecimal must be used when given decimal is invalid
	ErrInvalidDecimal = errors.Error("invalid decimal")

	// ErrInvalidInt256 must be used when given 256 bit signed integer is invalid
	ErrInvalidInt256 = errors.Error("invalid 256 bit signed integer")

	// ErrIdentityNotOwner must be used when an identity which does not own the entity relationship attempts to update the document
	ErrIdentityNotOwner = errors.Error("identity attempting to update the document does not own this entity relationship")

	// ErrIdentityInvalid must be used when the provided identity does not exist
	ErrIdentityInvalid = errors.Error("invalid identity")

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

	// ErrInvalidAttrTimestamp is a sentinel error when the attribute timestamp is invalid
	ErrInvalidAttrTimestamp = errors.Error("invalid attribute timestamp")

	// ErrDocumentValidation must be used when document validation fails
	ErrDocumentValidation = errors.Error("document validation failure")

	// ErrRoleNotExist must be used when role doesn't exist in the document.
	ErrRoleNotExist = errors.Error("role doesn't exist")

	// ErrRoleExist must be used when role exist in the document.
	ErrRoleExist = errors.Error("role already exists")

	// ErrEmptyRoleKey must be used when role key is empty
	ErrEmptyRoleKey = errors.Error("empty role key")

	// ErrEmptyCollaborators must be used when collaborator list is empty
	ErrEmptyCollaborators = errors.Error("empty collaborators")

	// ErrInvalidRoleKey must be used when role key is not 32 bytes long
	ErrInvalidRoleKey = errors.Error("role key is invalid")

	// ErrTransitionRuleMissing is a sentinel error used when transition rule is missing from the document.
	ErrTransitionRuleMissing = errors.Error("transition rule missing")

	// ErrTemplateAttributeMissing is an error when the template attribute is missing
	ErrTemplateAttributeMissing = errors.Error("template attribute missing")

	// ErrP2PDocumentSend is sent when the document cannot be sent by the p2p client
	ErrP2PDocumentSend = errors.Error("couldn't send document to recipient")

	// ErrP2PDocumentRetrieval is sent when the document cannot be retrieved by the p2p client
	ErrP2PDocumentRetrieval = errors.Error("couldn't get document")

	// ErrDocumentAddUpdateLog is sent when a document update log cannot be added
	ErrDocumentAddUpdateLog = errors.Error("couldn't add update log")

	// ErrDocumentExecuteComputeFields is sent when compute fields cannot be executed
	ErrDocumentExecuteComputeFields = errors.Error("couldn't execute compute fields")

	// ErrDocumentCalculateSigningRoot is sent when the signing root cannot be calculated
	ErrDocumentCalculateSigningRoot = errors.Error("couldn't calculate signing root")

	// ErrDocumentCalculateSignaturesRoot is sent when the signatures root cannot be calculated
	ErrDocumentCalculateSignaturesRoot = errors.Error("couldn't calculate signatures root")

	// ErrDocumentCalculateDocumentRoot is sent when the document root cannot be calculated
	ErrDocumentCalculateDocumentRoot = errors.Error("couldn't calculate document root")

	// ErrAccountSignMessage is sent when a message cannot be signed
	ErrAccountSignMessage = errors.Error("couldn't sign message")

	// ErrDocumentSignaturesRetrieval is sent when document signatures cannot be retrieved
	ErrDocumentSignaturesRetrieval = errors.Error("couldn't retrieve signatures for document")

	// ErrAnchorIDCreation is sent when an anchor ID cannot be created
	ErrAnchorIDCreation = errors.Error("couldn't create anchor ID")

	// ErrDocumentRootCreation is sent when the document root cannot be created
	ErrDocumentRootCreation = errors.Error("couldn't create document root")

	// ErrPreCommitAnchor is sent when an anchor cannot be pre-committed
	ErrPreCommitAnchor = errors.Error("couldn't pre-commit anchor")

	// ErrCommitAnchor is sent when an anchor cannot be committed
	ErrCommitAnchor = errors.Error("couldn't commit anchor")

	// ErrSignaturesRootProofConversion is sent when signatures root proof cannot be converted to a 32 byte slice
	ErrSignaturesRootProofConversion = errors.Error("couldn't convert signatures root proof")

	// ErrDocumentCollaboratorsRetrieval is sent when the document collaborators cannot be retrieved
	ErrDocumentCollaboratorsRetrieval = errors.Error("couldn't get document collaborators")

	// ErrInvalidSigningRoot is sent when the signing root is invalid
	ErrInvalidSigningRoot = errors.Error("invalid signing root")

	// ErrInvalidDocumentRoot is sent when the document root is invalid
	ErrInvalidDocumentRoot = errors.Error("invalid document root")

	// ErrDocumentAuthorRetrieval is sent when the document author cannot be retrieved
	ErrDocumentAuthorRetrieval = errors.Error("couldn't get document author")

	// ErrDocumentSenderNotAuthor is sent when the document sender is not the author
	ErrDocumentSenderNotAuthor = errors.Error("document sender is not the author")

	// ErrDocumentTimestampRetrieval is sent when the document timestamp cannot be retrieved
	ErrDocumentTimestampRetrieval = errors.Error("couldn't get document timestamp")

	// ErrDocumentTooOldToSign is sent when the document's timestamp is too old
	ErrDocumentTooOldToSign = errors.Error("document is too old to sign")

	// ErrDocumentNoSignatures is sent when the document does not have any signatures
	ErrDocumentNoSignatures = errors.Error("document has no signatures")

	// ErrDocumentSignerCollaboratorsRetrieval is sent when the document signer collaborators cannot be retrieved
	ErrDocumentSignerCollaboratorsRetrieval = errors.Error("couldn't get document signer collaborators")

	// ErrDocumentAnchorDataRetrieval is sent when the anchor data cannot be retrieved from the chain
	ErrDocumentAnchorDataRetrieval = errors.Error("couldn't retrieve document anchor data")

	// ErrDocumentRootsMismatch is sent when the document roots do no match
	ErrDocumentRootsMismatch = errors.Error("document roots do not match")

	// ErrDocumentInvalidAnchorTime is sent when a document is anchored after the MaxAuthoredToCommitDuration
	ErrDocumentInvalidAnchorTime = errors.Error("document anchor time is invalid")

	// ErrInvalidDocumentStateTransition is sent when a document state transition cannot be done by the collaborator
	ErrInvalidDocumentStateTransition = errors.Error("invalid document state transition")

	// ErrDocumentAddNFT is sent when an NFT cannot be added to a document
	ErrDocumentAddNFT = errors.Error("couldn't add NFT")

	// ErrAccountIDBytesParsing is sent when account ID bytes cannot be parsed
	ErrAccountIDBytesParsing = errors.Error("couldn't parse account ID bytes")

	// ErrDocumentDataMarshalling is sent when the document data cannot be marshalled
	ErrDocumentDataMarshalling = errors.Error("couldn't marshal document data")

	// ErrDocumentDataUnmarshalling is sent when the document data cannot be unmarshalled
	ErrDocumentDataUnmarshalling = errors.Error("couldn't unmarshal document data")

	// 	ErrCoreDocumentNil is sent when the core document is nil
	ErrCoreDocumentNil = errors.Error("core document is nil")

	// ErrDocumentPatch is sent when the document cannot be patched
	ErrDocumentPatch = errors.Error("couldn't patch document")
)

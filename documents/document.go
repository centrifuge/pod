package documents

import (
	"context"
	"time"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	v2 "github.com/centrifuge/go-centrifuge/identity/v2"
	"github.com/centrifuge/go-centrifuge/storage"
	"github.com/centrifuge/go-substrate-rpc-client/v4/types"
	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("documents")

//go:generate mockery --name Document --structname DocumentMock --filename document_mock.go --inpackage

// Document is an interface to abstract away model specificness like invoice or purchaseOrder
// The interface can cast into the type specified by the document if required
// It should only handle protocol-level document actions
type Document interface {
	storage.Model

	// ID returns the document identifier
	ID() []byte

	// CurrentVersion returns the current version identifier of the document
	CurrentVersion() []byte

	// CurrentVersionPreimage returns the current version pre-image of the document. This is intended to hide the next version of an updated version of the document.
	CurrentVersionPreimage() []byte

	// PreviousVersion returns the previous version identifier of the document
	PreviousVersion() []byte

	// NextVersion returns the next version identifier of the document.
	NextVersion() []byte

	NextPreimage() []byte

	// PackCoreDocument packs the implementing document into a core document
	// Should only be called when the document is about to be put on wire.
	PackCoreDocument() (*coredocumentpb.CoreDocument, error)

	// UnpackCoreDocument takes a core document protobuf and loads the data into the model.
	UnpackCoreDocument(cd *coredocumentpb.CoreDocument) error

	// DocumentType returns the type of the document
	DocumentType() string

	// Scheme returns the scheme of the document data.
	// TODO(ved): remove once the DocumentType is not used anymore.
	Scheme() string

	// CalculateSigningRoot calculates the signing root of the model.
	CalculateSigningRoot() ([]byte, error)

	// CalculateDocumentRoot returns the document root of the model.
	CalculateDocumentRoot() ([]byte, error)

	// CalculateSignaturesRoot returns signatures root of the model.
	CalculateSignaturesRoot() ([]byte, error)

	// AppendSignatures appends the signatures to the model.
	AppendSignatures(signatures ...*coredocumentpb.Signature)

	// Signatures returns a copy of the signatures on the document
	Signatures() []*coredocumentpb.Signature

	// CreateProofs creates precise-proofs for given fields
	CreateProofs(fields []string) (prf *DocumentProof, err error)

	// AddNFT adds an NFT to the document.
	// Note: The document should be anchored after successfully adding the NFT.
	AddNFT(grantReadAccess bool, collectionID types.U64, itemID types.U128) error

	// NFTs returns the list of NFTs created for this model
	NFTs() []*coredocumentpb.NFT

	// GetCollaborators returns the collaborators of this document.
	// filter ids should not be returned
	// Note: returns all the collaborators with Read and Read_Sign permission
	GetCollaborators(filterIDs ...*types.AccountID) (CollaboratorsAccess, error)

	// GetSignerCollaborators works like GetCollaborators except it returns only those with Read_Sign permission.
	GetSignerCollaborators(filterIDs ...*types.AccountID) ([]*types.AccountID, error)

	// AccountCanRead returns true if the account can read the document
	AccountCanRead(accountID *types.AccountID) bool

	// NFTCanRead returns true if the NFT can read the document
	NFTCanRead(registryID []byte, tokenID []byte) bool

	// ATGranteeCanRead returns error if the access token grantee cannot read the document.
	ATGranteeCanRead(ctx context.Context, docSrv Service, identityService v2.Service, tokenID, docID []byte, grantee *types.AccountID) (err error)

	// AddUpdateLog adds a log to the model to persist an update related metadata such as author
	AddUpdateLog(accountID *types.AccountID)

	// Author is the author of the document version represented by the model
	Author() (*types.AccountID, error)

	// Timestamp is the time of update in UTC of the document version represented by the model
	Timestamp() (time.Time, error)

	// CollaboratorCanUpdate returns an error if indicated identity does not have the capacity to update the document.
	CollaboratorCanUpdate(updated Document, collaborator *types.AccountID) error

	// IsCollaborator returns true if the account ID is a collaborator of the document
	IsCollaborator(accountID *types.AccountID) (bool, error)

	// AddAttributes adds a custom attribute to the model with the given value. If an attribute with the given name already exists, it's updated.
	AddAttributes(ca CollaboratorsAccess, prepareNewVersion bool, attrs ...Attribute) error

	// GetAttribute gets the attribute with the given name from the model, it returns a non-nil error if the attribute doesn't exist or can't be retrieved.
	GetAttribute(key AttrKey) (Attribute, error)

	// GetAttributes returns all the attributes in the current document
	GetAttributes() []Attribute

	// DeleteAttribute deletes a custom attribute from the model
	DeleteAttribute(key AttrKey, prepareNewVersion bool) error

	// AttributeExists checks if the attribute with the key exists
	AttributeExists(key AttrKey) bool

	// GetAccessTokens returns the access tokens of a core document
	GetAccessTokens() []*coredocumentpb.AccessToken

	// GetData returns the document data. Ex: invoice.Data
	GetData() interface{}

	// GetStatus returns the status of the document.
	GetStatus() Status

	// SetStatus set the status of the document.
	SetStatus(st Status) error

	// RemoveCollaborators removes collaborators from the current document.
	RemoveCollaborators(collaboratorAccountIDs []*types.AccountID) error

	// GetRole returns the role associated with key.
	GetRole(key []byte) (*coredocumentpb.Role, error)

	// AddRole adds a nw role to the document.
	AddRole(key string, collabs []*types.AccountID) (*coredocumentpb.Role, error)

	// UpdateRole updates existing role with provided collaborators
	UpdateRole(rk []byte, collabs []*types.AccountID) (*coredocumentpb.Role, error)

	// AddTransitionRuleForAttribute creates a new transition rule to edit an attribute.
	// The access is only given to the roleKey which is expected to be present already.
	AddTransitionRuleForAttribute(roleID []byte, key AttrKey) (*coredocumentpb.TransitionRule, error)

	// GetTransitionRule returns the transition rule associated with ruleID in the document.
	GetTransitionRule(ruleID []byte) (*coredocumentpb.TransitionRule, error)

	// DeleteTransitionRule deletes the rule associated with ruleID.
	DeleteTransitionRule(ruleID []byte) error

	// CalculateTransitionRulesFingerprint creates a fingerprint from the transition rules and roles of a document.
	CalculateTransitionRulesFingerprint() ([]byte, error)

	// AddComputeFieldsRule adds a new compute field rule
	AddComputeFieldsRule(wasm []byte, fields []string, targetField string) (*coredocumentpb.TransitionRule, error)

	// ExecuteComputeFields executes all the compute fields and updates the document with target attributes.
	ExecuteComputeFields(timeout time.Duration) error

	// GetComputeFieldsRules returns all the compute fields rules from the document.
	GetComputeFieldsRules() []*coredocumentpb.TransitionRule

	// DeriveFromCreatePayload loads the payload into self.
	DeriveFromCreatePayload(ctx context.Context, payload CreatePayload) error

	// DeriveFromUpdatePayload create the next version of the document.
	// Patches the old data with Payload data
	DeriveFromUpdatePayload(ctx context.Context, payload UpdatePayload) (Document, error)

	// DeriveFromClonePayload clones the transition rules and roles from another document
	// and loads the payload into self
	DeriveFromClonePayload(ctx context.Context, m Document) error

	Patch(payload UpdatePayload) error
}

// CreatePayload holds the scheme, CollaboratorsAccess, Attributes, and Data of the document.
type CreatePayload struct {
	Scheme        string
	Collaborators CollaboratorsAccess
	Attributes    map[AttrKey]Attribute
	Data          []byte
}

// UpdatePayload holds the scheme, CollaboratorsAccess, Attributes, Data and document identifier.
type UpdatePayload struct {
	CreatePayload
	DocumentID []byte
}

// ClonePayload holds the scheme, CollaboratorsAccess, Attributes, Data and document identifier.
type ClonePayload struct {
	Scheme     string
	TemplateID []byte
}

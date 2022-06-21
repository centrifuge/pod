package documents

import (
	"bytes"
	"context"
	"time"

	coredocumentpb "github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/notification"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/gocelery/v2"
	proofspb "github.com/centrifuge/precise-proofs/proofs/proto"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	logging "github.com/ipfs/go-log"
)

// DocumentProof is a value to represent a document and its field proofs
type DocumentProof struct {
	DocumentID     []byte
	VersionID      []byte
	State          string
	FieldProofs    []*proofspb.Proof
	SigningRoot    []byte
	SignaturesRoot []byte
}

// Patcher interface defines a Patch method for inner Models
type Patcher interface {
	// Patch merges payload data into doc
	Patch(payload UpdatePayload) error
}

// Service provides an interface for functions common to all document types
type Service interface {

	// GetCurrentVersion reads a document from the database
	GetCurrentVersion(ctx context.Context, documentID []byte) (Document, error)

	// GetVersion reads a document from the database
	GetVersion(ctx context.Context, documentID []byte, version []byte) (Document, error)

	// DeriveFromCoreDocument derives a doc given the core document.
	DeriveFromCoreDocument(cd coredocumentpb.CoreDocument) (Document, error)

	// CreateProofs creates proofs for the latest version document given the fields
	CreateProofs(ctx context.Context, documentID []byte, fields []string) (*DocumentProof, error)

	// CreateProofsForVersion creates proofs for a particular version of the document given the fields
	CreateProofsForVersion(ctx context.Context, documentID, version []byte, fields []string) (*DocumentProof, error)

	// RequestDocumentSignature Validates and Signs document received over the p2p layer
	RequestDocumentSignature(ctx context.Context, doc Document, collaborator identity.DID) ([]*coredocumentpb.Signature, error)

	// ReceiveAnchoredDocument receives a new anchored document over the p2p layer, validates and updates the document in DB
	ReceiveAnchoredDocument(ctx context.Context, doc Document, collaborator identity.DID) error

	// Derive derives the Document from the Payload.
	// If document_id is provided, it will prepare a new version of the document
	// Document Data will be patched from the old and attributes and collaborators are imported
	// If not provided, it is a fresh document.
	Derive(ctx context.Context, payload UpdatePayload) (Document, error)

	// DeriveClone derives the Document from the Payload, taking the provided template ID as the clone base
	DeriveClone(ctx context.Context, payload ClonePayload) (Document, error)

	// Commit triggers validations, state change and anchor job
	Commit(ctx context.Context, doc Document) (gocelery.JobID, error)

	// Validate takes care of document validation
	Validate(ctx context.Context, doc Document, old Document) error

	// New returns a new uninitialised document.
	New(scheme string) (Document, error)
}

// service implements Service
type service struct {
	config     Config
	repo       Repository
	notifier   notification.Sender
	anchorSrv  anchors.Service
	registry   *ServiceRegistry
	idService  identity.Service
	dispatcher jobs.Dispatcher
}

var srvLog = logging.Logger("document-service")

// DefaultService returns the default implementation of the service
func DefaultService(
	config Config,
	repo Repository,
	anchorSrv anchors.Service,
	registry *ServiceRegistry,
	idService identity.Service,
	dispatcher jobs.Dispatcher) Service {
	return service{
		config:     config,
		repo:       repo,
		anchorSrv:  anchorSrv,
		notifier:   notification.NewWebhookSender(),
		registry:   registry,
		idService:  idService,
		dispatcher: dispatcher,
	}
}

func (s service) GetCurrentVersion(ctx context.Context, documentID []byte) (Document, error) {
	acc, err := contextutil.Account(ctx)
	if err != nil {
		return nil, ErrDocumentConfigAccountID
	}

	accID := acc.GetIdentityID()
	m, err := s.repo.GetLatest(accID, documentID)
	if err != nil {
		return nil, errors.NewTypedError(ErrDocumentNotFound, err)
	}

	return m, nil
}

func (s service) GetVersion(ctx context.Context, documentID []byte, version []byte) (Document, error) {
	return s.getVersion(ctx, documentID, version)
}

func (s service) CreateProofs(ctx context.Context, documentID []byte, fields []string) (*DocumentProof, error) {
	doc, err := s.GetCurrentVersion(ctx, documentID)
	if err != nil {
		return nil, err
	}
	return s.createProofs(doc, fields)
}

func (s service) createProofs(doc Document, fields []string) (*DocumentProof, error) {
	if err := PostAnchoredValidator(s.idService, s.anchorSrv).Validate(nil, doc); err != nil {
		return nil, errors.NewTypedError(ErrDocumentInvalid, err)
	}

	docProof, err := doc.CreateProofs(fields)
	if err != nil {
		return nil, errors.NewTypedError(ErrDocumentProof, err)
	}

	docProof.DocumentID = doc.ID()
	docProof.VersionID = doc.CurrentVersion()
	return docProof, nil
}

func (s service) CreateProofsForVersion(ctx context.Context, documentID, version []byte, fields []string) (*DocumentProof, error) {
	doc, err := s.getVersion(ctx, documentID, version)
	if err != nil {
		return nil, errors.NewTypedError(ErrDocumentNotFound, err)
	}
	return s.createProofs(doc, fields)
}

func (s service) RequestDocumentSignature(ctx context.Context, doc Document, collaborator identity.DID) ([]*coredocumentpb.Signature, error) {
	acc, err := contextutil.Account(ctx)
	if err != nil {
		return nil, ErrDocumentConfigAccountID
	}
	idBytes := acc.GetIdentityID()
	did, err := identity.NewDIDFromBytes(idBytes)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, ErrDocumentNil
	}

	var old Document
	if !utils.IsEmptyByteSlice(doc.PreviousVersion()) {
		old, err = s.repo.Get(did[:], doc.PreviousVersion())
		if err != nil {
			// TODO: should pull old document from peer
			log.Infof("failed to fetch previous document: %v", err)
		}
	}

	if err := RequestDocumentSignatureValidator(s.anchorSrv, s.idService, collaborator, s.config.GetContractAddress(config.AnchorRepo)).Validate(old, doc); err != nil {
		return nil, errors.NewTypedError(ErrDocumentInvalid, err)
	}

	sr, err := doc.CalculateSigningRoot()
	if err != nil {
		return nil, errors.New("failed to get signing root: %v", err)
	}

	srvLog.Infof("document received %x with signing root %x", doc.ID(), sr)

	// If there is a previous version and we have successfully validated the transition then set the signature flag
	sig, err := acc.SignMsg(ConsensusSignaturePayload(sr, old != nil))
	if err != nil {
		return nil, err
	}
	sig.TransitionValidated = old != nil
	doc.AppendSignatures(sig)

	// set the status to committing since we are at requesting signatures stage.
	if err := doc.SetStatus(Committing); err != nil {
		return nil, err
	}

	// Logic for receiving version n (n > 1) of the document for the first time
	// TODO(ved): we should not save the new doc with old identifier. We should sync from the peer.
	if !s.repo.Exists(did[:], doc.ID()) && !utils.IsSameByteSlice(doc.ID(), doc.CurrentVersion()) {
		err = s.repo.Create(did[:], doc.ID(), doc)
		if err != nil {
			return nil, errors.NewTypedError(ErrDocumentPersistence, err)
		}
	}

	err = s.repo.Create(did[:], doc.CurrentVersion(), doc)
	if err != nil {
		return nil, errors.NewTypedError(ErrDocumentPersistence, err)
	}

	srvLog.Infof("signed document %x with version %x", doc.ID(), doc.CurrentVersion())
	return []*coredocumentpb.Signature{sig}, nil
}

func (s service) ReceiveAnchoredDocument(ctx context.Context, doc Document, collaborator identity.DID) error {
	acc, err := contextutil.Account(ctx)
	if err != nil {
		return ErrDocumentConfigAccountID
	}

	idBytes := acc.GetIdentityID()
	did, err := identity.NewDIDFromBytes(idBytes)
	if err != nil {
		return err
	}

	if doc == nil {
		return ErrDocumentNil
	}

	var old Document
	// lets pick the old version of the document from the repo and pass this to the validator
	if !utils.IsEmptyByteSlice(doc.PreviousVersion()) {
		old, err = s.repo.Get(did[:], doc.PreviousVersion())
		if err != nil {
			// TODO(ved): we should pull the old document from the peer
			log.Infof("failed to fetch previous document: %v", err)
		}
	}

	if err := ReceivedAnchoredDocumentValidator(s.idService, s.anchorSrv, collaborator).Validate(old, doc); err != nil {
		return errors.NewTypedError(ErrDocumentInvalid, err)
	}

	// set the status to committed since the document is anchored already.
	if err := doc.SetStatus(Committed); err != nil {
		return err
	}

	err = s.repo.Update(did[:], doc.CurrentVersion(), doc)
	if err != nil {
		return errors.NewTypedError(ErrDocumentPersistence, err)
	}

	notificationMsg := notification.Message{
		EventType:  notification.EventTypeDocument,
		RecordedAt: time.Now().UTC(),
		Document: &notification.DocumentMessage{
			ID:        doc.ID(),
			VersionID: doc.CurrentVersion(),
			From:      collaborator[:],
			To:        did[:],
		},
	}

	// async so that we don't return an error as the p2p reply
	go func() {
		err = s.notifier.Send(ctx, notificationMsg)
		if err != nil {
			log.Error(err)
		}
	}()

	return nil
}

func (s service) Exists(ctx context.Context, documentID []byte) bool {
	acc, err := contextutil.Account(ctx)
	if err != nil {
		return false
	}
	idBytes := acc.GetIdentityID()
	return s.repo.Exists(idBytes, documentID)
}

func (s service) getVersion(ctx context.Context, documentID, version []byte) (Document, error) {
	acc, err := contextutil.Account(ctx)
	if err != nil {
		return nil, ErrDocumentConfigAccountID
	}
	idBytes := acc.GetIdentityID()
	doc, err := s.repo.Get(idBytes, version)
	if err != nil {
		return nil, errors.NewTypedError(ErrDocumentVersionNotFound, err)
	}

	if !bytes.Equal(doc.ID(), documentID) {
		return nil, errors.NewTypedError(ErrDocumentVersionNotFound, errors.New("version is not valid for this identifier"))
	}

	return doc, nil
}

func (s service) DeriveFromCoreDocument(cd coredocumentpb.CoreDocument) (Document, error) {
	if cd.EmbeddedData == nil {
		return nil, errors.New("core document embed data is nil")
	}

	srv, err := s.registry.LocateService(cd.EmbeddedData.TypeUrl)
	if err != nil {
		return nil, err
	}

	return srv.DeriveFromCoreDocument(cd)
}

// Derive looks for specific document type service based in the schema and delegates the Derivation to that service.˜
func (s service) Derive(ctx context.Context, payload UpdatePayload) (Document, error) {
	if len(payload.DocumentID) == 0 {
		doc, err := s.New(payload.Scheme)
		if err != nil {
			return nil, err
		}

		if err := doc.(Deriver).DeriveFromCreatePayload(ctx, payload.CreatePayload); err != nil {
			return nil, errors.NewTypedError(ErrDocumentInvalid, err)
		}
		return doc, nil
	}

	old, err := s.GetCurrentVersion(ctx, payload.DocumentID)
	if err != nil {
		return nil, err
	}

	// check if the scheme is correct
	if old.Scheme() != payload.Scheme {
		return nil, errors.NewTypedError(ErrDocumentInvalidType, errors.New("%v is not an %s", hexutil.Encode(payload.DocumentID), payload.Scheme))
	}

	doc, err := old.(Deriver).DeriveFromUpdatePayload(ctx, payload)
	if err != nil {
		return nil, errors.NewTypedError(ErrDocumentInvalid, err)
	}

	return doc, nil
}

// DeriveClone looks for specific document type service based in the schema and delegates the Derivation of a cloned document to that service.˜
func (s service) DeriveClone(ctx context.Context, payload ClonePayload) (Document, error) {
	_, err := contextutil.AccountDID(ctx)
	if err != nil {
		return nil, ErrDocumentConfigAccountID
	}

	doc, err := s.New(payload.Scheme)
	if err != nil {
		return nil, err
	}

	m, err := s.GetCurrentVersion(ctx, payload.TemplateID)
	if err != nil {
		return nil, err
	}
	if err := doc.(Deriver).DeriveFromClonePayload(ctx, m); err != nil {
		return nil, errors.NewTypedError(ErrDocumentInvalid, err)
	}
	return doc, nil
}

// Validate takes care of document validation
func (s service) Validate(ctx context.Context, doc Document, old Document) error {
	srv, err := s.registry.LocateService(doc.Scheme())
	if err != nil {
		return errors.NewTypedError(ErrDocumentSchemeUnknown, err)
	}

	// If old version provided
	if old != nil {
		if err := UpdateVersionValidator(s.anchorSrv).Validate(old, doc); err != nil {
			return errors.NewTypedError(ErrDocumentValidation, err)
		}
	} else {
		if err := CreateVersionValidator(s.anchorSrv).Validate(nil, doc); err != nil {
			return errors.NewTypedError(ErrDocumentValidation, err)
		}
	}

	// Run document specific validations if any
	return srv.Validate(ctx, doc, old)
}

// Commit triggers validations, state change and anchor job
func (s service) Commit(ctx context.Context, doc Document) (gocelery.JobID, error) {
	acc, err := contextutil.Account(ctx)
	if err != nil {
		return nil, ErrDocumentConfigAccountID
	}
	did := identity.NewDID(common.BytesToAddress(acc.GetIdentityID()))

	// Get latest committed version
	old, err := s.GetCurrentVersion(ctx, doc.ID())
	if err != nil && !errors.IsOfType(ErrDocumentNotFound, err) {
		return nil, err
	}

	if err := s.Validate(ctx, doc, old); err != nil {
		return nil, errors.NewTypedError(ErrDocumentValidation, err)
	}

	if err := doc.SetStatus(Committing); err != nil {
		return nil, err
	}

	if s.repo.Exists(did[:], doc.CurrentVersion()) {
		err = s.repo.Update(did[:], doc.CurrentVersion(), doc)
	} else {
		err = s.repo.Create(did[:], doc.CurrentVersion(), doc)
	}

	if err != nil {
		return nil, errors.NewTypedError(ErrDocumentPersistence, err)
	}

	return initiateAnchorJob(s.dispatcher, did, doc.CurrentVersion(), acc.GetPrecommitEnabled())
}

// New returns a new uninitialised document for the scheme.
func (s service) New(scheme string) (Document, error) {
	srv, err := s.registry.LocateService(scheme)
	if err != nil {
		return nil, ErrDocumentSchemeUnknown
	}

	return srv.New(scheme)
}

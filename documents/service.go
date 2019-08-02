package documents

import (
	"bytes"
	"context"
	"time"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/jobs"
	"github.com/centrifuge/go-centrifuge/notification"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/precise-proofs/proofs/proto"
	"github.com/ethereum/go-ethereum/common/hexutil"
	logging "github.com/ipfs/go-log"
)

// DocumentProof is a value to represent a document and its field proofs
type DocumentProof struct {
	DocumentID  []byte
	VersionID   []byte
	State       string
	FieldProofs []*proofspb.Proof
}

// Patcher interface defines a Patch method for inner Models
type Patcher interface {
	// Patch merges payload data into model
	Patch(payload UpdatePayload) error
}

// Service provides an interface for functions common to all document types
type Service interface {

	// GetCurrentVersion reads a document from the database
	GetCurrentVersion(ctx context.Context, documentID []byte) (Model, error)

	// Exists checks if a document exists
	// Deprecated
	Exists(ctx context.Context, documentID []byte) bool

	// GetVersion reads a document from the database
	GetVersion(ctx context.Context, documentID []byte, version []byte) (Model, error)

	// DeriveFromCoreDocument derives a model given the core document.
	DeriveFromCoreDocument(cd coredocumentpb.CoreDocument) (Model, error)

	// CreateProofs creates proofs for the latest version document given the fields
	CreateProofs(ctx context.Context, documentID []byte, fields []string) (*DocumentProof, error)

	// CreateProofsForVersion creates proofs for a particular version of the document given the fields
	CreateProofsForVersion(ctx context.Context, documentID, version []byte, fields []string) (*DocumentProof, error)

	// RequestDocumentSignature Validates and Signs document received over the p2p layer
	RequestDocumentSignature(ctx context.Context, model Model, collaborator identity.DID) (*coredocumentpb.Signature, error)

	// ReceiveAnchoredDocument receives a new anchored document over the p2p layer, validates and updates the document in DB
	ReceiveAnchoredDocument(ctx context.Context, model Model, collaborator identity.DID) error

	// Create validates and persists Model and returns a Updated model
	// Deprecated
	Create(ctx context.Context, model Model) (Model, jobs.JobID, chan error, error)

	// Update validates and updates the model and return the updated model
	// Deprecated
	Update(ctx context.Context, model Model) (Model, jobs.JobID, chan error, error)

	// CreateModel creates a new model from the payload and initiates the anchor process.
	// Deprecated
	CreateModel(ctx context.Context, payload CreatePayload) (Model, jobs.JobID, error)

	// UpdateModel prepares the next version from the payload and initiates the anchor process.
	// Deprecated
	UpdateModel(ctx context.Context, payload UpdatePayload) (Model, jobs.JobID, error)

	// Derive derives the Model from the Payload.
	// If document_id is provided, it will prepare a new version of the document
	// Document Data will be patched from the old and attributes and collaborators are imported
	// If not provided, it is a fresh document.
	Derive(ctx context.Context, payload UpdatePayload) (Model, error)

	// Commit triggers validations, state change and anchor job
	Commit(ctx context.Context, model Model) (jobs.JobID, error)

	// Validate takes care of document validation
	Validate(ctx context.Context, model Model) error
}

// service implements Service
type service struct {
	config     Config
	repo       Repository
	notifier   notification.Sender
	anchorRepo anchors.AnchorRepository
	registry   *ServiceRegistry
	idService  identity.Service
	queueSrv   queue.TaskQueuer
	jobManager jobs.Manager
}

var srvLog = logging.Logger("document-service")

// DefaultService returns the default implementation of the service
func DefaultService(
	config Config,
	repo Repository,
	anchorRepo anchors.AnchorRepository,
	registry *ServiceRegistry,
	idService identity.Service,
	queueSrv queue.TaskQueuer,
	jobManager jobs.Manager) Service {
	return service{
		config:     config,
		repo:       repo,
		anchorRepo: anchorRepo,
		notifier:   notification.NewWebhookSender(),
		registry:   registry,
		idService:  idService,
		queueSrv:   queueSrv,
		jobManager: jobManager,
	}
}

func (s service) GetCurrentVersion(ctx context.Context, documentID []byte) (Model, error) {
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

func (s service) GetVersion(ctx context.Context, documentID []byte, version []byte) (Model, error) {
	return s.getVersion(ctx, documentID, version)
}

func (s service) CreateProofs(ctx context.Context, documentID []byte, fields []string) (*DocumentProof, error) {
	model, err := s.GetCurrentVersion(ctx, documentID)
	if err != nil {
		return nil, err
	}
	return s.createProofs(model, fields)

}

func (s service) createProofs(model Model, fields []string) (*DocumentProof, error) {
	if err := PostAnchoredValidator(s.idService, s.anchorRepo).Validate(nil, model); err != nil {
		return nil, errors.NewTypedError(ErrDocumentInvalid, err)
	}

	proofs, err := model.CreateProofs(fields)
	if err != nil {
		return nil, errors.NewTypedError(ErrDocumentProof, err)
	}

	return &DocumentProof{
		DocumentID:  model.ID(),
		VersionID:   model.CurrentVersion(),
		FieldProofs: proofs,
	}, nil

}

func (s service) CreateProofsForVersion(ctx context.Context, documentID, version []byte, fields []string) (*DocumentProof, error) {
	model, err := s.getVersion(ctx, documentID, version)
	if err != nil {
		return nil, errors.NewTypedError(ErrDocumentNotFound, err)
	}
	return s.createProofs(model, fields)
}

func (s service) RequestDocumentSignature(ctx context.Context, model Model, collaborator identity.DID) (*coredocumentpb.Signature, error) {
	acc, err := contextutil.Account(ctx)
	if err != nil {
		return nil, ErrDocumentConfigAccountID
	}
	idBytes := acc.GetIdentityID()
	did, err := identity.NewDIDFromBytes(idBytes)
	if err != nil {
		return nil, err
	}
	if model == nil {
		return nil, ErrDocumentNil
	}

	var old Model
	if !utils.IsEmptyByteSlice(model.PreviousVersion()) {
		old, err = s.repo.Get(did[:], model.PreviousVersion())
		if err != nil {
			// TODO: should pull old document from peer
			log.Infof("failed to fetch previous document: %v", err)
		}
	}

	if err := RequestDocumentSignatureValidator(s.anchorRepo, s.idService, collaborator, s.config.GetContractAddress(config.AnchorRepo)).Validate(old, model); err != nil {
		return nil, errors.NewTypedError(ErrDocumentInvalid, err)
	}

	sr, err := model.CalculateSigningRoot()
	if err != nil {
		return nil, errors.New("failed to get signing root: %v", err)
	}

	srvLog.Infof("document received %x with signing root %x", model.ID(), sr)

	sig, err := acc.SignMsg(sr)
	if err != nil {
		return nil, err
	}
	model.AppendSignatures(sig)

	// set the status to committing since we are at requesting signatures stage.
	if err := model.SetStatus(Committing); err != nil {
		return nil, err
	}

	// Logic for receiving version n (n > 1) of the document for the first time
	// TODO(ved): we should not save the new model with old identifier. We should sync from the peer.
	if !s.repo.Exists(did[:], model.ID()) && !utils.IsSameByteSlice(model.ID(), model.CurrentVersion()) {
		err = s.repo.Create(did[:], model.ID(), model)
		if err != nil {
			return nil, errors.NewTypedError(ErrDocumentPersistence, err)
		}
	}

	err = s.repo.Create(did[:], model.CurrentVersion(), model)
	if err != nil {
		return nil, errors.NewTypedError(ErrDocumentPersistence, err)
	}

	srvLog.Infof("signed document %x with version %x", model.ID(), model.CurrentVersion())
	return sig, nil
}

func (s service) ReceiveAnchoredDocument(ctx context.Context, model Model, collaborator identity.DID) error {
	acc, err := contextutil.Account(ctx)
	if err != nil {
		return ErrDocumentConfigAccountID
	}

	idBytes := acc.GetIdentityID()
	did, err := identity.NewDIDFromBytes(idBytes)
	if err != nil {
		return err
	}

	if model == nil {
		return ErrDocumentNil
	}

	var old Model
	// lets pick the old version of the document from the repo and pass this to the validator
	if !utils.IsEmptyByteSlice(model.PreviousVersion()) {
		old, err = s.repo.Get(did[:], model.PreviousVersion())
		if err != nil {
			// TODO(ved): we should pull the old document from the peer
			log.Infof("failed to fetch previous document: %v", err)
		}
	}

	if err := ReceivedAnchoredDocumentValidator(s.idService, s.anchorRepo, collaborator).Validate(old, model); err != nil {
		return errors.NewTypedError(ErrDocumentInvalid, err)
	}

	// set the status to committed since the document is anchored already.
	if err := model.SetStatus(Committed); err != nil {
		return err
	}

	err = s.repo.Update(did[:], model.CurrentVersion(), model)
	if err != nil {
		return errors.NewTypedError(ErrDocumentPersistence, err)
	}

	notificationMsg := notification.Message{
		EventType:    notification.ReceivedPayload,
		AccountID:    did.String(),
		FromID:       hexutil.Encode(collaborator[:]),
		ToID:         did.String(),
		Recorded:     time.Now().UTC(),
		DocumentType: model.DocumentType(),
		DocumentID:   hexutil.Encode(model.ID()),
	}

	// async so that we don't return an error as the p2p reply
	go func() {
		_, err = s.notifier.Send(ctx, notificationMsg)
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

func (s service) getVersion(ctx context.Context, documentID, version []byte) (Model, error) {
	acc, err := contextutil.Account(ctx)
	if err != nil {
		return nil, ErrDocumentConfigAccountID
	}
	idBytes := acc.GetIdentityID()
	model, err := s.repo.Get(idBytes, version)
	if err != nil {
		return nil, errors.NewTypedError(ErrDocumentVersionNotFound, err)
	}

	if !bytes.Equal(model.ID(), documentID) {
		return nil, errors.NewTypedError(ErrDocumentVersionNotFound, errors.New("version is not valid for this identifier"))
	}

	return model, nil
}

func (s service) DeriveFromCoreDocument(cd coredocumentpb.CoreDocument) (Model, error) {
	if cd.EmbeddedData == nil {
		return nil, errors.New("core document embed data is nil")
	}

	srv, err := s.registry.LocateService(cd.EmbeddedData.TypeUrl)
	if err != nil {
		return nil, err
	}

	return srv.DeriveFromCoreDocument(cd)
}

func (s service) Create(ctx context.Context, model Model) (Model, jobs.JobID, chan error, error) {
	srv, err := s.getService(model)
	if err != nil {
		return nil, jobs.NilJobID(), nil, errors.New("failed to get service: %v", err)
	}

	return srv.Create(ctx, model)
}

func (s service) Update(ctx context.Context, model Model) (Model, jobs.JobID, chan error, error) {
	srv, err := s.getService(model)
	if err != nil {
		return nil, jobs.NilJobID(), nil, errors.New("failed to get service: %v", err)
	}

	return srv.Update(ctx, model)
}

func (s service) getService(model Model) (Service, error) {
	return s.registry.LocateService(model.DocumentType())
}

func (s service) CreateModel(ctx context.Context, payload CreatePayload) (Model, jobs.JobID, error) {
	srv, err := s.registry.LocateService(payload.Scheme)
	if err != nil {
		return nil, jobs.NilJobID(), errors.NewTypedError(ErrDocumentSchemeUnknown, err)
	}

	return srv.CreateModel(ctx, payload)
}

func (s service) UpdateModel(ctx context.Context, payload UpdatePayload) (Model, jobs.JobID, error) {
	srv, err := s.registry.LocateService(payload.Scheme)
	if err != nil {
		return nil, jobs.NilJobID(), errors.NewTypedError(ErrDocumentSchemeUnknown, err)
	}

	return srv.UpdateModel(ctx, payload)
}

// Derive looks for specific document type service based in the schema and delegates the Derivation to that service.˜
func (s service) Derive(ctx context.Context, payload UpdatePayload) (Model, error) {
	srv, err := s.registry.LocateService(payload.Scheme)
	if err != nil {
		return nil, errors.NewTypedError(ErrDocumentSchemeUnknown, err)
	}

	return srv.Derive(ctx, payload)
}

// Validate takes care of document validation
func (s service) Validate(ctx context.Context, model Model) error {
	srv, err := s.registry.LocateService(model.Scheme())
	if err != nil {
		return errors.NewTypedError(ErrDocumentSchemeUnknown, err)
	}

	if old, err := s.GetCurrentVersion(ctx, model.ID()); err != nil {
		if !errors.IsOfType(ErrDocumentVersionNotFound, err) {
			return err
		}
		if err := CreateVersionValidator(s.anchorRepo).Validate(nil, model); err != nil {
			return errors.NewTypedError(ErrDocumentValidation, err)
		}
	} else {
		if err := UpdateVersionValidator(s.anchorRepo).Validate(old, model); err != nil {
			return errors.NewTypedError(ErrDocumentValidation, err)
		}
	}
	// Run document specific validations if any
	return srv.Validate(ctx, model)
}

// Commit triggers validations, state change and anchor job
func (s service) Commit(ctx context.Context, model Model) (jobs.JobID, error) {
	did, err := contextutil.AccountDID(ctx)
	if err != nil {
		return jobs.NilJobID(), ErrDocumentConfigAccountID
	}

	if err := s.Validate(ctx, model); err != nil {
		return jobs.NilJobID(), errors.NewTypedError(ErrDocumentValidation, err)
	}

	if err := model.SetStatus(Committing); err != nil {
		return jobs.NilJobID(), err
	}

	err = s.repo.Create(did[:], model.CurrentVersion(), model)
	if err != nil {
		return jobs.NilJobID(), errors.NewTypedError(ErrDocumentPersistence, err)
	}

	jobID := contextutil.Job(ctx)
	jobID, _, err = CreateAnchorJob(ctx, s.jobManager, s.queueSrv, did, jobID, model.CurrentVersion())
	if err != nil {
		return jobs.NilJobID(), err
	}

	return jobID, nil
}

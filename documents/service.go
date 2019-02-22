package documents

import (
	"bytes"
	"context"
	"time"

	"github.com/centrifuge/go-centrifuge/transactions"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/notification"
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/crypto"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/notification"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/centrifuge/precise-proofs/proofs/proto"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/golang/protobuf/ptypes"
	logging "github.com/ipfs/go-log"
)

// DocumentProof is a value to represent a document and its field proofs
type DocumentProof struct {
	DocumentID  []byte
	VersionID   []byte
	State       string
	FieldProofs []*proofspb.Proof
}

// Service provides an interface for functions common to all document types
type Service interface {

	// GetCurrentVersion reads a document from the database
	GetCurrentVersion(ctx context.Context, documentID []byte) (Model, error)

	// Exists checks if a document exists
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
	RequestDocumentSignature(ctx context.Context, model Model) (*coredocumentpb.Signature, error)

	// ReceiveAnchoredDocument receives a new anchored document over the p2p layer, validates and updates the document in DB
	ReceiveAnchoredDocument(ctx context.Context, model Model, senderID []byte) error

	// Create validates and persists Model and returns a Updated model
	Create(ctx context.Context, model Model) (Model, transactions.TxID, chan bool, error)

	// Update validates and updates the model and return the updated model
	Update(ctx context.Context, model Model) (Model, transactions.TxID, chan bool, error)
}

// service implements Service
type service struct {
	repo             Repository
	identityService  identity.Service
	notifier         notification.Sender
	anchorRepository anchors.AnchorRepository
	registry         *ServiceRegistry
}

var srvLog = logging.Logger("document-service")

// DefaultService returns the default implementation of the service
func DefaultService(
	repo Repository,
	idService identity.Service,
	anchorRepo anchors.AnchorRepository,
	registry *ServiceRegistry) Service {
	return service{
		repo:             repo,
		anchorRepository: anchorRepo,
		notifier:         notification.NewWebhookSender(),
		identityService:  idService,
		registry:         registry,
	}
}

func (s service) searchVersion(ctx context.Context, m Model) (Model, error) {
	id, next := m.ID(), m.NextVersion()
	if !s.Exists(ctx, next) {
		// at the latest version
		return m, nil
	}

	m, err := s.getVersion(ctx, id, next)
	if err != nil {
		return nil, err
	}
	return s.searchVersion(ctx, m)

}

func (s service) GetCurrentVersion(ctx context.Context, documentID []byte) (Model, error) {
	model, err := s.getVersion(ctx, documentID, documentID)
	if err != nil {
		return nil, errors.NewTypedError(ErrDocumentNotFound, err)
	}
	return s.searchVersion(ctx, model)
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
	if err := PostAnchoredValidator(s.identityService, s.anchorRepository).Validate(nil, model); err != nil {
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

func (s service) RequestDocumentSignature(ctx context.Context, model Model) (*coredocumentpb.Signature, error) {
	idConf, err := contextutil.Self(ctx)
	if err != nil {
		return nil, ErrDocumentConfigAccountID
	}

	if err := SignatureRequestValidator(s.identityService).Validate(nil, model); err != nil {
		return nil, errors.NewTypedError(ErrDocumentInvalid, err)
	}

	sr, err := model.SigningRoot()
	if err != nil {
		return nil, errors.New("failed to get signing root: %v", err)
	}

	srvLog.Infof("document received %x with signing root %x", model.ID(), sr)
	idKeys, ok := idConf.Keys[identity.KeyPurposeSigning]
	if !ok {
		return nil, errors.NewTypedError(ErrDocumentSigning, errors.New("missing signing key"))
	}

	sig := crypto.Sign(idConf.ID[:], idKeys.PrivateKey, idKeys.PublicKey, sr)
	model.AppendSignatures(sig)

	tenantID := idConf.ID[:]
	// Logic for receiving version n (n > 1) of the document for the first time
	// TODO(ved): we should not save the new model with old identifier. We should sync from the peer.
	if !s.repo.Exists(tenantID, model.ID()) && !utils.IsSameByteSlice(model.ID(), model.CurrentVersion()) {
		err = s.repo.Create(tenantID, model.ID(), model)
		if err != nil {
			return nil, errors.NewTypedError(ErrDocumentPersistence, err)
		}
	}

	err = s.repo.Create(tenantID, model.CurrentVersion(), model)
	if err != nil {
		return nil, errors.NewTypedError(ErrDocumentPersistence, err)
	}

	srvLog.Infof("signed document %x with version %x", model.ID(), model.CurrentVersion())
	return sig, nil
}

func (s service) ReceiveAnchoredDocument(ctx context.Context, model Model, senderID []byte) error {
	idConf, err := contextutil.Self(ctx)
	if err != nil {
		return ErrDocumentConfigAccountID
	}

	if err := PostAnchoredValidator(s.identityService, s.anchorRepository).Validate(nil, model); err != nil {
		return errors.NewTypedError(ErrDocumentInvalid, err)
	}

	err = s.repo.Update(idConf.ID[:], model.CurrentVersion(), model)
	if err != nil {
		return errors.NewTypedError(ErrDocumentPersistence, err)
	}

	ts, _ := ptypes.TimestampProto(time.Now().UTC())
	notificationMsg := &notificationpb.NotificationMessage{
		EventType:    uint32(notification.ReceivedPayload),
		AccountId:    idConf.ID.String(),
		FromId:       hexutil.Encode(senderID),
		ToId:         idConf.ID.String(),
		Recorded:     ts,
		DocumentType: model.DocumentType(),
		DocumentId:   hexutil.Encode(model.ID()),
	}

	// Async until we add queuing
	go s.notifier.Send(ctx, notificationMsg)

	return nil
}

func (s service) Exists(ctx context.Context, documentID []byte) bool {
	idConf, err := contextutil.Self(ctx)
	if err != nil {
		return false
	}
	return s.repo.Exists(idConf.ID[:], documentID)
}

func (s service) getVersion(ctx context.Context, documentID, version []byte) (Model, error) {
	idConf, err := contextutil.Self(ctx)
	if err != nil {
		return nil, ErrDocumentConfigAccountID
	}
	model, err := s.repo.Get(idConf.ID[:], version)
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

func (s service) Create(ctx context.Context, model Model) (Model, transactions.TxID, chan bool, error) {
	srv, err := s.getService(model)
	if err != nil {
		return nil, transactions.NilTxID(), nil, errors.New("failed to get service: %v", err)
	}

	return srv.Create(ctx, model)
}

func (s service) Update(ctx context.Context, model Model) (Model, transactions.TxID, chan bool, error) {
	srv, err := s.getService(model)
	if err != nil {
		return nil, transactions.NilTxID(), nil, errors.New("failed to get service: %v", err)
	}

	return srv.Update(ctx, model)
}

func (s service) getService(model Model) (Service, error) {
	return s.registry.LocateService(model.DocumentType())
}

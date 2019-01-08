package invoice

import (
	"bytes"
	"context"
	"time"

	"github.com/centrifuge/go-centrifuge/documents/genericdoc"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/notification"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/contextutil"
	"github.com/centrifuge/go-centrifuge/coredocument"
	"github.com/centrifuge/go-centrifuge/crypto"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/errors"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/notification"
	clientinvoicepb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/invoice"
	"github.com/centrifuge/go-centrifuge/queue"
	"github.com/centrifuge/go-centrifuge/transactions"
	"github.com/centrifuge/go-centrifuge/utils"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/golang/protobuf/ptypes"
	logging "github.com/ipfs/go-log"
	"github.com/satori/go.uuid"
)

var srvLog = logging.Logger("invoice-service")

// Service defines specific functions for invoice
type Service interface {
	documents.Service

	// DeriverFromPayload derives Invoice from clientPayload
	DeriveFromCreatePayload(ctx context.Context, payload *clientinvoicepb.InvoiceCreatePayload) (documents.Model, error)

	// DeriveFromUpdatePayload derives invoice model from update payload
	DeriveFromUpdatePayload(ctx context.Context, payload *clientinvoicepb.InvoiceUpdatePayload) (documents.Model, error)

	// Create validates and persists invoice Model and returns a Updated model
	Create(ctx context.Context, inv documents.Model) (documents.Model, uuid.UUID, error)

	// Update validates and updates the invoice model and return the updated model
	Update(ctx context.Context, inv documents.Model) (documents.Model, uuid.UUID, error)

	// DeriveInvoiceData returns the invoice data as client data
	DeriveInvoiceData(inv documents.Model) (*clientinvoicepb.InvoiceData, error)

	// DeriveInvoiceResponse returns the invoice model in our standard client format
	DeriveInvoiceResponse(inv documents.Model) (*clientinvoicepb.InvoiceResponse, error)
}

// service implements Service and handles all invoice related persistence and validations
// service always returns errors of type `errors.Error` or `errors.TypedError`
type service struct {
	repo             documents.Repository
	notifier         notification.Sender
	anchorRepository anchors.AnchorRepository
	identityService  identity.Service
	queueSrv         queue.TaskQueuer
	txService        transactions.Service
	genService       genericdoc.Service
}

// DefaultService returns the default implementation of the service.
func DefaultService(
	repo documents.Repository,
	anchorRepository anchors.AnchorRepository,
	identityService identity.Service,
	queueSrv queue.TaskQueuer,
	txService transactions.Service,
	genService genericdoc.Service,
) Service {
	return service{
		repo:             repo,
		notifier:         notification.NewWebhookSender(),
		anchorRepository: anchorRepository,
		identityService:  identityService,
		queueSrv:         queueSrv,
		txService:        txService,
		genService:       genService}
}

// CreateProofs creates proofs for the latest version document given the fields
func (s service) CreateProofs(ctx context.Context, documentID []byte, fields []string) (*documents.DocumentProof, error) {
	model, err := s.GetCurrentVersion(ctx, documentID)
	if err != nil {
		return nil, err
	}

	return s.invoiceProof(model, fields)
}

// CreateProofsForVersion creates proofs for a particular version of the document given the fields
func (s service) CreateProofsForVersion(ctx context.Context, documentID, version []byte, fields []string) (*documents.DocumentProof, error) {
	model, err := s.GetVersion(ctx, documentID, version)
	if err != nil {
		return nil, err
	}

	return s.invoiceProof(model, fields)
}

// invoiceProof creates proofs for invoice model fields
func (s service) invoiceProof(model documents.Model, fields []string) (*documents.DocumentProof, error) {
	inv, ok := model.(*Invoice)
	if !ok {
		return nil, documents.ErrDocumentInvalidType
	}

	if err := coredocument.PostAnchoredValidator(s.identityService, s.anchorRepository).Validate(nil, inv); err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentInvalid, err)
	}
	coreDoc, proofs, err := inv.CreateProofs(fields)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentProof, err)
	}
	return &documents.DocumentProof{
		DocumentID:  coreDoc.DocumentIdentifier,
		VersionID:   coreDoc.CurrentVersion,
		FieldProofs: proofs,
	}, nil
}

// DeriveFromCoreDocument unpacks the core document into a model
func (s service) DeriveFromCoreDocument(cd *coredocumentpb.CoreDocument) (documents.Model, error) {
	var model documents.Model = new(Invoice)
	err := model.UnpackCoreDocument(cd)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentUnPackingCoreDocument, err)
	}

	return model, nil
}

// UnpackFromCreatePayload initializes the model with parameters provided from the rest-api call
func (s service) DeriveFromCreatePayload(ctx context.Context, payload *clientinvoicepb.InvoiceCreatePayload) (documents.Model, error) {
	if payload == nil || payload.Data == nil {
		return nil, documents.ErrDocumentNil
	}

	id, err := contextutil.Self(ctx)
	if err != nil {
		return nil, documents.ErrDocumentConfigTenantID
	}

	invoiceModel := new(Invoice)
	err = invoiceModel.InitInvoiceInput(payload, id.ID.String())
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentInvalid, err)
	}

	return invoiceModel, nil
}

// calculateDataRoot validates the document, calculates the data root, and persists to DB
func (s service) calculateDataRoot(ctx context.Context, old, new documents.Model, validator documents.Validator) (documents.Model, error) {
	self, err := contextutil.Self(ctx)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentConfigTenantID, err)
	}

	inv, ok := new.(*Invoice)
	if !ok {
		return nil, errors.NewTypedError(documents.ErrDocumentInvalidType, errors.New("unknown document type: %T", new))
	}

	// create data root, has to be done at the model level to access fields
	err = inv.CalculateDataRoot()
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentInvalid, err)
	}

	// validate the invoice
	err = validator.Validate(old, inv)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentInvalid, err)
	}

	// we use CurrentVersion as the id since that will be unique across multiple versions of the same document
	err = s.repo.Create(self.ID[:], inv.CoreDocument.CurrentVersion, inv)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentPersistence, err)
	}

	return inv, nil
}

// Create takes and invoice model and does required validation checks, tries to persist to DB
func (s service) Create(ctx context.Context, inv documents.Model) (documents.Model, uuid.UUID, error) {
	self, err := contextutil.Self(ctx)
	if err != nil {
		return nil, uuid.Nil, errors.NewTypedError(documents.ErrDocumentConfigTenantID, err)
	}

	inv, err = s.calculateDataRoot(ctx, nil, inv, CreateValidator())
	if err != nil {
		return nil, uuid.Nil, err
	}

	cd, err := inv.PackCoreDocument()
	if err != nil {
		return nil, uuid.Nil, err
	}

	txID, err := documents.InitDocumentAnchorTask(
		s.queueSrv,
		s.txService,
		self.ID,
		cd.CurrentVersion)
	if err != nil {
		return nil, uuid.Nil, err
	}

	return inv, txID, nil
}

// Update finds the old document, validates the new version and persists the updated document
func (s service) Update(ctx context.Context, inv documents.Model) (documents.Model, uuid.UUID, error) {
	self, err := contextutil.Self(ctx)
	if err != nil {
		return nil, uuid.Nil, errors.NewTypedError(documents.ErrDocumentConfigTenantID, err)
	}

	cd, err := inv.PackCoreDocument()
	if err != nil {
		return nil, uuid.Nil, errors.NewTypedError(documents.ErrDocumentPackingCoreDocument, err)
	}

	old, err := s.GetCurrentVersion(ctx, cd.DocumentIdentifier)
	if err != nil {
		return nil, uuid.Nil, errors.NewTypedError(documents.ErrDocumentNotFound, err)
	}

	inv, err = s.calculateDataRoot(ctx, old, inv, UpdateValidator())
	if err != nil {
		return nil, uuid.Nil, err
	}

	txID, err := documents.InitDocumentAnchorTask(
		s.queueSrv,
		s.txService,
		self.ID,
		cd.CurrentVersion)
	if err != nil {
		return nil, uuid.Nil, err
	}

	return inv, txID, nil
}

// GetVersion returns an invoice for a given version
func (s service) GetVersion(ctx context.Context, documentID []byte, version []byte) (documents.Model, error) {
	inv, err := s.getInvoiceVersion(ctx, documentID, version)
	if err != nil {
		return nil, err
	}
	return inv, nil
}

// GetCurrentVersion returns the last known version of an invoice
func (s service) GetCurrentVersion(ctx context.Context, documentID []byte) (documents.Model, error) {
	inv, err := s.getInvoiceVersion(ctx, documentID, documentID)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentNotFound, err)
	}
	nextVersion := inv.CoreDocument.NextVersion
	for nextVersion != nil {
		temp, err := s.getInvoiceVersion(ctx, documentID, nextVersion)
		if err != nil {
			// here the err is returned as nil because it is expected that the nextVersion is not available in the db at some stage of the iteration
			return inv, nil
		}

		inv = temp
		nextVersion = inv.CoreDocument.NextVersion
	}
	return inv, nil
}

func (s service) getInvoiceVersion(ctx context.Context, documentID, version []byte) (inv *Invoice, err error) {
	self, err := contextutil.Self(ctx)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentConfigTenantID, err)
	}

	doc, err := s.repo.Get(self.ID[:], version)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentVersionNotFound, err)
	}
	inv, ok := doc.(*Invoice)
	if !ok {
		return nil, documents.ErrDocumentInvalidType
	}

	if !bytes.Equal(inv.CoreDocument.DocumentIdentifier, documentID) {
		return nil, errors.NewTypedError(documents.ErrDocumentVersionNotFound, errors.New("version is not valid for this identifier"))
	}
	return inv, nil
}

// DeriveInvoiceResponse returns create response from invoice model
func (s service) DeriveInvoiceResponse(doc documents.Model) (*clientinvoicepb.InvoiceResponse, error) {
	cd, err := doc.PackCoreDocument()
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentPackingCoreDocument, err)
	}

	collaborators := make([]string, len(cd.Collaborators))
	for i, c := range cd.Collaborators {
		cid, err := identity.ToCentID(c)
		if err != nil {
			return nil, errors.NewTypedError(documents.ErrDocumentCollaborator, err)
		}
		collaborators[i] = cid.String()
	}

	h := &clientinvoicepb.ResponseHeader{
		DocumentId:    hexutil.Encode(cd.DocumentIdentifier),
		VersionId:     hexutil.Encode(cd.CurrentVersion),
		Collaborators: collaborators,
	}

	data, err := s.DeriveInvoiceData(doc)
	if err != nil {
		return nil, err
	}

	return &clientinvoicepb.InvoiceResponse{
		Header: h,
		Data:   data,
	}, nil

}

// DeriveInvoiceData returns create response from invoice model
func (s service) DeriveInvoiceData(doc documents.Model) (*clientinvoicepb.InvoiceData, error) {
	inv, ok := doc.(*Invoice)
	if !ok {
		return nil, documents.ErrDocumentInvalidType
	}

	return inv.getClientData(), nil
}

// DeriveFromUpdatePayload returns a new version of the old invoice identified by identifier in payload
func (s service) DeriveFromUpdatePayload(ctx context.Context, payload *clientinvoicepb.InvoiceUpdatePayload) (documents.Model, error) {
	if payload == nil || payload.Data == nil {
		return nil, documents.ErrDocumentNil
	}

	// get latest old version of the document
	id, err := hexutil.Decode(payload.Identifier)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentIdentifier, errors.New("failed to decode identifier: %v", err))
	}

	old, err := s.GetCurrentVersion(ctx, id)
	if err != nil {
		return nil, err
	}

	// load invoice data
	inv := new(Invoice)
	err = inv.initInvoiceFromData(payload.Data)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentInvalid, errors.New("failed to load invoice from data: %v", err))
	}

	// update core document
	oldCD, err := old.PackCoreDocument()
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentPackingCoreDocument, err)
	}

	idConf, err := contextutil.Self(ctx)
	if err != nil {
		return nil, documents.ErrDocumentConfigTenantID
	}

	collaborators := append([]string{idConf.ID.String()}, payload.Collaborators...)
	inv.CoreDocument, err = coredocument.PrepareNewVersion(*oldCD, collaborators)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentPrepareCoreDocument, err)
	}

	return inv, nil
}

// RequestDocumentSignature Validates, Signs document received over the p2p layer and returns Signature
func (s service) RequestDocumentSignature(ctx context.Context, model documents.Model) (*coredocumentpb.Signature, error) {
	self, err := contextutil.Self(ctx)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentConfigTenantID, err)
	}

	if err := coredocument.SignatureRequestValidator(s.identityService).Validate(nil, model); err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentInvalid, err)
	}

	doc, err := model.PackCoreDocument()
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentPackingCoreDocument, err)
	}

	srvLog.Infof("coredoc received %x with signing root %x", doc.DocumentIdentifier, doc.SigningRoot)

	idConf, err := contextutil.Self(ctx)
	if err != nil {
		return nil, documents.ErrDocumentConfigTenantID
	}

	idKeys, ok := idConf.Keys[identity.KeyPurposeSigning]
	if !ok {
		return nil, errors.NewTypedError(documents.ErrDocumentSigning, errors.New("missing signing key"))
	}
	sig := crypto.Sign(idConf.ID[:], idKeys.PrivateKey, idKeys.PublicKey, doc.SigningRoot)
	doc.Signatures = append(doc.Signatures, sig)
	err = model.UnpackCoreDocument(doc)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentUnPackingCoreDocument, err)
	}

	// Logic for receiving version n (n > 1) of the document for the first time
	if !s.repo.Exists(self.ID[:], doc.DocumentIdentifier) && !utils.IsSameByteSlice(doc.DocumentIdentifier, doc.CurrentVersion) {
		err = s.repo.Create(self.ID[:], doc.DocumentIdentifier, model)
		if err != nil {
			return nil, errors.NewTypedError(documents.ErrDocumentPersistence, err)
		}
	}

	err = s.repo.Create(self.ID[:], doc.CurrentVersion, model)
	if err != nil {
		return nil, errors.NewTypedError(documents.ErrDocumentPersistence, err)
	}

	srvLog.Infof("signed coredoc %x with version %x", doc.DocumentIdentifier, doc.CurrentVersion)
	return sig, nil
}

// ReceiveAnchoredDocument receives a new anchored document, validates and updates the document in DB
func (s service) ReceiveAnchoredDocument(ctx context.Context, model documents.Model, headers *p2ppb.CentrifugeHeader) error {
	self, err := contextutil.Self(ctx)
	if err != nil {
		return errors.NewTypedError(documents.ErrDocumentConfigTenantID, err)
	}

	if err := coredocument.PostAnchoredValidator(s.identityService, s.anchorRepository).Validate(nil, model); err != nil {
		return errors.NewTypedError(documents.ErrDocumentInvalid, err)
	}

	doc, err := model.PackCoreDocument()
	if err != nil {
		return errors.NewTypedError(documents.ErrDocumentPackingCoreDocument, err)
	}

	err = s.repo.Update(self.ID[:], doc.CurrentVersion, model)
	if err != nil {
		return errors.NewTypedError(documents.ErrDocumentPersistence, err)
	}

	ts, _ := ptypes.TimestampProto(time.Now().UTC())
	notificationMsg := &notificationpb.NotificationMessage{
		EventType:    uint32(notification.ReceivedPayload),
		CentrifugeId: hexutil.Encode(headers.SenderCentrifugeId),
		Recorded:     ts,
		DocumentType: doc.EmbeddedData.TypeUrl,
		DocumentId:   hexutil.Encode(doc.DocumentIdentifier),
	}

	// Async until we add queuing
	go s.notifier.Send(ctx, notificationMsg)

	return nil
}

// Exists checks if an invoice exists
func (s service) Exists(ctx context.Context, documentID []byte) bool {
	self, err := contextutil.Self(ctx)
	if err != nil {
		return false
	}
	return s.repo.Exists(self.ID[:], documentID)
}

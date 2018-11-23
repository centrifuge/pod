package invoice

import (
	"bytes"
	"fmt"
	"time"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/notification"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	"github.com/centrifuge/go-centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/centerrors"
	"github.com/centrifuge/go-centrifuge/code"
	"github.com/centrifuge/go-centrifuge/config"
	"github.com/centrifuge/go-centrifuge/coredocument"
	"github.com/centrifuge/go-centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/header"
	"github.com/centrifuge/go-centrifuge/identity"
	"github.com/centrifuge/go-centrifuge/notification"
	clientinvoicepb "github.com/centrifuge/go-centrifuge/protobufs/gen/go/invoice"
	"github.com/centrifuge/go-centrifuge/signatures"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/golang/protobuf/ptypes"
	logging "github.com/ipfs/go-log"
)

var srvLog = logging.Logger("invoice-service")

// Service defines specific functions for invoice
type Service interface {
	documents.Service

	// DeriverFromPayload derives Invoice from clientPayload
	DeriveFromCreatePayload(*clientinvoicepb.InvoiceCreatePayload, *header.ContextHeader) (documents.Model, error)

	// DeriveFromUpdatePayload derives invoice model from update payload
	DeriveFromUpdatePayload(*clientinvoicepb.InvoiceUpdatePayload, *header.ContextHeader) (documents.Model, error)

	// Create validates and persists invoice Model and returns a Updated model
	Create(ctx *header.ContextHeader, inv documents.Model) (documents.Model, error)

	// Update validates and updates the invoice model and return the updated model
	Update(ctx *header.ContextHeader, inv documents.Model) (documents.Model, error)

	// DeriveInvoiceData returns the invoice data as client data
	DeriveInvoiceData(inv documents.Model) (*clientinvoicepb.InvoiceData, error)

	// DeriveInvoiceResponse returns the invoice model in our standard client format
	DeriveInvoiceResponse(inv documents.Model) (*clientinvoicepb.InvoiceResponse, error)
}

// service implements Service and handles all invoice related persistence and validations
// service always returns errors of type `centerrors` with proper error code
type service struct {
	repo             documents.Repository
	coreDocProcessor coredocument.Processor
	notifier         notification.Sender
	anchorRepository anchors.AnchorRepository
	identityService  identity.Service
}

// DefaultService returns the default implementation of the service
func DefaultService(config config.Config, repo documents.Repository, processor coredocument.Processor, anchorRepository anchors.AnchorRepository, identityService identity.Service) Service {
	return service{repo: repo, coreDocProcessor: processor, notifier: notification.NewWebhookSender(config.(notification.Config)), anchorRepository: anchorRepository, identityService: identityService}
}

// CreateProofs creates proofs for the latest version document given the fields
func (s service) CreateProofs(documentID []byte, fields []string) (*documents.DocumentProof, error) {
	model, err := s.GetCurrentVersion(documentID)
	if err != nil {
		return nil, err
	}

	return s.invoiceProof(model, fields)
}

// CreateProofsForVersion creates proofs for a particular version of the document given the fields
func (s service) CreateProofsForVersion(documentID, version []byte, fields []string) (*documents.DocumentProof, error) {
	model, err := s.GetVersion(documentID, version)
	if err != nil {
		return nil, err
	}

	return s.invoiceProof(model, fields)
}

// invoiceProof creates proofs for invoice model fields
func (s service) invoiceProof(model documents.Model, fields []string) (*documents.DocumentProof, error) {
	inv, ok := model.(*Invoice)
	if !ok {
		return nil, centerrors.New(code.DocumentInvalid, "document of invalid type")
	}

	if err := coredocument.PostAnchoredValidator(s.identityService, s.anchorRepository).Validate(nil, inv); err != nil {
		return nil, centerrors.New(code.DocumentInvalid, err.Error())
	}
	coreDoc, proofs, err := inv.createProofs(fields)
	if err != nil {
		return nil, err
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
		return nil, centerrors.New(code.DocumentInvalid, err.Error())
	}

	return model, nil
}

// UnpackFromCreatePayload initializes the model with parameters provided from the rest-api call
func (s service) DeriveFromCreatePayload(payload *clientinvoicepb.InvoiceCreatePayload, contextHeader *header.ContextHeader) (documents.Model, error) {
	if payload == nil || payload.Data == nil {
		return nil, centerrors.New(code.DocumentInvalid, "input is nil")
	}

	invoiceModel := new(Invoice)
	err := invoiceModel.InitInvoiceInput(payload, contextHeader)
	if err != nil {
		return nil, centerrors.New(code.DocumentInvalid, err.Error())
	}

	return invoiceModel, nil
}

// calculateDataRoot validates the document, calculates the data root, and persists to DB
func (s service) calculateDataRoot(old, new documents.Model, validator documents.Validator) (documents.Model, error) {
	inv, ok := new.(*Invoice)
	if !ok {
		return nil, centerrors.New(code.DocumentInvalid, fmt.Sprintf("unknown document type: %T", new))
	}

	// create data root, has to be done at the model level to access fields
	err := inv.calculateDataRoot()
	if err != nil {
		return nil, centerrors.New(code.DocumentInvalid, err.Error())
	}

	// validate the invoice
	err = validator.Validate(old, inv)
	if err != nil {
		return nil, centerrors.NewWithErrors(code.DocumentInvalid, "validations failed", documents.ConvertToMap(err))
	}

	// we use CurrentVersion as the id since that will be unique across multiple versions of the same document
	err = s.repo.Create(inv.CoreDocument.CurrentVersion, inv)
	if err != nil {
		return nil, centerrors.New(code.Unknown, err.Error())
	}

	return inv, nil
}

// Create takes and invoice model and does required validation checks, tries to persist to DB
func (s service) Create(ctx *header.ContextHeader, inv documents.Model) (documents.Model, error) {
	inv, err := s.calculateDataRoot(nil, inv, CreateValidator())
	if err != nil {
		return nil, err
	}

	inv, err = documents.AnchorDocument(ctx, inv, s.coreDocProcessor, s.repo.Update)
	if err != nil {
		return nil, centerrors.New(code.Unknown, err.Error())
	}

	return inv, nil
}

// Update finds the old document, validates the new version and persists the updated document
func (s service) Update(ctx *header.ContextHeader, inv documents.Model) (documents.Model, error) {
	cd, err := inv.PackCoreDocument()
	if err != nil {
		return nil, centerrors.New(code.Unknown, err.Error())
	}

	old, err := s.GetCurrentVersion(cd.DocumentIdentifier)
	if err != nil {
		return nil, centerrors.New(code.DocumentNotFound, err.Error())
	}

	inv, err = s.calculateDataRoot(old, inv, UpdateValidator())
	if err != nil {
		return nil, err
	}

	inv, err = documents.AnchorDocument(ctx, inv, s.coreDocProcessor, s.repo.Update)
	if err != nil {
		return nil, centerrors.New(code.Unknown, err.Error())
	}

	return inv, nil
}

// GetVersion returns an invoice for a given version
func (s service) GetVersion(documentID []byte, version []byte) (doc documents.Model, err error) {
	inv, err := s.getInvoiceVersion(documentID, version)
	if err != nil {
		return nil, centerrors.Wrap(err, "document not found for the given version")
	}
	return inv, nil
}

// GetCurrentVersion returns the last known version of an invoice
func (s service) GetCurrentVersion(documentID []byte) (doc documents.Model, err error) {
	inv, err := s.getInvoiceVersion(documentID, documentID)
	if err != nil {
		return nil, centerrors.Wrap(err, "document not found")
	}
	nextVersion := inv.CoreDocument.NextVersion
	for nextVersion != nil {
		temp, err := s.getInvoiceVersion(documentID, nextVersion)
		if err != nil {
			return inv, nil
		}

		inv = temp
		nextVersion = inv.CoreDocument.NextVersion
	}
	return inv, nil
}

func (s service) getInvoiceVersion(documentID, version []byte) (inv *Invoice, err error) {
	var doc documents.Model = new(Invoice)
	err = s.repo.LoadByID(version, doc)
	if err != nil {
		return nil, err
	}
	inv, ok := doc.(*Invoice)
	if !ok {
		return nil, err
	}

	if !bytes.Equal(inv.CoreDocument.DocumentIdentifier, documentID) {
		return nil, centerrors.New(code.DocumentInvalid, "version is not valid for this identifier")
	}
	return inv, nil
}

// DeriveInvoiceResponse returns create response from invoice model
func (s service) DeriveInvoiceResponse(doc documents.Model) (*clientinvoicepb.InvoiceResponse, error) {
	cd, err := doc.PackCoreDocument()
	if err != nil {
		return nil, centerrors.New(code.DocumentInvalid, err.Error())
	}

	collaborators := make([]string, len(cd.Collaborators))
	for i, c := range cd.Collaborators {
		cid, err := identity.ToCentID(c)
		if err != nil {
			return nil, centerrors.New(code.Unknown, err.Error())
		}
		collaborators[i] = cid.String()
	}

	header := &clientinvoicepb.ResponseHeader{
		DocumentId:    hexutil.Encode(cd.DocumentIdentifier),
		VersionId:     hexutil.Encode(cd.CurrentVersion),
		Collaborators: collaborators,
	}

	data, err := s.DeriveInvoiceData(doc)
	if err != nil {
		return nil, err
	}

	return &clientinvoicepb.InvoiceResponse{
		Header: header,
		Data:   data,
	}, nil

}

// DeriveInvoiceData returns create response from invoice model
func (s service) DeriveInvoiceData(doc documents.Model) (*clientinvoicepb.InvoiceData, error) {
	inv, ok := doc.(*Invoice)
	if !ok {
		return nil, centerrors.New(code.DocumentInvalid, "document of invalid type")
	}

	return inv.getClientData(), nil
}

// DeriveFromUpdatePayload returns a new version of the old invoice identified by identifier in payload
func (s service) DeriveFromUpdatePayload(payload *clientinvoicepb.InvoiceUpdatePayload, contextHeader *header.ContextHeader) (documents.Model, error) {
	if payload == nil || payload.Data == nil {
		return nil, centerrors.New(code.DocumentInvalid, "invalid payload")
	}

	// get latest old version of the document
	id, err := hexutil.Decode(payload.Identifier)
	if err != nil {
		return nil, centerrors.New(code.DocumentInvalid, fmt.Sprintf("failed to decode identifier: %v", err))
	}

	old, err := s.GetCurrentVersion(id)
	if err != nil {
		return nil, centerrors.New(code.DocumentInvalid, fmt.Sprintf("failed to fetch old version: %v", err))
	}

	// load invoice data
	inv := new(Invoice)
	err = inv.initInvoiceFromData(payload.Data)
	if err != nil {
		return nil, centerrors.New(code.DocumentInvalid, fmt.Sprintf("failed to load invoice from data: %v", err))
	}

	// update core document
	oldCD, err := old.PackCoreDocument()
	if err != nil {
		return nil, centerrors.New(code.Unknown, err.Error())
	}

	collaborators := append([]string{contextHeader.Self().ID.String()}, payload.Collaborators...)
	inv.CoreDocument, err = coredocument.PrepareNewVersion(*oldCD, collaborators)
	if err != nil {
		return nil, centerrors.New(code.DocumentInvalid, fmt.Sprintf("failed to prepare new version: %v", err))
	}

	return inv, nil
}

// RequestDocumentSignature Validates, Signs document received over the p2p layer and returns Signature
func (s service) RequestDocumentSignature(contextHeader *header.ContextHeader, model documents.Model) (*coredocumentpb.Signature, error) {
	if err := coredocument.SignatureRequestValidator(s.identityService).Validate(nil, model); err != nil {
		return nil, centerrors.New(code.DocumentInvalid, err.Error())
	}

	doc, err := model.PackCoreDocument()
	if err != nil {
		return nil, centerrors.New(code.DocumentInvalid, err.Error())
	}

	srvLog.Infof("coredoc received %x with signing root %x", doc.DocumentIdentifier, doc.SigningRoot)

	idKeys, ok := contextHeader.Self().Keys[identity.KeyPurposeSigning]
	if !ok {
		return nil, centerrors.New(code.Unknown, fmt.Sprintf("missing signing key"))
	}
	sig := signatures.Sign(contextHeader.Self().ID[:], idKeys.PrivateKey, idKeys.PublicKey, doc.SigningRoot)
	doc.Signatures = append(doc.Signatures, sig)
	err = model.UnpackCoreDocument(doc)
	if err != nil {
		return nil, centerrors.New(code.Unknown, fmt.Sprintf("failed to Unpack CoreDocument: %v", err))
	}

	err = s.repo.Create(doc.CurrentVersion, model)
	if err != nil {
		return nil, centerrors.New(code.Unknown, fmt.Sprintf("failed to store document: %v", err))
	}

	srvLog.Infof("signed coredoc %x with version %x", doc.DocumentIdentifier, doc.CurrentVersion)
	return sig, nil
}

// ReceiveAnchoredDocument receives a new anchored document, validates and updates the document in DB
func (s service) ReceiveAnchoredDocument(model documents.Model, headers *p2ppb.CentrifugeHeader) error {
	if err := coredocument.PostAnchoredValidator(s.identityService, s.anchorRepository).Validate(nil, model); err != nil {
		return centerrors.New(code.DocumentInvalid, err.Error())
	}

	doc, err := model.PackCoreDocument()
	if err != nil {
		return centerrors.New(code.DocumentInvalid, err.Error())
	}

	err = s.repo.Update(doc.CurrentVersion, model)
	if err != nil {
		return centerrors.New(code.Unknown, err.Error())
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
	go s.notifier.Send(notificationMsg)

	return nil
}

// Exists checks if an invoice exists
func (s service) Exists(documentID []byte) bool {
	return s.repo.Exists(documentID)
}

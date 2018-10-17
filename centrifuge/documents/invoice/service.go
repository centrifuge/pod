package invoice

import (
	"bytes"
	"context"
	"fmt"

	"time"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/notification"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	"github.com/centrifuge/go-centrifuge/centrifuge/centerrors"
	"github.com/centrifuge/go-centrifuge/centrifuge/code"
	"github.com/centrifuge/go-centrifuge/centrifuge/coredocument"
	"github.com/centrifuge/go-centrifuge/centrifuge/coredocument/processor"
	"github.com/centrifuge/go-centrifuge/centrifuge/coredocument/repository"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents/common"
	"github.com/centrifuge/go-centrifuge/centrifuge/identity"
	centED25519 "github.com/centrifuge/go-centrifuge/centrifuge/keytools/ed25519keys"
	"github.com/centrifuge/go-centrifuge/centrifuge/notification"
	clientinvoicepb "github.com/centrifuge/go-centrifuge/centrifuge/protobufs/gen/go/invoice"
	"github.com/centrifuge/go-centrifuge/centrifuge/signatures"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/golang/protobuf/ptypes"
)

// Service defines specific functions for invoice
type Service interface {
	documents.Service

	// DeriverFromPayload derives InvoiceModel from clientPayload
	DeriveFromCreatePayload(*clientinvoicepb.InvoiceCreatePayload) (documents.Model, error)

	// DeriveFromUpdatePayload derives invoice model from update payload
	DeriveFromUpdatePayload(*clientinvoicepb.InvoiceUpdatePayload) (documents.Model, error)

	// Create validates and persists invoice Model and returns a Updated model
	Create(ctx context.Context, inv documents.Model) (documents.Model, error)

	// Update validates and updates the invoice model and return the updated model
	Update(ctx context.Context, inv documents.Model) (documents.Model, error)

	// DeriveInvoiceData returns the invoice data as client data
	DeriveInvoiceData(inv documents.Model) (*clientinvoicepb.InvoiceData, error)

	// DeriveInvoiceResponse returns the invoice model in our standard client format
	DeriveInvoiceResponse(inv documents.Model) (*clientinvoicepb.InvoiceResponse, error)

	// SaveState updates the model in DB
	SaveState(inv documents.Model) error
}

// service implements Service and handles all invoice related persistence and validations
// service always returns errors of type `centerrors` with proper error code
type service struct {
	repo             documents.Repository
	coreDocProcessor coredocumentprocessor.Processor
	notifier         notification.Sender
}

// DefaultService returns the default implementation of the service
func DefaultService(repo documents.Repository, processor coredocumentprocessor.Processor) Service {
	return &service{repo: repo, coreDocProcessor: processor, notifier: &notification.WebhookSender{}}
}

// CreateProofs creates proofs for the latest version document given the fields
func (s service) CreateProofs(documentID []byte, fields []string) (common.DocumentProof, error) {
	doc, err := s.GetCurrentVersion(documentID)
	if err != nil {
		return common.DocumentProof{}, err
	}
	inv, ok := doc.(*InvoiceModel)
	if !ok {
		return common.DocumentProof{}, centerrors.New(code.DocumentInvalid, "document of invalid type")
	}
	return s.invoiceProof(inv, fields)
}

// CreateProofsForVersion creates proofs for a particular version of the document given the fields
func (s service) CreateProofsForVersion(documentID, version []byte, fields []string) (common.DocumentProof, error) {
	doc, err := s.GetVersion(documentID, version)
	if err != nil {
		return common.DocumentProof{}, err
	}
	inv, ok := doc.(*InvoiceModel)
	if !ok {
		return common.DocumentProof{}, centerrors.New(code.DocumentInvalid, "document of invalid type")
	}
	return s.invoiceProof(inv, fields)
}

// invoiceProof creates proofs for invoice model fields
func (s service) invoiceProof(inv *InvoiceModel, fields []string) (common.DocumentProof, error) {
	coreDoc, proofs, err := inv.createProofs(fields)
	if err != nil {
		return common.DocumentProof{}, err
	}
	return common.DocumentProof{
		DocumentId:  coreDoc.DocumentIdentifier,
		VersionId:   coreDoc.CurrentVersion,
		FieldProofs: proofs}, nil
}

// DeriveFromCoreDocument unpacks the core document into a model
func (s service) DeriveFromCoreDocument(cd *coredocumentpb.CoreDocument) (documents.Model, error) {
	var model documents.Model = new(InvoiceModel)
	err := model.UnpackCoreDocument(cd)
	if err != nil {
		return nil, centerrors.New(code.DocumentInvalid, err.Error())
	}

	return model, nil
}

// UnpackFromCreatePayload initializes the model with parameters provided from the rest-api call
func (s service) DeriveFromCreatePayload(invoiceInput *clientinvoicepb.InvoiceCreatePayload) (documents.Model, error) {
	if invoiceInput == nil {
		return nil, centerrors.New(code.DocumentInvalid, "input is nil")
	}

	invoiceModel := new(InvoiceModel)
	err := invoiceModel.InitInvoiceInput(invoiceInput)
	if err != nil {
		return nil, centerrors.New(code.DocumentInvalid, err.Error())
	}

	return invoiceModel, nil
}

// validateAndPersist validate models, persist the new model and anchors the document
func (s service) validateAndPersist(ctx context.Context, old, new documents.Model, validator documents.ValidatorGroup) (documents.Model, error) {
	inv, ok := new.(*InvoiceModel)
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

	coreDoc, err := inv.PackCoreDocument()
	if err != nil {
		return nil, centerrors.New(code.Unknown, err.Error())
	}

	saveState := func(coreDoc *coredocumentpb.CoreDocument) error {
		err := inv.UnpackCoreDocument(coreDoc)
		if err != nil {
			return err
		}

		return s.SaveState(inv)
	}

	err = s.coreDocProcessor.Anchor(ctx, coreDoc, saveState)
	if err != nil {
		return nil, centerrors.New(code.Unknown, err.Error())
	}

	coreDoc, err = inv.PackCoreDocument()
	if err != nil {
		return nil, centerrors.New(code.Unknown, err.Error())
	}

	for _, id := range coreDoc.Collaborators {
		cid, err := identity.ToCentID(id)
		if err != nil {
			return nil, centerrors.New(code.Unknown, err.Error())
		}
		err = s.coreDocProcessor.Send(ctx, coreDoc, cid)
		if err != nil {
			log.Infof("failed to send anchored document: %v\n", err)
		}
	}

	return inv, nil
}

// Create takes and invoice model and does required validation checks, tries to persist to DB
func (s service) Create(ctx context.Context, model documents.Model) (documents.Model, error) {
	return s.validateAndPersist(ctx, nil, model, CreateValidator())
}

// Update finds the old document, validates the new version and persists the updated document
func (s service) Update(ctx context.Context, model documents.Model) (documents.Model, error) {
	cd, err := model.PackCoreDocument()
	if err != nil {
		return nil, centerrors.New(code.Unknown, err.Error())
	}

	old, err := s.GetCurrentVersion(cd.DocumentIdentifier)
	if err != nil {
		return nil, centerrors.New(code.DocumentNotFound, err.Error())
	}

	return s.validateAndPersist(ctx, old, model, UpdateValidator())
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
		} else {
			inv = temp
			nextVersion = inv.CoreDocument.NextVersion
		}
	}
	return inv, nil
}

func (s service) getInvoiceVersion(documentID, version []byte) (inv *InvoiceModel, err error) {
	var doc documents.Model = new(InvoiceModel)
	err = s.repo.LoadByID(version, doc)
	if err != nil {
		return nil, err
	}
	inv, ok := doc.(*InvoiceModel)
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
	inv, ok := doc.(*InvoiceModel)
	if !ok {
		return nil, centerrors.New(code.DocumentInvalid, "document of invalid type")
	}

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
		DocumentId:    hexutil.Encode(inv.CoreDocument.DocumentIdentifier),
		VersionId:     hexutil.Encode(inv.CoreDocument.CurrentVersion),
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
	inv, ok := doc.(*InvoiceModel)
	if !ok {
		return nil, centerrors.New(code.DocumentInvalid, "document of invalid type")
	}

	data := inv.getClientData()
	return data, nil
}

// SaveState updates the model on DB
// This will disappear once we have common DB for every document
func (s service) SaveState(doc documents.Model) error {
	inv, ok := doc.(*InvoiceModel)
	if !ok {
		return centerrors.New(code.DocumentInvalid, "document of invalid type")
	}

	if inv.CoreDocument == nil {
		return centerrors.New(code.DocumentInvalid, "core document missing")
	}

	err := s.repo.Update(inv.CoreDocument.CurrentVersion, inv)
	if err != nil {
		return centerrors.New(code.Unknown, err.Error())
	}

	return nil
}

// DeriveFromUpdatePayload returns a new version of the old invoice identified by identifier in payload
func (s service) DeriveFromUpdatePayload(payload *clientinvoicepb.InvoiceUpdatePayload) (documents.Model, error) {
	if payload == nil {
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
	inv := new(InvoiceModel)
	err = inv.initInvoiceFromData(payload.Data)
	if err != nil {
		return nil, centerrors.New(code.DocumentInvalid, fmt.Sprintf("failed to load invoice from data: %v", err))
	}

	// update core document
	oldCD, err := old.PackCoreDocument()
	if err != nil {
		return nil, centerrors.New(code.Unknown, err.Error())
	}

	inv.CoreDocument, err = coredocument.PrepareNewVersion(*oldCD, payload.Collaborators)
	if err != nil {
		return nil, centerrors.New(code.DocumentInvalid, fmt.Sprintf("failed to prepare new version: %v", err))
	}

	return inv, nil
}

// RequestDocumentSignature Validates, Signs document received over the p2p layer and returs Signature
func (s service) RequestDocumentSignature(model documents.Model) (*coredocumentpb.Signature, error) {
	doc, err := model.PackCoreDocument()
	if err != nil {
		return nil, centerrors.New(code.DocumentInvalid, err.Error())
	}

	log.Infof("coredoc received %x with signing root %x", doc.DocumentIdentifier, doc.SigningRoot)

	// TODO(mig) Invoke validation as part of service call
	if err := coredocument.ValidateWithSignature(doc); err != nil {
		return nil, centerrors.New(code.DocumentInvalid, err.Error())
	}

	idConfig, err := centED25519.GetIDConfig()
	if err != nil {
		return nil, centerrors.New(code.Unknown, fmt.Sprintf("failed to get ID Config: %v", err))
	}

	sig := signatures.Sign(idConfig, doc.SigningRoot)
	doc.Signatures = append(doc.Signatures, sig)
	err = model.UnpackCoreDocument(doc)
	if err != nil {
		return nil, centerrors.New(code.Unknown, fmt.Sprintf("failed to Unpack CoreDocument: %v", err))
	}

	// TODO temporary until we deprecate old document version
	err = coredocumentrepository.GetRepository().Create(doc.DocumentIdentifier, doc)
	if err != nil {
		return nil, centerrors.New(code.Unknown, fmt.Sprintf("failed to Create legacy CoreDocument: %v", err))
	}

	err = s.repo.Create(doc.DocumentIdentifier, model)
	if err != nil {
		return nil, centerrors.New(code.Unknown, fmt.Sprintf("failed to store document: %v", err))
	}
	log.Infof("signed coredoc %x", doc.DocumentIdentifier)
	return sig, nil
}

// ReceiveAnchoredDocument receives a new anchored document, validates and updates the document in DB
func (s service) ReceiveAnchoredDocument(model documents.Model, headers *p2ppb.CentrifugeHeader) error {
	doc, err := model.PackCoreDocument()
	if err != nil {
		return centerrors.New(code.DocumentInvalid, err.Error())
	}

	// TODO temporary until we deprecate old document version
	err = coredocumentrepository.GetRepository().Update(doc.DocumentIdentifier, doc)
	if err != nil {
		return centerrors.New(code.Unknown, fmt.Sprintf("failed to Create legacy CoreDocument: %v", err))
	}

	// TODO(ved): post anchoring validations should be done before deriving model
	err = repo.Update(doc.CurrentVersion, model)
	if err != nil {
		return centerrors.New(code.Unknown, err.Error())
	}

	ts, _ := ptypes.TimestampProto(time.Now().UTC())
	notificationMsg := &notificationpb.NotificationMessage{
		EventType:          uint32(notification.RECEIVED_PAYLOAD),
		CentrifugeId:       headers.SenderCentrifugeId,
		Recorded:           ts,
		DocumentType:       doc.EmbeddedData.TypeUrl,
		DocumentIdentifier: doc.DocumentIdentifier,
	}

	// Async until we add queuing
	go s.notifier.Send(notificationMsg)

	return nil
}

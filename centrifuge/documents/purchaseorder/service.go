package purchaseorder

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/notification"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	"github.com/centrifuge/go-centrifuge/centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/centrifuge/centerrors"
	"github.com/centrifuge/go-centrifuge/centrifuge/code"
	"github.com/centrifuge/go-centrifuge/centrifuge/coredocument"
	"github.com/centrifuge/go-centrifuge/centrifuge/coredocument/processor"
	"github.com/centrifuge/go-centrifuge/centrifuge/coredocument/repository"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/centrifuge/identity"
	centED25519 "github.com/centrifuge/go-centrifuge/centrifuge/keytools/ed25519keys"
	"github.com/centrifuge/go-centrifuge/centrifuge/notification"
	clientpopb "github.com/centrifuge/go-centrifuge/centrifuge/protobufs/gen/go/purchaseorder"
	"github.com/centrifuge/go-centrifuge/centrifuge/signatures"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/golang/protobuf/ptypes"
)

// Service defines specific functions for purchase order
type Service interface {
	documents.Service

	// DeriverFromPayload derives purchase order from clientPayload
	DeriveFromCreatePayload(payload *clientpopb.PurchaseOrderCreatePayload, hdr *documents.ContextHeader) (documents.Model, error)

	// DeriveFromUpdatePayload derives purchase order from update payload
	DeriveFromUpdatePayload(payload *clientpopb.PurchaseOrderUpdatePayload, hdr *documents.ContextHeader) (documents.Model, error)

	// Create validates and persists purchase order and returns a Updated model
	Create(ctx context.Context, po documents.Model) (documents.Model, error)

	// Update validates and updates the purchase order and return the updated model
	Update(ctx context.Context, po documents.Model) (documents.Model, error)

	// DerivePurchaseOrderData returns the purchase order data as client data
	DerivePurchaseOrderData(po documents.Model) (*clientpopb.PurchaseOrderData, error)

	// DerivePurchaseOrderResponse returns the purchase order in our standard client format
	DerivePurchaseOrderResponse(po documents.Model) (*clientpopb.PurchaseOrderResponse, error)
}

// service implements Service and handles all purchase order related persistence and validations
// service always returns errors of type `centerrors` with proper error code
type service struct {
	repo             documents.Repository
	coreDocProcessor coredocumentprocessor.Processor
	notifier         notification.Sender
	anchorRepository anchors.AnchorRepository
}

// DefaultService returns the default implementation of the service
func DefaultService(repo documents.Repository, processor coredocumentprocessor.Processor, anchorRepository anchors.AnchorRepository) Service {
	return service{repo: repo, coreDocProcessor: processor, notifier: &notification.WebhookSender{}, anchorRepository: anchorRepository}
}

// DeriveFromCoreDocument takes a core document and returns a purchase order
func (s service) DeriveFromCoreDocument(cd *coredocumentpb.CoreDocument) (documents.Model, error) {
	var model documents.Model = new(PurchaseOrderModel)
	err := model.UnpackCoreDocument(cd)
	if err != nil {
		return nil, centerrors.New(code.DocumentInvalid, err.Error())
	}

	return model, nil
}

// calculateDataRoot validates the document, calculates the data root, and persists to DB
func (s service) calculateDataRoot(old, new documents.Model, validator documents.Validator) (documents.Model, error) {
	po, ok := new.(*PurchaseOrderModel)
	if !ok {
		return nil, centerrors.New(code.DocumentInvalid, fmt.Sprintf("unknown document type: %T", new))
	}

	// create data root, has to be done at the model level to access fields
	err := po.calculateDataRoot()
	if err != nil {
		return nil, centerrors.New(code.DocumentInvalid, err.Error())
	}

	// validate the purchase order
	err = validator.Validate(old, po)
	if err != nil {
		return nil, centerrors.NewWithErrors(code.DocumentInvalid, "validations failed", documents.ConvertToMap(err))
	}

	// we use CurrentVersion as the id since that will be unique across multiple versions of the same document
	err = s.repo.Create(po.CoreDocument.CurrentVersion, po)
	if err != nil {
		return nil, centerrors.New(code.Unknown, err.Error())
	}

	return po, nil
}

// Create validates, persists, and anchors a purchase order
func (s service) Create(ctx context.Context, po documents.Model) (documents.Model, error) {
	po, err := s.calculateDataRoot(nil, po, CreateValidator())
	if err != nil {
		return nil, err
	}

	po, err = documents.AnchorDocument(ctx, po, s.coreDocProcessor, s.repo.Update)
	if err != nil {
		return nil, centerrors.New(code.Unknown, err.Error())
	}

	return po, nil
}

// Update validates, persists, and anchors a new version of purchase order
func (s service) Update(ctx context.Context, po documents.Model) (documents.Model, error) {
	return nil, fmt.Errorf("implement me")
}

// DeriveFromCreatePayload derives purchase order from create payload
func (s service) DeriveFromCreatePayload(payload *clientpopb.PurchaseOrderCreatePayload, ctxH *documents.ContextHeader) (documents.Model, error) {
	if payload == nil {
		return nil, centerrors.New(code.DocumentInvalid, "input is nil")
	}

	po := new(PurchaseOrderModel)
	err := po.InitPurchaseOrderInput(payload, ctxH)
	if err != nil {
		return nil, centerrors.New(code.DocumentInvalid, fmt.Sprintf("purchase order init failed: %v", err))
	}

	return po, nil
}

// DeriveFromUpdatePayload derives purchase order from update payload
func (s service) DeriveFromUpdatePayload(payload *clientpopb.PurchaseOrderUpdatePayload, ctxH *documents.ContextHeader) (documents.Model, error) {
	return nil, fmt.Errorf("implement me")
}

// DerivePurchaseOrderData returns po data from the model
func (s service) DerivePurchaseOrderData(doc documents.Model) (*clientpopb.PurchaseOrderData, error) {
	po, ok := doc.(*PurchaseOrderModel)
	if !ok {
		return nil, centerrors.New(code.DocumentInvalid, "document of invalid type")
	}

	return po.getClientData(), nil
}

// DerivePurchaseOrderResponse returns po response from the model
func (s service) DerivePurchaseOrderResponse(doc documents.Model) (*clientpopb.PurchaseOrderResponse, error) {
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

	header := &clientpopb.ResponseHeader{
		DocumentId:    hexutil.Encode(cd.DocumentIdentifier),
		VersionId:     hexutil.Encode(cd.CurrentVersion),
		Collaborators: collaborators,
	}

	data, err := s.DerivePurchaseOrderData(doc)
	if err != nil {
		return nil, err
	}

	return &clientpopb.PurchaseOrderResponse{
		Header: header,
		Data:   data,
	}, nil
}

func (s service) getPurchaseOrderVersion(documentID, version []byte) (model *PurchaseOrderModel, err error) {
	var doc documents.Model = new(PurchaseOrderModel)
	err = s.repo.LoadByID(version, doc)
	if err != nil {
		return nil, err
	}
	model, ok := doc.(*PurchaseOrderModel)
	if !ok {
		return nil, err
	}

	if !bytes.Equal(model.CoreDocument.DocumentIdentifier, documentID) {
		return nil, centerrors.New(code.DocumentInvalid, "version is not valid for this identifier")
	}
	return model, nil
}

// GetLastVersion returns the latest version of the document
func (s service) GetCurrentVersion(documentID []byte) (documents.Model, error) {
	model, err := s.getPurchaseOrderVersion(documentID, documentID)
	if err != nil {
		return nil, centerrors.Wrap(err, "document not found")
	}
	nextVersion := model.CoreDocument.NextVersion
	for nextVersion != nil {
		temp, err := s.getPurchaseOrderVersion(documentID, nextVersion)
		if err != nil {
			return model, nil
		} else {
			model = temp
			nextVersion = model.CoreDocument.NextVersion
		}
	}
	return model, nil
}

// GetVersion returns the specific version of the document
func (s service) GetVersion(documentID []byte, version []byte) (documents.Model, error) {
	po, err := s.getPurchaseOrderVersion(documentID, version)
	if err != nil {
		return nil, centerrors.Wrap(err, "document not found for the given version")
	}
	return po, nil

}

// purchaseOrderProof creates proofs for purchaseOrder model fields
func (s service) purchaseOrderProof(model documents.Model, fields []string) (*documents.DocumentProof, error) {
	po, ok := model.(*PurchaseOrderModel)
	if !ok {
		return nil, centerrors.New(code.DocumentInvalid, "document of invalid type")
	}
	if err := coredocument.PostAnchoredValidator(s.anchorRepository).Validate(nil, po); err != nil {
		return nil, centerrors.New(code.DocumentInvalid, err.Error())
	}
	coreDoc, proofs, err := po.createProofs(fields)
	if err != nil {
		return nil, err
	}
	return &documents.DocumentProof{
		DocumentId:  coreDoc.DocumentIdentifier,
		VersionId:   coreDoc.CurrentVersion,
		FieldProofs: proofs,
	}, nil
}

// CreateProofs generates proofs for given document
func (s service) CreateProofs(documentID []byte, fields []string) (*documents.DocumentProof, error) {
	model, err := s.GetCurrentVersion(documentID)
	if err != nil {
		return nil, err
	}
	return s.purchaseOrderProof(model, fields)
}

// CreateProofsForVersion generates proofs for specific version of the document
func (s service) CreateProofsForVersion(documentID, version []byte, fields []string) (*documents.DocumentProof, error) {
	model, err := s.GetVersion(documentID, version)
	if err != nil {
		return nil, err
	}
	return s.purchaseOrderProof(model, fields)
}

// RequestDocumentSignature validates the document and returns the signature
// Note: this is document agnostic. But since we do not have a common implementation, adding it here.
// will remove this once we have a common implementation for documents.Service
func (s service) RequestDocumentSignature(model documents.Model) (*coredocumentpb.Signature, error) {
	if err := coredocument.SignatureRequestValidator().Validate(nil, model); err != nil {
		return nil, centerrors.New(code.DocumentInvalid, err.Error())
	}

	cd, err := model.PackCoreDocument()
	if err != nil {
		return nil, centerrors.New(code.DocumentInvalid, err.Error())
	}

	log.Infof("coredoc received %x with signing root %x", cd.DocumentIdentifier, cd.SigningRoot)

	idConfig, err := centED25519.GetIDConfig()
	if err != nil {
		return nil, centerrors.New(code.Unknown, fmt.Sprintf("failed to get ID Config: %v", err))
	}

	sig := signatures.Sign(idConfig, cd.SigningRoot)
	cd.Signatures = append(cd.Signatures, sig)
	err = model.UnpackCoreDocument(cd)
	if err != nil {
		return nil, centerrors.New(code.Unknown, fmt.Sprintf("failed to Unpack CoreDocument: %v", err))
	}

	err = coredocumentrepository.GetRepository().Create(cd.DocumentIdentifier, cd)
	if err != nil {
		return nil, centerrors.New(code.Unknown, fmt.Sprintf("failed to Create legacy CoreDocument: %v", err))
	}

	err = s.repo.Create(cd.DocumentIdentifier, model)
	if err != nil {
		return nil, centerrors.New(code.Unknown, fmt.Sprintf("failed to store document: %v", err))
	}
	log.Infof("signed coredoc %x", cd.DocumentIdentifier)
	return sig, nil
}

// ReceiveAnchoredDocument validates the anchored document and updates it on DB
// Note: this is document agnostic. But since we do not have a common implementation, adding it here.
// will remove this once we have a common implementation for documents.Service
func (s service) ReceiveAnchoredDocument(model documents.Model, headers *p2ppb.CentrifugeHeader) error {
	if err := coredocument.PostAnchoredValidator(s.anchorRepository).Validate(nil, model); err != nil {
		return centerrors.New(code.DocumentInvalid, err.Error())
	}

	doc, err := model.PackCoreDocument()
	if err != nil {
		return centerrors.New(code.DocumentInvalid, err.Error())
	}

	err = coredocumentrepository.GetRepository().Update(doc.DocumentIdentifier, doc)
	if err != nil {
		return centerrors.New(code.Unknown, fmt.Sprintf("failed to Create legacy CoreDocument: %v", err))
	}

	err = s.repo.Update(doc.CurrentVersion, model)
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

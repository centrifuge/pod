package purchaseorder

import (
	"context"
	"fmt"
	"github.com/centrifuge/go-centrifuge/centrifuge/anchors"
	"github.com/centrifuge/go-centrifuge/centrifuge/centerrors"
	"github.com/centrifuge/go-centrifuge/centrifuge/code"
	"github.com/centrifuge/go-centrifuge/centrifuge/coredocument"

	"github.com/centrifuge/centrifuge-protobufs/gen/go/coredocument"
	"github.com/centrifuge/centrifuge-protobufs/gen/go/p2p"
	"github.com/centrifuge/go-centrifuge/centrifuge/coredocument/processor"
	"github.com/centrifuge/go-centrifuge/centrifuge/documents"
	"github.com/centrifuge/go-centrifuge/centrifuge/notification"
	clientpopb "github.com/centrifuge/go-centrifuge/centrifuge/protobufs/gen/go/purchaseorder"
)

// Service defines specific functions for purchase order
type Service interface {
	documents.Service

	// DeriverFromPayload derives purchase order from clientPayload
	DeriveFromCreatePayload(payload *clientpopb.PurchaseOrderCreatePayload) (documents.Model, error)

	// DeriveFromUpdatePayload derives purchase order from update payload
	DeriveFromUpdatePayload(payload *clientpopb.PurchaseOrderUpdatePayload) (documents.Model, error)

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
	return nil, fmt.Errorf("implement me")
}

// Create validates, persists, and anchors a purchase order
func (s service) Create(ctx context.Context, po documents.Model) (documents.Model, error) {
	return nil, fmt.Errorf("implement me")
}

// Update validates, persists, and anchors a new version of purchase order
func (s service) Update(ctx context.Context, po documents.Model) (documents.Model, error) {
	return nil, fmt.Errorf("implement me")
}

// DeriveFromCreatePayload derives purchase order from create payload
func (s service) DeriveFromCreatePayload(payload *clientpopb.PurchaseOrderCreatePayload) (documents.Model, error) {
	return nil, fmt.Errorf("implement me")
}

// DeriveFromUpdatePayload derives purchase order from update payload
func (s service) DeriveFromUpdatePayload(payload *clientpopb.PurchaseOrderUpdatePayload) (documents.Model, error) {
	return nil, fmt.Errorf("implement me")
}

// DerivePurchaseOrderData returns po data from the model
func (s service) DerivePurchaseOrderData(po documents.Model) (*clientpopb.PurchaseOrderData, error) {
	return nil, fmt.Errorf("implement me")
}

// DerivePurchaseOrderResponse returns po response from the model
func (s service) DerivePurchaseOrderResponse(inv documents.Model) (*clientpopb.PurchaseOrderResponse, error) {
	return nil, fmt.Errorf("implement me")
}

// GetLastVersion returns the latest version of the document
func (s service) GetCurrentVersion(documentID []byte) (documents.Model, error) {
	return nil, fmt.Errorf("implement me")
}

// GetVersion returns the specific version of the document
func (s service) GetVersion(documentID []byte, version []byte) (documents.Model, error) {
	return nil, fmt.Errorf("implement me")
}

// purchaseOrderProof creates proofs for purchaseOrder model fields
func (s service) purchaseOrderProof(po *PurchaseOrderModel, fields []string) (*documents.DocumentProof, error) {
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
	po, ok := model.(*PurchaseOrderModel)
	if !ok {
		return nil, centerrors.New(code.DocumentInvalid, "document of invalid type")
	}
	return s.purchaseOrderProof(po, fields)
}

// CreateProofsForVersion generates proofs for specific version of the document
func (s service) CreateProofsForVersion(documentID, version []byte, fields []string) (*documents.DocumentProof, error) {
	model, err := s.GetVersion(documentID, version)
	if err != nil {
		return nil, err
	}
	po, ok := model.(*PurchaseOrderModel)
	if !ok {
		return nil, centerrors.New(code.DocumentInvalid, "document of invalid type")
	}
	return s.purchaseOrderProof(po, fields)
}

// RequestDocumentSignature validates the document and returns the signature
func (s service) RequestDocumentSignature(model documents.Model) (*coredocumentpb.Signature, error) {
	return nil, fmt.Errorf("implement me")
}

// ReceiveAnchoredDocument validates the anchored document and updates it on DB
func (s service) ReceiveAnchoredDocument(model documents.Model, headers *p2ppb.CentrifugeHeader) error {
	return fmt.Errorf("implement me")
}
